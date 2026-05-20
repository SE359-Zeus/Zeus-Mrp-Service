package repository

import (
	"context"
	"errors"
	"sort"
	"strings"
	"sync"
	"time"

	"zeus-sales-service/internal/models"

	"github.com/google/uuid"
)

var ErrNotFound = errors.New("not found")
var ErrInsufficientInventory = errors.New("insufficient inventory")

type MemoryRepository struct {
	mu            sync.RWMutex
	clients       map[uuid.UUID]*models.Client
	clientsByName map[string]uuid.UUID
	orders        map[uuid.UUID]*models.SalesOrder
	orderItems    map[uuid.UUID][]models.SalesOrderItem
	inventory     map[string]int
	reservations  map[uuid.UUID]*models.InventoryReservation
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		clients:       map[uuid.UUID]*models.Client{},
		clientsByName: map[string]uuid.UUID{},
		orders:        map[uuid.UUID]*models.SalesOrder{},
		orderItems:    map[uuid.UUID][]models.SalesOrderItem{},
		inventory:     map[string]int{},
		reservations:  map[uuid.UUID]*models.InventoryReservation{},
	}
}

func (repo *MemoryRepository) SeedInventory(sku string, qty int) {
	repo.mu.Lock()
	defer repo.mu.Unlock()
	repo.inventory[strings.ToUpper(sku)] = qty
}

func (repo *MemoryRepository) SnapshotInventory() map[string]int {
	repo.mu.RLock()
	defer repo.mu.RUnlock()
	snapshot := make(map[string]int, len(repo.inventory))
	for sku, qty := range repo.inventory {
		snapshot[sku] = qty
	}
	return snapshot
}

func (repo *MemoryRepository) CreateClient(ctx context.Context, client *models.Client) error {
	repo.mu.Lock()
	defer repo.mu.Unlock()
	if client.ID == uuid.Nil {
		client.ID = uuid.New()
	}
	now := time.Now().UTC()
	if client.CreatedAt.IsZero() {
		client.CreatedAt = now
	}
	client.UpdatedAt = now
	cloned := *client
	repo.clients[client.ID] = &cloned
	repo.clientsByName[strings.ToLower(client.Name)] = client.ID
	return nil
}

func (repo *MemoryRepository) GetClient(ctx context.Context, id uuid.UUID) (*models.Client, error) {
	repo.mu.RLock()
	defer repo.mu.RUnlock()
	client, ok := repo.clients[id]
	if !ok {
		return nil, ErrNotFound
	}
	cloned := *client
	return &cloned, nil
}

func (repo *MemoryRepository) GetClientByName(ctx context.Context, name string) (*models.Client, error) {
	repo.mu.RLock()
	defer repo.mu.RUnlock()
	id, ok := repo.clientsByName[strings.ToLower(name)]
	if !ok {
		return nil, ErrNotFound
	}
	client := repo.clients[id]
	if client == nil {
		return nil, ErrNotFound
	}
	cloned := *client
	return &cloned, nil
}

func (repo *MemoryRepository) ListClients(ctx context.Context) ([]models.Client, error) {
	repo.mu.RLock()
	defer repo.mu.RUnlock()
	clients := make([]models.Client, 0, len(repo.clients))
	for _, client := range repo.clients {
		clients = append(clients, *client)
	}
	sort.Slice(clients, func(i, j int) bool { return clients[i].Name < clients[j].Name })
	return clients, nil
}

func (repo *MemoryRepository) UpdateClient(ctx context.Context, client *models.Client) error {
	repo.mu.Lock()
	defer repo.mu.Unlock()
	_, ok := repo.clients[client.ID]
	if !ok {
		return ErrNotFound
	}
	client.UpdatedAt = time.Now().UTC()
	cloned := *client
	repo.clients[client.ID] = &cloned
	repo.clientsByName[strings.ToLower(client.Name)] = client.ID
	return nil
}

func (repo *MemoryRepository) CreateOrder(ctx context.Context, order *models.SalesOrder) error {
	repo.mu.Lock()
	defer repo.mu.Unlock()
	if order.ID == uuid.Nil {
		order.ID = uuid.New()
	}
	now := time.Now().UTC()
	if order.CreatedAt.IsZero() {
		order.CreatedAt = now
	}
	order.UpdatedAt = now
	cloned := *order
	repo.orders[order.ID] = &cloned
	return nil
}

func (repo *MemoryRepository) GetOrder(ctx context.Context, id uuid.UUID) (*models.SalesOrder, error) {
	repo.mu.RLock()
	defer repo.mu.RUnlock()
	order, ok := repo.orders[id]
	if !ok {
		return nil, ErrNotFound
	}
	cloned := *order
	return &cloned, nil
}

