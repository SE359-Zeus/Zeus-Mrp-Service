package models

import "github.com/google/uuid"

type BomEntry struct {
	ID                       int       `json:"id"`
	ParentModelCode          string    `json:"parent_model_code"`
	ComponentPartID          uuid.UUID `json:"component_part_id"`
	RequiredQuantityPerUnit  int       `json:"required_quantity_per_unit"`
}
