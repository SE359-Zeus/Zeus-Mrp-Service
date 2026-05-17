package models

import (
	"time"

	"github.com/google/uuid"
)

type Warranty struct {
	ID             uuid.UUID  `gorm:"type:uuid;primaryKey"`
	CustomerID     uuid.UUID  `gorm:"type:uuid;not null"`
	ProductID      uuid.UUID  `gorm:"type:uuid;not null"`
	StartDate      time.Time  `gorm:"not null"`
	EndDate        time.Time  `gorm:"not null"`
	WarrantyStatus string     `gorm:"type:varchar;not null"`
	CreatedAt      time.Time  `gorm:"not null"`
	UpdatedAt      time.Time  `gorm:"not null"`
	DeletedAt      *time.Time
}
