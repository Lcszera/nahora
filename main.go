package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime/debug"
	"strings"
	"sync"
	"time"
)

const (
	appName = "naHora"
	port    = 47923
)

//go:embed ui
var uiFS embed.FS

var wallIDRe = regexp.MustCompile(`^(|custom|[a-z0-9-]{1,40})$`)

var (
	store    *Store
	listener net.Listener
	exePath  string
	uiURL    = fmt.Sprintf("http://127.0.0.1:%d/", port)
	notesQ   = make(chan Note, 64)
	hub      = &Hub{clients: map[chan string]bool{}}
)

func dataDir() string {
	if d := os.Getenv("NAHORA_DIR"); d != "" {
		return d
	}
	base, err := os.UserConfigDir()
	if err != nil {
		base = "."
	}
	return filepath.Join(base, appName)
}

func main() {
	// Pouca memória: GC mais agressivo e teto suave.
	debug.SetGCPercent(25)
	debug.SetMemoryLimit(24 << 20)

	trayOnly, afterUpdate := false, false
	for _, a := range os.Args[1:] {
		switch a {
		case "--tray", "/tray":
			trayOnly = true
		case "--after-update":
			afterUpdate = true
		}
	}
	exePath, _ = os.Executable()
	if p, err := filepath.EvalSymlinks(exePath); err == nil {
		exePath = p
	}

	ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	for i := 0; err != nil && afterUpdate && i < 60; i++ {
		// a versão anterior ainda está fechando
		time.Sleep(250 * time.Millisecond)
		ln, err = net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	}
	if err != nil {
		// Já está rodando: só abre a janela da instância existente.
		if !trayOnly {
			openUI(uiURL)
		}
		return
	}

	store, err = openStore(dataDir())
	if err != nil {
		fatal("Não consegui abrir a pasta de dados:\n" + err.Error())
		return
	}

	store.mu.Lock()
	if !store.d.Settings.Initialized {
		store.d.Settings.Initialized = true
		store.d.Settings.AutoStart = true
		store.d.Settings.OpenOnBoot = true
		store.saveLocked()
	}
	auto := store.d.Settings.AutoStart
	openOnBoot := store.d.Settings.OpenOnBoot
	store.mu.Unlock()
	if auto {
		setAutoStart(true) // reaplica (caso o .exe tenha mudado de pasta)
	}

	listener = ln
	afterStart()
	go serve(ln)
	go scheduler()
	go notifier()
	go updateLoop()

	if afterUpdate {
		// a janela (se estava aberta) se reconecta sozinha
	} else if !trayOnly {
		go func() { time.Sleep(150 * time.Millisecond); openUI(uiURL) }()
	} else if openOnBoot {
		// ligou o PC: abre a janela para você montar a rotina do dia
		go func() { time.Sleep(6 * time.Second); openUI(uiURL) }()
	}
	platformRun() // laço da bandeja; retorna ao clicar em "Sair"
}

func scheduler() {
	tick := func() {
		store.mu.Lock()
		notes, changed := store.check(time.Now())
		if changed {
			store.saveLocked()
		}
		store.mu.Unlock()
		for _, n := range notes {
			notesQ <- n
		}
		if changed {
			hub.broadcast("changed")
		}
	}
	tick()
	// alinha ao início de cada 15 s para não perder minutos
	for {
		now := time.Now()
		time.Sleep(15*time.Second - time.Duration(now.UnixNano())%(15*time.Second) + 50*time.Millisecond)
		tick()
		if hub.count() == 0 {
			trimMemory()
		}
	}
}

// notifier espaça as notificações para o Windows não engolir uma pela outra.
func notifier() {
	for n := range notesQ {
		deliver(n)
		time.Sleep(5 * time.Second)
	}
}

// ---------- SSE (atualiza a janela quando algo muda) ----------

type Hub struct {
	mu      sync.Mutex
	clients map[chan string]bool
}

