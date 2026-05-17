package models

import (
	"time"

	"github.com/google/uuid"
)

type Product struct {
	ID               uuid.UUID  `gorm:"type:uuid;primaryKey"`
	ProductModelCode string     `gorm:"type:varchar;not null"`
	CustomerID       uuid.UUID  `gorm:"type:uuid;not null"`
	ProductName      string     `gorm:"type:varchar;not null"`
	SerialNumber     string     `gorm:"type:varchar;not null"`
	CreatedAt        time.Time  `gorm:"not null"`
	UpdatedAt        time.Time  `gorm:"not null"`
	DeletedAt        *time.Time
}
