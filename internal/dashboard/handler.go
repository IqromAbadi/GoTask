package dashboard

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/iqromabadi/gotask/internal/middleware"
	"github.com/iqromabadi/gotask/internal/platform/response"
)

// Handler handles HTTP requests for dashboard.
type Handler struct {
	service *Service
	logger  *slog.Logger
}

// NewHandler creates a new dashboard Handler.
func NewHandler(service *Service, logger *slog.Logger) *Handler {
	return &Handler{service: service, logger: logger}
}

// RegisterRoutes registers dashboard routes on the given router.
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/dashboard/summary", h.Summary)
	r.Get("/dashboard/progress", h.Progress)
	r.Get("/dashboard/upcoming-deadlines", h.UpcomingDeadlines)
	r.Get("/dashboard/overdue-tasks", h.OverdueTasks)
	r.Get("/dashboard/priority-distribution", h.PriorityDistribution)
}

func (h *Handler) getUserID(r *http.Request) (uuid.UUID, error) {
	uid := middleware.GetUserID(r.Context())
	return uuid.Parse(uid)
}

// Summary handles GET /dashboard/summary.
func (h *Handler) Summary(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserID(r)
	if err != nil {
		response.Unauthorized(w, "User tidak terautentikasi")
		return
	}

	var listID *uuid.UUID
	if lid := r.URL.Query().Get("list_id"); lid != "" {
		id, err := uuid.Parse(lid)
		if err != nil {
			response.BadRequest(w, "list_id tidak valid")
			return
		}
		listID = &id
	}

	data, err := h.service.GetSummary(r.Context(), userID, listID)
	if err != nil {
		h.logger.Error("dashboard summary failed", slog.String("error", err.Error()))
		response.InternalError(w)
		return
	}
	response.OK(w, "Dashboard berhasil diambil", data)
}

// Progress handles GET /dashboard/progress.
func (h *Handler) Progress(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserID(r)
	if err != nil {
		response.Unauthorized(w, "User tidak terautentikasi")
		return
	}

	period := r.URL.Query().Get("period")
	if period == "" {
		period = "week"
	}

	data, err := h.service.GetProgressAnalytics(r.Context(), userID, period)
	if err != nil {
		h.logger.Error("dashboard progress failed", slog.String("error", err.Error()))
		response.InternalError(w)
		return
	}
	if data == nil {
		data = []ProgressPoint{}
	}
	response.OK(w, "Data berhasil diambil", data)
}

// UpcomingDeadlines handles GET /dashboard/upcoming-deadlines.
func (h *Handler) UpcomingDeadlines(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserID(r)
	if err != nil {
		response.Unauthorized(w, "User tidak terautentikasi")
		return
	}

	limit := 10
	if l := r.URL.Query().Get("limit"); l != "" {
		limit, _ = strconv.Atoi(l)
	}

	data, err := h.service.GetUpcomingDeadlines(r.Context(), userID, limit)
	if err != nil {
		h.logger.Error("dashboard deadlines failed", slog.String("error", err.Error()))
		response.InternalError(w)
		return
	}
	if data == nil {
		data = []map[string]any{}
	}
	response.OK(w, "Data berhasil diambil", data)
}

// OverdueTasks handles GET /dashboard/overdue-tasks.
func (h *Handler) OverdueTasks(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserID(r)
	if err != nil {
		response.Unauthorized(w, "User tidak terautentikasi")
		return
	}

	data, err := h.service.GetOverdueTasks(r.Context(), userID)
	if err != nil {
		h.logger.Error("dashboard overdue failed", slog.String("error", err.Error()))
		response.InternalError(w)
		return
	}
	if data == nil {
		data = []map[string]any{}
	}
	response.OK(w, "Data berhasil diambil", data)
}

// PriorityDistribution handles GET /dashboard/priority-distribution.
func (h *Handler) PriorityDistribution(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserID(r)
	if err != nil {
		response.Unauthorized(w, "User tidak terautentikasi")
		return
	}

	var listID *uuid.UUID
	if lid := r.URL.Query().Get("list_id"); lid != "" {
		id, err := uuid.Parse(lid)
		if err != nil {
			response.BadRequest(w, "list_id tidak valid")
			return
		}
		listID = &id
	}

	data, err := h.service.GetPriorityDistribution(r.Context(), userID, listID)
	if err != nil {
		h.logger.Error("dashboard priority failed", slog.String("error", err.Error()))
		response.InternalError(w)
		return
	}
	if data == nil {
		data = []PriorityCount{}
	}
	response.OK(w, "Data berhasil diambil", data)
}
