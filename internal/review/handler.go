package review

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/iqromabadi/gotask/internal/middleware"
	"github.com/iqromabadi/gotask/internal/platform/response"
)

// Handler handles HTTP requests for reviews.
type Handler struct {
	service *Service
	logger  *slog.Logger
}

// NewHandler creates a new review Handler.
func NewHandler(service *Service, logger *slog.Logger) *Handler {
	return &Handler{service: service, logger: logger}
}

// RegisterRoutes registers review routes on the given router.
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Post("/tasks/{taskId}/submit-review", h.SubmitReview)
	r.Get("/tasks/{taskId}/reviews", h.List)
	r.Get("/tasks/{taskId}/reviews/{reviewId}", h.GetByID)
	r.Post("/tasks/{taskId}/reviews/{reviewId}/approve", h.Approve)
	r.Post("/tasks/{taskId}/reviews/{reviewId}/request-changes", h.RequestChanges)
}

// SubmitReview handles POST /tasks/{taskId}/submit-review.
func (h *Handler) SubmitReview(w http.ResponseWriter, r *http.Request) {
	userID, taskID, err := getReviewIDs(r)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	var req SubmitReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Format request tidak valid")
		return
	}

	result, err := h.service.SubmitReview(r.Context(), taskID, userID, req.SubmissionNote)
	if err != nil {
		if strings.Contains(err.Error(), "harus berstatus") ||
			strings.Contains(err.Error(), "progress harus") {
			response.BadRequest(w, err.Error())
			return
		}
		if strings.Contains(err.Error(), "tidak ditemukan") {
			response.NotFound(w, err.Error())
			return
		}
		h.logger.Error("submit review failed", slog.String("error", err.Error()))
		response.InternalError(w)
		return
	}

	response.Created(w, "Review berhasil disubmit", result)
}

// List handles GET /tasks/{taskId}/reviews.
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	taskID, err := uuid.Parse(chi.URLParam(r, "taskId"))
	if err != nil {
		response.BadRequest(w, "ID task tidak valid")
		return
	}

	reviews, err := h.service.List(r.Context(), taskID)
	if err != nil {
		h.logger.Error("list reviews failed", slog.String("error", err.Error()))
		response.InternalError(w)
		return
	}
	if reviews == nil {
		reviews = []ReviewResponse{}
	}

	response.OK(w, "Data berhasil diambil", reviews)
}

// GetByID handles GET /tasks/{taskId}/reviews/{reviewId}.
func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	taskID, reviewID, err := getReviewDetailIDs(r)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	review, err := h.service.GetByID(r.Context(), reviewID, taskID)
	if err != nil {
		if strings.Contains(err.Error(), "tidak ditemukan") {
			response.NotFound(w, err.Error())
			return
		}
		h.logger.Error("get review failed", slog.String("error", err.Error()))
		response.InternalError(w)
		return
	}

	response.OK(w, "Data berhasil diambil", review)
}

// Approve handles POST /tasks/{taskId}/reviews/{reviewId}/approve.
func (h *Handler) Approve(w http.ResponseWriter, r *http.Request) {
	userID, taskID, reviewID, err := getFullIDs(r)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	var req ApproveReviewRequest
	json.NewDecoder(r.Body).Decode(&req)

	result, err := h.service.Approve(r.Context(), reviewID, taskID, userID, req.ReviewNote)
	if err != nil {
		if strings.Contains(err.Error(), "sudah diproses") {
			response.Conflict(w, err.Error())
			return
		}
		if strings.Contains(err.Error(), "tidak ditemukan") {
			response.NotFound(w, err.Error())
			return
		}
		h.logger.Error("approve review failed", slog.String("error", err.Error()))
		response.InternalError(w)
		return
	}

	response.OK(w, "Review berhasil disetujui", result)
}

// RequestChanges handles POST /tasks/{taskId}/reviews/{reviewId}/request-changes.
func (h *Handler) RequestChanges(w http.ResponseWriter, r *http.Request) {
	userID, taskID, reviewID, err := getFullIDs(r)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	var req RequestChangesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Format request tidak valid")
		return
	}
	if req.ReviewNote == "" {
		response.BadRequest(w, "Review note wajib diisi")
		return
	}

	result, err := h.service.RequestChanges(r.Context(), reviewID, taskID, userID, req.ReviewNote)
	if err != nil {
		if strings.Contains(err.Error(), "sudah diproses") {
			response.Conflict(w, err.Error())
			return
		}
		if strings.Contains(err.Error(), "tidak ditemukan") {
			response.NotFound(w, err.Error())
			return
		}
		h.logger.Error("request changes failed", slog.String("error", err.Error()))
		response.InternalError(w)
		return
	}

	response.OK(w, "Perubahan diminta", result)
}

func getReviewIDs(r *http.Request) (userID, taskID uuid.UUID, err error) {
	uid := middleware.GetUserID(r.Context())
	if uid == "" {
		return uuid.Nil, uuid.Nil, fmt.Errorf("user not authenticated")
	}
	userID, _ = uuid.Parse(uid)
	taskID, err = uuid.Parse(chi.URLParam(r, "taskId"))
	return
}

func getReviewDetailIDs(r *http.Request) (taskID, reviewID uuid.UUID, err error) {
	taskID, err = uuid.Parse(chi.URLParam(r, "taskId"))
	if err != nil {
		return
	}
	reviewID, err = uuid.Parse(chi.URLParam(r, "reviewId"))
	return
}

func getFullIDs(r *http.Request) (userID, taskID, reviewID uuid.UUID, err error) {
	uid := middleware.GetUserID(r.Context())
	if uid == "" {
		return uuid.Nil, uuid.Nil, uuid.Nil, fmt.Errorf("user not authenticated")
	}
	userID, _ = uuid.Parse(uid)
	taskID, err = uuid.Parse(chi.URLParam(r, "taskId"))
	if err != nil {
		return
	}
	reviewID, err = uuid.Parse(chi.URLParam(r, "reviewId"))
	return
}
