package main

import (
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

var soundExts = map[string]bool{".mp3": true, ".wav": true, ".wma": true}

// soundConfig devolve como tocar o aviso: modo, caminho do arquivo e volume.
func soundConfig() (mode, path string, vol int) {
	store.mu.Lock()
	defer store.mu.Unlock()
	st := store.d.Settings
	mode, vol = st.Sound, st.Volume
	if mode == "custom" {
		path = filepath.Join(dataDir(), st.SoundFile)
		if _, err := os.Stat(path); err != nil {
			mode = "windows"
		}
	}
	return
}

// deliver mostra a notificação e toca o som escolhido.
func deliver(n Note) {
	mode, path, vol := soundConfig()
	switch mode {
	case "custom":
		notify(n.Title, n.Msg, true)
		playSound(path, vol)
	case "none":
		notify(n.Title, n.Msg, true)
	default:
		notify(n.Title, n.Msg, false)
	}
}

func apiSoundUpload(w http.ResponseWriter, r *http.Request) {
	name, _ := url.QueryUnescape(r.Header.Get("X-Filename"))
	name = filepath.Base(strings.ReplaceAll(name, "\\", "/"))
	ext := strings.ToLower(filepath.Ext(name))
	if !soundExts[ext] {
		http.Error(w, "Use um arquivo MP3, WAV ou WMA.", 400)
		return
	}
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 15<<20))
	if err != nil {
		http.Error(w, "Arquivo grande demais (máximo 15 MB).", 400)
		return
	}
	if len(body) < 100 {
		http.Error(w, "Arquivo de áudio vazio.", 400)
		return
	}
	file := "som-notificacao" + ext
	if err := os.WriteFile(filepath.Join(dataDir(), file), body, 0o644); err != nil {
		http.Error(w, "Não consegui salvar o áudio: "+err.Error(), 500)
		return
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	if old := store.d.Settings.SoundFile; old != "" && old != file {
		os.Remove(filepath.Join(dataDir(), old))
	}
	store.d.Settings.SoundFile = file
	store.d.Settings.SoundName = clip(name, 80)
	store.d.Settings.Sound = "custom"
	changedAndSave()
	writeJSON(w, store.d.Settings)
}

func apiSoundDelete(w http.ResponseWriter, r *http.Request) {
	store.mu.Lock()
	defer store.mu.Unlock()
	if f := store.d.Settings.SoundFile; f != "" {
		os.Remove(filepath.Join(dataDir(), f))
	}
	store.d.Settings.SoundFile, store.d.Settings.SoundName = "", ""
	if store.d.Settings.Sound == "custom" {
		store.d.Settings.Sound = "windows"
	}
	changedAndSave()
	w.WriteHeader(204)
}

func apiSoundPlay(w http.ResponseWriter, r *http.Request) {
	mode, path, vol := soundConfig()
	if mode != "custom" {
		http.Error(w, "Escolha um áudio primeiro.", 400)
		return
	}
	playSound(path, vol)
	w.WriteHeader(204)
}
