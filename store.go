package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	dateFmt  = "2006-01-02"
	stampFmt = "2006-01-02T15:04"
	clockFmt = "15:04"
)

// Task: existe até ser concluída; ao concluir, é apagada.
type Task struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Type    string `json:"type"`  // legado (v1.0); hoje usa Area
	Area    string `json:"area"`  // id do catálogo
	Start   string `json:"start"` // "2006-01-02T15:04" (opcional)
	Due     string `json:"due"`   // "2006-01-02T15:04"
	Created string `json:"created"`
}

// Routine: item da rotina. Fixa (repete nos dias da semana marcados) ou só de um dia.
type Routine struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Time  string `json:"time"` // "15:04"
	End   string `json:"end"`  // opcional
	Fixed bool   `json:"fixed"`
	Days  []int  `json:"days"` // 0=dom ... 6=sáb (só para fixas)
	Date  string `json:"date"` // "2006-01-02" (só para as do dia)
	Note  string `json:"note"`
	Area  string `json:"area"` // id do catálogo (opcional)
}

// Catalog: uma área do dia a dia (Faculdade, Trabalho, Plano DevJr...).
type Catalog struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Emoji string `json:"emoji"`
	Color string `json:"color"`
	Order int    `json:"order"`
}

// Group: uma etapa/seção de conteúdos dentro de um catálogo.
type Group struct {
	ID      string `json:"id"`
	Catalog string `json:"catalog"`
	Title   string `json:"title"`
	Gate    string `json:"gate"` // o que dominar antes de avançar
	Order   int    `json:"order"`
}

// Topic: um assunto para estudar.
type Topic struct {
	ID     string `json:"id"`
	Group  string `json:"group"`
	Title  string `json:"title"`
	Note   string `json:"note"`
	Src    string `json:"src"`
	Prio   int    `json:"prio"`   // 1 essencial, 2 importante, 3 diferencial, 4 depois
	Status int    `json:"status"` // 0 a estudar, 1 estudando, 2 dominado
	Order  int    `json:"order"`
}

type Settings struct {
	AutoStart   bool `json:"autoStart"`
	OpenOnBoot  bool `json:"openOnBoot"` // abre a janela ao ligar o PC
	Lead        int  `json:"lead"`       // minutos de antecedência das rotinas
	Initialized bool `json:"initialized"`

	Theme        string `json:"theme"`     // dark | light | system
	Sound        string `json:"sound"`     // windows | custom | none
	SoundFile    string `json:"soundFile"` // arquivo salvo na pasta de dados
	SoundName    string `json:"soundName"` // nome original do arquivo
	Volume       int    `json:"volume"`    // 5..100
	NoAutoUpdate bool   `json:"noAutoUpdate"`
	UpdateRepo   string `json:"updateRepo"` // dono/repositório no GitHub

	Wallpaper string `json:"wallpaper"` // "" sem, "custom" ou id de um tema (ex.: "lofi-janela")
	WallFile  string `json:"wallFile"`  // imagem própria salva na pasta de dados
	WallName  string `json:"wallName"`
	WallRev   int    `json:"wallRev"`  // muda a cada envio (evita cache)
	WallDim   int    `json:"wallDim"`  // 0..90 — quanto escurecer o fundo
	WallBlur  int    `json:"wallBlur"` // 0..20 px
	Accent    string `json:"accent"`   // cor de destaque ("" = padrão)
}

type Data struct {
	Tasks    []Task              `json:"tasks"`
	Routines []Routine           `json:"routines"`
	Done     map[string][]string `json:"done"` // data -> ids de rotina concluídos
	Sent     map[string]string   `json:"sent"` // chave de notificação -> data
	Settings Settings            `json:"settings"`
	Catalogs []Catalog           `json:"catalogs"`
	Groups   []Group             `json:"groups"`
	Topics   []Topic             `json:"topics"`
	Seed     int                 `json:"seed"` // versão do conteúdo inicial já aplicado
}

type Store struct {
	mu   sync.Mutex
	path string
	d    Data
}

