package service

import (
	"context"
	"strings"
	"testing"
	"zeus-mrp-service/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProductionService_PlanProduction_TableDriven(t *testing.T) {
	svc := NewProductionService(nil)

	tests := []struct {
		name          string
		request       models.CreateProductionOrderRequest
		expectError   bool
		expectedModel string
	}{
		{
			name: "Valid Request",
			request: models.CreateProductionOrderRequest{
				ProductModelCode: "MODEL-A",
				TargetQuantity:   100,
			},
			expectError:   false,
			expectedModel: "MODEL-A",
		},
		{
			name: "Zero Quantity - Should Fail",
			request: models.CreateProductionOrderRequest{
				ProductModelCode: "MODEL-B",
				TargetQuantity:   0,
			},
			expectError: true,
		},
		{
			name: "Negative Quantity - Should Fail",
			request: models.CreateProductionOrderRequest{
				ProductModelCode: "MODEL-C",
				TargetQuantity:   -50,
			},
			expectError: true,
		},
		{
			name: "Extreme Quantity - Should Fail or Cap",
			request: models.CreateProductionOrderRequest{
				ProductModelCode: "MODEL-D",
				TargetQuantity:   999999999, // Max int check
			},
			expectError: true, // Assuming business rule caps order sizes
		},
		{
			name: "Missing Model Code - Should Fail",
			request: models.CreateProductionOrderRequest{
				ProductModelCode: "",
				TargetQuantity:   10,
			},
			expectError: true,
		},
		{
			name: "Invalid Characters in Model Code - Should Fail",
			request: models.CreateProductionOrderRequest{
				ProductModelCode: "MODEL*&^%",
				TargetQuantity:   10,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := svc.PlanProduction(context.Background(), tt.request)
			if tt.expectError {
				// We expect the real implementation to return an error here
				assert.Error(t, err, "Business Rule: Target quantity must be > 0 and ModelCode must be valid")
			} else {
				assert.NoError(t, err)
				require.NotNil(t, res)
				assert.Equal(t, tt.expectedModel, res.ProductModelCode)
			}
		})
	}
}

// TestPlanProduction_ResponseEchoesRequestFields verifies that every field in
// the response is derived from the request — not hard-coded or zeroed out.
// Tests a matrix of valid codes × quantities to catch any field that is
// accidentally fixed to a single value.
func TestPlanProduction_ResponseEchoesRequestFields(t *testing.T) {
	svc := NewProductionService(nil)

	cases := []struct {
		code string
		qty  int
	}{
		{"MODEL-A", 1},
		{"MODEL-Z", 50},
		{"X1", 500},
		{"PROD-9999", 1000},
	}

	for _, c := range cases {
		req := models.CreateProductionOrderRequest{
			ProductModelCode: c.code,
			TargetQuantity:   c.qty,
		}
		res, err := svc.PlanProduction(context.Background(), req)
		require.NoError(t, err, "code=%s qty=%d must be valid", c.code, c.qty)
		require.NotNil(t, res)
		assert.Equal(t, c.code, res.ProductModelCode,
			"ProductModelCode must echo the request exactly, got %q want %q", res.ProductModelCode, c.code)
		assert.Equal(t, c.qty, res.TargetQuantity,
			"TargetQuantity must equal request quantity, got %d want %d", res.TargetQuantity, c.qty)
	}
}

// TestPlanProduction_EachCallMintsFreshOrderID ensures the service generates
// a distinct UUID per call. Reusing an ID across calls would cause collisions
// in any downstream store or message bus.
func TestPlanProduction_EachCallMintsFreshOrderID(t *testing.T) {
	svc := NewProductionService(nil)
	req := models.CreateProductionOrderRequest{
		ProductModelCode: "MODEL-A",
		TargetQuantity:   10,
	}

	seen := make(map[uuid.UUID]bool)
	const calls = 10
	for i := 0; i < calls; i++ {
		res, err := svc.PlanProduction(context.Background(), req)
		require.NoError(t, err)
		require.NotNil(t, res)
		assert.False(t, seen[res.ID],
			"call %d produced a duplicate OrderID %s", i, res.ID)
		assert.NotEqual(t, uuid.Nil, res.ID,
			"call %d produced a nil OrderID", i)
		seen[res.ID] = true
	}
}

