package models

import (
	"time"
)

type PartMfgStatus struct {
	ID        int32      `gorm:"primaryKey;autoIncrement:false"`
	Name      string     `gorm:"type:varchar;not null"`
	CreatedAt time.Time  `gorm:"not null"`
	UpdatedAt time.Time  `gorm:"not null"`
	DeletedAt *time.Time
}
