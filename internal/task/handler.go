package task

import (
	"github.com/iqromabadi/gotask/internal/platform/util"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/iqromabadi/gotask/internal/middleware"
	"github.com/iqromabadi/gotask/internal/platform/response"
	"github.com/iqromabadi/gotask/internal/platform/validator"
)

// Handler handles HTTP requests for tasks.
type Handler struct {
	service *Service
	logger  *slog.Logger
}

// NewHandler creates a new task Handler.
func NewHandler(service *Service, logger *slog.Logger) *Handler {
	return &Handler{service: service, logger: logger}
}

// RegisterRoutes registers task routes on the given router.
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Post("/lists/{listId}/tasks", h.Create)
	r.Get("/lists/{listId}/tasks", h.List)

	r.Get("/tasks/{taskId}", h.GetByID)
	r.Patch("/tasks/{taskId}", h.Update)
	r.Delete("/tasks/{taskId}", h.Delete)

	r.Patch("/tasks/{taskId}/status", h.UpdateStatus)
	r.Patch("/tasks/{taskId}/priority", h.UpdatePriority)
	r.Post("/tasks/{taskId}/reopen", h.Reopen)
}

// Create handles POST /lists/{listId}/tasks.
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID, listID, err := getListUserIDs(r)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	var req CreateTaskRequest
	if err := util.DecodeBody(r, &req); err != nil {
		response.BadRequest(w, "Format request tidak valid")
		return
	}

	if errs := validator.ValidateStruct(req); errs != nil {
		response.ValidationError(w, "Data tidak valid", errs)
		return
	}

	task, err := h.service.Create(r.Context(), listID, userID, req)
	if err != nil {
		if strings.Contains(err.Error(), "tidak valid") {
			response.BadRequest(w, err.Error())
			return
		}
		h.logger.Error("create task failed", slog.String("error", err.Error()))
		response.InternalError(w)
		return
	}

	response.Created(w, "Task berhasil dibuat", task)
}

// List handles GET /lists/{listId}/tasks.
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID, listID, err := getListUserIDs(r)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	filter := parseTaskFilter(r)

	tasks, total, err := h.service.List(r.Context(), listID, userID, filter)
	if err != nil {
		h.logger.Error("list tasks failed", slog.String("error", err.Error()))
		response.InternalError(w)
		return
	}

	totalPages := (total + filter.Limit - 1) / filter.Limit
	meta := response.Meta{
		Page:       filter.Page,
		Limit:      filter.Limit,
		Total:      total,
		TotalPages: totalPages,
	}

	response.SuccessWithMeta(w, http.StatusOK, "Data berhasil diambil", tasks, meta)
}

// GetByID handles GET /tasks/{taskId}.
func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	userID, taskID, err := getTaskUserID(r)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	task, err := h.service.GetByID(r.Context(), taskID, userID)
	if err != nil {
		if strings.Contains(err.Error(), "tidak ditemukan") {
			response.NotFound(w, err.Error())
			return
		}
		h.logger.Error("get task failed", slog.String("error", err.Error()))
		response.InternalError(w)
		return
	}

	response.OK(w, "Data berhasil diambil", task)
}

// Update handles PATCH /tasks/{taskId}.
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	userID, taskID, err := getTaskUserID(r)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	var req UpdateTaskRequest
	if err := util.DecodeBody(r, &req); err != nil {
		response.BadRequest(w, "Format request tidak valid")
		return
	}

	if errs := validator.ValidateStruct(req); errs != nil {
		response.ValidationError(w, "Data tidak valid", errs)
		return
	}

	// Get task to know its list_id
	current, err := h.service.GetByID(r.Context(), taskID, userID)
	if err != nil {
		response.NotFound(w, "Task tidak ditemukan")
		return
	}

	listID, _ := uuid.Parse(current.ListID)
	task, err := h.service.Update(r.Context(), taskID, listID, req)
	if err != nil {
		h.logger.Error("update task failed", slog.String("error", err.Error()))
		response.InternalError(w)
		return
	}

	response.OK(w, "Task berhasil diperbarui", task)
}

// Delete handles DELETE /tasks/{taskId}.
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, taskID, err := getTaskUserID(r)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	// Get task to know its list_id
	current, err := h.service.GetByID(r.Context(), taskID, userID)
	if err != nil {
		response.NotFound(w, "Task tidak ditemukan")
		return
	}

	listID, _ := uuid.Parse(current.ListID)
	if err := h.service.Delete(r.Context(), taskID, listID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			response.NotFound(w, "Task tidak ditemukan")
			return
		}
		h.logger.Error("delete task failed", slog.String("error", err.Error()))
		response.InternalError(w)
		return
	}

	response.OK(w, "Task berhasil dihapus", nil)
}

