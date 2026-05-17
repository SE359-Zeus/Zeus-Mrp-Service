package models

import (
	"time"

	"github.com/google/uuid"
)

type PartCatalog struct {
	ID            uuid.UUID  `gorm:"type:uuid;primaryKey"`
	PartNumber    string     `gorm:"type:varchar;not null"`
	PartTypesID   int32      `gorm:"not null"`
	MfgNumber     string     `gorm:"type:varchar;not null"`
	Description   *string    `gorm:"type:text"`
	PartMfgStatus int32      `gorm:"not null"`
	CreatedAt     time.Time  `gorm:"not null"`
	UpdatedAt     time.Time  `gorm:"not null"`
	DeletedAt     *time.Time
}