func newID() string {
	b := make([]byte, 6)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func openStore(dir string) (*Store, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	s := &Store{path: filepath.Join(dir, "dados.json")}
	if b, err := os.ReadFile(s.path); err == nil {
		if err := json.Unmarshal(b, &s.d); err != nil {
			// arquivo corrompido: guarda cópia e começa do zero
			os.Rename(s.path, s.path+".corrompido-"+time.Now().Format("20060102-150405"))
			s.d = Data{}
		}
	}
	if s.d.Done == nil {
		s.d.Done = map[string][]string{}
	}
	if s.d.Sent == nil {
		s.d.Sent = map[string]string{}
	}
	if s.d.Settings.Lead <= 0 {
		s.d.Settings.Lead = 10
	}
	s.migrate()
	s.prune(time.Now())
	return s, nil
}

// saveLocked grava de forma atômica (arquivo temporário + rename).
func (s *Store) saveLocked() error {
	b, err := json.MarshalIndent(&s.d, "", "  ")
	if err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}

// prune remove o que já não serve: rotinas avulsas antigas e registros velhos.
func (s *Store) prune(now time.Time) {
	limit := now.AddDate(0, 0, -7).Format(dateFmt)
	keep := s.d.Routines[:0]
	for _, r := range s.d.Routines {
		if r.Fixed || r.Date >= limit {
			keep = append(keep, r)
		}
	}
	s.d.Routines = keep
	for d := range s.d.Done {
		if d < limit {
			delete(s.d.Done, d)
		}
	}
	for k, d := range s.d.Sent {
		if d < limit {
			delete(s.d.Sent, k)
		}
	}
}

func parseStamp(v string) (time.Time, bool) {
	if v == "" {
		return time.Time{}, false
	}
	t, err := time.ParseInLocation(stampFmt, v, time.Local)
	return t, err == nil
}

func atClock(day time.Time, clock string) (time.Time, bool) {
	c, err := time.ParseInLocation(clockFmt, clock, time.Local)
	if err != nil {
		return time.Time{}, false
	}
	return time.Date(day.Year(), day.Month(), day.Day(), c.Hour(), c.Minute(), 0, 0, time.Local), true
}

func (r Routine) occursOn(day time.Time) bool {
	if !r.Fixed {
		return r.Date == day.Format(dateFmt)
	}
	wd := int(day.Weekday())
	for _, d := range r.Days {
		if d == wd {
			return true
		}
	}
	return false
}

func (s *Store) isDone(date, id string) bool {
	for _, x := range s.d.Done[date] {
		if x == id {
			return true
		}
	}
	return false
}

// ---- operações (todas chamadas com s.mu travado) ----

type DayItem struct {
	Routine
	Done bool `json:"done"`
}

func (s *Store) dayItems(day time.Time) []DayItem {
	date := day.Format(dateFmt)
	out := []DayItem{}
	for _, r := range s.d.Routines {
		if r.occursOn(day) {
			out = append(out, DayItem{Routine: r, Done: s.isDone(date, r.ID)})
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Time < out[j].Time })
	return out
}

func (s *Store) fixedRoutines() []Routine {
	out := []Routine{}
	for _, r := range s.d.Routines {
		if r.Fixed {
			out = append(out, r)
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Time < out[j].Time })
	return out
}

func (s *Store) sortedTasks() []Task {
	out := append([]Task{}, s.d.Tasks...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].Due < out[j].Due })
	return out
}

func cleanTask(t *Task) bool {
	t.Name = strings.TrimSpace(t.Name)
	t.Type = strings.TrimSpace(t.Type)
	if t.Name == "" {
		return false
	}
	if _, ok := parseStamp(t.Due); !ok {
		return false
	}
	if t.Start != "" {
		if _, ok := parseStamp(t.Start); !ok {
			t.Start = ""
		}
	}
	return true
}

func cleanRoutine(r *Routine) bool {
	r.Title = strings.TrimSpace(r.Title)
	r.Note = strings.TrimSpace(r.Note)
	if r.Title == "" {
		return false
	}
	if _, err := time.Parse(clockFmt, r.Time); err != nil {
		return false
	}
	if r.End != "" {
		if _, err := time.Parse(clockFmt, r.End); err != nil {
			r.End = ""
		}
	}
	if r.Fixed {
		seen := map[int]bool{}
		days := []int{}
		for _, d := range r.Days {
			if d >= 0 && d <= 6 && !seen[d] {
				seen[d] = true
				days = append(days, d)
			}
		}
		sort.Ints(days)
		if len(days) == 0 {
			return false
		}
		r.Days, r.Date = days, ""
	} else {
		if _, err := time.Parse(dateFmt, r.Date); err != nil {
			return false
		}
		r.Days = nil
	}
	return true
}

