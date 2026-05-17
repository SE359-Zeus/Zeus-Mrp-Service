package tests

import (
	"context"
	"testing"

	"zeus-scm-service/internal/service"

	"github.com/stretchr/testify/assert"
)

func TestShipment_DispatchLockingProcedure(t *testing.T) {
	svc := service.ShipmentService()

	// Test acquiring the 30-minute packing lock to prevent duplicate shipments
	err := svc.AcquireDispatchLock(context.Background(), "SHP-2024-201", "Operator-B")
	assert.NoError(t, err, "Should acquire dispatch lock successfully")
}

func TestShipment_InventoryDeductionTrigger(t *testing.T) {
	svc := service.ShipmentService()

	// Dispatching a shipment should trigger an atomic ledger update
	err := svc.DispatchShipment(context.Background(), "SHP-2024-201", "Operator-B")
	assert.NoError(t, err, "Should dispatch shipment successfully")
}
