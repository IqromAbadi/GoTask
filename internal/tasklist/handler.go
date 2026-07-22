package tasklist

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/iqromabadi/gotask/internal/middleware"
	"github.com/iqromabadi/gotask/internal/platform/response"
	"github.com/iqromabadi/gotask/internal/platform/validator"
	"github.com/iqromabadi/gotask/internal/task"
)

// Handler handles HTTP requests for task lists.
type Handler struct {
	service     *Service
	taskService *task.Service
	logger      *slog.Logger
}

// NewHandler creates a new task list Handler.
func NewHandler(service *Service, taskService *task.Service, logger *slog.Logger) *Handler {
	return &Handler{service: service, taskService: taskService, logger: logger}
}

// RegisterRoutes registers task list routes on the given router.
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Post("/lists", h.Create)
	r.Get("/lists", h.List)
	r.Get("/lists/{listId}", h.GetByID)
	r.Patch("/lists/{listId}", h.Update)
	r.Delete("/lists/{listId}", h.Delete)
	r.Patch("/lists/{listId}/archive", h.Archive)
	r.Patch("/lists/{listId}/restore", h.Restore)
	r.Get("/lists/{listId}/board", h.GetBoard)
}

// Create handles POST /lists.
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r)
	if err != nil {
		response.Unauthorized(w, "User tidak terautentikasi")
		return
	}

	var req CreateTaskListRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Format request tidak valid")
		return
	}

	if errs := validator.ValidateStruct(req); errs != nil {
		response.ValidationError(w, "Data tidak valid", errs)
		return
	}

	tl, err := h.service.Create(r.Context(), userID, req)
	if err != nil {
		h.logger.Error("create task list failed", slog.String("error", err.Error()))
		response.InternalError(w)
		return
	}

	response.Created(w, "Task list berhasil dibuat", tl)
}

// List handles GET /lists.
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r)
	if err != nil {
		response.Unauthorized(w, "User tidak terautentikasi")
		return
	}

	lists, err := h.service.List(r.Context(), userID)
	if err != nil {
		h.logger.Error("list task lists failed", slog.String("error", err.Error()))
		response.InternalError(w)
		return
	}

	response.OK(w, "Data berhasil diambil", lists)
}

// GetByID handles GET /lists/{listId}.
func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	userID, listID, err := getIDs(r)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	tl, err := h.service.GetByID(r.Context(), listID, userID)
	if err != nil {
		if strings.Contains(err.Error(), "tidak ditemukan") {
			response.NotFound(w, err.Error())
			return
		}
		h.logger.Error("get task list failed", slog.String("error", err.Error()))
		response.InternalError(w)
		return
	}

	response.OK(w, "Data berhasil diambil", tl)
}

// Update handles PATCH /lists/{listId}.
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	userID, listID, err := getIDs(r)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	var req UpdateTaskListRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Format request tidak valid")
		return
	}

	if errs := validator.ValidateStruct(req); errs != nil {
		response.ValidationError(w, "Data tidak valid", errs)
		return
	}

	tl, err := h.service.Update(r.Context(), listID, userID, req)
	if err != nil {
		h.logger.Error("update task list failed", slog.String("error", err.Error()))
		response.InternalError(w)
		return
	}

	response.OK(w, "Task list berhasil diperbarui", tl)
}

// Delete handles DELETE /lists/{listId}.
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, listID, err := getIDs(r)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	if err := h.service.Delete(r.Context(), listID, userID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			response.NotFound(w, "Task list tidak ditemukan")
			return
		}
		h.logger.Error("delete task list failed", slog.String("error", err.Error()))
		response.InternalError(w)
		return
	}

	response.OK(w, "Task list berhasil dihapus", nil)
}

// Archive handles PATCH /lists/{listId}/archive.
func (h *Handler) Archive(w http.ResponseWriter, r *http.Request) {
	userID, listID, err := getIDs(r)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	tl, err := h.service.Archive(r.Context(), listID, userID)
	if err != nil {
		if strings.Contains(err.Error(), "tidak ditemukan") {
			response.NotFound(w, err.Error())
			return
		}
		h.logger.Error("archive task list failed", slog.String("error", err.Error()))
		response.InternalError(w)
		return
	}

	response.OK(w, "Task list berhasil diarsipkan", tl)
}

// Restore handles PATCH /lists/{listId}/restore.
func (h *Handler) Restore(w http.ResponseWriter, r *http.Request) {
	userID, listID, err := getIDs(r)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	tl, err := h.service.Restore(r.Context(), listID, userID)
	if err != nil {
		if strings.Contains(err.Error(), "tidak ditemukan") {
			response.NotFound(w, err.Error())
			return
		}
		h.logger.Error("restore task list failed", slog.String("error", err.Error()))
		response.InternalError(w)
		return
	}

	response.OK(w, "Task list berhasil dipulihkan", tl)
}

// GetBoard handles GET /lists/{listId}/board.
func (h *Handler) GetBoard(w http.ResponseWriter, r *http.Request) {
	userID, listID, err := getIDs(r)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	board, err := h.taskService.GetBoard(r.Context(), listID, userID)
	if err != nil {
		h.logger.Error("get board failed", slog.String("error", err.Error()))
		response.InternalError(w)
		return
	}

	response.OK(w, "Board berhasil diambil", board)
}

func getUserID(r *http.Request) (uuid.UUID, error) {
	idStr := middleware.GetUserID(r.Context())
	return uuid.Parse(idStr)
}

func getIDs(r *http.Request) (userID, listID uuid.UUID, err error) {
	userID, err = getUserID(r)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	listID, err = uuid.Parse(chi.URLParam(r, "listId"))
	return
}
