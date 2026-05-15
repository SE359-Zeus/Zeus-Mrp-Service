package models

import (
	"time"

	"github.com/google/uuid"
)

type CreateProductionOrderRequest struct {
	ProductModelCode string    `json:"product_model_code" validate:"required"`
	TargetQuantity   int       `json:"target_quantity" validate:"gt=0"`
	ScheduledAt      time.Time `json:"scheduled_at"`
}

type ProductionOrderResponse struct {
	ID               uuid.UUID             `json:"id"`
	ProductModelCode string                `json:"product_model_code"`
	Status           ProductionOrderStatus `json:"status"`
	Shortages        []ShortageLog         `json:"shortages,omitempty"`
}

type BOMExplosionResult struct {
	PartID           uuid.UUID `json:"part_id"`
	TotalRequiredQty int       `json:"total_required_qty"`
	AvailableQty     int       `json:"available_qty"`
	IsShortage       bool      `json:"is_shortage"`
}
