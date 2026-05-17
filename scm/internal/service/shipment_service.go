package service

import (
	"context"
)

type IShipmentService interface {
	// AcquireDispatchLock executes the Dispatch-Locking Procedure to prevent duplicate shipments.
	AcquireDispatchLock(ctx context.Context, shipmentID string, operatorID string) error

	// DispatchShipment finalizes packing, triggers inventory deduction, and advances state.
	DispatchShipment(ctx context.Context, shipmentID string, operatorID string) error
}

type shipmentService struct{}

func ShipmentService() IShipmentService {
	return &shipmentService{}
}

func (s *shipmentService) AcquireDispatchLock(ctx context.Context, shipmentID string, operatorID string) error {
	return ErrNotImplemented
}

func (s *shipmentService) DispatchShipment(ctx context.Context, shipmentID string, operatorID string) error {
	return ErrNotImplemented
}
