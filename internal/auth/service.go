package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Service handles authentication business logic.
type Service struct {
	repo         Repository
	tokenService *TokenService
	refreshTTL   time.Duration
}

// NewService creates a new auth Service.
func NewService(repo Repository, tokenService *TokenService, refreshTTL time.Duration) *Service {
	return &Service{
		repo:         repo,
		tokenService: tokenService,
		refreshTTL:   refreshTTL,
	}
}

// Register creates a new user account.
func (s *Service) Register(ctx context.Context, req RegisterRequest) (*UserResponse, error) {
	email := NormalizeEmail(req.Email)

	// Check password strength
	if err := ValidatePasswordStrength(req.Password); err != nil {
		return nil, fmt.Errorf("password validation: %w", err)
	}

	// Check if email is already taken
	existing, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("check existing user: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("email sudah terdaftar")
	}

	// Hash password
	passwordHash, err := HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	// Create user
	user := &User{
		Name:         req.Name,
		Email:        email,
		PasswordHash: passwordHash,
	}

	if err := s.repo.CreateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	resp := ToUserResponse(user)
	return &resp, nil
}

// Login authenticates a user and returns tokens.
func (s *Service) Login(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	email := NormalizeEmail(req.Email)

	// Find user
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	if user == nil {
		// Generic error to prevent email enumeration
		return nil, fmt.Errorf("email atau password salah")
	}

	// Check password
	if !CheckPassword(req.Password, user.PasswordHash) {
		return nil, fmt.Errorf("email atau password salah")
	}

	// Generate access token
	accessToken, expiresIn, err := s.tokenService.GenerateAccessToken(user.ID, user.Email)
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}

	// Generate refresh token
	refreshToken, tokenHash, err := GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}

	// Store refresh token hash
	now := time.Now().UTC()
	rt := &RefreshToken{
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: now.Add(s.refreshTTL),
	}
	if err := s.repo.CreateRefreshToken(ctx, rt); err != nil {
		return nil, fmt.Errorf("store refresh token: %w", err)
	}

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresIn,
		TokenType:    "Bearer",
	}, nil
}

// RefreshAccessToken rotates the refresh token and returns new tokens.
func (s *Service) RefreshAccessToken(ctx context.Context, refreshTokenStr string) (*LoginResponse, error) {
	// Hash the incoming refresh token
	tokenHash := hashToken(refreshTokenStr)

	// Find stored token
	stored, err := s.repo.GetRefreshTokenByHash(ctx, tokenHash)
	if err != nil {
		return nil, fmt.Errorf("get refresh token: %w", err)
	}
	if stored == nil {
		return nil, fmt.Errorf("refresh token tidak valid")
	}

	// Check if revoked
	if stored.RevokedAt != nil {
		// Revoke all tokens for this user (token reuse indicates possible theft)
		uid, _ := uuid.Parse(stored.UserID)
		_ = s.repo.RevokeAllUserRefreshTokens(ctx, uid)
		return nil, fmt.Errorf("refresh token telah dicabut")
	}

	// Check if expired
	if time.Now().UTC().After(stored.ExpiresAt) {
		return nil, fmt.Errorf("refresh token telah kadaluarsa")
	}

	// Revoke the old refresh token (rotation)
	oldID, _ := uuid.Parse(stored.ID)
	if err := s.repo.RevokeRefreshToken(ctx, oldID); err != nil {
		return nil, fmt.Errorf("revoke old refresh token: %w", err)
	}

	// Get user for new access token
	userID, _ := uuid.Parse(stored.UserID)
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user tidak ditemukan")
	}

	// Generate new access token
	accessToken, expiresIn, err := s.tokenService.GenerateAccessToken(user.ID, user.Email)
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}

	// Generate new refresh token
	newRefreshToken, newTokenHash, err := GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}

	// Store new refresh token
	now := time.Now().UTC()
	rt := &RefreshToken{
		UserID:    user.ID,
		TokenHash: newTokenHash,
		ExpiresAt: now.Add(s.refreshTTL),
	}
	if err := s.repo.CreateRefreshToken(ctx, rt); err != nil {
		return nil, fmt.Errorf("store new refresh token: %w", err)
	}

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    expiresIn,
		TokenType:    "Bearer",
	}, nil
}

// Logout revokes a specific refresh token.
func (s *Service) Logout(ctx context.Context, refreshTokenStr string) error {
	tokenHash := hashToken(refreshTokenStr)

	stored, err := s.repo.GetRefreshTokenByHash(ctx, tokenHash)
	if err != nil {
		return fmt.Errorf("get refresh token: %w", err)
	}
	if stored == nil {
		// Token not found, consider it already logged out
		return nil
	}

	id, _ := uuid.Parse(stored.ID)
	if err := s.repo.RevokeRefreshToken(ctx, id); err != nil {
		return fmt.Errorf("revoke refresh token: %w", err)
	}

	return nil
}

// GetProfile returns the user profile for the given user ID.
func (s *Service) GetProfile(ctx context.Context, userID uuid.UUID) (*UserResponse, error) {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user tidak ditemukan")
	}

	resp := ToUserResponse(user)
	return &resp, nil
}

// UpdateProfile updates a user's profile.
func (s *Service) UpdateProfile(ctx context.Context, userID uuid.UUID, req UpdateProfileRequest) (*UserResponse, error) {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user tidak ditemukan")
	}

	user.Name = req.Name
	user.AvatarURL = req.AvatarURL

	if err := s.repo.UpdateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("update user: %w", err)
	}

	resp := ToUserResponse(user)
	return &resp, nil
}

// ChangePassword changes a user's password after verifying the current password.
func (s *Service) ChangePassword(ctx context.Context, userID uuid.UUID, req ChangePasswordRequest) error {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("get user: %w", err)
	}
	if user == nil {
		return fmt.Errorf("user tidak ditemukan")
	}

	// Verify current password
	if !CheckPassword(req.CurrentPassword, user.PasswordHash) {
		return fmt.Errorf("password saat ini salah")
	}

	// Validate new password strength
	if err := ValidatePasswordStrength(req.NewPassword); err != nil {
		return fmt.Errorf("password baru: %w", err)
	}

	// Hash new password
	newHash, err := HashPassword(req.NewPassword)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	if err := s.repo.UpdateUserPassword(ctx, userID, newHash); err != nil {
		return fmt.Errorf("update password: %w", err)
	}

	return nil
}
