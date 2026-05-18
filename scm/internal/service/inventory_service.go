package service

import (
	"context"
	"time"

	"zeus-scm-service/internal/messaging"
	"zeus-scm-service/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type IInventoryService interface {
	GetProduct(ctx context.Context, id uuid.UUID) (*models.Product, error)
	ListProducts(ctx context.Context) ([]models.Product, error)
	CreateProduct(ctx context.Context, p *models.Product) error

	GetProductModel(ctx context.Context, code string) (*models.ProductModel, error)
	CreateProductModel(ctx context.Context, m *models.ProductModel) error

	GetPart(ctx context.Context, id uuid.UUID) (*models.Part, error)
	ListParts(ctx context.Context, catalogID *uuid.UUID, productID *uuid.UUID, conditionID *int32) ([]models.Part, error)
	CreatePart(ctx context.Context, p *models.Part) error

	UpdatePartCondition(ctx context.Context, partID uuid.UUID, conditionID int32) error
	MarkPartScrapped(ctx context.Context, partID uuid.UUID) error
	InstallPart(ctx context.Context, partID uuid.UUID, productID uuid.UUID) error
	RemovePart(ctx context.Context, partID uuid.UUID) error

	GetPartCatalog(ctx context.Context, id uuid.UUID) (*models.PartCatalog, error)
	ListPartCatalog(ctx context.Context, typeID *int32) ([]models.PartCatalog, error)

	GetUser(ctx context.Context, id uuid.UUID) (*models.User, error)
}

type inventoryService struct {
	db *gorm.DB
	mq *messaging.RabbitMQ
}

func NewInventoryService(db *gorm.DB, mq *messaging.RabbitMQ) IInventoryService {
	return &inventoryService{db: db, mq: mq}
}

func (s *inventoryService) GetProduct(ctx context.Context, id uuid.UUID) (*models.Product, error) {
	var p models.Product
	if err := s.db.WithContext(ctx).First(&p, "id = ?", id).Error; err != nil {
		return nil, ErrNotFound
	}
	return &p, nil
}

func (s *inventoryService) ListProducts(ctx context.Context) ([]models.Product, error) {
	var products []models.Product
	if err := s.db.WithContext(ctx).Find(&products).Error; err != nil {
		return nil, err
	}
	return products, nil
}

func (s *inventoryService) CreateProduct(ctx context.Context, p *models.Product) error {
	return s.db.WithContext(ctx).Create(p).Error
}

func (s *inventoryService) GetProductModel(ctx context.Context, code string) (*models.ProductModel, error) {
	var m models.ProductModel
	if err := s.db.WithContext(ctx).First(&m, "model_code = ?", code).Error; err != nil {
		return nil, ErrNotFound
	}
	return &m, nil
}

func (s *inventoryService) CreateProductModel(ctx context.Context, m *models.ProductModel) error {
	return s.db.WithContext(ctx).Create(m).Error
}

func (s *inventoryService) GetPart(ctx context.Context, id uuid.UUID) (*models.Part, error) {
	var p models.Part
	if err := s.db.WithContext(ctx).First(&p, "id = ?", id).Error; err != nil {
		return nil, ErrNotFound
	}
	return &p, nil
}

func (s *inventoryService) ListParts(ctx context.Context, catalogID *uuid.UUID, productID *uuid.UUID, conditionID *int32) ([]models.Part, error) {
	query := s.db.WithContext(ctx).Model(&models.Part{})
	if catalogID != nil {
		query = query.Where("part_catalog_id = ?", *catalogID)
	}
	if productID != nil {
		query = query.Where("product_id = ?", *productID)
	}
	if conditionID != nil {
		query = query.Where("part_condition_id = ?", *conditionID)
	}
	var parts []models.Part
	if err := query.Find(&parts).Error; err != nil {
		return nil, err
	}
	return parts, nil
}

func (s *inventoryService) CreatePart(ctx context.Context, p *models.Part) error {
	return s.db.WithContext(ctx).Create(p).Error
}

func (s *inventoryService) UpdatePartCondition(ctx context.Context, partID uuid.UUID, conditionID int32) error {
	result := s.db.WithContext(ctx).Model(&models.Part{}).
		Where("id = ?", partID).
		Update("part_condition_id", conditionID)
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return result.Error
}

func (s *inventoryService) MarkPartScrapped(ctx context.Context, partID uuid.UUID) error {
	now := time.Now()
	result := s.db.WithContext(ctx).Model(&models.Part{}).
		Where("id = ?", partID).
		Update("scrapped_date", now)
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return result.Error
}

func (s *inventoryService) InstallPart(ctx context.Context, partID uuid.UUID, productID uuid.UUID) error {
	now := time.Now()
	result := s.db.WithContext(ctx).Model(&models.Part{}).
		Where("id = ?", partID).
		Updates(map[string]interface{}{
			"product_id":        productID,
			"installation_date": now,
		})
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return result.Error
}

func (s *inventoryService) RemovePart(ctx context.Context, partID uuid.UUID) error {
	now := time.Now()
	result := s.db.WithContext(ctx).Model(&models.Part{}).
		Where("id = ?", partID).
		Updates(map[string]interface{}{
			"product_id":   nil,
			"removal_date": now,
		})
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return result.Error
}

func (s *inventoryService) GetPartCatalog(ctx context.Context, id uuid.UUID) (*models.PartCatalog, error) {
	var c models.PartCatalog
	if err := s.db.WithContext(ctx).First(&c, "id = ?", id).Error; err != nil {
		return nil, ErrNotFound
	}
	return &c, nil
}

func (s *inventoryService) ListPartCatalog(ctx context.Context, typeID *int32) ([]models.PartCatalog, error) {
	query := s.db.WithContext(ctx).Model(&models.PartCatalog{})
	if typeID != nil {
		query = query.Where("part_types_id = ?", *typeID)
	}
	var catalogs []models.PartCatalog
	if err := query.Find(&catalogs).Error; err != nil {
		return nil, err
	}
	return catalogs, nil
}

func (s *inventoryService) GetUser(ctx context.Context, id uuid.UUID) (*models.User, error) {
	var u models.User
	if err := s.db.WithContext(ctx).First(&u, "id = ?", id).Error; err != nil {
		return nil, ErrNotFound
	}
	return &u, nil
}