func (repo *MemoryRepository) ListOrders(ctx context.Context) ([]models.SalesOrder, error) {
	repo.mu.RLock()
	defer repo.mu.RUnlock()
	orders := make([]models.SalesOrder, 0, len(repo.orders))
	for _, order := range repo.orders {
		orders = append(orders, *order)
	}
	sort.Slice(orders, func(i, j int) bool { return orders[i].CreatedAt.Before(orders[j].CreatedAt) })
	return orders, nil
}

func (repo *MemoryRepository) ListPendingOrders(ctx context.Context) ([]models.SalesOrder, error) {
	orders, err := repo.ListOrders(ctx)
	if err != nil {
		return nil, err
	}
	pending := make([]models.SalesOrder, 0)
	for _, order := range orders {
		if order.Status != nil && order.Status.Code == models.SalesOrderStatusPendingCode {
			pending = append(pending, order)
		}
	}
	return pending, nil
}

func (repo *MemoryRepository) UpdateOrder(ctx context.Context, order *models.SalesOrder) error {
	repo.mu.Lock()
	defer repo.mu.Unlock()
	_, ok := repo.orders[order.ID]
	if !ok {
		return ErrNotFound
	}
	order.UpdatedAt = time.Now().UTC()
	cloned := *order
	repo.orders[order.ID] = &cloned
	return nil
}

func (repo *MemoryRepository) CreateOrderItem(ctx context.Context, item *models.SalesOrderItem) error {
	repo.mu.Lock()
	defer repo.mu.Unlock()
	if item.ID == uuid.Nil {
		item.ID = uuid.New()
	}
	if item.CreatedAt.IsZero() {
		item.CreatedAt = time.Now().UTC()
	}
	item.UpdatedAt = item.CreatedAt
	repo.orderItems[item.OrderID] = append(repo.orderItems[item.OrderID], *item)
	return nil
}

func (repo *MemoryRepository) ReplaceOrderItems(ctx context.Context, orderID uuid.UUID, items []models.SalesOrderItem) error {
	repo.mu.Lock()
	defer repo.mu.Unlock()
	cloned := make([]models.SalesOrderItem, len(items))
	copy(cloned, items)
	repo.orderItems[orderID] = cloned
	return nil
}

func (repo *MemoryRepository) GetOrderItems(ctx context.Context, orderID uuid.UUID) ([]models.SalesOrderItem, error) {
	repo.mu.RLock()
	defer repo.mu.RUnlock()
	items := repo.orderItems[orderID]
	cloned := make([]models.SalesOrderItem, len(items))
	copy(cloned, items)
	return cloned, nil
}

func (repo *MemoryRepository) GetATP(ctx context.Context, sku string) (int, error) {
	repo.mu.RLock()
	defer repo.mu.RUnlock()
	return repo.inventory[strings.ToUpper(sku)], nil
}

func (repo *MemoryRepository) ReserveInventory(ctx context.Context, orderID uuid.UUID, items []models.ReservationItem) error {
	repo.mu.Lock()
	defer repo.mu.Unlock()
	for _, item := range items {
		sku := strings.ToUpper(item.SKU)
		if repo.inventory[sku] < item.Quantity {
			return ErrInsufficientInventory
		}
	}
	for _, item := range items {
		sku := strings.ToUpper(item.SKU)
		repo.inventory[sku] -= item.Quantity
	}
	return nil
}

func (repo *MemoryRepository) ReleaseInventory(ctx context.Context, orderID uuid.UUID) error {
	repo.mu.Lock()
	defer repo.mu.Unlock()
	reservation, ok := repo.reservations[orderID]
	if !ok {
		return nil
	}
	for _, item := range reservation.Items {
		repo.inventory[strings.ToUpper(item.SKU)] += item.Quantity
	}
	return nil
}

func (repo *MemoryRepository) CreateReservation(ctx context.Context, reservation *models.InventoryReservation) error {
	repo.mu.Lock()
	defer repo.mu.Unlock()
	if reservation.ID == uuid.Nil {
		reservation.ID = uuid.New()
	}
	if reservation.ReservedAt.IsZero() {
		reservation.ReservedAt = time.Now().UTC()
	}
	cloned := *reservation
	cloned.Items = append([]models.ReservationItem(nil), reservation.Items...)
	repo.reservations[reservation.OrderID] = &cloned
	return nil
}

func (repo *MemoryRepository) GetReservation(ctx context.Context, orderID uuid.UUID) (*models.InventoryReservation, error) {
	repo.mu.RLock()
	defer repo.mu.RUnlock()
	reservation, ok := repo.reservations[orderID]
	if !ok {
		return nil, ErrNotFound
	}
	cloned := *reservation
	cloned.Items = append([]models.ReservationItem(nil), reservation.Items...)
	return &cloned, nil
}

func (repo *MemoryRepository) DeleteReservation(ctx context.Context, orderID uuid.UUID) error {
	repo.mu.Lock()
	defer repo.mu.Unlock()
	delete(repo.reservations, orderID)
	return nil
}
