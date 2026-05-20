package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"zeus-sales-service/internal/middlewares"
	"zeus-sales-service/internal/models"
	"zeus-sales-service/internal/repository"

	"github.com/google/uuid"
)

type OrderService struct {
	repo    repository.DbRepository
	clients *ClientService
}

func NewOrderService(repo repository.DbRepository, clients *ClientService) *OrderService {
	return &OrderService{repo: repo, clients: clients}
}

func (svc *OrderService) CreateOrder(ctx context.Context, req models.CreateOrderRequest) (*models.OrderResponse, error) {
	if strings.TrimSpace(req.ClientName) == "" {
		return nil, fmt.Errorf("%w: client name is required", middlewares.ErrValidation)
	}
	if req.RequiredDate.IsZero() {
		return nil, fmt.Errorf("%w: required date is required", middlewares.ErrValidation)
	}
	if len(req.Items) == 0 {
		return nil, fmt.Errorf("%w: at least one order item is required", middlewares.ErrValidation)
	}
	seen := make(map[string]struct{}, len(req.Items))
	totalValue := 0.0
	for _, item := range req.Items {
		sku := strings.TrimSpace(item.SKU)
		if sku == "" {
			return nil, fmt.Errorf("%w: sku is required", middlewares.ErrValidation)
		}
		if item.RequestedQty <= 0 {
			return nil, fmt.Errorf("%w: requested quantity for %s must be positive", middlewares.ErrValidation, sku)
		}
		if item.UnitPrice < 0 {
			return nil, fmt.Errorf("%w: unit price for %s cannot be negative", middlewares.ErrValidation, sku)
		}
		key := strings.ToUpper(sku)
		if _, exists := seen[key]; exists {
			return nil, fmt.Errorf("%w: duplicate sku %s", middlewares.ErrValidation, sku)
		}
		seen[key] = struct{}{}
		totalValue += float64(item.RequestedQty) * item.UnitPrice
	}
	client, err := svc.clients.ResolveOrCreateClient(ctx, req.ClientName, req.DestinationAddress, req.ClientTier)
	if err != nil {
		return nil, err
	}
	pendingStatus, err := svc.repo.GetOrderStatusByCode(ctx, models.SalesOrderStatusPendingCode)
	if err != nil {
		return nil, err
	}
	order := &models.SalesOrder{
		ID:                 uuid.New(),
		ClientID:           client.ID,
		ClientName:         client.Name,
		DestinationAddress: strings.TrimSpace(req.DestinationAddress),
		RequiredDate:       req.RequiredDate,
		StatusID:           pendingStatus.ID,
		Status:             pendingStatus,
		TotalValue:         totalValue,
		Locked:             false,
		CreatedAt:          time.Now().UTC(),
	}
	if order.DestinationAddress == "" {
		order.DestinationAddress = client.DefaultDestinationAddress
	}
	if err := svc.repo.CreateOrder(ctx, order); err != nil {
		return nil, err
	}
	items := make([]models.SalesOrderItem, 0, len(req.Items))
	for _, item := range req.Items {
		salesItem := models.SalesOrderItem{
			ID:           uuid.New(),
			OrderID:      order.ID,
			SKU:          strings.TrimSpace(item.SKU),
			RequestedQty: item.RequestedQty,
			AllocatedQty: 0,
			UnitPrice:    item.UnitPrice,
			CreatedAt:    time.Now().UTC(),
		}
		items = append(items, salesItem)
		if err := svc.repo.CreateOrderItem(ctx, &salesItem); err != nil {
			return nil, err
		}
	}
	client.TotalLifetimeOrders++
	if err := svc.clients.repo.UpdateClient(ctx, client); err != nil {
		return nil, err
	}
	return svc.buildResponse(ctx, order, items)
}

func (svc *OrderService) GetOrder(ctx context.Context, id uuid.UUID) (*models.OrderResponse, error) {
	if id == uuid.Nil {
		return nil, fmt.Errorf("%w: order id is required", middlewares.ErrValidation)
	}
	order, err := svc.repo.GetOrder(ctx, id)
	if err != nil {
		return nil, err
	}
	items, err := svc.repo.GetOrderItems(ctx, id)
	if err != nil {
		return nil, err
	}
	return svc.buildResponse(ctx, order, items)
}

func (svc *OrderService) ListOrders(ctx context.Context) ([]models.OrderResponse, error) {
	orders, err := svc.repo.ListOrders(ctx)
	if err != nil {
		return nil, err
	}
	responses := make([]models.OrderResponse, 0, len(orders))
	for _, order := range orders {
		items, _ := svc.repo.GetOrderItems(ctx, order.ID)
		client, _ := svc.repo.GetClient(ctx, order.ClientID)
		if client == nil {
			client = &models.Client{}
		}
		responses = append(responses, models.OrderResponse{Order: order, Client: *client, Items: items})
	}
	return responses, nil
}

