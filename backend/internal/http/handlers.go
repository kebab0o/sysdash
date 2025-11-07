package http

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/kebab0o/sysdash/backend/internal/store"
	"github.com/kebab0o/sysdash/backend/internal/types"
)

type App struct {
	Store *store.Memory
}

func (a *App) Routes() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID, middleware.RealIP, middleware.Recoverer)
	r.Use(CORS, Auth)

	r.Get("/api/health", a.getHealth)

	r.Route("/api/metrics", func(r chi.Router) {
		r.Get("/cpu", a.getCPU)
		r.Get("/mem", a.getMem)
		r.Get("/disk", a.getDisk)
		r.Get("/diskio", a.getDiskIO)
		r.Get("/net", a.getNet)
	})

	r.Route("/api/tasks", func(r chi.Router) {
		r.Get("/", a.listTasks)
		r.Post("/", a.createTask)
		r.Post("/{id}/run", a.runTask)
		r.Delete("/{id}", a.deleteTask)
	})

	r.Get("/api/logs", a.listLogs)

	return r
}

func (a *App) getHealth(w http.ResponseWriter, r *http.Request) {
	out := map[string]any{
		"status":          "ok",
		"now":             time.Now().UTC(),
		"lastCollectorAt": a.Store.LastCollector(),
	}
	writeJSON(w, out)
}

func parseRange(r *http.Request, def string) time.Duration {
	q := r.URL.Query().Get("range")
	if q == "" {
		q = def
	}
	d, err := time.ParseDuration(q)
	if err != nil {
		switch q {
		case "1h":
			return time.Hour
		case "24h":
			return 24 * time.Hour
		default:
			return time.Hour
		}
	}
	return d
}

func (a *App) getCPU(w http.ResponseWriter, r *http.Request) {
	d := parseRange(r, "1h")
	since := time.Now().Add(-d)
	pts := a.Store.CPUSince(since)
	avg, p95 := calcAvgP(pts)

	out := struct {
		Range  string           `json:"range"`
		Points []types.CPUPoint `json:"points"`
		Avg    float64          `json:"avg"`
		P95    float64          `json:"p95"`
	}{
		Range:  r.URL.Query().Get("range"),
		Points: pts,
		Avg:    avg,
		P95:    p95,
	}
	writeJSON(w, out)
}

func (a *App) getMem(w http.ResponseWriter, r *http.Request) {
	d := parseRange(r, "1h")
	since := time.Now().Add(-d)
	pts := a.Store.MemSince(since)
	latest := 0.0
	if n := len(pts); n > 0 {
		latest = pts[n-1].V
	}
	out := struct {
		Range  string           `json:"range"`
		Points []types.MemPoint `json:"points"`
		Latest float64          `json:"latest"`
	}{
		Range:  r.URL.Query().Get("range"),
		Points: pts,
		Latest: latest,
	}
	writeJSON(w, out)
}

func (a *App) getDisk(w http.ResponseWriter, r *http.Request) {
	d := parseRange(r, "24h")
	since := time.Now().Add(-d)
	series := a.Store.DiskSince(since)
	out := struct {
		Range  string             `json:"range"`
		Mounts []store.DiskSeries `json:"mounts"`
	}{
		Range:  r.URL.Query().Get("range"),
		Mounts: series,
	}
	writeJSON(w, out)
}

func (a *App) getDiskIO(w http.ResponseWriter, r *http.Request) {
	d := parseRange(r, "1h")
	since := time.Now().Add(-d)
	pts := a.Store.DiskIOSince(since)
	out := struct {
		Range  string              `json:"range"`
		Points []types.DiskIOPoint `json:"points"`
	}{
		Range:  r.URL.Query().Get("range"),
		Points: pts,
	}
	writeJSON(w, out)
}

func (a *App) getNet(w http.ResponseWriter, r *http.Request) {
	d := parseRange(r, "1h")
	since := time.Now().Add(-d)
	pts := a.Store.NetSince(since)
	out := struct {
		Range  string           `json:"range"`
		Points []types.NetPoint `json:"points"`
	}{
		Range:  r.URL.Query().Get("range"),
		Points: pts,
	}
	writeJSON(w, out)
}

func (a *App) listTasks(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, a.Store.ListTasks())
}
func (a *App) createTask(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name         string `json:"name"`
		EveryMinutes int    `json:"everyMinutes"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	if body.Name == "" || body.EveryMinutes <= 0 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	t := a.Store.CreateTask(body.Name, body.EveryMinutes)
	writeJSON(w, t)
}
func (a *App) runTask(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if err := a.Store.RunTaskNow(id); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, map[string]string{"status": "ok"})
}
func (a *App) deleteTask(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if err := a.Store.DeleteTask(id); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *App) listLogs(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, a.Store.ListLogs(300, r.URL.Query().Get("q")))
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func calcAvgP(pts []types.CPUPoint) (avg, p95 float64) {
	n := len(pts)
	if n == 0 {
		return 0, 0
	}
	var sum float64
	values := make([]float64, n)
	for i, p := range pts {
		sum += p.V
		values[i] = p.V
	}
	avg = sum / float64(n)
	idx := int(0.95 * float64(n-1))
	for i := 0; i <= idx; i++ {
		min := i
		for j := i + 1; j < n; j++ {
			if values[j] < values[min] {
				min = j
			}
		}
		values[i], values[min] = values[min], values[i]
	}
	p95 = values[idx]
	return
}
