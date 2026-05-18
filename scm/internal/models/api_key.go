package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ApiKey struct {
	ID         uuid.UUID      `gorm:"type:uuid;primaryKey"`
	Name       string         `gorm:"type:varchar;not null"`
	KeyPrefix  string         `gorm:"type:varchar(8);not null;uniqueIndex"`
	KeyHash    string         `gorm:"type:varchar;not null;uniqueIndex"`
	Active     bool           `gorm:"not null;default:true"`
	ExpiresAt  *time.Time
	LastUsedAt *time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  gorm.DeletedAt
}
