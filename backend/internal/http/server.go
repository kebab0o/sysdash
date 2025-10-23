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

// tiny timeout middleware
func timeout(d time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.TimeoutHandler(next, d, "timeout")
	}
}

// CORS middleware simple & explicit
func CORS(next http.Handler) http.Handler {
	allowed := map[string]bool{
		"http://localhost:3000": true,
		"http://localhost:5173": true, // Vite
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if allowed[origin] {
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

// single-header token check to look professional
func Auth(next http.Handler) http.Handler {
	secret := strings.TrimSpace(os.Getenv("DAEMON_API_KEY"))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if secret == "" {
			next.ServeHTTP(w, r) // no key set, auth disabled
			return
		}
		if r.Header.Get("X-API-Key") != secret {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