// UpdateStatus handles PATCH /tasks/{taskId}/status.
func (h *Handler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	userID, taskID, err := getTaskUserID(r)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	var req UpdateStatusRequest
	if err := util.DecodeBody(r, &req); err != nil {
		response.BadRequest(w, "Format request tidak valid")
		return
	}

	if !IsValidStatus(req.Status) {
		response.BadRequest(w, "Status tidak valid")
		return
	}

	// Get task to know its list_id
	current, err := h.service.GetByID(r.Context(), taskID, userID)
	if err != nil {
		response.NotFound(w, "Task tidak ditemukan")
		return
	}

	listID, _ := uuid.Parse(current.ListID)
	task, err := h.service.UpdateStatus(r.Context(), taskID, listID, req.Status, userID)
	if err != nil {
		if strings.Contains(err.Error(), "tidak diizinkan") || strings.Contains(err.Error(), "progress harus") || strings.Contains(err.Error(), "hanya dapat") {
			response.BadRequest(w, err.Error())
			return
		}
		h.logger.Error("update status failed", slog.String("error", err.Error()))
		response.InternalError(w)
		return
	}

	response.OK(w, "Status task berhasil diperbarui", task)
}

// UpdatePriority handles PATCH /tasks/{taskId}/priority.
func (h *Handler) UpdatePriority(w http.ResponseWriter, r *http.Request) {
	userID, taskID, err := getTaskUserID(r)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	var req UpdatePriorityRequest
	if err := util.DecodeBody(r, &req); err != nil {
		response.BadRequest(w, "Format request tidak valid")
		return
	}

	if !IsValidPriority(req.Priority) {
		response.BadRequest(w, "Prioritas tidak valid")
		return
	}

	current, err := h.service.GetByID(r.Context(), taskID, userID)
	if err != nil {
		response.NotFound(w, "Task tidak ditemukan")
		return
	}

	listID, _ := uuid.Parse(current.ListID)
	task, err := h.service.UpdatePriority(r.Context(), taskID, listID, req.Priority, userID)
	if err != nil {
		h.logger.Error("update priority failed", slog.String("error", err.Error()))
		response.InternalError(w)
		return
	}

	response.OK(w, "Prioritas task berhasil diperbarui", task)
}

// Reopen handles POST /tasks/{taskId}/reopen.
func (h *Handler) Reopen(w http.ResponseWriter, r *http.Request) {
	userID, taskID, err := getTaskUserID(r)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	current, err := h.service.GetByID(r.Context(), taskID, userID)
	if err != nil {
		response.NotFound(w, "Task tidak ditemukan")
		return
	}

	listID, _ := uuid.Parse(current.ListID)
	task, err := h.service.Reopen(r.Context(), taskID, listID, userID)
	if err != nil {
		if strings.Contains(err.Error(), "hanya task") {
			response.BadRequest(w, err.Error())
			return
		}
		h.logger.Error("reopen task failed", slog.String("error", err.Error()))
		response.InternalError(w)
		return
	}

	response.OK(w, "Task berhasil dibuka kembali", task)
}

// Helper functions

func getListUserIDs(r *http.Request) (userID, listID uuid.UUID, err error) {
	uid := middleware.GetUserID(r.Context())
	if uid == "" {
		return uuid.Nil, uuid.Nil, fmt.Errorf("user not authenticated")
	}
	userID, err = uuid.Parse(uid)
	if err != nil {
		return
	}
	listID, err = uuid.Parse(chi.URLParam(r, "listId"))
	return
}

func getTaskUserID(r *http.Request) (userID, taskID uuid.UUID, err error) {
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

func parseTaskFilter(r *http.Request) TaskFilter {
	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	limit, _ := strconv.Atoi(q.Get("limit"))

	filter := TaskFilter{
		Status:    q.Get("status"),
		Priority:  q.Get("priority"),
		Search:    q.Get("search"),
		SortBy:    q.Get("sort_by"),
		SortOrder: q.Get("sort_order"),
		Page:      page,
		Limit:     limit,
	}

	if df := q.Get("due_date_from"); df != "" {
		t, err := time.Parse("2006-01-02", df)
		if err == nil {
			filter.DueDateFrom = &t
		}
	}
	if dt := q.Get("due_date_to"); dt != "" {
		t, err := time.Parse("2006-01-02", dt)
		if err == nil {
			filter.DueDateTo = &t
		}
	}
	filter.IsOverdue = q.Get("is_overdue") == "true"

	return filter
}
