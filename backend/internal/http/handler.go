package httphandler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"leadtrack/backend/internal/domain"
	"leadtrack/backend/internal/service"
)

type contextKey string

const employeeContextKey contextKey = "employeeId"

type Handler struct {
	service *service.Service
	loc     *time.Location
}

func New(svc *service.Service, loc *time.Location) *Handler {
	return &Handler{service: svc, loc: loc}
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, code string, message string, status int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	response := map[string]map[string]string{
		"error": {
			"code":    code,
			"message": message,
		},
	}
	json.NewEncoder(w).Encode(response)
}

func handleError(w http.ResponseWriter, err error) {
	if errors.Is(err, domain.ErrNotFound) {
		writeError(w, "NOT_FOUND", err.Error(), http.StatusNotFound)
		return
	}
	if errors.Is(err, domain.ErrUnauthorized) {
		writeError(w, "UNAUTHORIZED", err.Error(), http.StatusUnauthorized)
		return
	}
	if errors.Is(err, domain.ErrForbidden) || errors.Is(err, domain.ErrOutOfHierarchy) || errors.Is(err, domain.ErrSelfEvaluation) {
		writeError(w, "FORBIDDEN", err.Error(), http.StatusForbidden)
		return
	}
	if errors.Is(err, domain.ErrAlreadyEvaluated) {
		writeError(w, "ALREADY_EVALUATED_THIS_WEEK", err.Error(), http.StatusConflict)
		return
	}
	if errors.Is(err, domain.ErrInvalidAnswers) {
		writeError(w, "INVALID_ANSWERS", err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// Check Postgres unique violation code 23505
	if err.Error() != "" && (errors.Is(err, domain.ErrAlreadyEvaluated) || containsCode(err.Error(), "23505")) {
		writeError(w, "ALREADY_EVALUATED_THIS_WEEK", "Already evaluated this week", http.StatusConflict)
		return
	}

	writeError(w, "INTERNAL_ERROR", "Internal server error", http.StatusInternalServerError)
}

func containsCode(s, code string) bool {
	return len(s) >= len(code) && (s == code || containsSubstring(s, code))
}

func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s[:len(substr)] == substr || containsSubstring(s[1:], substr))
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) ListEmployees(w http.ResponseWriter, r *http.Request) {
	emps, err := h.service.ListEmployees(r.Context())
	if err != nil {
		handleError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, emps)
}

func (h *Handler) GetMe(w http.ResponseWriter, r *http.Request) {
	empID, ok := r.Context().Value(employeeContextKey).(int)
	if !ok {
		writeError(w, "UNAUTHORIZED", "Unauthorized", http.StatusUnauthorized)
		return
	}

	emp, err := h.service.GetMe(r.Context(), empID)
	if err != nil {
		handleError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, emp)
}

func (h *Handler) ListQuestions(w http.ResponseWriter, r *http.Request) {
	questions, err := h.service.ListQuestions(r.Context())
	if err != nil {
		handleError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, questions)
}

func (h *Handler) ListSubordinates(w http.ResponseWriter, r *http.Request) {
	empID, ok := r.Context().Value(employeeContextKey).(int)
	if !ok {
		writeError(w, "UNAUTHORIZED", "Unauthorized", http.StatusUnauthorized)
		return
	}

	subs, err := h.service.ListSubordinates(r.Context(), empID, h.loc)
	if err != nil {
		handleError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, subs)
}

func (h *Handler) CreateEvaluation(w http.ResponseWriter, r *http.Request) {
	empID, ok := r.Context().Value(employeeContextKey).(int)
	if !ok {
		writeError(w, "UNAUTHORIZED", "Unauthorized", http.StatusUnauthorized)
		return
	}

	evaluatedIDStr := chi.URLParam(r, "id")
	evaluatedID, err := strconv.Atoi(evaluatedIDStr)
	if err != nil {
		writeError(w, "INVALID_ID", "Invalid employee id", http.StatusBadRequest)
		return
	}

	var input domain.EvaluationInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, "INVALID_JSON", "Invalid request body", http.StatusBadRequest)
		return
	}

	evalID, err := h.service.CreateEvaluation(r.Context(), empID, evaluatedID, input, h.loc)
	if err != nil {
		handleError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]int{"id": evalID})
}

func (h *Handler) GetEvaluations(w http.ResponseWriter, r *http.Request) {
	empID, ok := r.Context().Value(employeeContextKey).(int)
	if !ok {
		writeError(w, "UNAUTHORIZED", "Unauthorized", http.StatusUnauthorized)
		return
	}

	evaluatedIDStr := chi.URLParam(r, "id")
	evaluatedID, err := strconv.Atoi(evaluatedIDStr)
	if err != nil {
		writeError(w, "INVALID_ID", "Invalid employee id", http.StatusBadRequest)
		return
	}

	evals, err := h.service.GetEvaluations(r.Context(), empID, evaluatedID)
	if err != nil {
		handleError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, evals)
}

func (h *Handler) GetLatestEvaluation(w http.ResponseWriter, r *http.Request) {
	empID, ok := r.Context().Value(employeeContextKey).(int)
	if !ok {
		writeError(w, "UNAUTHORIZED", "Unauthorized", http.StatusUnauthorized)
		return
	}

	evaluatedIDStr := chi.URLParam(r, "id")
	evaluatedID, err := strconv.Atoi(evaluatedIDStr)
	if err != nil {
		writeError(w, "INVALID_ID", "Invalid employee id", http.StatusBadRequest)
		return
	}

	eval, err := h.service.GetLatestEvaluation(r.Context(), empID, evaluatedID)
	if err != nil {
		handleError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, eval)
}

func (h *Handler) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		empIDHeader := r.Header.Get("X-Employee-Id")
		if empIDHeader == "" {
			writeError(w, "UNAUTHORIZED", "Missing X-Employee-Id header", http.StatusUnauthorized)
			return
		}

		empID, err := strconv.Atoi(empIDHeader)
		if err != nil {
			writeError(w, "UNAUTHORIZED", "Invalid X-Employee-Id header", http.StatusUnauthorized)
			return
		}

		// Verify employee exists
		_, err = h.service.GetMe(r.Context(), empID)
		if err != nil {
			writeError(w, "UNAUTHORIZED", "Employee not found", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), employeeContextKey, empID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
