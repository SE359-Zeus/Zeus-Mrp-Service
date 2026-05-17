package service

import (
	"context"

	"zeus-scm-service/internal/models"
)

type InventoryService interface {
	// GetStock Retrieves the component stock based on SKU.
	GetStock(ctx context.Context, sku string) (*models.ComponentStock, error)
	
	// AdjustStock performs an atomic ledger update upon GR completion or Shipment dispatch.
	AdjustStock(ctx context.Context, sku string, quantityDelta int) error
	
	// CheckAgingQuarantine forces chronological age check based on production date if component is aging sensitive.
	CheckAgingQuarantine(ctx context.Context, sku string, productionDate string) (bool, error)
}

type inventoryService struct{}

func NewInventoryService() InventoryService {
	return &inventoryService{}
}

func (s *inventoryService) GetStock(ctx context.Context, sku string) (*models.ComponentStock, error) {
	return nil, ErrNotImplemented
}

func (s *inventoryService) AdjustStock(ctx context.Context, sku string, quantityDelta int) error {
	return ErrNotImplemented
}

func (s *inventoryService) CheckAgingQuarantine(ctx context.Context, sku string, productionDate string) (bool, error) {
	return false, ErrNotImplemented
}
