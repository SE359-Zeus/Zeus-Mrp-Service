package service

import (
	"context"
	"testing"
	"time"

	"zeus-sales-service/internal/models"
	rootrepo "zeus-sales-service/internal/repository"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestOrderService_CreateOrder_ValidatesHardCases(t *testing.T) {
	db := setupMockDbRepo()
	cache := setupMockCacheRepo()
	svc := newTestServicesWithMocks(db, cache).Orders
	baseReq := models.CreateOrderRequest{
		ClientName:         "Acme Manufacturing",
		DestinationAddress: "Dock 9",
		ClientTier:         models.ClientTierB2B,
		RequiredDate:       time.Now().Add(24 * time.Hour).UTC(),
		Items:              []models.OrderItemRequest{{SKU: "SKU-1", RequestedQty: 2, UnitPrice: 10}},
	}

	tests := []struct {
		name    string
		req     models.CreateOrderRequest
		wantErr string
	}{
		{name: "missing client name", req: func() models.CreateOrderRequest { req := baseReq; req.ClientName = ""; return req }(), wantErr: "validation error"},
		{name: "missing required date", req: func() models.CreateOrderRequest { req := baseReq; req.RequiredDate = time.Time{}; return req }(), wantErr: "validation error"},
		{name: "missing items", req: func() models.CreateOrderRequest { req := baseReq; req.Items = nil; return req }(), wantErr: "validation error"},
		{name: "duplicate sku", req: func() models.CreateOrderRequest {
			req := baseReq
			req.Items = []models.OrderItemRequest{{SKU: "SKU-1", RequestedQty: 2, UnitPrice: 10}, {SKU: "sku-1", RequestedQty: 1, UnitPrice: 5}}
			return req
		}(), wantErr: "validation error"},
		{name: "zero quantity", req: func() models.CreateOrderRequest {
			req := baseReq
			req.Items = []models.OrderItemRequest{{SKU: "SKU-2", RequestedQty: 0, UnitPrice: 10}}
			return req
		}(), wantErr: "validation error"},
		{name: "negative price", req: func() models.CreateOrderRequest {
			req := baseReq
			req.Items = []models.OrderItemRequest{{SKU: "SKU-2", RequestedQty: 1, UnitPrice: -1}}
			return req
		}(), wantErr: "validation error"},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			response, err := svc.CreateOrder(context.Background(), testCase.req)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), testCase.wantErr)
			assert.Nil(t, response)
		})
	}
}

func TestOrderService_CreateOrder_SetsDerivedFields(t *testing.T) {
	db := setupMockDbRepo()
	cache := setupMockCacheRepo()
	svc := newTestServicesWithMocks(db, cache).Orders
	pendingStatus := defaultPendingStatus()
	clientID := uuid.New()

	db.On("GetClientByName", mock.Anything, "Orbit Co").Return(nil, rootrepo.ErrNotFound)
	db.On("CreateClient", mock.Anything, mock.AnythingOfType("*models.Client")).Return(nil)
	db.On("GetOrderStatusByCode", mock.Anything, models.SalesOrderStatusPendingCode).Return(pendingStatus, nil)
	db.On("CreateOrder", mock.Anything, mock.AnythingOfType("*models.SalesOrder")).Return(nil)
	db.On("CreateOrderItem", mock.Anything, mock.AnythingOfType("*models.SalesOrderItem")).Return(nil).Times(2)
	db.On("UpdateClient", mock.Anything, mock.MatchedBy(func(client *models.Client) bool {
		return client.Name == "Orbit Co" && client.TotalLifetimeOrders == 1
	})).Return(nil)
	db.On("GetClient", mock.Anything, mock.Anything).Return(&models.Client{
		ID:                        clientID,
		Name:                      "Orbit Co",
		Tier:                      models.ClientTierB2C,
		DefaultDestinationAddress: "",
		TotalLifetimeOrders:       1,
	}, nil)

	response, err := svc.CreateOrder(context.Background(), models.CreateOrderRequest{
		ClientName:         "Orbit Co",
		DestinationAddress: "",
		RequiredDate:       time.Now().Add(72 * time.Hour).UTC(),
		Items: []models.OrderItemRequest{
			{SKU: "SKU-100", RequestedQty: 3, UnitPrice: 15},
			{SKU: "SKU-200", RequestedQty: 2, UnitPrice: 8},
		},
	})
	require.NoError(t, err)
	require.NotNil(t, response)
	require.NotNil(t, response.Order.Status)
	assert.Equal(t, models.SalesOrderStatusPendingCode, response.Order.Status.Code)
	assert.Equal(t, models.ClientTierB2C, response.Client.Tier)
	assert.Len(t, response.Items, 2)
	assert.InDelta(t, 61.0, response.Order.TotalValue, 0.0001)
	assert.Equal(t, response.Client.DefaultDestinationAddress, response.Order.DestinationAddress)
	db.AssertExpectations(t)
}

func TestOrderService_UpdateAndCancel_RespectLockAndStatus(t *testing.T) {
	db := setupMockDbRepo()
	cache := setupMockCacheRepo()
	svc := newTestServicesWithMocks(db, cache).Orders
	orderID := uuid.New()
	clientID := uuid.New()
	processingStatus := defaultProcessingStatus()
	lockedOrder := &models.SalesOrder{
		ID:                 orderID,
		ClientID:           clientID,
		ClientName:         "Lockstep Ltd",
		DestinationAddress: "Bay 2",
		RequiredDate:       time.Now().Add(24 * time.Hour).UTC(),
		StatusID:           processingStatus.ID,
		Status:             processingStatus,
		Locked:             true,
	}

	db.On("GetOrder", mock.Anything, orderID).Return(lockedOrder, nil).Twice()

	updated, err := svc.UpdateOrder(context.Background(), orderID, models.UpdateOrderRequest{DestinationAddress: ptrString("New Bay")})
	assert.Error(t, err)
	assert.Nil(t, updated)
	assert.Contains(t, err.Error(), "conflict")

	err = svc.CancelOrder(context.Background(), orderID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "conflict")
	db.AssertExpectations(t)
}

func ptrString(value string) *string {
	return &value
}
