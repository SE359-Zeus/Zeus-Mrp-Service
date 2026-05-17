package service

import (
	"context"

	"zeus-scm-service/internal/models"

	"github.com/google/uuid"
)

type IPOService interface {
	// CreateDraft initializes a new PO draft. Enforces mono-vendor constraint.
	CreateDraft(ctx context.Context, vendorID uuid.UUID, targetBuild string) (*models.PurchaseOrder, error)

	// AddLineItemWithLock executes the Eager Slot-Locking concurrency control. Deducts from global deficit immediately.
	AddLineItemWithLock(ctx context.Context, poID string, sku string, qty int) error

	// ApprovePO advances Draft to Approved state. Fails if 30-minute lock expired.
	ApprovePO(ctx context.Context, poID string) error

	// TransitionState advances PO state while preventing state regression
	TransitionState(ctx context.Context, poID string, newState models.POStatus) error
}

type poService struct{}

func POService() IPOService {
	return &poService{}
}

func (s *poService) CreateDraft(ctx context.Context, vendorID uuid.UUID, targetBuild string) (*models.PurchaseOrder, error) {
	return nil, ErrNotImplemented
}

func (s *poService) AddLineItemWithLock(ctx context.Context, poID string, sku string, qty int) error {
	return ErrNotImplemented
}

func (s *poService) ApprovePO(ctx context.Context, poID string) error {
	return ErrNotImplemented
}

func (s *poService) TransitionState(ctx context.Context, poID string, newState models.POStatus) error {
	return ErrNotImplemented
}
