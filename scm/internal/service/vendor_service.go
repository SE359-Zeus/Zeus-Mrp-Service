package service

import (
	"context"

	"github.com/google/uuid"
	"zeus-scm-service/internal/models"
)

type VendorService interface {
	// GetOptimalSupplier routes a component shortage to the best active supplier based on QualityScore and UnitPrice
	GetOptimalSupplier(ctx context.Context, sku string) (*models.Supplier, *models.SkuMapping, error)
	
	// UpdateSupplierMetrics recalculates OnTimeRate and DefectRate dynamically based on historical Goods Receipts
	UpdateSupplierMetrics(ctx context.Context, supplierID uuid.UUID) error
}

type vendorService struct{}

func NewVendorService() VendorService {
	return &vendorService{}
}

func (s *vendorService) GetOptimalSupplier(ctx context.Context, sku string) (*models.Supplier, *models.SkuMapping, error) {
	return nil, nil, ErrNotImplemented
}

func (s *vendorService) UpdateSupplierMetrics(ctx context.Context, supplierID uuid.UUID) error {
	return ErrNotImplemented
}
