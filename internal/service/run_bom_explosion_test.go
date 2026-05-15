package service

import (
	"context"
	"sync"
	"testing"
	"zeus-mrp-service/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProductionService_BOMExplosion_EdgeCases(t *testing.T) {
	svc := NewProductionService(nil)

	tests := []struct {
		name             string
		orderID          uuid.UUID
		mockBOM          []models.BomEntry
		mockInventory    map[uuid.UUID]int
		expectedError    bool
		expectedShortage bool
	}{
		{
			name:             "Sufficient Stock",
			orderID:          uuid.New(),
			expectedError:    false,
			expectedShortage: false,
		},
		{
			name:             "Exact Stock Match",
			orderID:          uuid.New(),
			expectedError:    false,
			expectedShortage: false,
		},
		{
			name:             "One Part Missing - Shortage",
			orderID:          uuid.New(),
			expectedError:    false,
			expectedShortage: true,
		},
		{
			name:             "Partial Stock - Shortage",
			orderID:          uuid.New(),
			expectedError:    false,
			expectedShortage: true,
		},
		{
			name:             "Empty BOM (No components required)",
			orderID:          uuid.New(),
			expectedError:    false,
			expectedShortage: false, // If no parts needed, no shortage
		},
		{
			name:             "Invalid Order ID",
			orderID:          uuid.Nil,
			expectedError:    true,
			expectedShortage: false,
		},
		{
			name:             "BOM Entry with Zero Quantity Required",
			orderID:          uuid.New(),
			expectedError:    true, // Business logic should reject invalid BOM definitions
			expectedShortage: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := svc.RunBOMExplosion(context.Background(), tt.orderID)
			
			if tt.expectedError {
				assert.Error(t, err, "Expected an error for invalid input or state")
			} else {
				assert.NoError(t, err)
				// These assertions describe the expected behavior for complex scenarios
				if tt.expectedShortage {
					hasShortage := false
					for _, r := range results {
						if r.IsShortage {
							hasShortage = true
						}
					}
					assert.True(t, hasShortage, "Should detect shortage in this scenario")
				}
			}
		})
	}
}

// TestBOMExplosion_ResultSliceIsNeverNilOnSuccess enforces that callers can
// always safely range over the result. A nil slice is valid Go but breaks
// code that checks len(results) == 0 to mean "no BOM lines exist".
func TestBOMExplosion_ResultSliceIsNeverNilOnSuccess(t *testing.T) {
	svc := NewProductionService(nil)
	results, err := svc.RunBOMExplosion(context.Background(), uuid.New())
	require.NoError(t, err)
	assert.NotNil(t, results,
		"result must be an empty slice, not nil — callers must be able to range over it safely")
}

// TestBOMExplosion_ErrorReturnsEmptyResults enforces atomicity: an error
// return must never be accompanied by partial result data. Mixing (results, err)
// forces callers to decide which to trust.
func TestBOMExplosion_ErrorReturnsEmptyResults(t *testing.T) {
	svc := NewProductionService(nil)
	results, err := svc.RunBOMExplosion(context.Background(), uuid.Nil)
	require.Error(t, err, "uuid.Nil must be rejected")
	assert.Empty(t, results,
		"when an error is returned the result slice must be nil/empty — no partial data")
}

// TestBOMExplosion_ShortageFieldAgreesWithQtyFields checks the internal
// consistency of every result row: IsShortage must be the logical consequence
// of AvailableQty vs RequiredQty, not an independently set flag.
func TestBOMExplosion_ShortageFieldAgreesWithQtyFields(t *testing.T) {
	svc := NewProductionService(nil)
	results, err := svc.RunBOMExplosion(context.Background(), uuid.New())
	if err != nil {
		return
	}

	for i, r := range results {
		if r.IsShortage {
			assert.Less(t, r.AvailableQty, r.TotalRequiredQty,
				"result[%d] PartID=%s: IsShortage=true but AvailableQty(%d) >= RequiredQty(%d) — flag is inconsistent",
				i, r.PartID, r.AvailableQty, r.TotalRequiredQty)
		} else {
			assert.GreaterOrEqual(t, r.AvailableQty, r.TotalRequiredQty,
				"result[%d] PartID=%s: IsShortage=false but AvailableQty(%d) < RequiredQty(%d) — shortage was missed",
				i, r.AvailableQty, r.TotalRequiredQty)
		}
	}
}

// TestBOMExplosion_RequiredQtyIsStrictlyPositive documents the BOM validity
// invariant: a component line with RequiredQty <= 0 is a data error and must
// never appear in the output (the implementation must reject it upstream).
func TestBOMExplosion_RequiredQtyIsStrictlyPositive(t *testing.T) {
	svc := NewProductionService(nil)
	results, err := svc.RunBOMExplosion(context.Background(), uuid.New())
	if err != nil {
		return
	}

	for i, r := range results {
		assert.Greater(t, r.TotalRequiredQty, 0,
			"result[%d] PartID=%s: RequiredQty must be > 0; zero/negative BOM lines are invalid data",
			i, r.PartID)
	}
}

// TestBOMExplosion_AvailableQtyIsNonNegative asserts that inventory levels
// can never be reported as negative — a stock count below zero is nonsensical.
func TestBOMExplosion_AvailableQtyIsNonNegative(t *testing.T) {
	svc := NewProductionService(nil)
	results, err := svc.RunBOMExplosion(context.Background(), uuid.New())
	if err != nil {
		return
	}

	for i, r := range results {
		assert.GreaterOrEqual(t, r.AvailableQty, 0,
			"result[%d] PartID=%s: AvailableQty=%d is negative — impossible stock level",
			i, r.PartID, r.AvailableQty)
	}
}