func (h *Hub) add() chan string {
	c := make(chan string, 4)
	h.mu.Lock()
	h.clients[c] = true
	h.mu.Unlock()
	return c
}
func (h *Hub) remove(c chan string) {
	h.mu.Lock()
	delete(h.clients, c)
	n := len(h.clients)
	h.mu.Unlock()
	if n == 0 {
		go func() {
			time.Sleep(5 * time.Second)
			if h.count() > 0 {
				return
			}
			upd.Lock()
			ready := upd.info.Status == "ready"
			upd.Unlock()
			if ready {
				applyUpdate() // janela fechada: bom momento para atualizar
				return
			}
			debug.FreeOSMemory()
			trimMemory()
		}()
	}
}
func (h *Hub) count() int { h.mu.Lock(); defer h.mu.Unlock(); return len(h.clients) }
func (h *Hub) broadcast(msg string) {
	h.mu.Lock()
	for c := range h.clients {
		select {
		case c <- msg:
		default:
		}
	}
	h.mu.Unlock()
}

// ---------- HTTP ----------

func serve(ln net.Listener) {
	mux := http.NewServeMux()
	sub, _ := fs.Sub(uiFS, "ui")
	mux.Handle("GET /", http.FileServer(http.FS(sub)))

	mux.HandleFunc("GET /api/state", apiState)
	mux.HandleFunc("GET /api/events", apiEvents)
	mux.HandleFunc("POST /api/tasks", apiTaskSave)
	mux.HandleFunc("PUT /api/tasks/{id}", apiTaskSave)
	mux.HandleFunc("DELETE /api/tasks/{id}", apiTaskDelete)
	mux.HandleFunc("POST /api/routines", apiRoutineSave)
	mux.HandleFunc("PUT /api/routines/{id}", apiRoutineSave)
	mux.HandleFunc("DELETE /api/routines/{id}", apiRoutineDelete)
	mux.HandleFunc("POST /api/routines/{id}/done", apiRoutineDone)
	mux.HandleFunc("POST /api/settings", apiSettings)
	mux.HandleFunc("POST /api/test", func(w http.ResponseWriter, r *http.Request) {
		notesQ <- Note{"⚡ naHora funcionando", "É assim que os avisos vão aparecer."}
		w.WriteHeader(204)
	})
	mux.HandleFunc("POST /api/catalogs", apiCatalogSave)
	mux.HandleFunc("PUT /api/catalogs/{id}", apiCatalogSave)
	mux.HandleFunc("DELETE /api/catalogs/{id}", apiCatalogDelete)
	mux.HandleFunc("POST /api/groups", apiGroupSave)
	mux.HandleFunc("PUT /api/groups/{id}", apiGroupSave)
	mux.HandleFunc("DELETE /api/groups/{id}", apiGroupDelete)
	mux.HandleFunc("POST /api/topics", apiTopicSave)
	mux.HandleFunc("PUT /api/topics/{id}", apiTopicSave)
	mux.HandleFunc("POST /api/topics/{id}/status", apiTopicStatus)
	mux.HandleFunc("DELETE /api/topics/{id}", apiTopicDelete)
	mux.HandleFunc("POST /api/sound", apiSoundUpload)
	mux.HandleFunc("DELETE /api/sound", apiSoundDelete)
	mux.HandleFunc("POST /api/sound/play", apiSoundPlay)
	mux.HandleFunc("POST /api/update/check", apiUpdateCheck)
	mux.HandleFunc("POST /api/update/apply", apiUpdateApply)
	mux.HandleFunc("POST /api/update/file", apiUpdateFile)
	mux.HandleFunc("POST /api/wallpaper", apiWallUpload)
	mux.HandleFunc("DELETE /api/wallpaper", apiWallDelete)
	mux.HandleFunc("GET /wall/custom", serveCustomWall)

	srv := &http.Server{Handler: guard(mux), ReadHeaderTimeout: 5 * time.Second}
	srv.Serve(ln)
}

