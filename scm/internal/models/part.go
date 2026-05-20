package models

import (
	"time"

	"github.com/google/uuid"
)

type Part struct {
	ID               uuid.UUID  `gorm:"type:uuid;primaryKey"`
	PartCatalogID    uuid.UUID  `gorm:"type:uuid;not null"`
	ProductID        *uuid.UUID `gorm:"type:uuid"`
	SerialNumber     string     `gorm:"type:varchar;not null"`
	PartConditionID  int32      `gorm:"not null"`
	ManufacturedDate time.Time  `gorm:"not null"`
	InstallationDate *time.Time
	RemovalDate      *time.Time
	ScrappedDate     *time.Time
	CreatedAt        time.Time `gorm:"not null"`
	UpdatedAt        time.Time `gorm:"not null"`
	DeletedAt        *time.Time
}
