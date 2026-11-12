package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

// appVersion é definido na compilação: -ldflags "-X main.appVersion=1.1.0"
var appVersion = "dev"

// Repositório público no GitHub onde ficam as versões (Releases).
// Pode ser trocado em Ajustes → Atualizações.
var defaultUpdateRepo = "Lcszera/nahora"

const exeAsset = "naHora.exe"

var repoRe = regexp.MustCompile(`^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$`)

type UpdateInfo struct {
	Version string `json:"version"`
	Latest  string `json:"latest"`
	Status  string `json:"status"` // idle checking downloading ready uptodate error norepo applying
	Message string `json:"message"`
	Notes   string `json:"notes"`
	Checked string `json:"checked"`
	Repo    string `json:"repo"`
}

var upd struct {
	sync.Mutex
	info     UpdateInfo
	path     string // exe baixado e verificado, pronto para instalar
	busy     bool
	applying bool
}

func updateRepo() string {
	store.mu.Lock()
	defer store.mu.Unlock()
	if r := store.d.Settings.UpdateRepo; r != "" {
		return r
	}
	return defaultUpdateRepo
}

func autoUpdateOn() bool {
	store.mu.Lock()
	defer store.mu.Unlock()
	return !store.d.Settings.NoAutoUpdate
}

func updateSnapshot() UpdateInfo {
	upd.Lock()
	defer upd.Unlock()
	i := upd.info
	i.Version = appVersion
	i.Repo = updateRepo()
	if i.Status == "" {
		i.Status = "idle"
	}
	return i
}

func setUpd(status, msg string) {
	upd.Lock()
	upd.info.Status, upd.info.Message = status, msg
	upd.Unlock()
	hub.broadcast("update")
}

// parseVer transforma "v1.2.3" em [1 2 3].
func parseVer(v string) []int {
	v = strings.TrimPrefix(strings.TrimSpace(v), "v")
	out := []int{}
	for _, p := range strings.Split(v, ".") {
		n, err := strconv.Atoi(strings.TrimFunc(p, func(r rune) bool { return r < '0' || r > '9' }))
		if err != nil {
			return []int{0}
		}
		out = append(out, n)
	}
	return out
}

func newerThan(a, b string) bool {
	x, y := parseVer(a), parseVer(b)
	for i := 0; i < len(x) || i < len(y); i++ {
		var p, q int
		if i < len(x) {
			p = x[i]
		}
		if i < len(y) {
			q = y[i]
		}
		if p != q {
			return p > q
		}
	}
	return false
}

func updateLoop() {
	time.Sleep(90 * time.Second)
	for {
		if autoUpdateOn() {
			checkUpdate()
		}
		time.Sleep(6 * time.Hour)
	}
}

type ghRelease struct {
	Tag    string `json:"tag_name"`
	Name   string `json:"name"`
	Body   string `json:"body"`
	Draft  bool   `json:"draft"`
	Assets []struct {
		Name string `json:"name"`
		URL  string `json:"browser_download_url"`
		Size int64  `json:"size"`
	} `json:"assets"`
}

var httpc = &http.Client{Timeout: 3 * time.Minute}

func apiBase() string {
	if b := os.Getenv("NAHORA_UPDATE_API"); b != "" { // usado nos testes
		return strings.TrimRight(b, "/")
	}
	return "https://api.github.com"
}

func get(url string) (*http.Response, error) {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "naHora/"+appVersion)
	req.Header.Set("Accept", "application/vnd.github+json")
	res, err := httpc.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != 200 {
		res.Body.Close()
		return nil, fmt.Errorf("HTTP %d", res.StatusCode)
	}
	return res, nil
}

