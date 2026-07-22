package auth

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

// mockRepository implements Repository for testing.
type mockRepository struct {
	users         map[string]*User
	refreshTokens map[string]*RefreshToken
}

func newMockRepository() *mockRepository {
	return &mockRepository{
		users:         make(map[string]*User),
		refreshTokens: make(map[string]*RefreshToken),
	}
}

func (m *mockRepository) CreateUser(ctx context.Context, user *User) error {
	id := uuid.New().String()
	user.ID = id
	user.CreatedAt = time.Now().UTC()
	user.UpdatedAt = time.Now().UTC()
	m.users[id] = user
	return nil
}

func (m *mockRepository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	for _, u := range m.users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, nil
}

func (m *mockRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*User, error) {
	if u, ok := m.users[id.String()]; ok {
		return u, nil
	}
	return nil, nil
}

func (m *mockRepository) UpdateUser(ctx context.Context, user *User) error {
	if _, ok := m.users[user.ID]; ok {
		user.UpdatedAt = time.Now().UTC()
		m.users[user.ID] = user
		return nil
	}
	return nil
}

func (m *mockRepository) UpdateUserPassword(ctx context.Context, userID uuid.UUID, passwordHash string) error {
	if u, ok := m.users[userID.String()]; ok {
		u.PasswordHash = passwordHash
		u.UpdatedAt = time.Now().UTC()
		return nil
	}
	return nil
}

func (m *mockRepository) CreateRefreshToken(ctx context.Context, token *RefreshToken) error {
	id := uuid.New().String()
	token.ID = id
	token.CreatedAt = time.Now().UTC()
	m.refreshTokens[id] = token
	return nil
}

func (m *mockRepository) GetRefreshTokenByHash(ctx context.Context, hash string) (*RefreshToken, error) {
	for _, t := range m.refreshTokens {
		if t.TokenHash == hash {
			return t, nil
		}
	}
	return nil, nil
}

func (m *mockRepository) RevokeRefreshToken(ctx context.Context, id uuid.UUID) error {
	if t, ok := m.refreshTokens[id.String()]; ok {
		now := time.Now().UTC()
		t.RevokedAt = &now
		return nil
	}
	return nil
}

func (m *mockRepository) RevokeAllUserRefreshTokens(ctx context.Context, userID uuid.UUID) error {
	for _, t := range m.refreshTokens {
		if t.UserID == userID.String() {
			now := time.Now().UTC()
			t.RevokedAt = &now
		}
	}
	return nil
}

func newTestService() *Service {
	repo := newMockRepository()
	tokenService := NewTokenService(
		"test-access-secret-key",
		"test-refresh-secret-key",
		15*time.Minute,
		720*time.Hour,
	)
	return NewService(repo, tokenService, 720*time.Hour)
}

func TestRegister_Success(t *testing.T) {
	svc := newTestService()

	req := RegisterRequest{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "Password123!",
	}

	user, err := svc.Register(context.Background(), req)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if user.Name != req.Name {
		t.Errorf("expected name %q, got %q", req.Name, user.Name)
	}
	if user.Email != "test@example.com" {
		t.Errorf("expected email %q, got %q", "test@example.com", user.Email)
	}
	if user.ID == "" {
		t.Error("expected non-empty user ID")
	}
}

func TestRegister_DuplicateEmail(t *testing.T) {
	svc := newTestService()

	req := RegisterRequest{
		Name:     "Test User",
		Email:    "duplicate@example.com",
		Password: "Password123!",
	}

	_, err := svc.Register(context.Background(), req)
	if err != nil {
		t.Fatalf("first register should succeed: %v", err)
	}

	_, err = svc.Register(context.Background(), req)
	if err == nil {
		t.Fatal("expected error for duplicate email")
	}
	if err.Error() != "email sudah terdaftar" {
		t.Errorf("expected 'email sudah terdaftar', got: %v", err)
	}
}

func TestRegister_WeakPassword(t *testing.T) {
	svc := newTestService()

	tests := []struct {
		name     string
		password string
	}{
		{"too short", "Ab1"},
		{"no uppercase", "password123"},
		{"no lowercase", "PASSWORD123"},
		{"no digit", "PasswordOnly"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := RegisterRequest{
				Name:     "Test",
				Email:    "test@example.com",
				Password: tt.password,
			}
			_, err := svc.Register(context.Background(), req)
			if err == nil {
				t.Errorf("expected error for password %q", tt.password)
			}
		})
	}
}

func TestRegister_EmailNormalized(t *testing.T) {
	svc := newTestService()

	req := RegisterRequest{
		Name:     "Test",
		Email:    "Test@Example.COM",
		Password: "Password123!",
	}

	user, err := svc.Register(context.Background(), req)
	if err != nil {
		t.Fatalf("expected no error: %v", err)
	}

	if user.Email != "test@example.com" {
		t.Errorf("expected normalized email, got: %s", user.Email)
	}
}