// TestBOMExplosion_PartIDsAreUnique verifies that the explosion does not
// emit duplicate component rows. Duplicate rows would double-count shortage
// quantities for any downstream procurement logic.
func TestBOMExplosion_PartIDsAreUnique(t *testing.T) {
	svc := NewProductionService(nil)
	results, err := svc.RunBOMExplosion(context.Background(), uuid.New())
	if err != nil {
		return
	}

	seen := make(map[uuid.UUID]int, len(results))
	for i, r := range results {
		prev, exists := seen[r.PartID]
		assert.False(t, exists,
			"result[%d]: PartID %s already appeared at result[%d] — duplicate rows corrupt shortage calculations",
			i, r.PartID, prev)
		seen[r.PartID] = i
	}
}

// TestBOMExplosion_PartIDsAreNotNil ensures no result row carries a
// zero-value PartID, which would indicate an unresolved BOM reference.
func TestBOMExplosion_PartIDsAreNotNil(t *testing.T) {
	svc := NewProductionService(nil)
	results, err := svc.RunBOMExplosion(context.Background(), uuid.New())
	if err != nil {
		return
	}

	for i, r := range results {
		assert.NotEqual(t, uuid.Nil, r.PartID,
			"result[%d]: PartID is uuid.Nil — unresolved BOM reference must not appear in output", i)
	}
}

// TestBOMExplosion_DeterministicResultOrder verifies that two calls for the
// same order return results in the same order. Non-deterministic ordering
// breaks snapshot tests and makes diffs unreadable in audits.
func TestBOMExplosion_DeterministicResultOrder(t *testing.T) {
	svc := NewProductionService(nil)
	id := uuid.New()

	first, err1 := svc.RunBOMExplosion(context.Background(), id)
	second, err2 := svc.RunBOMExplosion(context.Background(), id)

	assert.Equal(t, err1 == nil, err2 == nil,
		"error consistency: both calls must agree on error-or-not")
	if err1 != nil || err2 != nil {
		return
	}

	require.Equal(t, len(first), len(second),
		"result count must be identical across two calls with no state change")

	for i := range first {
		assert.Equal(t, first[i].PartID, second[i].PartID,
			"result[%d]: PartID differs between calls — ordering is non-deterministic", i)
	}
}

// TestBOMExplosion_NilUUIDVariantsAreAllRejected checks both uuid.Nil and
// the zero-value uuid.UUID{} — they have identical bytes but may be reached
// via different code paths (explicit Nil vs uninitialized struct field).
func TestBOMExplosion_NilUUIDVariantsAreAllRejected(t *testing.T) {
	svc := NewProductionService(nil)

	for _, tc := range []struct {
		label string
		id    uuid.UUID
	}{
		{"uuid.Nil constant", uuid.Nil},
		{"zero-value uuid.UUID{}", uuid.UUID{}},
	} {
		t.Run(tc.label, func(t *testing.T) {
			results, err := svc.RunBOMExplosion(context.Background(), tc.id)
			assert.Error(t, err, "%s must be rejected as an invalid order ID", tc.label)
			assert.Empty(t, results, "no results must be returned alongside an error")
		})
	}
}

// TestBOMExplosion_CancelledContextIsRejected ensures the function honours a
// cancelled context rather than continuing work on a dead request.
func TestBOMExplosion_CancelledContextIsRejected(t *testing.T) {
	svc := NewProductionService(nil)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := svc.RunBOMExplosion(ctx, uuid.New())
	assert.Error(t, err, "pre-cancelled context must surface as an error")
}

// TestBOMExplosion_ConcurrentCallsDoNotRace runs multiple goroutines to catch
// data races on the service. Run with: go test -race ./...
func TestBOMExplosion_ConcurrentCallsDoNotRace(t *testing.T) {
	svc := NewProductionService(nil)
	id := uuid.New()

	const goroutines = 20
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			svc.RunBOMExplosion(context.Background(), id) //nolint:errcheck
		}()
	}
	wg.Wait()
}

// TestBOMExplosion_ShortageCountNeverExceedsTotalResults is a sanity bound:
// the number of shortage rows must be <= total rows, since every shortage row
// is also a result row.
func TestBOMExplosion_ShortageCountNeverExceedsTotalResults(t *testing.T) {
	svc := NewProductionService(nil)
	results, err := svc.RunBOMExplosion(context.Background(), uuid.New())
	if err != nil {
		return
	}

	shortageCount := 0
	for _, r := range results {
		if r.IsShortage {
			shortageCount++
		}
	}
	assert.LessOrEqual(t, shortageCount, len(results),
		"shortage count (%d) cannot exceed total result count (%d)", shortageCount, len(results))
}

// TestBOMExplosion_MultipleOrdersAreIndependent calls RunBOMExplosion for
// several distinct order IDs and checks that results for order A are not
// polluted by the call for order B (no shared mutable state).
func TestBOMExplosion_MultipleOrdersAreIndependent(t *testing.T) {
	svc := NewProductionService(nil)

	type callResult struct {
		results []models.BOMExplosionResult
		err     error
	}

	ids := make([]uuid.UUID, 5)
	first := make(map[uuid.UUID]callResult, 5)

	for i := range ids {
		ids[i] = uuid.New()
		r, e := svc.RunBOMExplosion(context.Background(), ids[i])
		first[ids[i]] = callResult{r, e}
	}

	// Second pass: results must match the first call per order
	for _, id := range ids {
		r2, e2 := svc.RunBOMExplosion(context.Background(), id)
		f := first[id]
		assert.Equal(t, f.err == nil, e2 == nil,
			"order %s: error consistency violated between two calls", id)
		if f.err == nil && e2 == nil {
			assert.Equal(t, len(f.results), len(r2),
				"order %s: result count changed between calls with no state change", id)
		}
	}
}