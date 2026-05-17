package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"zeus-scm-service/internal/service"
)

func TestInventory_GetStock(t *testing.T) {
	svc := service.NewInventoryService()
	
	stock, err := svc.GetStock(context.Background(), "SOC-XM100-PRO")
	assert.NoError(t, err)
	assert.NotNil(t, stock)
}

func TestInventory_AdjustStock(t *testing.T) {
	svc := service.NewInventoryService()
	
	// Adjusting stock should increase or decrease the quantity atomically
	err := svc.AdjustStock(context.Background(), "SOC-XM100-PRO", 100)
	assert.NoError(t, err)
}

func TestInventory_CheckAgingQuarantine(t *testing.T) {
	svc := service.NewInventoryService()
	
	// Test stability exemption and chronological age checks
	quarantined, err := svc.CheckAgingQuarantine(context.Background(), "BATT-LIPO-99W", "2019-01-01")
	assert.NoError(t, err)
	assert.True(t, quarantined, "Component is > 5 years old, should be quarantined")
}
