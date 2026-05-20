package sqlite

import (
	"context"
	"errors"
	"strings"
	"time"

	"zeus-sales-service/internal/models"
	rootrepo "zeus-sales-service/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository struct {
	db *gorm.DB
}

func Open(dsn string) (*Repository, error) {
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	repo := &Repository{db: db}
	if err := repo.EnsureSchema(context.Background()); err != nil {
		return nil, err
	}
	return repo, nil
}

func New(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (repo *Repository) Close() error {
	if repo == nil || repo.db == nil {
		return nil
	}
	sqlDB, err := repo.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func (repo *Repository) EnsureSchema(ctx context.Context) error {
	if err := repo.db.WithContext(ctx).AutoMigrate(
		&clientRecord{},
		&salesOrderStatusRecord{},
		&salesOrderRecord{},
		&salesOrderItemRecord{},
		&inventoryReservationRecord{},
		&inventoryReservationItemRecord{},
	); err != nil {
		return err
	}
	return repo.seedStatuses(ctx)
}

func (repo *Repository) seedStatuses(ctx context.Context) error {
	statuses := defaultStatuses()
	records := make([]salesOrderStatusRecord, 0, len(statuses))
	now := time.Now().UTC()
	for _, status := range statuses {
		records = append(records, salesOrderStatusRecord{
			ID:         status.ID.String(),
			Code:       status.Code,
			Label:      status.Label,
			SortOrder:  status.SortOrder,
			IsTerminal: status.IsTerminal,
			CreatedAt:  now,
			UpdatedAt:  now,
		})
	}
	return repo.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoNothing: true,
	}).Create(&records).Error
}

func defaultStatuses() []models.SalesOrderStatusLUT {
	return []models.SalesOrderStatusLUT{
		{ID: salesStatusID(models.SalesOrderStatusPendingCode), Code: models.SalesOrderStatusPendingCode, Label: "Pending", SortOrder: 1, IsTerminal: false},
		{ID: salesStatusID(models.SalesOrderStatusProcessingCode), Code: models.SalesOrderStatusProcessingCode, Label: "Processing", SortOrder: 2, IsTerminal: false},
		{ID: salesStatusID(models.SalesOrderStatusDeliveringCode), Code: models.SalesOrderStatusDeliveringCode, Label: "Delivering", SortOrder: 3, IsTerminal: false},
		{ID: salesStatusID(models.SalesOrderStatusCompletedCode), Code: models.SalesOrderStatusCompletedCode, Label: "Completed", SortOrder: 4, IsTerminal: true},
		{ID: salesStatusID(models.SalesOrderStatusCancelledCode), Code: models.SalesOrderStatusCancelledCode, Label: "Cancelled", SortOrder: 5, IsTerminal: true},
	}
}

func salesStatusID(code string) uuid.UUID {
	return uuid.NewSHA1(uuid.NameSpaceURL, []byte("sales-order-status:"+code))
}

func (repo *Repository) CreateClient(ctx context.Context, client *models.Client) error {
	if client.ID == uuid.Nil {
		client.ID = uuid.New()
	}
	now := time.Now().UTC()
	if client.CreatedAt.IsZero() {
		client.CreatedAt = now
	}
	client.UpdatedAt = now
	return repo.db.WithContext(ctx).Create(clientRecordFromModel(client)).Error
}

func (repo *Repository) GetClient(ctx context.Context, id uuid.UUID) (*models.Client, error) {
	var record clientRecord
	if err := repo.db.WithContext(ctx).First(&record, "id = ?", id.String()).Error; err != nil {
		return nil, mapRecordError(err)
	}
	model := record.toModel()
	return &model, nil
}

func (repo *Repository) GetClientByName(ctx context.Context, name string) (*models.Client, error) {
	var record clientRecord
	if err := repo.db.WithContext(ctx).Where("lower(name) = lower(?)", strings.TrimSpace(name)).First(&record).Error; err != nil {
		return nil, mapRecordError(err)
	}
	model := record.toModel()
	return &model, nil
}

