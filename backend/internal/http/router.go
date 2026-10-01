package httphandler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(h *Handler, corsOrigin string) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// CORS middleware
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", corsOrigin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Employee-Id")
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			next.ServeHTTP(w, r)
		})
	})

	r.Get("/health", h.Health)

	r.Route("/api", func(r chi.Router) {
		r.Get("/employees", h.ListEmployees)
		r.Get("/questions", h.ListQuestions)

		r.Group(func(r chi.Router) {
			r.Use(h.AuthMiddleware)
			r.Get("/me", h.GetMe)
			r.Get("/subordinates", h.ListSubordinates)
			r.Get("/employees/{id}/evaluations", h.GetEvaluations)
			r.Get("/employees/{id}/evaluations/latest", h.GetLatestEvaluation)
			r.Post("/employees/{id}/evaluations", h.CreateEvaluation)
		})
	})

	return r
}
