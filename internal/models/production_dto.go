package models

import (
	"time"

	"github.com/google/uuid"
)

type CreateProductionOrderRequest struct {
	ProductModelCode string    `json:"product_model_code" validate:"required"`
	TargetQuantity   int       `json:"target_quantity" validate:"gt=0"`
	ScheduledAt      time.Time `json:"scheduled_at"`
}

type ProductionOrderResponse struct {
	ID               uuid.UUID             `json:"id"`
	ProductModelCode string                `json:"product_model_code"`
	TargetQuantity   int                   `json:"target_quantity"`
	Status           ProductionOrderStatus `json:"status"`
	Shortages        []ShortageLog         `json:"shortages,omitempty"`
}

type BOMExplosionResult struct {
	PartID           uuid.UUID `json:"part_id"`
	TotalRequiredQty int       `json:"total_required_qty"`
	AvailableQty     int       `json:"available_qty"`
	IsShortage       bool      `json:"is_shortage"`
}

type ReadinessMatrixRow struct {
	OrderID           uuid.UUID            `json:"order_id"`
	TargetBuild       string               `json:"target_build"`
	Quantity          int                  `json:"quantity"`
	Status            string               `json:"status"`
	DeficitBreakdown  []BOMExplosionResult `json:"deficit_breakdown,omitempty"`
}

type ReadinessMetrics struct {
	TotalOpenOrders       int     `json:"total_open_orders"`
	ComponentsInShortage  int     `json:"components_in_shortage"`
	BlockedOrders         int     `json:"blocked_orders"`
	SupplyReadinessRate   float64 `json:"supply_readiness_rate"`
}

type InventoryTransactionDTO struct {
	ID             string    `json:"id"`
	SKU            string    `json:"sku"`
	Type           string    `json:"type"`
	QtyChange      int       `json:"qty_change"`
	RunningBalance int       `json:"running_balance"`
	Location       string    `json:"location"`
	Timestamp      time.Time `json:"timestamp"`
	Operator       string    `json:"operator"`
	Reference      string    `json:"reference"`
}

type PickListDTO struct {
	OrderID    uuid.UUID      `json:"order_id"`
	Components []PickListItem `json:"components"`
}

type PickListItem struct {
	PartID      uuid.UUID `json:"part_id"`
	SKU         string    `json:"sku"`
	Quantity    int       `json:"quantity"`
	BinLocation string    `json:"bin_location"`
}

// --- BOM & Assembly DTOs ---

type CreateAssemblyRequest struct {
	Name       string               `json:"name" validate:"required"`
	Components []ComponentReference `json:"components" validate:"required,dive"`
}

type UpdateAssemblyRequest struct {
	Name       string               `json:"name"`
	Components []ComponentReference `json:"components"`
}

type ComponentReference struct {
	SKU      string `json:"sku" validate:"required"`
	Quantity int    `json:"qty" validate:"gt=0"`
}

// --- Pagination & Filtering ---

type PaginationParams struct {
	Page    int `json:"page"`
	PerPage int `json:"per_page"`
}

type ReadinessFilter struct {
	Status string `json:"status"`
	Search string `json:"search"` // Order ID or SKU
}

// --- Inventory & Ledger DTOs ---

type InventoryMetrics struct {
	ActiveSKUs        int     `json:"active_skus"`
	StockAccuracy     float64 `json:"stock_accuracy"`
	InventoryTurnover float64 `json:"inventory_turnover"`
	CycleCountGaps    int     `json:"cycle_count_gaps"`
}

// --- Demand & POs DTOs ---

type DemandPOSummary struct {
	OrderID      string `json:"order_id"`
	TargetBuild  string `json:"target_build"`
	Quantity     int    `json:"quantity"`
	Status       string `json:"status"`
	MissingCount int    `json:"missing_count"`
}
