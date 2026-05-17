package tests

import (
	"context"
	"testing"

	"zeus-scm-service/internal/models"
	"zeus-scm-service/internal/service"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestPOOrchestration_CreateDraft(t *testing.T) {
	svc := service.POService()

	po, err := svc.CreateDraft(context.Background(), uuid.New(), "Build-X1")
	assert.NoError(t, err, "Should successfully create draft")
	assert.NotNil(t, po)
	if po != nil {
		assert.Equal(t, models.POStatusDraft, po.Status)
	}
}

func TestPOOrchestration_EagerSlotLocking(t *testing.T) {
	svc := service.POService()

	// Adding an item to a draft should immediately deduct from global deficit pool (lock)
	err := svc.AddLineItemWithLock(context.Background(), "PO-2024-108", "MOD-WIFI7-AX", 200)
	assert.NoError(t, err, "Should acquire slot lock successfully")
}

func TestPOOrchestration_ApprovePO(t *testing.T) {
	svc := service.POService()

	// Should fail if the 30-minute lock expired before approval
	err := svc.ApprovePO(context.Background(), "PO-2024-108")
	assert.NoError(t, err, "Should transition Draft to Approved")
}

func TestPOOrchestration_StateRegressionPrevention(t *testing.T) {
	svc := service.POService()

	// State should not be allowed to revert from 'Received' or 'In Transit' back to 'Draft'
	err := svc.TransitionState(context.Background(), "PO-2024-101", models.POStatusDraft)
	assert.Error(t, err, "Should block state regression")
}
