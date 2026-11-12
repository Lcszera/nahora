package main

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"
	"unicode/utf8"
)

var colorRe = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

func clip(s string, n int) string {
	s = strings.TrimSpace(s)
	if utf8.RuneCountInString(s) <= n {
		return s
	}
	return string([]rune(s)[:n])
}

func decode(w http.ResponseWriter, r *http.Request, v any) bool {
	if json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<16)).Decode(v) != nil {
		http.Error(w, "Dados inválidos.", 400)
		return false
	}
	return true
}

// ---------- catálogos ----------

func apiCatalogSave(w http.ResponseWriter, r *http.Request) {
	var c Catalog
	if !decode(w, r, &c) {
		return
	}
	c.Name, c.Emoji = clip(c.Name, 40), clip(c.Emoji, 4)
	if c.Name == "" {
		http.Error(w, "Dê um nome ao catálogo.", 400)
		return
	}
	if c.Emoji == "" {
		c.Emoji = "📁"
	}
	if !colorRe.MatchString(c.Color) {
		c.Color = "#3d8bff"
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	if id := r.PathValue("id"); id != "" {
		old := store.catalog(id)
		if old == nil {
			http.Error(w, "não encontrado", 404)
			return
		}
		c.ID, c.Order = id, old.Order
		*old = c
	} else {
		if dup := store.catalogByName(c.Name); dup != nil {
			http.Error(w, "Já existe um catálogo com esse nome.", 409)
			return
		}
		c.ID = newID()
		for _, x := range store.d.Catalogs {
			if x.Order >= c.Order {
				c.Order = x.Order + 1
			}
		}
		store.d.Catalogs = append(store.d.Catalogs, c)
	}
	changedAndSave()
	writeJSON(w, c)
}

func apiCatalogDelete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	store.mu.Lock()
	defer store.mu.Unlock()
	cats := store.d.Catalogs[:0]
	for _, c := range store.d.Catalogs {
		if c.ID != id {
			cats = append(cats, c)
		}
	}
	store.d.Catalogs = cats
	gone := map[string]bool{}
	groups := store.d.Groups[:0]
	for _, g := range store.d.Groups {
		if g.Catalog == id {
			gone[g.ID] = true
		} else {
			groups = append(groups, g)
		}
	}
	store.d.Groups = groups
	topics := store.d.Topics[:0]
	for _, t := range store.d.Topics {
		if !gone[t.Group] {
			topics = append(topics, t)
		}
	}
	store.d.Topics = topics
	// tarefas e rotinas continuam existindo, só ficam sem catálogo
	for i := range store.d.Tasks {
		if store.d.Tasks[i].Area == id {
			store.d.Tasks[i].Area = ""
		}
	}
	for i := range store.d.Routines {
		if store.d.Routines[i].Area == id {
			store.d.Routines[i].Area = ""
		}
	}
	changedAndSave()
	w.WriteHeader(204)
}

// ---------- etapas / seções ----------

func apiGroupSave(w http.ResponseWriter, r *http.Request) {
	var g Group
	if !decode(w, r, &g) {
		return
	}
	g.Title, g.Gate = clip(g.Title, 80), clip(g.Gate, 300)
	if g.Title == "" {
		http.Error(w, "Dê um nome à seção.", 400)
		return
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	if id := r.PathValue("id"); id != "" {
		old := store.group(id)
		if old == nil {
			http.Error(w, "não encontrada", 404)
			return
		}
		g.ID, g.Catalog, g.Order = id, old.Catalog, old.Order
		*old = g
	} else {
		if store.catalog(g.Catalog) == nil {
			http.Error(w, "Catálogo não encontrado.", 400)
			return
		}
		g.ID, g.Order = newID(), 0
		for _, x := range store.d.Groups {
			if x.Catalog == g.Catalog && x.Order >= g.Order {
				g.Order = x.Order + 1
			}
		}
		store.d.Groups = append(store.d.Groups, g)
	}
	changedAndSave()
	writeJSON(w, g)
}

func apiGroupDelete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	store.mu.Lock()
	defer store.mu.Unlock()
	groups := store.d.Groups[:0]
	for _, g := range store.d.Groups {
		if g.ID != id {
			groups = append(groups, g)
		}
	}
	store.d.Groups = groups
	topics := store.d.Topics[:0]
	for _, t := range store.d.Topics {
		if t.Group != id {
			topics = append(topics, t)
		}
	}
	store.d.Topics = topics
	changedAndSave()
	w.WriteHeader(204)
}

// ---------- assuntos ----------

func apiTopicSave(w http.ResponseWriter, r *http.Request) {
	var t Topic
	if !decode(w, r, &t) {
		return
	}
	t.Title, t.Note, t.Src = clip(t.Title, 120), clip(t.Note, 400), clip(t.Src, 12)
	if t.Title == "" {
		http.Error(w, "Escreva o assunto.", 400)
		return
	}
	if t.Prio < 1 || t.Prio > 4 {
		t.Prio = 1
	}
	if t.Status < 0 || t.Status > 2 {
		t.Status = 0
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	if store.group(t.Group) == nil {
		http.Error(w, "Seção não encontrada.", 400)
		return
	}
	if id := r.PathValue("id"); id != "" {
		for i := range store.d.Topics {
			if store.d.Topics[i].ID == id {
				t.ID = id
				if t.Group == store.d.Topics[i].Group {
					t.Order = store.d.Topics[i].Order
				} else {
					t.Order = nextTopicOrder(t.Group)
				}
				store.d.Topics[i] = t
				changedAndSave()
				writeJSON(w, t)
				return
			}
		}
		http.Error(w, "não encontrado", 404)
		return
	}
	t.ID, t.Order = newID(), nextTopicOrder(t.Group)
	store.d.Topics = append(store.d.Topics, t)
	changedAndSave()
	writeJSON(w, t)
}

func nextTopicOrder(group string) int {
	n := 0
	for _, x := range store.d.Topics {
		if x.Group == group && x.Order >= n {
			n = x.Order + 1
		}
	}
	return n
}

func apiTopicStatus(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Status int `json:"status"`
	}
	if !decode(w, r, &body) {
		return
	}
	if body.Status < 0 || body.Status > 2 {
		http.Error(w, "inválido", 400)
		return
	}
	id := r.PathValue("id")
	store.mu.Lock()
	defer store.mu.Unlock()
	for i := range store.d.Topics {
		if store.d.Topics[i].ID == id {
			store.d.Topics[i].Status = body.Status
			changedAndSave()
			w.WriteHeader(204)
			return
		}
	}
	http.Error(w, "não encontrado", 404)
}

func apiTopicDelete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	store.mu.Lock()
	defer store.mu.Unlock()
	topics := store.d.Topics[:0]
	for _, t := range store.d.Topics {
		if t.ID != id {
			topics = append(topics, t)
		}
	}
	store.d.Topics = topics
	changedAndSave()
	w.WriteHeader(204)
}
