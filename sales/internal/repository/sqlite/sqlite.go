package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"zeus-sales-service/internal/models"
	rootrepo "zeus-sales-service/internal/repository"

	"github.com/google/uuid"
	_ "modernc.org/sqlite"
)

type Repository struct {
	db *sql.DB
}

func (repo *Repository) Close() error {
	if repo == nil || repo.db == nil {
		return nil
	}
	return repo.db.Close()
}

func Open(dsn string) (*Repository, error) {
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	db.SetConnMaxLifetime(0)
	repo := &Repository{db: db}
	if err := repo.EnsureSchema(context.Background()); err != nil {
		_ = db.Close()
		return nil, err
	}
	return repo, nil
}

func New(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (repo *Repository) EnsureSchema(ctx context.Context) error {
	statements := []string{
		`PRAGMA foreign_keys = ON`,
		`CREATE TABLE IF NOT EXISTS sales_order_status_lut (
			id TEXT PRIMARY KEY,
			code TEXT NOT NULL UNIQUE,
			label TEXT NOT NULL,
			sort_order INTEGER NOT NULL,
			is_terminal INTEGER NOT NULL DEFAULT 0,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS clients (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
			tier TEXT NOT NULL CHECK (tier IN ('B2B', 'B2C')),
			default_destination_address TEXT NOT NULL DEFAULT '',
			total_lifetime_orders INTEGER NOT NULL DEFAULT 0,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS sales_orders (
			id TEXT PRIMARY KEY,
			client_id TEXT NOT NULL,
			client_name TEXT NOT NULL,
			destination_address TEXT NOT NULL DEFAULT '',
			required_date TEXT NOT NULL,
			status_id TEXT NOT NULL,
			total_value REAL NOT NULL DEFAULT 0,
			locked INTEGER NOT NULL DEFAULT 0,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL,
			FOREIGN KEY (client_id) REFERENCES clients(id) ON UPDATE CASCADE ON DELETE RESTRICT,
			FOREIGN KEY (status_id) REFERENCES sales_order_status_lut(id) ON UPDATE CASCADE ON DELETE RESTRICT
		)`,
		`CREATE INDEX IF NOT EXISTS idx_sales_orders_client_id ON sales_orders(client_id)`,
		`CREATE INDEX IF NOT EXISTS idx_sales_orders_status_id ON sales_orders(status_id)`,
		`CREATE INDEX IF NOT EXISTS idx_sales_orders_created_at ON sales_orders(created_at)`,
		`CREATE TABLE IF NOT EXISTS sales_order_items (
			id TEXT PRIMARY KEY,
			order_id TEXT NOT NULL,
			sku TEXT NOT NULL,
			requested_qty INTEGER NOT NULL,
			allocated_qty INTEGER NOT NULL DEFAULT 0,
			unit_price REAL NOT NULL DEFAULT 0,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL,
			FOREIGN KEY (order_id) REFERENCES sales_orders(id) ON UPDATE CASCADE ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_sales_order_items_order_id ON sales_order_items(order_id)`,
		`CREATE TABLE IF NOT EXISTS inventory_reservations (
			id TEXT PRIMARY KEY,
			order_id TEXT NOT NULL UNIQUE,
			reserved_at TEXT NOT NULL,
			FOREIGN KEY (order_id) REFERENCES sales_orders(id) ON UPDATE CASCADE ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS inventory_reservation_items (
			id TEXT PRIMARY KEY,
			reservation_id TEXT NOT NULL,
			sku TEXT NOT NULL,
			quantity INTEGER NOT NULL,
			FOREIGN KEY (reservation_id) REFERENCES inventory_reservations(id) ON UPDATE CASCADE ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_inventory_reservation_items_reservation_id ON inventory_reservation_items(reservation_id)`,
	}
	for _, statement := range statements {
		if _, err := repo.db.ExecContext(ctx, statement); err != nil {
			return err
		}
	}
	return repo.seedStatuses(ctx)
}

func (repo *Repository) seedStatuses(ctx context.Context) error {
	for _, status := range defaultStatuses() {
		_, err := repo.db.ExecContext(ctx, `INSERT OR IGNORE INTO sales_order_status_lut (id, code, label, sort_order, is_terminal, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
			status.ID.String(), status.Code, status.Label, status.SortOrder, boolInt(status.IsTerminal), formatTime(time.Now().UTC()), formatTime(time.Now().UTC()))
		if err != nil {
			return err
		}
	}
	return nil
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

func formatTime(value time.Time) string {
	return value.UTC().Format(time.RFC3339Nano)
}

func parseTime(value string) (time.Time, error) {
	return time.Parse(time.RFC3339Nano, value)
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func intBool(value int64) bool {
	return value != 0
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
	_, err := repo.db.ExecContext(ctx, `INSERT INTO clients (id, name, tier, default_destination_address, total_lifetime_orders, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		client.ID.String(), client.Name, string(client.Tier), client.DefaultDestinationAddress, client.TotalLifetimeOrders, formatTime(client.CreatedAt), formatTime(client.UpdatedAt))
	return err
}

func (repo *Repository) GetClient(ctx context.Context, id uuid.UUID) (*models.Client, error) {
	row := repo.db.QueryRowContext(ctx, `SELECT id, name, tier, default_destination_address, total_lifetime_orders, created_at, updated_at FROM clients WHERE id = ?`, id.String())
	client, err := scanClient(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, rootrepo.ErrNotFound
		}
		return nil, err
	}
	return client, nil
}

func (repo *Repository) GetClientByName(ctx context.Context, name string) (*models.Client, error) {
	row := repo.db.QueryRowContext(ctx, `SELECT id, name, tier, default_destination_address, total_lifetime_orders, created_at, updated_at FROM clients WHERE lower(name) = lower(?)`, strings.TrimSpace(name))
	client, err := scanClient(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, rootrepo.ErrNotFound
		}
		return nil, err
	}
	return client, nil
}

func (repo *Repository) ListClients(ctx context.Context) ([]models.Client, error) {
	rows, err := repo.db.QueryContext(ctx, `SELECT id, name, tier, default_destination_address, total_lifetime_orders, created_at, updated_at FROM clients ORDER BY name ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	clients := make([]models.Client, 0)
	for rows.Next() {
		client, err := scanClient(rows)
		if err != nil {
			return nil, err
		}
		clients = append(clients, *client)
	}
	return clients, rows.Err()
}

func (repo *Repository) UpdateClient(ctx context.Context, client *models.Client) error {
	client.UpdatedAt = time.Now().UTC()
	result, err := repo.db.ExecContext(ctx, `UPDATE clients SET name = ?, tier = ?, default_destination_address = ?, total_lifetime_orders = ?, updated_at = ? WHERE id = ?`,
		client.Name, string(client.Tier), client.DefaultDestinationAddress, client.TotalLifetimeOrders, formatTime(client.UpdatedAt), client.ID.String())
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return rootrepo.ErrNotFound
	}
	return nil
}

func (repo *Repository) ListOrderStatuses(ctx context.Context) ([]models.SalesOrderStatusLUT, error) {
	rows, err := repo.db.QueryContext(ctx, `SELECT id, code, label, sort_order, is_terminal, created_at, updated_at FROM sales_order_status_lut ORDER BY sort_order ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	statuses := make([]models.SalesOrderStatusLUT, 0)
	for rows.Next() {
		status, err := scanStatus(rows)
		if err != nil {
			return nil, err
		}
		statuses = append(statuses, *status)
	}
	return statuses, rows.Err()
}

func (repo *Repository) GetOrderStatusByID(ctx context.Context, id uuid.UUID) (*models.SalesOrderStatusLUT, error) {
	row := repo.db.QueryRowContext(ctx, `SELECT id, code, label, sort_order, is_terminal, created_at, updated_at FROM sales_order_status_lut WHERE id = ?`, id.String())
	status, err := scanStatus(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, rootrepo.ErrNotFound
		}
		return nil, err
	}
	return status, nil
}

func (repo *Repository) GetOrderStatusByCode(ctx context.Context, code string) (*models.SalesOrderStatusLUT, error) {
	row := repo.db.QueryRowContext(ctx, `SELECT id, code, label, sort_order, is_terminal, created_at, updated_at FROM sales_order_status_lut WHERE code = ?`, strings.ToUpper(strings.TrimSpace(code)))
	status, err := scanStatus(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, rootrepo.ErrNotFound
		}
		return nil, err
	}
	return status, nil
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
	_, err := repo.db.ExecContext(ctx, `INSERT INTO sales_orders (id, client_id, client_name, destination_address, required_date, status_id, total_value, locked, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		order.ID.String(), order.ClientID.String(), order.ClientName, order.DestinationAddress, formatTime(order.RequiredDate), order.StatusID.String(), order.TotalValue, boolInt(order.Locked), formatTime(order.CreatedAt), formatTime(order.UpdatedAt))
	return err
}

func (repo *Repository) GetOrder(ctx context.Context, id uuid.UUID) (*models.SalesOrder, error) {
	row := repo.db.QueryRowContext(ctx, orderSelectByIDQuery, id.String())
	order, err := scanOrder(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, rootrepo.ErrNotFound
		}
		return nil, err
	}
	return order, nil
}

func (repo *Repository) ListOrders(ctx context.Context) ([]models.SalesOrder, error) {
	rows, err := repo.db.QueryContext(ctx, orderSelectBaseQuery+` ORDER BY o.created_at ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	orders := make([]models.SalesOrder, 0)
	for rows.Next() {
		order, err := scanOrder(rows)
		if err != nil {
			return nil, err
		}
		orders = append(orders, *order)
	}
	return orders, rows.Err()
}

func (repo *Repository) ListPendingOrders(ctx context.Context) ([]models.SalesOrder, error) {
	rows, err := repo.db.QueryContext(ctx, orderSelectBaseQuery+` WHERE s.code = ? ORDER BY o.created_at ASC`, models.SalesOrderStatusPendingCode)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	orders := make([]models.SalesOrder, 0)
	for rows.Next() {
		order, err := scanOrder(rows)
		if err != nil {
			return nil, err
		}
		orders = append(orders, *order)
	}
	return orders, rows.Err()
}

func (repo *Repository) UpdateOrder(ctx context.Context, order *models.SalesOrder) error {
	order.UpdatedAt = time.Now().UTC()
	result, err := repo.db.ExecContext(ctx, `UPDATE sales_orders SET client_id = ?, client_name = ?, destination_address = ?, required_date = ?, status_id = ?, total_value = ?, locked = ?, updated_at = ? WHERE id = ?`,
		order.ClientID.String(), order.ClientName, order.DestinationAddress, formatTime(order.RequiredDate), order.StatusID.String(), order.TotalValue, boolInt(order.Locked), formatTime(order.UpdatedAt), order.ID.String())
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
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
	_, err := repo.db.ExecContext(ctx, `INSERT INTO sales_order_items (id, order_id, sku, requested_qty, allocated_qty, unit_price, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		item.ID.String(), item.OrderID.String(), item.SKU, item.RequestedQty, item.AllocatedQty, item.UnitPrice, formatTime(item.CreatedAt), formatTime(item.UpdatedAt))
	return err
}

func (repo *Repository) GetOrderItems(ctx context.Context, orderID uuid.UUID) ([]models.SalesOrderItem, error) {
	rows, err := repo.db.QueryContext(ctx, `SELECT id, order_id, sku, requested_qty, allocated_qty, unit_price, created_at, updated_at FROM sales_order_items WHERE order_id = ? ORDER BY created_at ASC`, orderID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]models.SalesOrderItem, 0)
	for rows.Next() {
		item, err := scanOrderItem(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *item)
	}
	return items, rows.Err()
}

func (repo *Repository) ReplaceOrderItems(ctx context.Context, orderID uuid.UUID, items []models.SalesOrderItem) error {
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, `DELETE FROM sales_order_items WHERE order_id = ?`, orderID.String()); err != nil {
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
		if _, err := tx.ExecContext(ctx, `INSERT INTO sales_order_items (id, order_id, sku, requested_qty, allocated_qty, unit_price, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			item.ID.String(), orderID.String(), item.SKU, item.RequestedQty, item.AllocatedQty, item.UnitPrice, formatTime(item.CreatedAt), formatTime(item.UpdatedAt)); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (repo *Repository) CreateReservation(ctx context.Context, reservation *models.InventoryReservation) error {
	if reservation.ID == uuid.Nil {
		reservation.ID = uuid.New()
	}
	if reservation.ReservedAt.IsZero() {
		reservation.ReservedAt = time.Now().UTC()
	}
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, `INSERT INTO inventory_reservations (id, order_id, reserved_at) VALUES (?, ?, ?)`, reservation.ID.String(), reservation.OrderID.String(), formatTime(reservation.ReservedAt)); err != nil {
		return err
	}
	for _, item := range reservation.Items {
		if _, err := tx.ExecContext(ctx, `INSERT INTO inventory_reservation_items (id, reservation_id, sku, quantity) VALUES (?, ?, ?, ?)`, uuid.New().String(), reservation.ID.String(), item.SKU, item.Quantity); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (repo *Repository) GetReservation(ctx context.Context, orderID uuid.UUID) (*models.InventoryReservation, error) {
	row := repo.db.QueryRowContext(ctx, `SELECT id, order_id, reserved_at FROM inventory_reservations WHERE order_id = ?`, orderID.String())
	var reservationIDText, orderIDText, reservedAtText string
	if err := row.Scan(&reservationIDText, &orderIDText, &reservedAtText); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, rootrepo.ErrNotFound
		}
		return nil, err
	}
	rows, err := repo.db.QueryContext(ctx, `SELECT sku, quantity FROM inventory_reservation_items WHERE reservation_id = ?`, reservationIDText)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]models.ReservationItem, 0)
	for rows.Next() {
		var sku string
		var quantity int
		if err := rows.Scan(&sku, &quantity); err != nil {
			return nil, err
		}
		items = append(items, models.ReservationItem{SKU: sku, Quantity: quantity})
	}
	reservedAt, err := parseTime(reservedAtText)
	if err != nil {
		return nil, err
	}
	return &models.InventoryReservation{ID: uuid.MustParse(reservationIDText), OrderID: uuid.MustParse(orderIDText), Items: items, ReservedAt: reservedAt}, rows.Err()
}

