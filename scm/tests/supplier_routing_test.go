package tests

import (
	"context"
	"testing"

	"zeus-scm-service/internal/service"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestVendorRouting_GetOptimalSupplier(t *testing.T) {
	svc := service.VendorService()

	// Test the logic that resolves component shortage by sorting by QualityScore and UnitPrice
	supplier, mapping, err := svc.GetOptimalSupplier(context.Background(), "SOC-XM100-PRO")

	assert.NoError(t, err, "Should successfully route the shortage to the optimal supplier")
	assert.NotNil(t, supplier)
	assert.NotNil(t, mapping)
	// Example assertions for implementors:
	// assert.Equal(t, "Intel Corporation", supplier.Name)
}

func TestVendorRouting_UpdateSupplierMetrics(t *testing.T) {
	svc := service.VendorService()

	// Test recalculation of OnTimeRate and QualityScore based on Goods Receipt logs
	err := svc.UpdateSupplierMetrics(context.Background(), uuid.New())

	assert.NoError(t, err, "Should successfully recalculate metrics")
}
