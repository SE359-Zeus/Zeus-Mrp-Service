package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ApiKey struct {
	ID         uuid.UUID `gorm:"primaryKey"`
	Name       string    `gorm:"not null"`
	KeyPrefix  string    `gorm:"not null;uniqueIndex"`
	KeyHash    string    `gorm:"not null;uniqueIndex"`
	Active     bool      `gorm:"not null;default:true"`
	ExpiresAt  *time.Time
	LastUsedAt *time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  gorm.DeletedAt
}
