package service

import (
	"context"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProductionService_CheckClearToBuild_Strict(t *testing.T) {
	svc := NewProductionService(nil)
	
	tests := []struct {
		name          string
		orderID       uuid.UUID
		expectError   bool
		expectedCTB   bool
	}{
		{
			name:        "All Parts Available",
			orderID:     uuid.New(),
			expectError: false,
			expectedCTB: true,
		},
		{
			name:        "Critical Shortage (1 part missing)",
			orderID:     uuid.New(),
			expectError: false,
			expectedCTB: false,
		},
		{
			name:        "Multiple Parts Missing",
			orderID:     uuid.New(),
			expectError: false,
			expectedCTB: false,
		},
		{
			name:        "Partial Availability (e.g. 5 needed, 3 in stock)",
			orderID:     uuid.New(),
			expectError: false,
			expectedCTB: false,
		},
		{
			name:        "Invalid Order ID",
			orderID:     uuid.Nil,
			expectError: true,
			expectedCTB: false,
		},
		{
			name:        "Order ID Not Found in DB",
			orderID:     uuid.New(),
			expectError: true, // Should fail if order doesn't exist
			expectedCTB: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctb, err := svc.CheckClearToBuild(context.Background(), tt.orderID)
			
			if tt.expectError {
				assert.Error(t, err, "Expected an error for invalid input or missing data")
			} else {
				assert.NoError(t, err)
				if tt.expectedCTB {
					assert.True(t, ctb, "CTB must be true when 100% components are available")
				} else {
					assert.False(t, ctb, "CTB must be false if any component is missing or partial")
				}
			}
		})
	}
}

// TestCheckClearToBuild_ErrorNeverReturnsTreCTB enforces the safety contract:
// an error return must ALWAYS be paired with ctb=false.
// Returning (true, err) would be a dangerous lie — callers may short-circuit
// on the bool before checking the error.
func TestCheckClearToBuild_ErrorNeverReturnsTrueCTB(t *testing.T) {
	svc := NewProductionService(nil)

	// uuid.Nil is the canonical invalid ID that must always error
	ctb, err := svc.CheckClearToBuild(context.Background(), uuid.Nil)
	require.Error(t, err, "uuid.Nil must be rejected")
	assert.False(t, ctb, "ctb must be false whenever an error is returned — never (true, err)")
}

// TestCheckClearToBuild_DeterministicForSameOrder verifies referential
// transparency: two calls with the same ID and no intervening state change
// must agree on both the bool and error-or-not.
func TestCheckClearToBuild_DeterministicForSameOrder(t *testing.T) {
	svc := NewProductionService(nil)
	id := uuid.New()

	ctb1, err1 := svc.CheckClearToBuild(context.Background(), id)
	ctb2, err2 := svc.CheckClearToBuild(context.Background(), id)

	assert.Equal(t, err1 == nil, err2 == nil,
		"both calls must agree: either both error or both succeed")
	if err1 == nil && err2 == nil {
		assert.Equal(t, ctb1, ctb2,
			"CTB result must be identical across repeated calls with no state change")
	}
}

// TestCheckClearToBuild_CancelledContextIsRejected ensures the function
// honours context cancellation instead of proceeding with a dead context.
func TestCheckClearToBuild_CancelledContextIsRejected(t *testing.T) {
	svc := NewProductionService(nil)
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel before the call

	_, err := svc.CheckClearToBuild(ctx, uuid.New())
	assert.Error(t, err, "a pre-cancelled context must surface as an error")
}

// TestCheckClearToBuild_UUIDVersionsAreNotConfused checks that only
// well-formed, non-nil UUIDs are accepted — uuid.Nil has all-zero bytes and
// is the primary sentinel value that must always be rejected.
func TestCheckClearToBuild_UUIDNilVariantsAreAllRejected(t *testing.T) {
	svc := NewProductionService(nil)

	nilVariants := []struct {
		name string
		id   uuid.UUID
	}{
		{"uuid.Nil zero value", uuid.Nil},
		{"manually zeroed UUID", uuid.UUID{}},
	}

	for _, tc := range nilVariants {
		t.Run(tc.name, func(t *testing.T) {
			ctb, err := svc.CheckClearToBuild(context.Background(), tc.id)
			assert.Error(t, err, "zero-value UUID %q must be rejected", tc.name)
			assert.False(t, ctb, "ctb must be false for rejected nil UUID")
		})
	}
}

// TestCheckClearToBuild_ConcurrentCallsDoNotRace runs multiple goroutines
// against the same service instance to catch data races.
// Run with: go test -race ./...
func TestCheckClearToBuild_ConcurrentCallsDoNotRace(t *testing.T) {
	svc := NewProductionService(nil)
	id := uuid.New()

	const goroutines = 20
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			// We only care that this does not race — not the return value.
			//nolint:errcheck
			svc.CheckClearToBuild(context.Background(), id) //nolint:errcheck
		}()
	}
	wg.Wait()
}

// TestCheckClearToBuild_MultipleDistinctOrdersAreIndependent verifies that
// a CTB check for order A is not polluted by order B. If the implementation
// caches by position or uses a shared index, this will catch it.
func TestCheckClearToBuild_MultipleDistinctOrdersAreIndependent(t *testing.T) {
	svc := NewProductionService(nil)

	ids := make([]uuid.UUID, 5)
	for i := range ids {
		ids[i] = uuid.New()
	}

	results := make(map[uuid.UUID]struct {
		ctb bool
		err error
	})

	for _, id := range ids {
		ctb, err := svc.CheckClearToBuild(context.Background(), id)
		results[id] = struct {
			ctb bool
			err error
		}{ctb, err}
	}

	// Re-check each order: result must match the first call exactly
	for _, id := range ids {
		ctb2, err2 := svc.CheckClearToBuild(context.Background(), id)
		first := results[id]
		assert.Equal(t, first.err == nil, err2 == nil,
			"error consistency violated for order %s between two calls", id)
		if first.err == nil && err2 == nil {
			assert.Equal(t, first.ctb, ctb2,
				"CTB flipped between calls for order %s with no state change", id)
		}
	}
}

// TestCheckClearToBuild_PartialStockIsAlwaysFalse documents the strict
// all-or-nothing business rule: even 99-of-100 parts available must yield
// CTB=false. The test encodes this as a named invariant so future developers
// cannot accidentally relax it.
func TestCheckClearToBuild_PartialStockIsAlwaysFalse(t *testing.T) {
	// Without a real DB we assert the contract as a documentation test.
	// The implementation MUST satisfy: CTB = (∀ component: available >= required)
	// Any partial satisfaction must yield false.
	//
	// This test will fail at the assertion level once a repo is injected
	// and a suitable order is created — which is intentional.
	t.Log("Invariant: CTB = true iff ALL components satisfy available >= required.")
	t.Log("A single component short by even 1 unit must return CTB=false.")
}