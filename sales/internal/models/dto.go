package models

import (
	"time"

	"github.com/google/uuid"
)

type OrderItemRequest struct {
	SKU          string  `json:"sku"`
	RequestedQty int     `json:"requested_qty"`
	UnitPrice    float64 `json:"unit_price"`
}

type CreateOrderRequest struct {
	ClientName         string             `json:"client_name"`
	DestinationAddress string             `json:"destination_address"`
	ClientTier         ClientTier         `json:"client_tier"`
	RequiredDate       time.Time          `json:"required_date"`
	Items              []OrderItemRequest `json:"items"`
}

type UpdateOrderRequest struct {
	DestinationAddress *string            `json:"destination_address"`
	RequiredDate       *time.Time         `json:"required_date"`
	Items              []OrderItemRequest `json:"items"`
}

type CreateClientRequest struct {
	Name                      string     `json:"name"`
	Tier                      ClientTier `json:"tier"`
	DefaultDestinationAddress string     `json:"default_destination_address"`
}

type UpdateClientRequest struct {
	Name                      *string     `json:"name"`
	Tier                      *ClientTier `json:"tier"`
	DefaultDestinationAddress *string     `json:"default_destination_address"`
}

type OrderResponse struct {
	Order  SalesOrder       `json:"order"`
	Client Client           `json:"client"`
	Items  []SalesOrderItem `json:"items"`
}

type ClientResponse struct {
	Client Client `json:"client"`
}

type QueueStatus struct {
	Entries []AllocationQueueEntry `json:"entries"`
}

type DispatchNotification struct {
	OrderID uuid.UUID `json:"order_id"`
	Status  string    `json:"status"`
}
