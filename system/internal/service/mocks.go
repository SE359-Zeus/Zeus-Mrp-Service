package service

import (
	"context"
	"time"

	"zeus-system-service/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) != nil {
		return args.Get(0).(*models.User), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) != nil {
		return args.Get(0).(*models.User), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockUserRepository) List(ctx context.Context, page, limit int, q string) ([]models.User, int64, error) {
	args := m.Called(ctx, page, limit, q)
	var users []models.User
	if v := args.Get(0); v != nil {
		users = v.([]models.User)
	}
	return users, args.Get(1).(int64), args.Error(2)
}

func (m *MockUserRepository) Update(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) SetStatus(ctx context.Context, id uuid.UUID, status models.AccountStatus) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

type MockRefreshTokenRepository struct {
	mock.Mock
}

func (m *MockRefreshTokenRepository) Create(ctx context.Context, token *models.RefreshToken) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *MockRefreshTokenRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.RefreshToken, error) {
	args := m.Called(ctx, id)
	if args.Get(0) != nil {
		return args.Get(0).(*models.RefreshToken), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockRefreshTokenRepository) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockRefreshTokenRepository) DeleteExpired(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

type MockAuditRepository struct {
	mock.Mock
}

func (m *MockAuditRepository) Insert(ctx context.Context, log *models.AuditLog) error {
	args := m.Called(ctx, log)
	return args.Error(0)
}

func (m *MockAuditRepository) Query(ctx context.Context, filter models.AuditFilter, page, limit int) ([]models.AuditLog, int64, error) {
	args := m.Called(ctx, filter, page, limit)
	var logs []models.AuditLog
	if v := args.Get(0); v != nil {
		logs = v.([]models.AuditLog)
	}
	return logs, args.Get(1).(int64), args.Error(2)
}

func (m *MockAuditRepository) CountByAction(ctx context.Context, actionType models.ActionType, start, end time.Time) (int64, error) {
	args := m.Called(ctx, actionType, start, end)
	return args.Get(0).(int64), args.Error(1)
}