// TestPlanProduction_QuantityLowerBoundaryExact probes the fence at qty=1 vs
// qty=0. Off-by-one errors in the validator commonly appear exactly here.
func TestPlanProduction_QuantityLowerBoundaryExact(t *testing.T) {
	svc := NewProductionService(nil)
	base := models.CreateProductionOrderRequest{ProductModelCode: "MODEL-A"}

	for _, tc := range []struct {
		qty     int
		wantErr bool
		label   string
	}{
		{-2, true, "well below zero"},
		{-1, true, "one below zero"},
		{0, true, "zero — exclusive lower bound"},
		{1, false, "one — inclusive lower bound"},
		{2, false, "two — safely above lower bound"},
	} {
		req := base
		req.TargetQuantity = tc.qty
		_, err := svc.PlanProduction(context.Background(), req)
		if tc.wantErr {
			assert.Error(t, err, "qty=%d (%s) must be rejected", tc.qty, tc.label)
		} else {
			assert.NoError(t, err, "qty=%d (%s) must be accepted", tc.qty, tc.label)
		}
	}
}

// TestPlanProduction_QuantityUpperBoundaryExact probes the fence at the
// assumed business cap (1 000 000). Adjust the constant if the real cap differs.
func TestPlanProduction_QuantityUpperBoundaryExact(t *testing.T) {
	svc := NewProductionService(nil)
	const cap = 1_000_000
	base := models.CreateProductionOrderRequest{ProductModelCode: "MODEL-A"}

	for _, tc := range []struct {
		qty     int
		wantErr bool
		label   string
	}{
		{cap - 1, false, "one below cap — must be accepted"},
		{cap, false, "exactly at cap — must be accepted"},
		{cap + 1, true, "one above cap — must be rejected"},
		{999_999_999, true, "far above cap — must be rejected"},
	} {
		req := base
		req.TargetQuantity = tc.qty
		_, err := svc.PlanProduction(context.Background(), req)
		if tc.wantErr {
			assert.Error(t, err, "qty=%d (%s)", tc.qty, tc.label)
		} else {
			assert.NoError(t, err, "qty=%d (%s)", tc.qty, tc.label)
		}
	}
}

// TestPlanProduction_ModelCodeForbiddenCharsOneByOne tests each forbidden
// character individually. A single table-entry with "MODEL*&^%" hides which
// character actually breaks the validator.
func TestPlanProduction_ModelCodeForbiddenCharsOneByOne(t *testing.T) {
	svc := NewProductionService(nil)
	forbidden := `!@#$%^&*()+=[]{}|;':",.<>?/\ `

	for _, ch := range forbidden {
		code := "MODEL" + string(ch) + "X"
		req := models.CreateProductionOrderRequest{
			ProductModelCode: code,
			TargetQuantity:   10,
		}
		_, err := svc.PlanProduction(context.Background(), req)
		assert.Error(t, err,
			"model code %q (contains %q) must be rejected by the allowlist", code, string(ch))
	}
}

// TestPlanProduction_ModelCodeAllowedCharactersAreNeverRejected is the
// inverse of the forbidden-char test — it confirms the allowlist is not
// over-restrictive for alphanumeric + hyphen codes.
func TestPlanProduction_ModelCodeAllowedCharactersAreNeverRejected(t *testing.T) {
	svc := NewProductionService(nil)
	valid := []string{
		"A", "Z", "0", "9",
		"MODEL-1", "X-99", "PROD-ALPHA-3",
		"a", "z",            // lowercase (if allowed by business rule)
		strings.Repeat("A", 32), // long but character-valid
	}

	for _, code := range valid {
		res, err := svc.PlanProduction(context.Background(), models.CreateProductionOrderRequest{
			ProductModelCode: code,
			TargetQuantity:   1,
		})
		assert.NoError(t, err, "valid model code %q must not be rejected", code)
		assert.NotNil(t, res)
	}
}