// guard: só aceita acesso local e exige um cabeçalho próprio nas alterações,
// para nenhum site aberto no navegador mexer nos seus dados.
func guard(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host := r.Host
		if h, _, err := net.SplitHostPort(host); err == nil {
			host = h
		}
		if host != "127.0.0.1" && host != "localhost" {
			http.Error(w, "forbidden", 403)
			return
		}
		if r.Method != "GET" && r.Header.Get("X-NaHora") != "1" {
			http.Error(w, "forbidden", 403)
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(v)
}

func changedAndSave() {
	store.saveLocked()
	go hub.broadcast("changed")
}

func apiState(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	day := now
	if q := r.URL.Query().Get("date"); q != "" {
		if d, err := time.ParseInLocation(dateFmt, q, time.Local); err == nil {
			day = d
		}
	}
	up := updateSnapshot()
	store.mu.Lock()
	defer store.mu.Unlock()
	writeJSON(w, map[string]any{
		"update":   up,
		"now":      now.Format("2006-01-02T15:04:05"),
		"today":    now.Format(dateFmt),
		"date":     day.Format(dateFmt),
		"day":      store.dayItems(day),
		"fixed":    store.fixedRoutines(),
		"tasks":    store.sortedTasks(),
		"settings": store.d.Settings,
		"catalogs": store.sortedCatalogs(),
		"groups":   store.sortedGroups(),
		"topics":   store.sortedTopics(),
		"version":  appVersion,
	})
}

func apiEvents(w http.ResponseWriter, r *http.Request) {
	fl, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "no stream", 500)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	c := hub.add()
	defer hub.remove(c)
	fmt.Fprint(w, "retry: 2000\n\n")
	fl.Flush()
	ping := time.NewTicker(30 * time.Second)
	defer ping.Stop()
	for {
		select {
		case <-r.Context().Done():
			return
		case m := <-c:
			fmt.Fprintf(w, "data: %s\n\n", m)
			fl.Flush()
		case <-ping.C:
			fmt.Fprint(w, ": ping\n\n")
			fl.Flush()
		}
	}
}

func apiTaskSave(w http.ResponseWriter, r *http.Request) {
	var t Task
	if json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<16)).Decode(&t) != nil || !cleanTask(&t) {
		http.Error(w, "Preencha o nome e a data de entrega.", 400)
		return
	}
	now := time.Now()
	store.mu.Lock()
	defer store.mu.Unlock()
	t.Area = store.fixArea(t.Area)
	if t.Area != "" {
		t.Type = ""
	}
	if id := r.PathValue("id"); id != "" {
		found := false
		for i := range store.d.Tasks {
			if store.d.Tasks[i].ID == id {
				t.ID, t.Created = id, store.d.Tasks[i].Created
				store.d.Tasks[i] = t
				found = true
			}
		}
		if !found {
			http.Error(w, "não encontrada", 404)
			return
		}
		store.clearSent("t|" + id + "|")
	} else {
		t.ID, t.Created = newID(), now.Format(stampFmt)
		store.d.Tasks = append(store.d.Tasks, t)
	}
	store.markPassed(t, now)
	changedAndSave()
	writeJSON(w, t)
}

func apiTaskDelete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	store.mu.Lock()
	defer store.mu.Unlock()
	keep := store.d.Tasks[:0]
	for _, t := range store.d.Tasks {
		if t.ID != id {
			keep = append(keep, t)
		}
	}
	store.d.Tasks = keep
	store.clearSent("t|" + id + "|")
	changedAndSave()
	w.WriteHeader(204)
}

func apiRoutineSave(w http.ResponseWriter, r *http.Request) {
	var rt Routine
	if json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<16)).Decode(&rt) != nil || !cleanRoutine(&rt) {
		http.Error(w, "Preencha o título, o horário e os dias.", 400)
		return
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	rt.Area = store.fixArea(rt.Area)
	if id := r.PathValue("id"); id != "" {
		found := false
		for i := range store.d.Routines {
			if store.d.Routines[i].ID == id {
				rt.ID = id
				store.d.Routines[i] = rt
				found = true
			}
		}
		if !found {
			http.Error(w, "não encontrada", 404)
			return
		}
		// horário mudou: permite avisar de novo hoje
		store.clearSent("r|" + time.Now().Format(dateFmt) + "|" + id)
	} else {
		rt.ID = newID()
		store.d.Routines = append(store.d.Routines, rt)
	}
	changedAndSave()
	writeJSON(w, rt)
}

