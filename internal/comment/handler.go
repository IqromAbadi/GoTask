package comment

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

// Handler handles HTTP requests for comments.
type Handler struct {
	service *Service
	logger  *slog.Logger
}

// NewHandler creates a new comment Handler.
func NewHandler(service *Service, logger *slog.Logger) *Handler {
	return &Handler{service: service, logger: logger}
}

// RegisterRoutes registers comment routes on the given router.
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Post("/tasks/{taskId}/comments", h.Create)
	r.Get("/tasks/{taskId}/comments", h.List)
	r.Patch("/tasks/{taskId}/comments/{commentId}", h.Update)
	r.Delete("/tasks/{taskId}/comments/{commentId}", h.Delete)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID, taskID, err := getCommentIDs(r)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	var req CreateCommentRequest
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
		h.logger.Error("create comment failed", slog.String("error", err.Error()))
		response.InternalError(w)
		return
	}
	response.Created(w, "Komentar berhasil ditambahkan", result)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	taskID, err := uuid.Parse(chi.URLParam(r, "taskId"))
	if err != nil {
		response.BadRequest(w, "ID task tidak valid")
		return
	}

	comments, err := h.service.List(r.Context(), taskID)
	if err != nil {
		h.logger.Error("list comments failed", slog.String("error", err.Error()))
		response.InternalError(w)
		return
	}
	if comments == nil {
		comments = []CommentResponse{}
	}
	response.OK(w, "Data berhasil diambil", comments)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	userID, taskID, commentID, err := getFullCommentIDs(r)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	var req UpdateCommentRequest
	if err := util.DecodeBody(r, &req); err != nil {
		response.BadRequest(w, "Format request tidak valid")
		return
	}
	if errs := validator.ValidateStruct(req); errs != nil {
		response.ValidationError(w, "Data tidak valid", errs)
		return
	}

	result, err := h.service.Update(r.Context(), commentID, taskID, userID, req)
	if err != nil {
		if strings.Contains(err.Error(), "hanya dapat mengedit") {
			response.Forbidden(w, err.Error())
			return
		}
		if strings.Contains(err.Error(), "tidak ditemukan") {
			response.NotFound(w, err.Error())
			return
		}
		h.logger.Error("update comment failed", slog.String("error", err.Error()))
		response.InternalError(w)
		return
	}
	response.OK(w, "Komentar berhasil diperbarui", result)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, taskID, commentID, err := getFullCommentIDs(r)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	if err := h.service.Delete(r.Context(), commentID, taskID, userID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			response.NotFound(w, "Komentar tidak ditemukan")
			return
		}
		h.logger.Error("delete comment failed", slog.String("error", err.Error()))
		response.InternalError(w)
		return
	}
	response.OK(w, "Komentar berhasil dihapus", nil)
}

func getCommentIDs(r *http.Request) (userID, taskID uuid.UUID, err error) {
	uid := middleware.GetUserID(r.Context())
	if uid == "" {
		return uuid.Nil, uuid.Nil, fmt.Errorf("user not authenticated")
	}
	userID, _ = uuid.Parse(uid)
	taskID, err = uuid.Parse(chi.URLParam(r, "taskId"))
	return
}

func getFullCommentIDs(r *http.Request) (userID, taskID, commentID uuid.UUID, err error) {
	userID, taskID, err = getCommentIDs(r)
	if err != nil {
		return
	}
	commentID, err = uuid.Parse(chi.URLParam(r, "commentId"))
	return
}
