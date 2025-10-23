package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/kebab0o/sysdash/backend/internal/store"
)

type App struct {
	Store *store.Memory
}

func (a *App) Routes() http.Handler {
	r := chi.NewRouter()
	r.Use(CORS, Auth)

	r.Get("/api/health", a.health)

	r.Route("/api/items", func(r chi.Router) {
		r.Get("/", a.listItems)
		r.Post("/", a.createItem)
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", a.getItem)
			r.Put("/", a.updateItem)
			r.Delete("/", a.deleteItem)
		})
	})

	return r
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func (a *App) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (a *App) listItems(w http.ResponseWriter, r *http.Request) {
	items := a.Store.List()
	writeJSON(w, http.StatusOK, items)
}

type createReq struct {
	Title string `json:"title"`
	Notes string `json:"notes"`
}

func (a *App) createItem(w http.ResponseWriter, r *http.Request) {
	var body createReq
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}
	it := a.Store.Create(body.Title, body.Notes)
	writeJSON(w, http.StatusCreated, it)
}

func (a *App) getItem(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	it, err := a.Store.Get(id)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, it)
}

func (a *App) updateItem(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var body createReq
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}
	it, err := a.Store.Update(id, body.Title, body.Notes)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, it)
}

func (a *App) deleteItem(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := a.Store.Delete(id); err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