func apiRoutineDelete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	store.mu.Lock()
	defer store.mu.Unlock()
	keep := store.d.Routines[:0]
	for _, x := range store.d.Routines {
		if x.ID != id {
			keep = append(keep, x)
		}
	}
	store.d.Routines = keep
	changedAndSave()
	w.WriteHeader(204)
}

func apiRoutineDone(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Date string `json:"date"`
		Done bool   `json:"done"`
	}
	if json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<12)).Decode(&body) != nil {
		http.Error(w, "inválido", 400)
		return
	}
	if _, err := time.Parse(dateFmt, body.Date); err != nil {
		http.Error(w, "data inválida", 400)
		return
	}
	id := r.PathValue("id")
	store.mu.Lock()
	defer store.mu.Unlock()
	list := []string{}
	for _, x := range store.d.Done[body.Date] {
		if x != id {
			list = append(list, x)
		}
	}
	if body.Done {
		list = append(list, id)
	}
	store.d.Done[body.Date] = list
	changedAndSave()
	w.WriteHeader(204)
}

func apiSettings(w http.ResponseWriter, r *http.Request) {
	var s Settings
	if json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<12)).Decode(&s) != nil {
		http.Error(w, "inválido", 400)
		return
	}
	if s.Lead < 1 || s.Lead > 180 {
		s.Lead = 10
	}
	if s.Theme != "light" && s.Theme != "system" {
		s.Theme = "dark"
	}
	if s.Volume < 5 || s.Volume > 100 {
		s.Volume = 80
	}
	s.UpdateRepo = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(s.UpdateRepo), "https://github.com/"))
	s.UpdateRepo = strings.TrimSuffix(s.UpdateRepo, "/")
	if s.UpdateRepo != "" && !repoRe.MatchString(s.UpdateRepo) {
		http.Error(w, "Repositório inválido. Use o formato dono/repositorio.", 400)
		return
	}
	if err := setAutoStart(s.AutoStart); err != nil {
		http.Error(w, "Não consegui alterar a inicialização: "+err.Error(), 500)
		return
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	old := store.d.Settings
	s.Initialized = true
	s.SoundFile, s.SoundName = old.SoundFile, old.SoundName // controlados pelo envio do áudio
	s.WallFile, s.WallName, s.WallRev = old.WallFile, old.WallName, old.WallRev
	if s.Wallpaper == "custom" && s.WallFile == "" {
		s.Wallpaper = ""
	}
	if !wallIDRe.MatchString(s.Wallpaper) {
		s.Wallpaper = ""
	}
	s.WallDim = min(max(s.WallDim, 0), 90)
	s.WallBlur = min(max(s.WallBlur, 0), 20)
	if s.Accent != "" && !colorRe.MatchString(s.Accent) {
		s.Accent = ""
	}
	if s.Sound != "custom" && s.Sound != "none" {
		s.Sound = "windows"
	}
	if s.Sound == "custom" && s.SoundFile == "" {
		s.Sound = "windows"
	}
	repoChanged := s.UpdateRepo != old.UpdateRepo
	store.d.Settings = s
	if repoChanged {
		go checkUpdate()
	}
	changedAndSave()
	writeJSON(w, s)
}

// usado pelo menu da bandeja
func toggleAutoStart() bool {
	store.mu.Lock()
	defer store.mu.Unlock()
	on := !store.d.Settings.AutoStart
	if setAutoStart(on) == nil {
		store.d.Settings.AutoStart = on
		changedAndSave()
	}
	return store.d.Settings.AutoStart
}

func autoStartOn() bool {
	store.mu.Lock()
	defer store.mu.Unlock()
	return store.d.Settings.AutoStart
}
