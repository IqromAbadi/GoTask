package activity

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/iqromabadi/gotask/internal/middleware"
	"github.com/iqromabadi/gotask/internal/platform/response"
)

// Handler handles HTTP requests for activities.
type Handler struct {
	service *Service
	logger  *slog.Logger
}

// NewHandler creates a new activity Handler.
func NewHandler(service *Service, logger *slog.Logger) *Handler {
	return &Handler{service: service, logger: logger}
}

// RegisterRoutes registers activity routes on the given router.
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/tasks/{taskId}/activities", h.ListByTask)
	r.Get("/activities", h.ListByUser)
}

// ListByTask handles GET /tasks/{taskId}/activities.
func (h *Handler) ListByTask(w http.ResponseWriter, r *http.Request) {
	taskID, err := uuid.Parse(chi.URLParam(r, "taskId"))
	if err != nil {
		response.BadRequest(w, "ID task tidak valid")
		return
	}

	activities, err := h.service.ListByTask(r.Context(), taskID)
	if err != nil {
		h.logger.Error("list task activities failed", slog.String("error", err.Error()))
		response.InternalError(w)
		return
	}

	if activities == nil {
		activities = []ActivityResponse{}
	}

	response.OK(w, "Data berhasil diambil", activities)
}

// ListByUser handles GET /activities.
func (h *Handler) ListByUser(w http.ResponseWriter, r *http.Request) {
	uid := middleware.GetUserID(r.Context())
	userID, err := uuid.Parse(uid)
	if err != nil {
		response.Unauthorized(w, "User tidak terautentikasi")
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	activities, total, err := h.service.ListByUser(r.Context(), userID, page, limit)
	if err != nil {
		h.logger.Error("list user activities failed", slog.String("error", err.Error()))
		response.InternalError(w)
		return
	}

	if activities == nil {
		activities = []ActivityResponse{}
	}

	totalPages := (total + limit - 1) / limit
	meta := response.Meta{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	}

	response.SuccessWithMeta(w, http.StatusOK, "Data berhasil diambil", activities, meta)
}
