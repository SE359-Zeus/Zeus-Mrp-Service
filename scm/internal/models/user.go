package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID             uuid.UUID  `gorm:"type:uuid;primaryKey"`
	AccountStatus  int32      `gorm:"not null"`
	RoleID         int32      `gorm:"not null"`
	Email          string     `gorm:"type:varchar;uniqueIndex;not null"`
	FullName       string     `gorm:"type:varchar;not null"`
	PasswordHash   string     `gorm:"type:varchar;not null"`
	PhoneNumber    string     `gorm:"type:varchar;not null"`
	Province       *string    `gorm:"type:varchar"`
	FcmToken       *string    `gorm:"type:varchar"`
	InstallationID *string    `gorm:"type:varchar"`
	CreatedAt      time.Time  `gorm:"not null"`
	UpdatedAt      time.Time  `gorm:"not null"`
	DeletedAt      *time.Time
}