// checkUpdate procura uma versão nova, baixa e confere. Retorna quando termina.
func checkUpdate() {
	upd.Lock()
	if upd.busy || upd.info.Status == "ready" || upd.info.Status == "applying" {
		upd.Unlock()
		return
	}
	upd.busy = true
	upd.Unlock()
	defer func() { upd.Lock(); upd.busy = false; upd.Unlock(); hub.broadcast("update") }()

	repo := updateRepo()
	if repo == "" {
		setUpd("norepo", "Nenhum repositório de atualizações configurado.")
		return
	}
	setUpd("checking", "Procurando atualização…")
	upd.Lock()
	upd.info.Checked = time.Now().Format("02/01 15:04")
	upd.Unlock()

	res, err := get(apiBase() + "/repos/" + repo + "/releases/latest")
	if err != nil {
		setUpd("error", "Não consegui consultar as atualizações ("+err.Error()+").")
		return
	}
	var rel ghRelease
	err = json.NewDecoder(io.LimitReader(res.Body, 1<<20)).Decode(&rel)
	res.Body.Close()
	if err != nil || rel.Tag == "" {
		setUpd("error", "Resposta inválida do GitHub.")
		return
	}
	if !newerThan(rel.Tag, appVersion) {
		setUpd("uptodate", "Você está na versão mais recente.")
		return
	}
	var exeURL, shaURL string
	var size int64
	for _, a := range rel.Assets {
		switch a.Name {
		case exeAsset:
			exeURL, size = a.URL, a.Size
		case exeAsset + ".sha256":
			shaURL = a.URL
		}
	}
	if exeURL == "" || shaURL == "" || size > 80<<20 {
		setUpd("error", "A versão "+rel.Tag+" não tem os arquivos esperados.")
		return
	}
	upd.Lock()
	upd.info.Latest = strings.TrimPrefix(rel.Tag, "v")
	upd.info.Notes = strings.TrimSpace(firstNonEmpty(rel.Name, rel.Body))
	upd.Unlock()
	setUpd("downloading", "Baixando a versão "+rel.Tag+"…")

	path, err := download(exeURL, shaURL, strings.TrimPrefix(rel.Tag, "v"))
	if err != nil {
		setUpd("error", "Falha ao baixar a atualização: "+err.Error())
		return
	}
	upd.Lock()
	upd.path = path
	upd.Unlock()
	setUpd("ready", "A versão "+rel.Tag+" está pronta para instalar.")
	if hub.count() == 0 {
		go applyUpdate() // ninguém com a janela aberta: instala já
	}
}

func firstNonEmpty(a ...string) string {
	for _, s := range a {
		if strings.TrimSpace(s) != "" {
			return s
		}
	}
	return ""
}

func updateDir() string { return filepath.Join(dataDir(), "atualizacao") }

func download(exeURL, shaURL, ver string) (string, error) {
	res, err := get(shaURL)
	if err != nil {
		return "", err
	}
	shaTxt, _ := io.ReadAll(io.LimitReader(res.Body, 4096))
	res.Body.Close()
	fields := strings.Fields(string(shaTxt))
	if len(fields) == 0 || len(fields[0]) != 64 {
		return "", errors.New("arquivo de verificação inválido")
	}
	want := strings.ToLower(fields[0])

	os.MkdirAll(updateDir(), 0o755)
	path := filepath.Join(updateDir(), "naHora-"+ver+".exe")
	res, err = get(exeURL)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	f, err := os.Create(path)
	if err != nil {
		return "", err
	}
	h := sha256.New()
	_, err = io.Copy(io.MultiWriter(f, h), io.LimitReader(res.Body, 80<<20))
	f.Close()
	if err != nil {
		os.Remove(path)
		return "", err
	}
	if hex.EncodeToString(h.Sum(nil)) != want {
		os.Remove(path)
		return "", errors.New("o arquivo baixado não confere (SHA-256)")
	}
	return path, nil
}

// applyUpdate troca o executável e reinicia. Os dados ficam na pasta de dados e não são tocados.
func applyUpdate() {
	upd.Lock()
	if upd.applying || upd.path == "" {
		upd.Unlock()
		return
	}
	upd.applying = true
	path, notes, latest := upd.path, upd.info.Notes, upd.info.Latest
	upd.Unlock()
	setUpd("applying", "Instalando e reiniciando…")
	if err := swapExe(path, latest, notes); err != nil {
		upd.Lock()
		upd.applying, upd.path = false, ""
		upd.Unlock()
		os.Remove(path)
		setUpd("error", "Não consegui instalar: "+err.Error())
		notesQ <- Note{"⚠️ Atualização do naHora", "Não consegui instalar a nova versão.\n" + err.Error()}
		return
	}
	restartSelf()
}

