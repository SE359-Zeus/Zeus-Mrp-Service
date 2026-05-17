package service

import (
	"context"

	"zeus-scm-service/internal/models"

	"github.com/google/uuid"
)

// IInventoryService is the MAIN universal contract for the entire application ecosystem,
// upholding exactly the data models specified in the Zeus reference/ Rust directory.
type IInventoryService interface {
	// --- Products ---
	GetProduct(ctx context.Context, id uuid.UUID) (*models.Product, error)
	ListProducts(ctx context.Context) ([]models.Product, error)
	CreateProduct(ctx context.Context, p *models.Product) error

	// --- Product Models ---
	GetProductModel(ctx context.Context, code string) (*models.ProductModel, error)
	CreateProductModel(ctx context.Context, m *models.ProductModel) error

	// --- Parts ---
	GetPart(ctx context.Context, id uuid.UUID) (*models.Part, error)
	ListParts(ctx context.Context, catalogID *uuid.UUID, productID *uuid.UUID, conditionID *int32) ([]models.Part, error)
	CreatePart(ctx context.Context, p *models.Part) error

	// Operational endpoints for Parts (Lifecycle)
	UpdatePartCondition(ctx context.Context, partID uuid.UUID, conditionID int32) error
	MarkPartScrapped(ctx context.Context, partID uuid.UUID) error
	InstallPart(ctx context.Context, partID uuid.UUID, productID uuid.UUID) error
	RemovePart(ctx context.Context, partID uuid.UUID) error

	// --- Part Catalog ---
	GetPartCatalog(ctx context.Context, id uuid.UUID) (*models.PartCatalog, error)
	ListPartCatalog(ctx context.Context, typeID *int32) ([]models.PartCatalog, error)

	// --- Users ---
	GetUser(ctx context.Context, id uuid.UUID) (*models.User, error)
}

type inventoryService struct{}

func InventoryService() IInventoryService {
	return &inventoryService{}
}

func (s *inventoryService) GetProduct(ctx context.Context, id uuid.UUID) (*models.Product, error) {
	return nil, ErrNotImplemented
}
func (s *inventoryService) ListProducts(ctx context.Context) ([]models.Product, error) {
	return nil, ErrNotImplemented
}
func (s *inventoryService) CreateProduct(ctx context.Context, p *models.Product) error {
	return ErrNotImplemented
}
func (s *inventoryService) GetProductModel(ctx context.Context, code string) (*models.ProductModel, error) {
	return nil, ErrNotImplemented
}
func (s *inventoryService) CreateProductModel(ctx context.Context, m *models.ProductModel) error {
	return ErrNotImplemented
}
func (s *inventoryService) GetPart(ctx context.Context, id uuid.UUID) (*models.Part, error) {
	return nil, ErrNotImplemented
}
func (s *inventoryService) ListParts(ctx context.Context, catalogID *uuid.UUID, productID *uuid.UUID, conditionID *int32) ([]models.Part, error) {
	return nil, ErrNotImplemented
}
func (s *inventoryService) CreatePart(ctx context.Context, p *models.Part) error {
	return ErrNotImplemented
}
func (s *inventoryService) UpdatePartCondition(ctx context.Context, partID uuid.UUID, conditionID int32) error {
	return ErrNotImplemented
}
func (s *inventoryService) MarkPartScrapped(ctx context.Context, partID uuid.UUID) error {
	return ErrNotImplemented
}
func (s *inventoryService) InstallPart(ctx context.Context, partID uuid.UUID, productID uuid.UUID) error {
	return ErrNotImplemented
}
func (s *inventoryService) RemovePart(ctx context.Context, partID uuid.UUID) error {
	return ErrNotImplemented
}
func (s *inventoryService) GetPartCatalog(ctx context.Context, id uuid.UUID) (*models.PartCatalog, error) {
	return nil, ErrNotImplemented
}
func (s *inventoryService) ListPartCatalog(ctx context.Context, typeID *int32) ([]models.PartCatalog, error) {
	return nil, ErrNotImplemented
}
func (s *inventoryService) GetUser(ctx context.Context, id uuid.UUID) (*models.User, error) {
	return nil, ErrNotImplemented
}
