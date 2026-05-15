package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestProductionService_CheckClearToBuild_Strict(t *testing.T) {
	svc := NewProductionService(nil)
	orderID := uuid.New()

	t.Run("All Parts Available", func(t *testing.T) {
		ctb, err := svc.CheckClearToBuild(context.Background(), orderID)
		assert.NoError(t, err)
		assert.True(t, ctb, "CTB must be true when 100% components are available")
	})

	t.Run("Critical Shortage", func(t *testing.T) {
		ctb, err := svc.CheckClearToBuild(context.Background(), orderID)
		assert.NoError(t, err)
		assert.False(t, ctb, "CTB must be false if any component is missing")
	})
}
