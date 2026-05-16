package service

import (
	"context"
	"zeus-mrp-service/internal/models"
)

func (s *ProductionService) PlanProduction(ctx context.Context, req models.CreateProductionOrderRequest) (*models.ProductionOrderResponse, error) {
	// TODO: Implement planning logic
	// 1. Validate req
	// 2. Create ProductionOrder entity
	// 3. Save to repo
	return &models.ProductionOrderResponse{}, nil
}
