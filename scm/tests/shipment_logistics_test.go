package tests

import (
	"context"
	"testing"

	"zeus-scm-service/internal/models"
	"zeus-scm-service/internal/service"

	"github.com/stretchr/testify/assert"
)

func TestShipment_DispatchLockingProcedure(t *testing.T) {
	db := setupTestDB()
	db.AutoMigrate(&models.Shipment{}, &models.ShipmentItem{})
	svc := service.NewShipmentService(db, nil)

	err := svc.AcquireDispatchLock(context.Background(), "SHP-2024-201", "Operator-B")
	assert.Error(t, err, "Should fail when shipment does not exist")
}

func TestShipment_InventoryDeductionTrigger(t *testing.T) {
	db := setupTestDB()
	db.AutoMigrate(&models.Shipment{}, &models.ShipmentItem{}, &models.ComponentStock{})
	svc := service.NewShipmentService(db, nil)

	err := svc.DispatchShipment(context.Background(), "SHP-2024-201", "Operator-B")
	assert.Error(t, err, "Should fail when shipment does not exist")
}