func (svc *OrderService) ListPendingOrders(ctx context.Context) ([]models.OrderResponse, error) {
	orders, err := svc.repo.ListPendingOrders(ctx)
	if err != nil {
		return nil, err
	}
	responses := make([]models.OrderResponse, 0, len(orders))
	for _, order := range orders {
		items, _ := svc.repo.GetOrderItems(ctx, order.ID)
		client, _ := svc.repo.GetClient(ctx, order.ClientID)
		if client == nil {
			client = &models.Client{}
		}
		responses = append(responses, models.OrderResponse{Order: order, Client: *client, Items: items})
	}
	return responses, nil
}

func (svc *OrderService) UpdateOrder(ctx context.Context, id uuid.UUID, req models.UpdateOrderRequest) (*models.OrderResponse, error) {
	if id == uuid.Nil {
		return nil, fmt.Errorf("%w: order id is required", middlewares.ErrValidation)
	}
	order, err := svc.repo.GetOrder(ctx, id)
	if err != nil {
		return nil, err
	}
	if order.Locked || order.Status == nil || order.Status.Code != models.SalesOrderStatusPendingCode {
		return nil, fmt.Errorf("%w: order is locked or no longer editable", middlewares.ErrConflict)
	}
	if req.DestinationAddress == nil && req.RequiredDate == nil && len(req.Items) == 0 {
		return nil, fmt.Errorf("%w: update request is empty", middlewares.ErrValidation)
	}
	if req.DestinationAddress != nil {
		order.DestinationAddress = strings.TrimSpace(*req.DestinationAddress)
	}
	if req.RequiredDate != nil {
		order.RequiredDate = req.RequiredDate.UTC()
	}
	if len(req.Items) > 0 {
		seen := map[string]struct{}{}
		items := make([]models.SalesOrderItem, 0, len(req.Items))
		total := 0.0
		for _, item := range req.Items {
			sku := strings.TrimSpace(item.SKU)
			if sku == "" {
				return nil, fmt.Errorf("%w: sku is required", middlewares.ErrValidation)
			}
			if item.RequestedQty <= 0 {
				return nil, fmt.Errorf("%w: requested quantity for %s must be positive", middlewares.ErrValidation, sku)
			}
			key := strings.ToUpper(sku)
			if _, exists := seen[key]; exists {
				return nil, fmt.Errorf("%w: duplicate sku %s", middlewares.ErrValidation, sku)
			}
			seen[key] = struct{}{}
			total += float64(item.RequestedQty) * item.UnitPrice
			items = append(items, models.SalesOrderItem{
				ID:           uuid.New(),
				OrderID:      order.ID,
				SKU:          sku,
				RequestedQty: item.RequestedQty,
				UnitPrice:    item.UnitPrice,
				CreatedAt:    time.Now().UTC(),
			})
		}
		order.TotalValue = total
		if err := svc.repo.ReplaceOrderItems(ctx, order.ID, items); err != nil {
			return nil, err
		}
	}
	if err := svc.repo.UpdateOrder(ctx, order); err != nil {
		return nil, err
	}
	items, _ := svc.repo.GetOrderItems(ctx, order.ID)
	client, _ := svc.repo.GetClient(ctx, order.ClientID)
	if client == nil {
		client = &models.Client{}
	}
	return &models.OrderResponse{Order: *order, Client: *client, Items: items}, nil
}

func (svc *OrderService) CancelOrder(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return fmt.Errorf("%w: order id is required", middlewares.ErrValidation)
	}
	order, err := svc.repo.GetOrder(ctx, id)
	if err != nil {
		return err
	}
	if order.Locked || order.Status == nil || order.Status.Code != models.SalesOrderStatusPendingCode {
		return fmt.Errorf("%w: order cannot be cancelled once processing has started", middlewares.ErrConflict)
	}
	cancelledStatus, err := svc.repo.GetOrderStatusByCode(ctx, models.SalesOrderStatusCancelledCode)
	if err != nil {
		return err
	}
	order.StatusID = cancelledStatus.ID
	order.Status = cancelledStatus
	order.UpdatedAt = time.Now().UTC()
	return svc.repo.UpdateOrder(ctx, order)
}

func (svc *OrderService) buildResponse(ctx context.Context, order *models.SalesOrder, items []models.SalesOrderItem) (*models.OrderResponse, error) {
	if order.Status == nil && order.StatusID != uuid.Nil {
		status, err := svc.repo.GetOrderStatusByID(ctx, order.StatusID)
		if err != nil {
			return nil, err
		}
		order.Status = status
	}
	client, err := svc.repo.GetClient(ctx, order.ClientID)
	if err != nil {
		return nil, err
	}
	if client == nil {
		return nil, repository.ErrNotFound
	}
	return &models.OrderResponse{Order: *order, Client: *client, Items: items}, nil
}
