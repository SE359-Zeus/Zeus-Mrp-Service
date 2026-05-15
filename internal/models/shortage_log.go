package models

import "github.com/google/uuid"

type ShortageLog struct {
	ID                uuid.UUID `json:"id"`
	ProductionOrderID uuid.UUID `json:"production_order_id"`
	PartID            uuid.UUID `json:"part_id"`
	ShortageQty       int       `json:"shortage_qty"`
	ResolutionStatus  string    `json:"resolution_status"` // e.g., EMITTED, RESOLVED
}
