package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/kebab0o/sysdash/backend/internal/store"
)

type App struct{ Store *store.Memory }

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

	// metrics
	r.Get("/api/metrics/cpu", a.cpuMetrics)
	r.Get("/api/metrics/mem", a.memMetrics)
	r.Get("/api/metrics/disk", a.diskMetrics)
	r.Get("/api/metrics/diskio", a.diskIOMetrics)
	r.Get("/api/metrics/net", a.netMetrics)

	// maintenance
	r.Post("/api/tasks/prune", a.pruneRetention)
	return r
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func (a *App) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status": "ok", "now": time.Now().UTC().Format(time.RFC3339Nano),
		"lastCollectorAt": a.Store.LastCollector().Format(time.RFC3339Nano),
	})
}

/* ===== items ===== (unchanged) */
type createReq struct{ Title, Notes string }

func (a *App) listItems(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, a.Store.List())
}
func (a *App) createItem(w http.ResponseWriter, r *http.Request) {
	var b createReq
	if json.NewDecoder(r.Body).Decode(&b) != nil {
		http.Error(w, "bad json", 400)
		return
	}
	writeJSON(w, 201, a.Store.Create(b.Title, b.Notes))
}
func (a *App) getItem(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	it, err := a.Store.Get(id)
	if err != nil {
		http.Error(w, "not found", 404)
		return
	}
	writeJSON(w, 200, it)
}
func (a *App) updateItem(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var b createReq
	if json.NewDecoder(r.Body).Decode(&b) != nil {
		http.Error(w, "bad json", 400)
		return
	}
	it, err := a.Store.Update(id, b.Title, b.Notes)
	if err != nil {
		http.Error(w, "not found", 404)
		return
	}
	writeJSON(w, 200, it)
}
func (a *App) deleteItem(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if a.Store.Delete(id) != nil {
		http.Error(w, "not found", 404)
		return
	}
	w.WriteHeader(204)
}

/* ===== metrics ===== */
func parseRange(r *http.Request, def time.Duration) time.Duration {
	q := r.URL.Query().Get("range")
	if q == "" {
		return def
	}
	if len(q) > 1 && q[len(q)-1] == 'd' {
		if d, err := strconv.Atoi(q[:len(q)-1]); err == nil {
			return time.Duration(d) * 24 * time.Hour
		}
	}
	if d, err := time.ParseDuration(q); err == nil {
		return d
	}
	return def
}

func (a *App) cpuMetrics(w http.ResponseWriter, r *http.Request) {
	d := parseRange(r, time.Hour)
	since := time.Now().Add(-d)
	pts := a.Store.CPUSince(since)
	var avg, p95 float64
	n := len(pts)
	if n > 0 {
		vals := make([]float64, n)
		for i, p := range pts {
			avg += p.Usage
			vals[i] = p.Usage
		}
		avg /= float64(n)
		// sort vals
		for i := 0; i < n-1; i++ {
			for j := i + 1; j < n; j++ {
				if vals[j] < vals[i] {
					vals[i], vals[j] = vals[j], vals[i]
				}
			}
		}
		p95 = vals[int(float64(n-1)*0.95)]
	}
	writeJSON(w, 200, map[string]any{"range": d.String(), "points": pts, "avg": avg, "p95": p95})
}
func (a *App) memMetrics(w http.ResponseWriter, r *http.Request) {
	d := parseRange(r, time.Hour)
	since := time.Now().Add(-d)
	pts := a.Store.MemSince(since)
	var latest float64
	if len(pts) > 0 {
		latest = pts[len(pts)-1].UsedPct
	}
	writeJSON(w, 200, map[string]any{"range": d.String(), "points": pts, "latest": latest})
}
func (a *App) diskMetrics(w http.ResponseWriter, r *http.Request) {
	d := parseRange(r, 24*time.Hour)
	since := time.Now().Add(-d)
	writeJSON(w, 200, map[string]any{"range": d.String(), "mounts": a.Store.DiskSince(since)})
}
func (a *App) diskIOMetrics(w http.ResponseWriter, r *http.Request) {
	d := parseRange(r, time.Hour)
	since := time.Now().Add(-d)
	writeJSON(w, 200, map[string]any{"range": d.String(), "points": a.Store.DiskIOSince(since)})
}
func (a *App) netMetrics(w http.ResponseWriter, r *http.Request) {
	d := parseRange(r, time.Hour)
	since := time.Now().Add(-d)
	writeJSON(w, 200, map[string]any{"range": d.String(), "points": a.Store.NetSince(since)})
}

/* ===== maintenance ===== */
func (a *App) pruneRetention(w http.ResponseWriter, r *http.Request) {
	a.Store.PruneForRetention()
	writeJSON(w, 200, map[string]string{"status": "pruned"})
}
