package http

import (
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	chim "github.com/go-chi/chi/v5/middleware"
)

type Server struct {
	Router *chi.Mux
}

func NewServer(h http.Handler) *Server {
	r := chi.NewRouter()
	r.Use(chim.RequestID, chim.RealIP, chim.Logger, chim.Recoverer, timeout(30*time.Second))
	r.Mount("/", h)
	return &Server{Router: r}
}

func timeout(d time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler { return http.TimeoutHandler(next, d, "timeout") }
}

func CORS(next http.Handler) http.Handler {
	origins := map[string]bool{
		"http://localhost:3000": true,
		"http://localhost:5173": true,
	}
	if extra := strings.TrimSpace(os.Getenv("ALLOWED_ORIGINS")); extra != "" {
		for _, o := range strings.Split(extra, ",") {
			origins[strings.TrimSpace(o)] = true
		}
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origins[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-API-Key")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func Auth(next http.Handler) http.Handler {
	secret := strings.TrimSpace(os.Getenv("DAEMON_API_KEY"))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if secret == "" {
			next.ServeHTTP(w, r)
			return
		}
		if r.Header.Get("X-API-Key") != secret {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
