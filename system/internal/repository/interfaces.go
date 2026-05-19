package repository

import (
	"context"
	"time"

	"zeus-system-service/internal/models"

	"github.com/google/uuid"
)

type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	List(ctx context.Context, page, limit int, q string) ([]models.User, int64, error)
	Update(ctx context.Context, user *models.User) error
	SetStatus(ctx context.Context, id uuid.UUID, status models.AccountStatus) error
}

type RefreshTokenRepository interface {
	Create(ctx context.Context, token *models.RefreshToken) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.RefreshToken, error)
	DeleteByUserID(ctx context.Context, userID uuid.UUID) error
	DeleteExpired(ctx context.Context) error
}

type AuditRepository interface {
	Insert(ctx context.Context, log *models.AuditLog) error
	Query(ctx context.Context, filter models.AuditFilter, page, limit int) ([]models.AuditLog, int64, error)
	CountByAction(ctx context.Context, actionType models.ActionType, start, end time.Time) (int64, error)
}
