package service

import (
	"context"
	"testing"
	"zeus-mrp-service/internal/models"

	"github.com/stretchr/testify/assert"
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
			name: "Missing Model Code - Should Fail",
			request: models.CreateProductionOrderRequest{
				ProductModelCode: "",
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
				// Since it's a stub returning nil, nil, we assert against our desired "Strong" rule
				assert.Error(t, err, "Business Rule: Target quantity must be > 0 and ModelCode must be present")
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, res)
				assert.Equal(t, tt.expectedModel, res.ProductModelCode)
			}
		})
	}
}
