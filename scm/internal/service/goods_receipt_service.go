package service

import (
	"context"
)

type GoodsReceiptService interface {
	// AcquireLock executes the Parallel Conflict Resolution. Fails if locked by another operator.
	AcquireLock(ctx context.Context, grID string, operatorID string) error
	
	// ProcessBlindReceipt forces manual count entry and checks Aging Quarantines.
	ProcessBlindReceipt(ctx context.Context, grID string, operatorID string, counts map[string]struct{ Received int; Defective int }) error
	
	// ReleaseLock releases the 60-minute lock.
	ReleaseLock(ctx context.Context, grID string) error
}

type goodsReceiptService struct{}

func NewGoodsReceiptService() GoodsReceiptService {
	return &goodsReceiptService{}
}

func (s *goodsReceiptService) AcquireLock(ctx context.Context, grID string, operatorID string) error {
	return ErrNotImplemented
}

func (s *goodsReceiptService) ProcessBlindReceipt(ctx context.Context, grID string, operatorID string, counts map[string]struct{ Received int; Defective int }) error {
	return ErrNotImplemented
}

func (s *goodsReceiptService) ReleaseLock(ctx context.Context, grID string) error {
	return ErrNotImplemented
}
