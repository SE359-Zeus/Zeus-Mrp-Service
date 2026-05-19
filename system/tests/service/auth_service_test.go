package service_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"zeus-system-service/internal/models"
	"zeus-system-service/internal/service"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

type jwtAccessClaims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

type jwtRefreshClaims struct {
	JTI uuid.UUID `json:"jti"`
	SUB uuid.UUID `json:"sub"`
	jwt.RegisteredClaims
}

type mockRefreshRepo struct {
	mock.Mock
}

func (m *mockRefreshRepo) Create(ctx context.Context, token *models.RefreshToken) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *mockRefreshRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.RefreshToken, error) {
	args := m.Called(ctx, id)
	if args.Get(0) != nil {
		return args.Get(0).(*models.RefreshToken), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockRefreshRepo) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *mockRefreshRepo) DeleteExpired(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func generateTestKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err)
	return key
}

func assertTokenPair(t *testing.T, pair *models.TokenPair) {
	t.Helper()
	assert.NotEmpty(t, pair.AccessToken)
	assert.NotEmpty(t, pair.RefreshToken)
	assert.Equal(t, "Bearer", pair.TokenType)
	assert.Equal(t, int64(900), pair.ExpiresIn)
}

func setupAuthSvc(t *testing.T) (service.AuthService, *service.MockUserRepository, *mockRefreshRepo, *rsa.PrivateKey) {
	t.Helper()
	userRepo := new(service.MockUserRepository)
	refreshRepo := new(mockRefreshRepo)
	key := generateTestKey(t)

	userSvc := service.NewUserService(userRepo)
	svc := service.NewAuthService(userSvc, refreshRepo, key)

	return svc, userRepo, refreshRepo, key
}

func hashPassword(t *testing.T, password string) string {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	assert.NoError(t, err)
	return string(hash)
}

func TestAuthService_Login_Success(t *testing.T) {
	svc, userRepo, refreshRepo, _ := setupAuthSvc(t)
	email := "admin@zeus.com"
	password := "securepass123"
	userID := uuid.New()

	userRepo.On("GetByEmail", anyCtx, email).Return(&models.User{
		ID:           userID,
		Email:        email,
		PasswordHash: hashPassword(t, password),
		Role:         models.UserRoleAdmin,
		Status:       models.AccountStatusActive,
	}, nil)
	refreshRepo.On("Create", anyCtx, mock.AnythingOfType("*models.RefreshToken")).Return(nil)

	pair, err := svc.Login(context.Background(), models.LoginRequest{Email: email, Password: password})
	assert.NoError(t, err)
	assertTokenPair(t, pair)

	claims, err := svc.VerifyAccessToken(pair.AccessToken)
	assert.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, models.UserRoleAdmin, claims.Role)
	assert.Equal(t, email, claims.Email)

	userRepo.AssertExpectations(t)
	refreshRepo.AssertExpectations(t)
}

func TestAuthService_Login_InactiveUser(t *testing.T) {
	svc, userRepo, _, _ := setupAuthSvc(t)
	email := "inactive@zeus.com"

	userRepo.On("GetByEmail", anyCtx, email).Return(&models.User{
		Email:  email,
		Status: models.AccountStatusInactive,
	}, nil)

	pair, err := svc.Login(context.Background(), models.LoginRequest{Email: email, Password: "anypass"})
	assert.ErrorIs(t, err, service.ErrInactiveAccount)
	assert.Nil(t, pair)
	userRepo.AssertExpectations(t)
}

func TestAuthService_Login_InvalidCredentials(t *testing.T) {
	svc, userRepo, _, _ := setupAuthSvc(t)
	email := "admin@zeus.com"

	userRepo.On("GetByEmail", anyCtx, email).Return(&models.User{
		Email:        email,
		PasswordHash: hashPassword(t, "correctpass"),
		Role:         models.UserRoleAdmin,
		Status:       models.AccountStatusActive,
	}, nil)

	pair, err := svc.Login(context.Background(), models.LoginRequest{Email: email, Password: "wrongpass"})
	assert.ErrorIs(t, err, service.ErrUnauthorized)
	assert.Nil(t, pair)
	userRepo.AssertExpectations(t)
}

