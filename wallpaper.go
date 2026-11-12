package main

import (
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

var wallExts = map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".webp": true}

func apiWallUpload(w http.ResponseWriter, r *http.Request) {
	name, _ := url.QueryUnescape(r.Header.Get("X-Filename"))
	name = filepath.Base(strings.ReplaceAll(name, "\\", "/"))
	ext := strings.ToLower(filepath.Ext(name))
	if !wallExts[ext] {
		http.Error(w, "Use uma imagem PNG, JPG ou WEBP.", 400)
		return
	}
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 25<<20))
	if err != nil {
		http.Error(w, "Imagem grande demais (máximo 25 MB).", 400)
		return
	}
	if ct := http.DetectContentType(body); !strings.HasPrefix(ct, "image/") {
		http.Error(w, "Esse arquivo não parece ser uma imagem.", 400)
		return
	}
	file := "wallpaper" + ext
	if err := os.WriteFile(filepath.Join(dataDir(), file), body, 0o644); err != nil {
		http.Error(w, "Não consegui salvar a imagem: "+err.Error(), 500)
		return
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	if old := store.d.Settings.WallFile; old != "" && old != file {
		os.Remove(filepath.Join(dataDir(), old))
	}
	st := &store.d.Settings
	st.WallFile, st.WallName, st.Wallpaper = file, clip(name, 80), "custom"
	st.WallRev++
	changedAndSave()
	writeJSON(w, st)
}

func apiWallDelete(w http.ResponseWriter, r *http.Request) {
	store.mu.Lock()
	defer store.mu.Unlock()
	st := &store.d.Settings
	if st.WallFile != "" {
		os.Remove(filepath.Join(dataDir(), st.WallFile))
	}
	st.WallFile, st.WallName = "", ""
	if st.Wallpaper == "custom" {
		st.Wallpaper = ""
	}
	changedAndSave()
	w.WriteHeader(204)
}

func serveCustomWall(w http.ResponseWriter, r *http.Request) {
	store.mu.Lock()
	f := store.d.Settings.WallFile
	store.mu.Unlock()
	if f == "" {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, filepath.Join(dataDir(), f))
}
