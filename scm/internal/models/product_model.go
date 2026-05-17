package models

import (
	"time"
)

type ProductModel struct {
	ModelCode   string     `gorm:"type:varchar;primaryKey"`
	ModelName   string     `gorm:"type:varchar;not null"`
	Description *string    `gorm:"type:text"`
	CreatedAt   time.Time  `gorm:"not null"`
	UpdatedAt   time.Time  `gorm:"not null"`
	DeletedAt   *time.Time
}