func TestAuthService_Refresh_Success(t *testing.T) {
	svc, userRepo, refreshRepo, _ := setupAuthSvc(t)
	userID := uuid.New()

	userRepo.On("GetByEmail", anyCtx, "admin@zeus.com").Return(&models.User{
		ID:           userID,
		Email:        "admin@zeus.com",
		PasswordHash: hashPassword(t, "pass"),
		Role:         models.UserRoleAdmin,
		Status:       models.AccountStatusActive,
	}, nil)
	refreshRepo.On("Create", anyCtx, mock.AnythingOfType("*models.RefreshToken")).Return(nil)

	loginPair, err := svc.Login(context.Background(), models.LoginRequest{Email: "admin@zeus.com", Password: "pass"})
	assert.NoError(t, err)

	refreshClaims := &jwtRefreshClaims{}
	_, _, err = jwt.NewParser().ParseUnverified(loginPair.RefreshToken, refreshClaims)
	assert.NoError(t, err)

	refreshRepo.On("GetByID", anyCtx, refreshClaims.JTI).Return(&models.RefreshToken{
		ID:        refreshClaims.JTI,
		UserID:    userID,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}, nil)
	userRepo.On("GetByID", anyCtx, userID).Return(&models.User{
		ID:    userID,
		Email: "admin@zeus.com",
		Role:  models.UserRoleAdmin,
	}, nil)
	refreshRepo.On("Create", anyCtx, mock.AnythingOfType("*models.RefreshToken")).Return(nil)

	pair, err := svc.Refresh(context.Background(), models.RefreshRequest{RefreshToken: loginPair.RefreshToken})
	assert.NoError(t, err)
	assertTokenPair(t, pair)
}

func TestAuthService_Refresh_ExpiredToken(t *testing.T) {
	svc, userRepo, refreshRepo, _ := setupAuthSvc(t)
	userID := uuid.New()

	userRepo.On("GetByEmail", anyCtx, "admin@zeus.com").Return(&models.User{
		ID:           userID,
		Email:        "admin@zeus.com",
		PasswordHash: hashPassword(t, "pass"),
		Role:         models.UserRoleAdmin,
		Status:       models.AccountStatusActive,
	}, nil)
	refreshRepo.On("Create", anyCtx, mock.AnythingOfType("*models.RefreshToken")).Return(nil)

	loginPair, err := svc.Login(context.Background(), models.LoginRequest{Email: "admin@zeus.com", Password: "pass"})
	assert.NoError(t, err)

	refreshClaims := &jwtRefreshClaims{}
	_, _, err = jwt.NewParser().ParseUnverified(loginPair.RefreshToken, refreshClaims)
	assert.NoError(t, err)

	refreshRepo.On("GetByID", anyCtx, refreshClaims.JTI).Return(&models.RefreshToken{
		ID:        refreshClaims.JTI,
		UserID:    userID,
		ExpiresAt: time.Now().Add(-24 * time.Hour),
	}, nil)

	pair, err := svc.Refresh(context.Background(), models.RefreshRequest{RefreshToken: loginPair.RefreshToken})
	assert.Error(t, err)
	assert.Nil(t, pair)
}

func TestAuthService_Refresh_InvalidToken(t *testing.T) {
	svc, _, _, _ := setupAuthSvc(t)

	pair, err := svc.Refresh(context.Background(), models.RefreshRequest{RefreshToken: "not-a-valid-jwt"})
	assert.Error(t, err)
	assert.Nil(t, pair)
}

func TestAuthService_VerifyAccessToken_Success(t *testing.T) {
	svc, userRepo, refreshRepo, _ := setupAuthSvc(t)
	userID := uuid.New()

	userRepo.On("GetByEmail", anyCtx, "v@z.com").Return(&models.User{
		ID:           userID,
		Email:        "v@z.com",
		PasswordHash: hashPassword(t, "pass"),
		Role:         models.UserRoleAdmin,
		Status:       models.AccountStatusActive,
	}, nil)
	refreshRepo.On("Create", anyCtx, mock.AnythingOfType("*models.RefreshToken")).Return(nil)

	pair, err := svc.Login(context.Background(), models.LoginRequest{Email: "v@z.com", Password: "pass"})
	assert.NoError(t, err)

	claimsResult, err := svc.VerifyAccessToken(pair.AccessToken)
	assert.NoError(t, err)
	assert.Equal(t, userID, claimsResult.UserID)
	assert.Equal(t, models.UserRoleAdmin, claimsResult.Role)
	assert.Equal(t, "v@z.com", claimsResult.Email)
	userRepo.AssertExpectations(t)
	refreshRepo.AssertExpectations(t)
}

func TestAuthService_VerifyAccessToken_Expired(t *testing.T) {
	svc, _, _, _ := setupAuthSvc(t)

	_, err := svc.VerifyAccessToken("eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjEwMDAwMDAwMDAsInVzZXJfaWQiOiIxIn0.signature")
	assert.Error(t, err)
}

func TestAuthService_VerifyAccessToken_WrongKey(t *testing.T) {
	otherKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	claims := &jwtAccessClaims{
		UserID: uuid.New().String(),
		Role:   string(models.UserRoleViewer),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenString, err := token.SignedString(otherKey)
	assert.NoError(t, err)

	svc, _, _, _ := setupAuthSvc(t)
	_, err = svc.VerifyAccessToken(tokenString)
	assert.Error(t, err)
}
