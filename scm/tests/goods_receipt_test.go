package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"zeus-scm-service/internal/service"
)

func TestGoodsReceipt_ParallelLockProcedure(t *testing.T) {
	svc := service.NewGoodsReceiptService()
	
	// Test acquiring the lock on a Goods Receipt manifest for inspection
	err := svc.AcquireLock(context.Background(), "GR-2024-301", "Operator-A")
	assert.NoError(t, err, "Should successfully acquire lock")
}

func TestGoodsReceipt_BlindReceiving(t *testing.T) {
	svc := service.NewGoodsReceiptService()
	
	counts := map[string]struct{ Received int; Defective int }{
		"SOC-XM100-PRO": {Received: 100, Defective: 0},
	}
	
	// Test that processing requires manually typed counts
	err := svc.ProcessBlindReceipt(context.Background(), "GR-2024-301", "Operator-A", counts)
	assert.NoError(t, err, "Should process blind receipt successfully")
}

func TestGoodsReceipt_ReleaseLock(t *testing.T) {
	svc := service.NewGoodsReceiptService()
	
	// Test releasing the 60-minute lock
	err := svc.ReleaseLock(context.Background(), "GR-2024-301")
	assert.NoError(t, err, "Should release lock successfully")
}