func TestLogin_Success(t *testing.T) {
	svc := newTestService()

	// Register first
	regReq := RegisterRequest{
		Name:     "Test",
		Email:    "login@example.com",
		Password: "Password123!",
	}
	_, err := svc.Register(context.Background(), regReq)
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	// Login
	loginReq := LoginRequest{
		Email:    "login@example.com",
		Password: "Password123!",
	}

	result, err := svc.Login(context.Background(), loginReq)
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	if result.AccessToken == "" {
		t.Error("expected non-empty access token")
	}
	if result.RefreshToken == "" {
		t.Error("expected non-empty refresh token")
	}
	if result.TokenType != "Bearer" {
		t.Errorf("expected Bearer token type, got: %s", result.TokenType)
	}
	if result.ExpiresIn <= 0 {
		t.Errorf("expected positive expires_in, got: %d", result.ExpiresIn)
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	svc := newTestService()

	regReq := RegisterRequest{
		Name:     "Test",
		Email:    "wrong@example.com",
		Password: "Password123!",
	}
	_, err := svc.Register(context.Background(), regReq)
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	loginReq := LoginRequest{
		Email:    "wrong@example.com",
		Password: "WrongPassword1",
	}

	_, err = svc.Login(context.Background(), loginReq)
	if err == nil {
		t.Fatal("expected error for wrong password")
	}
	if err.Error() != "email atau password salah" {
		t.Errorf("expected generic error, got: %v", err)
	}
}

func TestLogin_NonexistentEmail(t *testing.T) {
	svc := newTestService()

	loginReq := LoginRequest{
		Email:    "notfound@example.com",
		Password: "Password123!",
	}

	_, err := svc.Login(context.Background(), loginReq)
	if err == nil {
		t.Fatal("expected error for nonexistent email")
	}
	// Should return generic error to prevent email enumeration
	if err.Error() != "email atau password salah" {
		t.Errorf("expected generic error, got: %v", err)
	}
}

func TestRefreshToken_Success(t *testing.T) {
	svc := newTestService()

	// Register and login to get refresh token
	regReq := RegisterRequest{
		Name:     "Test",
		Email:    "refresh@example.com",
		Password: "Password123!",
	}
	_, err := svc.Register(context.Background(), regReq)
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	loginReq := LoginRequest{
		Email:    "refresh@example.com",
		Password: "Password123!",
	}
	loginResult, err := svc.Login(context.Background(), loginReq)
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	// Refresh
	result, err := svc.RefreshAccessToken(context.Background(), loginResult.RefreshToken)
	if err != nil {
		t.Fatalf("refresh failed: %v", err)
	}

	if result.AccessToken == "" {
		t.Error("expected new access token")
	}
	if result.RefreshToken == "" {
		t.Error("expected new refresh token")
	}
	if result.RefreshToken == loginResult.RefreshToken {
		t.Error("expected new refresh token (rotation)")
	}
}

func TestRefreshToken_Expired(t *testing.T) {
	svc := newTestService()

	// Create expired refresh token
	repo := svc.repo.(*mockRepository)
	tokenHash := hashToken("expired-token")
	now := time.Now().UTC()
	rt := &RefreshToken{
		ID:        uuid.New().String(),
		UserID:    uuid.New().String(),
		TokenHash: tokenHash,
		ExpiresAt: now.Add(-1 * time.Hour), // expired
		CreatedAt: now.Add(-2 * time.Hour),
	}
	repo.refreshTokens[rt.ID] = rt

	_, err := svc.RefreshAccessToken(context.Background(), "expired-token")
	if err == nil {
		t.Fatal("expected error for expired refresh token")
	}
	if err.Error() != "refresh token telah kadaluarsa" {
		t.Errorf("got: %v", err)
	}
}

func TestRefreshToken_Revoked(t *testing.T) {
	svc := newTestService()

	repo := svc.repo.(*mockRepository)
	tokenHash := hashToken("revoked-token")
	now := time.Now().UTC()
	revokedAt := now.Add(-1 * time.Hour)
	rt := &RefreshToken{
		ID:        uuid.New().String(),
		UserID:    uuid.New().String(),
		TokenHash: tokenHash,
		ExpiresAt: now.Add(24 * time.Hour),
		RevokedAt: &revokedAt,
		CreatedAt: now.Add(-2 * time.Hour),
	}
	repo.refreshTokens[rt.ID] = rt

	_, err := svc.RefreshAccessToken(context.Background(), "revoked-token")
	if err == nil {
		t.Fatal("expected error for revoked refresh token")
	}
	if err.Error() != "refresh token telah dicabut" {
		t.Errorf("got: %v", err)
	}
}

func TestLogout_Success(t *testing.T) {
	svc := newTestService()

	regReq := RegisterRequest{
		Name:     "Test",
		Email:    "logout@example.com",
		Password: "Password123!",
	}
	_, err := svc.Register(context.Background(), regReq)
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	loginReq := LoginRequest{
		Email:    "logout@example.com",
		Password: "Password123!",
	}
	loginResult, err := svc.Login(context.Background(), loginReq)
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	err = svc.Logout(context.Background(), loginResult.RefreshToken)
	if err != nil {
		t.Fatalf("logout failed: %v", err)
	}

	// Try to use the same refresh token
	_, err = svc.RefreshAccessToken(context.Background(), loginResult.RefreshToken)
	if err == nil {
		t.Fatal("expected error when using logged-out refresh token")
	}
}

func TestGetProfile_Success(t *testing.T) {
	svc := newTestService()

	regReq := RegisterRequest{
		Name:     "Profile User",
		Email:    "profile@example.com",
		Password: "Password123!",
	}
	created, err := svc.Register(context.Background(), regReq)
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	userID, _ := uuid.Parse(created.ID)
	profile, err := svc.GetProfile(context.Background(), userID)
	if err != nil {
		t.Fatalf("get profile failed: %v", err)
	}

	if profile.Name != "Profile User" {
		t.Errorf("expected name 'Profile User', got: %s", profile.Name)
	}
	if profile.Email != "profile@example.com" {
		t.Errorf("expected email 'profile@example.com', got: %s", profile.Email)
	}
}

func TestUpdateProfile_Success(t *testing.T) {
	svc := newTestService()

	regReq := RegisterRequest{
		Name:     "Old Name",
		Email:    "update@example.com",
		Password: "Password123!",
	}
	created, err := svc.Register(context.Background(), regReq)
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	userID, _ := uuid.Parse(created.ID)
	avatarURL := "https://example.com/avatar.png"
	updateReq := UpdateProfileRequest{
		Name:      "New Name",
		AvatarURL: &avatarURL,
	}

	updated, err := svc.UpdateProfile(context.Background(), userID, updateReq)
	if err != nil {
		t.Fatalf("update profile failed: %v", err)
	}

	if updated.Name != "New Name" {
		t.Errorf("expected name 'New Name', got: %s", updated.Name)
	}
	if updated.AvatarURL == nil || *updated.AvatarURL != avatarURL {
		t.Error("avatar URL not updated correctly")
	}
}

func TestChangePassword_Success(t *testing.T) {
	svc := newTestService()

	regReq := RegisterRequest{
		Name:     "Test",
		Email:    "changepass@example.com",
		Password: "Password123!",
	}
	created, err := svc.Register(context.Background(), regReq)
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	userID, _ := uuid.Parse(created.ID)
	changeReq := ChangePasswordRequest{
		CurrentPassword: "Password123!",
		NewPassword:     "NewPassword456!",
	}

	err = svc.ChangePassword(context.Background(), userID, changeReq)
	if err != nil {
		t.Fatalf("change password failed: %v", err)
	}

	// Verify can login with new password
	loginReq := LoginRequest{
		Email:    "changepass@example.com",
		Password: "NewPassword456!",
	}
	_, err = svc.Login(context.Background(), loginReq)
	if err != nil {
		t.Fatalf("login with new password failed: %v", err)
	}
}

func TestChangePassword_WrongCurrent(t *testing.T) {
	svc := newTestService()

	regReq := RegisterRequest{
		Name:     "Test",
		Email:    "wrongpass@example.com",
		Password: "Password123!",
	}
	created, err := svc.Register(context.Background(), regReq)
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	userID, _ := uuid.Parse(created.ID)
	changeReq := ChangePasswordRequest{
		CurrentPassword: "WrongPassword1",
		NewPassword:     "NewPassword456!",
	}

	err = svc.ChangePassword(context.Background(), userID, changeReq)
	if err == nil {
		t.Fatal("expected error for wrong current password")
	}
	if err.Error() != "password saat ini salah" {
		t.Errorf("got: %v", err)
	}
}

func TestValidatePasswordStrength(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{"valid", "Password123!", false},
		{"too short", "Ab1", true},
		{"no uppercase", "password123!", true},
		{"no lowercase", "PASSWORD123!", true},
		{"no digit", "PasswordOnly!", true},
		{"minimal valid", "Abcdefg1", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePasswordStrength(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePasswordStrength(%q) error = %v, wantErr %v", tt.password, err, tt.wantErr)
			}
		})
	}
}

func TestNormalizeEmail(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Test@Example.COM", "test@example.com"},
		{"  TEST@EXAMPLE.COM  ", "test@example.com"},
		{"normal@example.com", "normal@example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := NormalizeEmail(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeEmail(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestHashPassword(t *testing.T) {
	password := "TestPassword123!"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	if hash == "" {
		t.Error("expected non-empty hash")
	}
	if hash == password {
		t.Error("hash should not equal password")
	}

	// Verify
	if !CheckPassword(password, hash) {
		t.Error("CheckPassword should return true for correct password")
	}
	if CheckPassword("WrongPassword", hash) {
		t.Error("CheckPassword should return false for wrong password")
	}
}
