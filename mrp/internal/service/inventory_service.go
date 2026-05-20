package service

import (
	"context"
	"zeus-mrp-service/internal/models"
)

func (s *ProductionService) GetInventoryLedger(ctx context.Context) ([]models.InventoryTransactionDTO, error) {
	// TODO: This service is read-only for MRP.
	// It should query the Product service or Audit service for transaction history.
	return nil, nil
}

func (s *ProductionService) GetInventoryMetrics(ctx context.Context) (*models.InventoryMetrics, error) {
	// TODO: Aggregate and return inventory KPIs
	return nil, nil
}

func (s *ProductionService) ExportInventoryCSV(ctx context.Context) ([]byte, error) {
	// TODO: Generate and return CSV data
	return nil, nil
}