func (repo *Repository) DeleteReservation(ctx context.Context, orderID uuid.UUID) error {
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	row := tx.QueryRowContext(ctx, `SELECT id FROM inventory_reservations WHERE order_id = ?`, orderID.String())
	var reservationID string
	if err := row.Scan(&reservationID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return rootrepo.ErrNotFound
		}
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM inventory_reservation_items WHERE reservation_id = ?`, reservationID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM inventory_reservations WHERE id = ?`, reservationID); err != nil {
		return err
	}
	return tx.Commit()
}

const orderSelectBaseQuery = `SELECT
	o.id, o.client_id, o.client_name, o.destination_address, o.required_date, o.status_id, o.total_value, o.locked, o.created_at, o.updated_at,
	s.id, s.code, s.label, s.sort_order, s.is_terminal, s.created_at, s.updated_at
FROM sales_orders o
JOIN sales_order_status_lut s ON s.id = o.status_id`

const orderSelectByIDQuery = orderSelectBaseQuery + ` WHERE o.id = ?`

func scanClient(scanner interface{ Scan(dest ...any) error }) (*models.Client, error) {
	var idText string
	var name string
	var tier string
	var defaultDestinationAddress string
	var totalLifetimeOrders int
	var createdAtText string
	var updatedAtText string
	if err := scanner.Scan(&idText, &name, &tier, &defaultDestinationAddress, &totalLifetimeOrders, &createdAtText, &updatedAtText); err != nil {
		return nil, err
	}
	createdAt, err := parseTime(createdAtText)
	if err != nil {
		return nil, err
	}
	updatedAt, err := parseTime(updatedAtText)
	if err != nil {
		return nil, err
	}
	return &models.Client{ID: uuid.MustParse(idText), Name: name, Tier: models.ClientTier(tier), DefaultDestinationAddress: defaultDestinationAddress, TotalLifetimeOrders: totalLifetimeOrders, CreatedAt: createdAt, UpdatedAt: updatedAt}, nil
}

