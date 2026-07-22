package auth

import (
	"github.com/iqromabadi/gotask/internal/platform/util"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/iqromabadi/gotask/internal/middleware"
	"github.com/iqromabadi/gotask/internal/platform/response"
	"github.com/iqromabadi/gotask/internal/platform/validator"
)

// Handler handles HTTP requests for authentication.
type Handler struct {
	service *Service
	logger  *slog.Logger
}

// NewHandler creates a new auth Handler.
func NewHandler(service *Service, logger *slog.Logger) *Handler {
	return &Handler{service: service, logger: logger}
}

// RegisterPublicRoutes registers public auth routes (no auth required).
func (h *Handler) RegisterPublicRoutes(r chi.Router) {
	r.Post("/auth/register", h.Register)
	r.Post("/auth/login", h.Login)
	r.Post("/auth/refresh", h.Refresh)
	r.Post("/auth/logout", h.Logout)
}

// RegisterProtectedRoutes registers protected user routes (auth required).
func (h *Handler) RegisterProtectedRoutes(r chi.Router) {
	r.Get("/users/me", h.GetProfile)
	r.Patch("/users/me", h.UpdateProfile)
	r.Patch("/users/me/password", h.ChangePassword)
}

// Register handles user registration.
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := util.DecodeBody(r, &req); err != nil {
		response.BadRequest(w, "Format request tidak valid")
		return
	}

	if errs := validator.ValidateStruct(req); errs != nil {
		// Additional password strength validation
		if req.Password != "" {
			if err := ValidatePasswordStrength(req.Password); err != nil {
				errs["password"] = "Password harus memiliki minimal 8 karakter, huruf besar, huruf kecil, dan angka"
			}
		}
		response.ValidationError(w, "Data tidak valid", errs)
		return
	}

	user, err := h.service.Register(r.Context(), req)
	if err != nil {
		if strings.Contains(err.Error(), "sudah terdaftar") {
			response.Conflict(w, err.Error())
			return
		}
		h.logger.Error("register failed", slog.String("error", err.Error()))
		response.InternalError(w)
		return
	}

	response.Created(w, "Registrasi berhasil", user)
}

// Login handles user login.
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := util.DecodeBody(r, &req); err != nil {
		response.BadRequest(w, "Format request tidak valid")
		return
	}

	if errs := validator.ValidateStruct(req); errs != nil {
		response.ValidationError(w, "Data tidak valid", errs)
		return
	}

	result, err := h.service.Login(r.Context(), req)
	if err != nil {
		if strings.Contains(err.Error(), "email atau password") {
			response.Unauthorized(w, "Email atau password salah")
			return
		}
		h.logger.Error("login failed", slog.String("error", err.Error()))
		response.InternalError(w)
		return
	}

	response.OK(w, "Login berhasil", result)
}

// Refresh handles refresh token rotation.
func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest
	if err := util.DecodeBody(r, &req); err != nil {
		response.BadRequest(w, "Format request tidak valid")
		return
	}

	if req.RefreshToken == "" {
		response.BadRequest(w, "Refresh token wajib diisi")
		return
	}

	result, err := h.service.RefreshAccessToken(r.Context(), req.RefreshToken)
	if err != nil {
		msg := err.Error()
		if strings.Contains(msg, "tidak valid") || strings.Contains(msg, "kadaluarsa") || strings.Contains(msg, "dicabut") {
			response.Unauthorized(w, msg)
			return
		}
		h.logger.Error("refresh failed", slog.String("error", msg))
		response.InternalError(w)
		return
	}

	response.OK(w, "Token berhasil diperbarui", result)
}

// Logout handles user logout.
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest
	if err := util.DecodeBody(r, &req); err != nil {
		response.BadRequest(w, "Format request tidak valid")
		return
	}

	if req.RefreshToken == "" {
		response.BadRequest(w, "Refresh token wajib diisi")
		return
	}

	if err := h.service.Logout(r.Context(), req.RefreshToken); err != nil {
		h.logger.Error("logout failed", slog.String("error", err.Error()))
		response.InternalError(w)
		return
	}

	response.OK(w, "Logout berhasil", nil)
}

// GetProfile returns the authenticated user's profile.
func (h *Handler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromContext(r)
	if err != nil {
		response.Unauthorized(w, "User tidak terautentikasi")
		return
	}

	user, err := h.service.GetProfile(r.Context(), userID)
	if err != nil {
		h.logger.Error("get profile failed", slog.String("error", err.Error()))
		response.InternalError(w)
		return
	}

	response.OK(w, "Profile berhasil diambil", user)
}

// UpdateProfile updates the authenticated user's profile.
func (h *Handler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromContext(r)
	if err != nil {
		response.Unauthorized(w, "User tidak terautentikasi")
		return
	}

	var req UpdateProfileRequest
	if err := util.DecodeBody(r, &req); err != nil {
		response.BadRequest(w, "Format request tidak valid")
		return
	}

	if errs := validator.ValidateStruct(req); errs != nil {
		response.ValidationError(w, "Data tidak valid", errs)
		return
	}

	user, err := h.service.UpdateProfile(r.Context(), userID, req)
	if err != nil {
		h.logger.Error("update profile failed", slog.String("error", err.Error()))
		response.InternalError(w)
		return
	}

	response.OK(w, "Profile berhasil diperbarui", user)
}

// ChangePassword changes the authenticated user's password.
func (h *Handler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromContext(r)
	if err != nil {
		response.Unauthorized(w, "User tidak terautentikasi")
		return
	}

	var req ChangePasswordRequest
	if err := util.DecodeBody(r, &req); err != nil {
		response.BadRequest(w, "Format request tidak valid")
		return
	}

	if errs := validator.ValidateStruct(req); errs != nil {
		response.ValidationError(w, "Data tidak valid", errs)
		return
	}

	if err := h.service.ChangePassword(r.Context(), userID, req); err != nil {
		if strings.Contains(err.Error(), "password saat ini") {
			response.BadRequest(w, err.Error())
			return
		}
		if strings.Contains(err.Error(), "password baru") {
			response.BadRequest(w, err.Error())
			return
		}
		h.logger.Error("change password failed", slog.String("error", err.Error()))
		response.InternalError(w)
		return
	}

	response.OK(w, "Password berhasil diubah", nil)
}

// getUserIDFromContext extracts the user ID from the request context.
func getUserIDFromContext(r *http.Request) (uuid.UUID, error) {
	idStr := middleware.GetUserID(r.Context())
	if idStr == "" {
		return uuid.Nil, errors.New("user not authenticated")
	}
	return uuid.Parse(idStr)
}