// restartSelf encerra esta instância e abre a nova versão na bandeja.
func restartSelf() {
	platformBeforeExit()
	store.mu.Lock() // impede gravações durante a troca
	if listener != nil {
		listener.Close()
	}
	cmd := exec.Command(exePath, "--tray", "--after-update")
	cmd.Dir = filepath.Dir(exePath)
	if err := cmd.Start(); err == nil {
		cmd.Process.Release()
	}
	os.Exit(0)
}

func swapExe(newPath, ver, notes string) error {
	exe := exePath
	old := oldExePath(exe)
	os.Remove(old)
	if err := os.Rename(exe, old); err != nil {
		return fmt.Errorf("sem permissão na pasta do programa (%v). Coloque o naHora.exe em uma pasta sua, como Documentos", err)
	}
	if err := copyFile(newPath, exe); err != nil {
		os.Remove(exe)
		os.Rename(old, exe)
		return err
	}
	os.Remove(newPath)
	marker, _ := json.Marshal(map[string]string{"from": appVersion, "to": ver, "notes": notes})
	os.WriteFile(filepath.Join(dataDir(), "atualizado.json"), marker, 0o644)
	return nil
}

func oldExePath(exe string) string {
	ext := filepath.Ext(exe)
	return strings.TrimSuffix(exe, ext) + ".antigo" + ext
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}

// afterStart: limpa o exe antigo e avisa se acabou de atualizar.
func afterStart() {
	if exePath != "" {
		go func() {
			old := oldExePath(exePath)
			for i := 0; i < 40; i++ {
				if err := os.Remove(old); err == nil || os.IsNotExist(err) {
					return
				}
				time.Sleep(500 * time.Millisecond)
			}
		}()
	}
	os.RemoveAll(updateDir())
	mf := filepath.Join(dataDir(), "atualizado.json")
	if b, err := os.ReadFile(mf); err == nil {
		os.Remove(mf)
		var m map[string]string
		json.Unmarshal(b, &m)
		msg := "Agora você está na versão " + appVersion + "."
		if n := strings.TrimSpace(m["notes"]); n != "" {
			if i := strings.IndexByte(n, '\n'); i > 0 {
				n = n[:i]
			}
			msg += "\n" + clip(n, 120)
		}
		go func() { time.Sleep(3 * time.Second); notesQ <- Note{"⚡ naHora atualizado", msg} }()
	}
}

// ---------- API ----------

func apiUpdateCheck(w http.ResponseWriter, r *http.Request) {
	go checkUpdate()
	w.WriteHeader(202)
}

func apiUpdateApply(w http.ResponseWriter, r *http.Request) {
	upd.Lock()
	ok := upd.path != ""
	upd.Unlock()
	if !ok {
		http.Error(w, "Nenhuma atualização pronta.", 400)
		return
	}
	go func() { time.Sleep(400 * time.Millisecond); applyUpdate() }()
	w.WriteHeader(202)
}

// apiUpdateFile instala um naHora.exe escolhido pelo usuário.
func apiUpdateFile(w http.ResponseWriter, r *http.Request) {
	os.MkdirAll(updateDir(), 0o755)
	path := filepath.Join(updateDir(), "naHora-arquivo.exe")
	f, err := os.Create(path)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	n, err := io.Copy(f, http.MaxBytesReader(w, r.Body, 80<<20))
	f.Close()
	if err != nil || n < 1<<20 {
		os.Remove(path)
		http.Error(w, "Arquivo inválido.", 400)
		return
	}
	head := make([]byte, 2)
	if f, err := os.Open(path); err == nil {
		f.Read(head)
		f.Close()
	}
	if !exeHeaderOK(head) {
		os.Remove(path)
		http.Error(w, "Esse arquivo não é um executável do Windows.", 400)
		return
	}
	upd.Lock()
	upd.path, upd.info.Latest, upd.info.Notes = path, "", "Instalado a partir de um arquivo."
	upd.Unlock()
	go func() { time.Sleep(400 * time.Millisecond); applyUpdate() }()
	w.WriteHeader(202)
}