func scanStatus(scanner interface{ Scan(dest ...any) error }) (*models.SalesOrderStatusLUT, error) {
	var idText string
	var code string
	var label string
	var sortOrder int
	var isTerminal int
	var createdAtText string
	var updatedAtText string
	if err := scanner.Scan(&idText, &code, &label, &sortOrder, &isTerminal, &createdAtText, &updatedAtText); err != nil {
		return nil, err
	}
	createdAt, err := parseTime(createdAtText)
	if err != nil {
		return nil, err
	}
	updatedAt, err := parseTime(updatedAtText)
	if err != nil {
		return nil, err
	}
	return &models.SalesOrderStatusLUT{ID: uuid.MustParse(idText), Code: code, Label: label, SortOrder: sortOrder, IsTerminal: intBool(int64(isTerminal)), CreatedAt: createdAt, UpdatedAt: updatedAt}, nil
}

func scanOrder(scanner interface{ Scan(dest ...any) error }) (*models.SalesOrder, error) {
	var orderIDText string
	var clientIDText string
	var clientName string
	var destinationAddress string
	var requiredDateText string
	var statusIDText string
	var totalValue float64
	var lockedInt int
	var createdAtText string
	var updatedAtText string
	var statusIDScan string
	var statusCode string
	var statusLabel string
	var statusSortOrder int
	var statusTerminalInt int
	var statusCreatedAtText string
	var statusUpdatedAtText string
	if err := scanner.Scan(&orderIDText, &clientIDText, &clientName, &destinationAddress, &requiredDateText, &statusIDText, &totalValue, &lockedInt, &createdAtText, &updatedAtText, &statusIDScan, &statusCode, &statusLabel, &statusSortOrder, &statusTerminalInt, &statusCreatedAtText, &statusUpdatedAtText); err != nil {
		return nil, err
	}
	requiredDate, err := parseTime(requiredDateText)
	if err != nil {
		return nil, err
	}
	createdAt, err := parseTime(createdAtText)
	if err != nil {
		return nil, err
	}
	updatedAt, err := parseTime(updatedAtText)
	if err != nil {
		return nil, err
	}
	statusCreatedAt, err := parseTime(statusCreatedAtText)
	if err != nil {
		return nil, err
	}
	statusUpdatedAt, err := parseTime(statusUpdatedAtText)
	if err != nil {
		return nil, err
	}
	status := &models.SalesOrderStatusLUT{ID: uuid.MustParse(statusIDScan), Code: statusCode, Label: statusLabel, SortOrder: statusSortOrder, IsTerminal: intBool(int64(statusTerminalInt)), CreatedAt: statusCreatedAt, UpdatedAt: statusUpdatedAt}
	return &models.SalesOrder{ID: uuid.MustParse(orderIDText), ClientID: uuid.MustParse(clientIDText), ClientName: clientName, DestinationAddress: destinationAddress, RequiredDate: requiredDate, StatusID: uuid.MustParse(statusIDText), Status: status, TotalValue: totalValue, Locked: lockedInt != 0, CreatedAt: createdAt, UpdatedAt: updatedAt}, nil
}

func scanOrderItem(scanner interface{ Scan(dest ...any) error }) (*models.SalesOrderItem, error) {
	var idText string
	var orderIDText string
	var sku string
	var requestedQty int
	var allocatedQty int
	var unitPrice float64
	var createdAtText string
	var updatedAtText string
	if err := scanner.Scan(&idText, &orderIDText, &sku, &requestedQty, &allocatedQty, &unitPrice, &createdAtText, &updatedAtText); err != nil {
		return nil, err
	}
	createdAt, err := parseTime(createdAtText)
	if err != nil {
		return nil, err
	}
	updatedAt, err := parseTime(updatedAtText)
	if err != nil {
		return nil, err
	}
	return &models.SalesOrderItem{ID: uuid.MustParse(idText), OrderID: uuid.MustParse(orderIDText), SKU: sku, RequestedQty: requestedQty, AllocatedQty: allocatedQty, UnitPrice: unitPrice, CreatedAt: createdAt, UpdatedAt: updatedAt}, nil
}

var _ rootrepo.SQLiteRepository = (*Repository)(nil)