func (repo *Repository) ListClients(ctx context.Context) ([]models.Client, error) {
	var records []clientRecord
	if err := repo.db.WithContext(ctx).Order("name ASC").Find(&records).Error; err != nil {
		return nil, err
	}
	clients := make([]models.Client, 0, len(records))
	for _, record := range records {
		clients = append(clients, record.toModel())
	}
	return clients, nil
}

func (repo *Repository) UpdateClient(ctx context.Context, client *models.Client) error {
	client.UpdatedAt = time.Now().UTC()
	result := repo.db.WithContext(ctx).Model(&clientRecord{}).Where("id = ?", client.ID.String()).Updates(map[string]any{
		"name":                        client.Name,
		"tier":                        string(client.Tier),
		"default_destination_address": client.DefaultDestinationAddress,
		"total_lifetime_orders":       client.TotalLifetimeOrders,
		"updated_at":                  client.UpdatedAt,
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return rootrepo.ErrNotFound
	}
	return nil
}

func (repo *Repository) ListOrderStatuses(ctx context.Context) ([]models.SalesOrderStatusLUT, error) {
	var records []salesOrderStatusRecord
	if err := repo.db.WithContext(ctx).Order("sort_order ASC").Find(&records).Error; err != nil {
		return nil, err
	}
	statuses := make([]models.SalesOrderStatusLUT, 0, len(records))
	for _, record := range records {
		statuses = append(statuses, record.toModel())
	}
	return statuses, nil
}

func (repo *Repository) GetOrderStatusByID(ctx context.Context, id uuid.UUID) (*models.SalesOrderStatusLUT, error) {
	var record salesOrderStatusRecord
	if err := repo.db.WithContext(ctx).First(&record, "id = ?", id.String()).Error; err != nil {
		return nil, mapRecordError(err)
	}
	model := record.toModel()
	return &model, nil
}

func (repo *Repository) GetOrderStatusByCode(ctx context.Context, code string) (*models.SalesOrderStatusLUT, error) {
	var record salesOrderStatusRecord
	if err := repo.db.WithContext(ctx).Where("code = ?", strings.ToUpper(strings.TrimSpace(code))).First(&record).Error; err != nil {
		return nil, mapRecordError(err)
	}
	model := record.toModel()
	return &model, nil
}

func (repo *Repository) CreateOrder(ctx context.Context, order *models.SalesOrder) error {
	if order.ID == uuid.Nil {
		order.ID = uuid.New()
	}
	now := time.Now().UTC()
	if order.CreatedAt.IsZero() {
		order.CreatedAt = now
	}
	order.UpdatedAt = now
	return repo.db.WithContext(ctx).Create(orderRecordFromModel(order)).Error
}

func (repo *Repository) GetOrder(ctx context.Context, id uuid.UUID) (*models.SalesOrder, error) {
	var record salesOrderRecord
	if err := repo.db.WithContext(ctx).Preload("Status").First(&record, "id = ?", id.String()).Error; err != nil {
		return nil, mapRecordError(err)
	}
	model := record.toModel()
	return &model, nil
}

func (repo *Repository) ListOrders(ctx context.Context) ([]models.SalesOrder, error) {
	var records []salesOrderRecord
	if err := repo.db.WithContext(ctx).Preload("Status").Order("created_at ASC").Find(&records).Error; err != nil {
		return nil, err
	}
	orders := make([]models.SalesOrder, 0, len(records))
	for _, record := range records {
		orders = append(orders, record.toModel())
	}
	return orders, nil
}

func (repo *Repository) ListPendingOrders(ctx context.Context) ([]models.SalesOrder, error) {
	pendingStatus, err := repo.GetOrderStatusByCode(ctx, models.SalesOrderStatusPendingCode)
	if err != nil {
		return nil, err
	}
	var records []salesOrderRecord
	if err := repo.db.WithContext(ctx).Preload("Status").Where("status_id = ?", pendingStatus.ID.String()).Order("created_at ASC").Find(&records).Error; err != nil {
		return nil, err
	}
	orders := make([]models.SalesOrder, 0, len(records))
	for _, record := range records {
		orders = append(orders, record.toModel())
	}
	return orders, nil
}

func (repo *Repository) UpdateOrder(ctx context.Context, order *models.SalesOrder) error {
	order.UpdatedAt = time.Now().UTC()
	result := repo.db.WithContext(ctx).Model(&salesOrderRecord{}).Where("id = ?", order.ID.String()).Updates(map[string]any{
		"client_id":           order.ClientID.String(),
		"client_name":         order.ClientName,
		"destination_address": order.DestinationAddress,
		"required_date":       order.RequiredDate,
		"status_id":           order.StatusID.String(),
		"total_value":         order.TotalValue,
		"locked":              order.Locked,
		"updated_at":          order.UpdatedAt,
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return rootrepo.ErrNotFound
	}
	return nil
}

func (repo *Repository) CreateOrderItem(ctx context.Context, item *models.SalesOrderItem) error {
	if item.ID == uuid.Nil {
		item.ID = uuid.New()
	}
	now := time.Now().UTC()
	if item.CreatedAt.IsZero() {
		item.CreatedAt = now
	}
	item.UpdatedAt = now
	return repo.db.WithContext(ctx).Create(orderItemRecordFromModel(item)).Error
}

func (repo *Repository) GetOrderItems(ctx context.Context, orderID uuid.UUID) ([]models.SalesOrderItem, error) {
	var records []salesOrderItemRecord
	if err := repo.db.WithContext(ctx).Where("order_id = ?", orderID.String()).Order("created_at ASC").Find(&records).Error; err != nil {
		return nil, err
	}
	items := make([]models.SalesOrderItem, 0, len(records))
	for _, record := range records {
		items = append(items, record.toModel())
	}
	return items, nil
}

func (repo *Repository) ReplaceOrderItems(ctx context.Context, orderID uuid.UUID, items []models.SalesOrderItem) error {
	return repo.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("order_id = ?", orderID.String()).Delete(&salesOrderItemRecord{}).Error; err != nil {
			return err
		}
		for _, item := range items {
			if item.ID == uuid.Nil {
				item.ID = uuid.New()
			}
			if item.CreatedAt.IsZero() {
				item.CreatedAt = time.Now().UTC()
			}
			item.UpdatedAt = item.CreatedAt
			record := orderItemRecordFromModel(&item)
			record.OrderID = orderID.String()
			if err := tx.Create(record).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (repo *Repository) CreateReservation(ctx context.Context, reservation *models.InventoryReservation) error {
	if reservation.ID == uuid.Nil {
		reservation.ID = uuid.New()
	}
	if reservation.ReservedAt.IsZero() {
		reservation.ReservedAt = time.Now().UTC()
	}
	return repo.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		record := reservationRecordFromModel(reservation)
		if err := tx.Create(record).Error; err != nil {
			return err
		}
		for _, item := range reservation.Items {
			reservationItem := inventoryReservationItemRecord{
				ID:            uuid.New().String(),
				ReservationID: reservation.ID.String(),
				SKU:           item.SKU,
				Quantity:      item.Quantity,
			}
			if err := tx.Create(&reservationItem).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (repo *Repository) GetReservation(ctx context.Context, orderID uuid.UUID) (*models.InventoryReservation, error) {
	var record inventoryReservationRecord
	if err := repo.db.WithContext(ctx).Preload("Items").Where("order_id = ?", orderID.String()).First(&record).Error; err != nil {
		return nil, mapRecordError(err)
	}
	model := record.toModel()
	return &model, nil
}

func (repo *Repository) DeleteReservation(ctx context.Context, orderID uuid.UUID) error {
	return repo.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var reservation inventoryReservationRecord
		if err := tx.Where("order_id = ?", orderID.String()).First(&reservation).Error; err != nil {
			return mapRecordError(err)
		}
		if err := tx.Where("reservation_id = ?", reservation.ID).Delete(&inventoryReservationItemRecord{}).Error; err != nil {
			return err
		}
		if err := tx.Delete(&reservation).Error; err != nil {
			return err
		}
		return nil
	})
}

func mapRecordError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return rootrepo.ErrNotFound
	}
	return err
}

type clientRecord struct {
	ID                        string    `gorm:"primaryKey;column:id"`
	Name                      string    `gorm:"column:name;uniqueIndex"`
	Tier                      string    `gorm:"column:tier"`
	DefaultDestinationAddress string    `gorm:"column:default_destination_address"`
	TotalLifetimeOrders       int       `gorm:"column:total_lifetime_orders"`
	CreatedAt                 time.Time `gorm:"column:created_at"`
	UpdatedAt                 time.Time `gorm:"column:updated_at"`
}

func (clientRecord) TableName() string { return "clients" }

func clientRecordFromModel(client *models.Client) *clientRecord {
	return &clientRecord{
		ID:                        client.ID.String(),
		Name:                      client.Name,
		Tier:                      string(client.Tier),
		DefaultDestinationAddress: client.DefaultDestinationAddress,
		TotalLifetimeOrders:       client.TotalLifetimeOrders,
		CreatedAt:                 client.CreatedAt,
		UpdatedAt:                 client.UpdatedAt,
	}
}

func (record clientRecord) toModel() models.Client {
	parsedID, _ := uuid.Parse(record.ID)
	return models.Client{
		ID:                        parsedID,
		Name:                      record.Name,
		Tier:                      models.ClientTier(record.Tier),
		DefaultDestinationAddress: record.DefaultDestinationAddress,
		TotalLifetimeOrders:       record.TotalLifetimeOrders,
		CreatedAt:                 record.CreatedAt,
		UpdatedAt:                 record.UpdatedAt,
	}
}

type salesOrderStatusRecord struct {
	ID         string    `gorm:"primaryKey;column:id"`
	Code       string    `gorm:"column:code;uniqueIndex"`
	Label      string    `gorm:"column:label"`
	SortOrder  int       `gorm:"column:sort_order"`
	IsTerminal bool      `gorm:"column:is_terminal"`
	CreatedAt  time.Time `gorm:"column:created_at"`
	UpdatedAt  time.Time `gorm:"column:updated_at"`
}

func (salesOrderStatusRecord) TableName() string { return "sales_order_status_lut" }

func (record salesOrderStatusRecord) toModel() models.SalesOrderStatusLUT {
	parsedID, _ := uuid.Parse(record.ID)
	return models.SalesOrderStatusLUT{
		ID:         parsedID,
		Code:       record.Code,
		Label:      record.Label,
		SortOrder:  record.SortOrder,
		IsTerminal: record.IsTerminal,
		CreatedAt:  record.CreatedAt,
		UpdatedAt:  record.UpdatedAt,
	}
}

type salesOrderRecord struct {
	ID                 string                      `gorm:"primaryKey;column:id"`
	ClientID           string                      `gorm:"column:client_id;index"`
	ClientName         string                      `gorm:"column:client_name"`
	DestinationAddress string                      `gorm:"column:destination_address"`
	RequiredDate       time.Time                   `gorm:"column:required_date"`
	StatusID           string                      `gorm:"column:status_id;index"`
	Status             salesOrderStatusRecord      `gorm:"foreignKey:StatusID;references:ID"`
	TotalValue         float64                     `gorm:"column:total_value"`
	Locked             bool                        `gorm:"column:locked"`
	CreatedAt          time.Time                   `gorm:"column:created_at"`
	UpdatedAt          time.Time                   `gorm:"column:updated_at"`
	Items              []salesOrderItemRecord      `gorm:"foreignKey:OrderID;references:ID"`
	Reservation        *inventoryReservationRecord `gorm:"foreignKey:OrderID;references:ID"`
}

func (salesOrderRecord) TableName() string { return "sales_orders" }

func orderRecordFromModel(order *models.SalesOrder) *salesOrderRecord {
	record := &salesOrderRecord{
		ID:                 order.ID.String(),
		ClientID:           order.ClientID.String(),
		ClientName:         order.ClientName,
		DestinationAddress: order.DestinationAddress,
		RequiredDate:       order.RequiredDate,
		StatusID:           order.StatusID.String(),
		TotalValue:         order.TotalValue,
		Locked:             order.Locked,
		CreatedAt:          order.CreatedAt,
		UpdatedAt:          order.UpdatedAt,
	}
	if order.Status != nil {
		record.Status = salesOrderStatusRecord{
			ID:         order.Status.ID.String(),
			Code:       order.Status.Code,
			Label:      order.Status.Label,
			SortOrder:  order.Status.SortOrder,
			IsTerminal: order.Status.IsTerminal,
			CreatedAt:  order.Status.CreatedAt,
			UpdatedAt:  order.Status.UpdatedAt,
		}
	}
	return record
}

func (record salesOrderRecord) toModel() models.SalesOrder {
	parsedID, _ := uuid.Parse(record.ID)
	clientID, _ := uuid.Parse(record.ClientID)
	statusID, _ := uuid.Parse(record.StatusID)
	var status *models.SalesOrderStatusLUT
	if record.Status.ID != "" {
		statusModel := record.Status.toModel()
		status = &statusModel
	}
	return models.SalesOrder{
		ID:                 parsedID,
		ClientID:           clientID,
		ClientName:         record.ClientName,
		DestinationAddress: record.DestinationAddress,
		RequiredDate:       record.RequiredDate,
		StatusID:           statusID,
		Status:             status,
		TotalValue:         record.TotalValue,
		Locked:             record.Locked,
		CreatedAt:          record.CreatedAt,
		UpdatedAt:          record.UpdatedAt,
	}
}

type salesOrderItemRecord struct {
	ID           string    `gorm:"primaryKey;column:id"`
	OrderID      string    `gorm:"column:order_id;index"`
	SKU          string    `gorm:"column:sku;index"`
	RequestedQty int       `gorm:"column:requested_qty"`
	AllocatedQty int       `gorm:"column:allocated_qty"`
	UnitPrice    float64   `gorm:"column:unit_price"`
	CreatedAt    time.Time `gorm:"column:created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at"`
}

func (salesOrderItemRecord) TableName() string { return "sales_order_items" }

func orderItemRecordFromModel(item *models.SalesOrderItem) *salesOrderItemRecord {
	return &salesOrderItemRecord{
		ID:           item.ID.String(),
		OrderID:      item.OrderID.String(),
		SKU:          item.SKU,
		RequestedQty: item.RequestedQty,
		AllocatedQty: item.AllocatedQty,
		UnitPrice:    item.UnitPrice,
		CreatedAt:    item.CreatedAt,
		UpdatedAt:    item.UpdatedAt,
	}
}

func (record salesOrderItemRecord) toModel() models.SalesOrderItem {
	parsedID, _ := uuid.Parse(record.ID)
	orderID, _ := uuid.Parse(record.OrderID)
	return models.SalesOrderItem{
		ID:           parsedID,
		OrderID:      orderID,
		SKU:          record.SKU,
		RequestedQty: record.RequestedQty,
		AllocatedQty: record.AllocatedQty,
		UnitPrice:    record.UnitPrice,
		CreatedAt:    record.CreatedAt,
		UpdatedAt:    record.UpdatedAt,
	}
}

type inventoryReservationRecord struct {
	ID         string                           `gorm:"primaryKey;column:id"`
	OrderID    string                           `gorm:"column:order_id;uniqueIndex"`
	ReservedAt time.Time                        `gorm:"column:reserved_at"`
	Items      []inventoryReservationItemRecord `gorm:"foreignKey:ReservationID;references:ID"`
}

func (inventoryReservationRecord) TableName() string { return "inventory_reservations" }

func reservationRecordFromModel(reservation *models.InventoryReservation) *inventoryReservationRecord {
	return &inventoryReservationRecord{
		ID:         reservation.ID.String(),
		OrderID:    reservation.OrderID.String(),
		ReservedAt: reservation.ReservedAt,
	}
}

func (record inventoryReservationRecord) toModel() models.InventoryReservation {
	parsedID, _ := uuid.Parse(record.ID)
	orderID, _ := uuid.Parse(record.OrderID)
	items := make([]models.ReservationItem, 0, len(record.Items))
	for _, item := range record.Items {
		items = append(items, models.ReservationItem{SKU: item.SKU, Quantity: item.Quantity})
	}
	return models.InventoryReservation{ID: parsedID, OrderID: orderID, Items: items, ReservedAt: record.ReservedAt}
}

type inventoryReservationItemRecord struct {
	ID            string `gorm:"primaryKey;column:id"`
	ReservationID string `gorm:"column:reservation_id;index"`
	SKU           string `gorm:"column:sku"`
	Quantity      int    `gorm:"column:quantity"`
}

func (inventoryReservationItemRecord) TableName() string { return "inventory_reservation_items" }
