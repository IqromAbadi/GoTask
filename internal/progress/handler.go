package progress

import (
	"github.com/iqromabadi/gotask/internal/platform/util"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/iqromabadi/gotask/internal/middleware"
	"github.com/iqromabadi/gotask/internal/platform/response"
	"github.com/iqromabadi/gotask/internal/platform/validator"
)

// Handler handles HTTP requests for progress updates.
type Handler struct {
	service *Service
	logger  *slog.Logger
}

// NewHandler creates a new progress Handler.
func NewHandler(service *Service, logger *slog.Logger) *Handler {
	return &Handler{service: service, logger: logger}
}

// RegisterRoutes registers progress routes on the given router.
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Post("/tasks/{taskId}/progress", h.Create)
	r.Get("/tasks/{taskId}/progress", h.ListByTask)
	r.Get("/tasks/{taskId}/progress/{progressId}", h.GetByID)
	r.Patch("/tasks/{taskId}/progress/{progressId}", h.UpdateNote)
	r.Delete("/tasks/{taskId}/progress/{progressId}", h.Delete)
}

// Create handles POST /tasks/{taskId}/progress.
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID, taskID, err := getIDs(r)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	var req CreateProgressRequest
	if err := util.DecodeBody(r, &req); err != nil {
		response.BadRequest(w, "Format request tidak valid")
		return
	}

	if errs := validator.ValidateStruct(req); errs != nil {
		response.ValidationError(w, "Data tidak valid", errs)
		return
	}

	result, err := h.service.Create(r.Context(), taskID, userID, req)
	if err != nil {
		if strings.Contains(err.Error(), "progress harus") ||
			strings.Contains(err.Error(), "note wajib") ||
			strings.Contains(err.Error(), "tidak boleh lebih kecil") ||
			strings.Contains(err.Error(), "hanya dapat ditambahkan") {
			response.BadRequest(w, err.Error())
			return
		}
		if strings.Contains(err.Error(), "tidak ditemukan") {
			response.NotFound(w, err.Error())
			return
		}
		h.logger.Error("create progress failed", slog.String("error", err.Error()))
		response.InternalError(w)
		return
	}

	response.Created(w, "Progress berhasil ditambahkan", result)
}

// ListByTask handles GET /tasks/{taskId}/progress.
func (h *Handler) ListByTask(w http.ResponseWriter, r *http.Request) {
	taskID, err := uuid.Parse(chi.URLParam(r, "taskId"))
	if err != nil {
		response.BadRequest(w, "ID task tidak valid")
		return
	}

	updates, err := h.service.ListByTask(r.Context(), taskID)
	if err != nil {
		h.logger.Error("list progress failed", slog.String("error", err.Error()))
		response.InternalError(w)
		return
	}

	if updates == nil {
		updates = []ProgressResponse{}
	}

	response.OK(w, "Data berhasil diambil", updates)
}

// GetByID handles GET /tasks/{taskId}/progress/{progressId}.
func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	taskID, progressID, err := getProgressIDs(r)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	update, err := h.service.GetByID(r.Context(), progressID, taskID)
	if err != nil {
		if strings.Contains(err.Error(), "tidak ditemukan") {
			response.NotFound(w, err.Error())
			return
		}
		h.logger.Error("get progress failed", slog.String("error", err.Error()))
		response.InternalError(w)
		return
	}

	response.OK(w, "Data berhasil diambil", update)
}

// UpdateNote handles PATCH /tasks/{taskId}/progress/{progressId}.
func (h *Handler) UpdateNote(w http.ResponseWriter, r *http.Request) {
	taskID, progressID, err := getProgressIDs(r)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	var req UpdateProgressNoteRequest
	if err := util.DecodeBody(r, &req); err != nil {
		response.BadRequest(w, "Format request tidak valid")
		return
	}

	if req.Note == "" {
		response.BadRequest(w, "Note wajib diisi")
		return
	}

	update, err := h.service.UpdateNote(r.Context(), progressID, taskID, req.Note)
	if err != nil {
		if strings.Contains(err.Error(), "tidak ditemukan") {
			response.NotFound(w, err.Error())
			return
		}
		h.logger.Error("update progress note failed", slog.String("error", err.Error()))
		response.InternalError(w)
		return
	}

	response.OK(w, "Note berhasil diperbarui", update)
}

// Delete handles DELETE /tasks/{taskId}/progress/{progressId}.
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	taskID, progressID, err := getProgressIDs(r)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	if err := h.service.Delete(r.Context(), progressID, taskID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			response.NotFound(w, "Progress tidak ditemukan")
			return
		}
		h.logger.Error("delete progress failed", slog.String("error", err.Error()))
		response.InternalError(w)
		return
	}

	response.OK(w, "Progress berhasil dihapus", nil)
}

func getIDs(r *http.Request) (userID, taskID uuid.UUID, err error) {
	uid := middleware.GetUserID(r.Context())
	if uid == "" {
		return uuid.Nil, uuid.Nil, fmt.Errorf("user not authenticated")
	}
	userID, err = uuid.Parse(uid)
	if err != nil {
		return
	}
	taskID, err = uuid.Parse(chi.URLParam(r, "taskId"))
	return
}

func getProgressIDs(r *http.Request) (taskID, progressID uuid.UUID, err error) {
	taskID, err = uuid.Parse(chi.URLParam(r, "taskId"))
	if err != nil {
		return
	}
	progressID, err = uuid.Parse(chi.URLParam(r, "progressId"))
	return
}
