package repository

import (
	"context"
	"zeus-mrp-service/internal/models"

	"github.com/google/uuid"
)

type MRPRepository interface {
	// Production Orders
	CreateProductionOrder(ctx context.Context, order *models.ProductionOrder) error
	GetProductionOrder(ctx context.Context, id uuid.UUID) (*models.ProductionOrder, error)
	UpdateProductionOrderStatus(ctx context.Context, id uuid.UUID, status models.ProductionOrderStatus) error

	// BOM
	GetBOMByModelCode(ctx context.Context, modelCode string) ([]models.BomEntry, error)

	// Shortages
	CreateShortageLog(ctx context.Context, log *models.ShortageLog) error
	GetShortagesByOrderID(ctx context.Context, orderID uuid.UUID) ([]models.ShortageLog, error)

	// External/Interop (Simplified for MRP)
	GetPartInventory(ctx context.Context, partID uuid.UUID) (int, error)
}