func (s *Store) clearSent(prefix string) {
	for k := range s.d.Sent {
		if strings.HasPrefix(k, prefix) {
			delete(s.d.Sent, k)
		}
	}
}

// ---- migração e conteúdo inicial ----

func (s *Store) migrate() {
	st := &s.d.Settings
	if st.Theme == "" {
		st.Theme = "dark"
	}
	if st.Sound == "" || (st.Sound == "custom" && st.SoundFile == "") {
		st.Sound = "windows"
	}
	if st.Volume <= 0 || st.Volume > 100 {
		st.Volume = 80
	}
	if st.WallDim == 0 && st.Wallpaper == "" {
		st.WallDim = 45
	}
	if s.d.Seed < 1 && len(s.d.Catalogs) == 0 && len(s.d.Tasks) == 0 && len(s.d.Routines) == 0 && editionSeed(s) {
		s.d.Seed = 1
		s.saveLocked()
	}
	if s.d.Seed < 1 {
		for i, c := range seedCatalogs {
			if s.catalogByName(c.Name) == nil {
				c.ID, c.Order = newID(), i
				s.d.Catalogs = append(s.d.Catalogs, c)
			}
		}
		dev := s.catalogByName("Plano DevJr")
		for gi, g := range seedDevJr {
			grp := Group{ID: newID(), Catalog: dev.ID, Title: g.title, Gate: g.gate, Order: gi}
			s.d.Groups = append(s.d.Groups, grp)
			for ti, t := range g.topics {
				s.d.Topics = append(s.d.Topics, Topic{ID: newID(), Group: grp.ID, Title: t.title,
					Note: t.note, Src: t.src, Prio: t.prio, Order: ti})
			}
		}
		// tarefas da v1.0: o "tipo" vira catálogo
		for i := range s.d.Tasks {
			t := &s.d.Tasks[i]
			if t.Area != "" || strings.TrimSpace(t.Type) == "" {
				continue
			}
			c := s.catalogByName(t.Type)
			if c == nil {
				s.d.Catalogs = append(s.d.Catalogs, Catalog{ID: newID(), Name: strings.TrimSpace(t.Type),
					Emoji: "📁", Color: "#94a3b8", Order: len(s.d.Catalogs)})
				c = &s.d.Catalogs[len(s.d.Catalogs)-1]
			}
			t.Area, t.Type = c.ID, ""
		}
		s.d.Seed = 1
		s.saveLocked()
	}
}

func (s *Store) catalogByName(name string) *Catalog {
	for i := range s.d.Catalogs {
		if strings.EqualFold(strings.TrimSpace(s.d.Catalogs[i].Name), strings.TrimSpace(name)) {
			return &s.d.Catalogs[i]
		}
	}
	return nil
}

func (s *Store) catalog(id string) *Catalog {
	for i := range s.d.Catalogs {
		if s.d.Catalogs[i].ID == id {
			return &s.d.Catalogs[i]
		}
	}
	return nil
}

func (s *Store) group(id string) *Group {
	for i := range s.d.Groups {
		if s.d.Groups[i].ID == id {
			return &s.d.Groups[i]
		}
	}
	return nil
}

// fixArea zera referências a catálogos que não existem mais.
func (s *Store) fixArea(id string) string {
	if id != "" && s.catalog(id) == nil {
		return ""
	}
	return id
}

func (s *Store) sortedCatalogs() []Catalog {
	out := append([]Catalog{}, s.d.Catalogs...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].Order < out[j].Order })
	return out
}

func (s *Store) sortedGroups() []Group {
	out := append([]Group{}, s.d.Groups...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].Order < out[j].Order })
	return out
}

func (s *Store) sortedTopics() []Topic {
	out := append([]Topic{}, s.d.Topics...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].Order < out[j].Order })
	return out
}