// TestPlanProduction_WhitespaceModelCodesAreRejected guards against silent
// strings.TrimSpace normalisation that could let " " slip through as "".
func TestPlanProduction_WhitespaceModelCodesAreRejected(t *testing.T) {
	svc := NewProductionService(nil)
	cases := []struct {
		code  string
		label string
	}{
		{" ", "single space"},
		{"\t", "tab"},
		{"\n", "newline"},
		{"   ", "multiple spaces"},
		{" MODEL-A", "leading space"},
		{"MODEL-A ", "trailing space"},
		{" MODEL-A ", "both sides"},
	}

	for _, tc := range cases {
		_, err := svc.PlanProduction(context.Background(), models.CreateProductionOrderRequest{
			ProductModelCode: tc.code,
			TargetQuantity:   10,
		})
		assert.Error(t, err,
			"model code with whitespace (%s) must be rejected, not silently trimmed", tc.label)
	}
}

// TestPlanProduction_CaseSensitivityIsPreserved checks that the service does
// not normalise case. "model-a" and "MODEL-A" are different product codes and
// must not be collapsed to the same value in the response.
func TestPlanProduction_CaseSensitivityIsPreserved(t *testing.T) {
	svc := NewProductionService(nil)

	resLower, errLower := svc.PlanProduction(context.Background(), models.CreateProductionOrderRequest{
		ProductModelCode: "model-a", TargetQuantity: 1,
	})
	resUpper, errUpper := svc.PlanProduction(context.Background(), models.CreateProductionOrderRequest{
		ProductModelCode: "MODEL-A", TargetQuantity: 1,
	})

	require.NoError(t, errLower)
	require.NoError(t, errUpper)
	assert.NotEqual(t, resLower.ProductModelCode, resUpper.ProductModelCode,
		"case must be preserved in the response — 'model-a' and 'MODEL-A' are different codes")
}

// TestPlanProduction_SQLInjectionPatternsAreRejectedByValidation ensures
// that injection payloads are caught by the input validator, not by the DB.
// If they reach the DB layer the validator has a gap.
func TestPlanProduction_SQLInjectionPatternsAreRejectedByValidation(t *testing.T) {
	svc := NewProductionService(nil)
	payloads := []string{
		"'; DROP TABLE orders; --",
		"1 OR 1=1",
		"MODEL' AND '1'='1",
		"MODEL-A; SELECT *",
		"\" OR \"\"=\"",
	}

	for _, code := range payloads {
		_, err := svc.PlanProduction(context.Background(), models.CreateProductionOrderRequest{
			ProductModelCode: code,
			TargetQuantity:   1,
		})
		assert.Error(t, err,
			"SQL-injection payload %q must be rejected by the input validator", code)
	}
}

// TestPlanProduction_NonASCIIModelCodesAreRejected ensures Unicode letters
// outside ASCII are blocked. unicode.IsLetter returns true for CJK/Cyrillic,
// so a naive letter-check allowlist would accept them silently.
func TestPlanProduction_NonASCIIModelCodesAreRejected(t *testing.T) {
	svc := NewProductionService(nil)
	nonASCII := []string{
		"MODÉL",  // accented É (U+00C9)
		"型号-1",  // CJK characters
		"МОДЕЛЬ", // Cyrillic
		"مودل",   // Arabic
	}

	for _, code := range nonASCII {
		_, err := svc.PlanProduction(context.Background(), models.CreateProductionOrderRequest{
			ProductModelCode: code,
			TargetQuantity:   1,
		})
		assert.Error(t, err, "non-ASCII model code %q must be rejected", code)
	}
}

// TestPlanProduction_CancelledContextIsRejected ensures context cancellation
// is checked — the function must not proceed with a dead context.
func TestPlanProduction_CancelledContextIsRejected(t *testing.T) {
	svc := NewProductionService(nil)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := svc.PlanProduction(ctx, models.CreateProductionOrderRequest{
		ProductModelCode: "MODEL-A",
		TargetQuantity:   10,
	})
	assert.Error(t, err, "pre-cancelled context must be surfaced as an error")
}