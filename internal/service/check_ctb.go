package service

import (
	"context"

	"github.com/google/uuid"
)

func (s *ProductionService) CheckClearToBuild(ctx context.Context, orderID uuid.UUID) (bool, error) {
	// TODO: Check if all components have >= required qty
	return false, nil
}
