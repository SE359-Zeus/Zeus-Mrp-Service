package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ComponentStatus string

const (
	ComponentStatusInStock     ComponentStatus = "In Stock"
	ComponentStatusLowStock    ComponentStatus = "Low Stock"
	ComponentStatusOutOfStock  ComponentStatus = "Out of Stock"
	ComponentStatusDiscontinued ComponentStatus = "Discontinued"
)

// ComponentStock represents the inventory ledger entity derived from the SCM UI requirements
// This complements the reference Product/Part entities by providing warehouse logistics data.
type ComponentStock struct {
	SKU               string          `gorm:"type:varchar(100);primary_key"`
	Name              string          `gorm:"type:varchar(255);not null"`
	Category          string          `gorm:"type:varchar(100);not null"`
	StockQty          int             `gorm:"not null;default:0"`
	ReorderPoint      int             `gorm:"not null;default:0"`
	UnitCost          float64         `gorm:"not null;default:0.0"`
	Status            ComponentStatus `gorm:"type:varchar(50);not null"`
	PrimarySupplierID uuid.UUID       `gorm:"type:uuid"`
	LeadTimeDays      int             `gorm:"not null;default:0"`
	Location          string          `gorm:"type:varchar(255)"`
	CreatedAt         time.Time       `gorm:"autoCreateTime"`
	UpdatedAt         time.Time       `gorm:"autoUpdateTime"`
	DeletedAt         gorm.DeletedAt  `gorm:"index"`
}
