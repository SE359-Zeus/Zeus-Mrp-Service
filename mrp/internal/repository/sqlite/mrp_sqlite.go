package sqlite

import (
	"context"
	"zeus-mrp-service/internal/models"
	"zeus-mrp-service/internal/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type sqliteMRPRepository struct {
	db *gorm.DB
}

func NewSqliteMRPRepository(db *gorm.DB) repository.MRPRepository {
	return &sqliteMRPRepository{db: db}
}

// Implementation of the interface (Stubs only, no logic)

func (r *sqliteMRPRepository) CreateProductionOrder(ctx context.Context, order *models.ProductionOrder) error {
	return nil
}

func (r *sqliteMRPRepository) GetProductionOrder(ctx context.Context, id uuid.UUID) (*models.ProductionOrder, error) {
	return nil, nil
}

func (r *sqliteMRPRepository) GetOpenProductionOrders(ctx context.Context) ([]models.ProductionOrder, error) {
	return nil, nil
}

func (r *sqliteMRPRepository) UpdateProductionOrderStatus(ctx context.Context, id uuid.UUID, status models.ProductionOrderStatus) error {
	return nil
}

func (r *sqliteMRPRepository) CreateBOMEntries(ctx context.Context, entries []models.BomEntry) error {
	return nil
}

func (r *sqliteMRPRepository) DeleteBOMEntriesByModelCode(ctx context.Context, modelCode string) error {
	return nil
}

func (r *sqliteMRPRepository) GetBOMByModelCode(ctx context.Context, modelCode string) ([]models.BomEntry, error) {
	return nil, nil
}

func (r *sqliteMRPRepository) GetAllBOMs(ctx context.Context) ([]models.BomEntry, error) {
	return nil, nil
}

func (r *sqliteMRPRepository) GetWhereUsedByPartID(ctx context.Context, partID uuid.UUID) ([]models.BomEntry, error) {
	return nil, nil
}

func (r *sqliteMRPRepository) CreateShortageLog(ctx context.Context, log *models.ShortageLog) error {
	return nil
}

func (r *sqliteMRPRepository) GetShortagesByOrderID(ctx context.Context, orderID uuid.UUID) ([]models.ShortageLog, error) {
	return nil, nil
}

func (r *sqliteMRPRepository) GetAggregatedShortages(ctx context.Context) ([]models.BOMExplosionResult, error) {
	return nil, nil
}

func (r *sqliteMRPRepository) GetPartInventory(ctx context.Context, partID uuid.UUID) (int, error) {
	return 0, nil
}

func (r *sqliteMRPRepository) GetInventoryTransactions(ctx context.Context) ([]models.InventoryTransactionDTO, error) {
	return nil, nil
}

func (r *sqliteMRPRepository) GetInventoryMetrics(ctx context.Context) (*models.InventoryMetrics, error) {
	return nil, nil
}
