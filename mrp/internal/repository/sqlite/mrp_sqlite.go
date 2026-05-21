package sqlite

import (
	"context"
	"fmt"
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
	if id == uuid.Nil {
		return nil, nil
	}

	type row struct {
		ID               string
		ProductModelCode string
		TargetQuantity   int
		Status           string
		ScheduledAt      *string
		CreatedAt        *string
	}

	var dbRow row
	err := r.db.WithContext(ctx).
		Table("production_orders").
		Select("id, product_model_code, target_quantity, status, scheduled_at, created_at").
		Where("id = ?", id.String()).
		Take(&dbRow).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	parsedID, err := uuid.Parse(dbRow.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid production order id in database: %w", err)
	}

	return &models.ProductionOrder{
		ID:               parsedID,
		ProductModelCode: dbRow.ProductModelCode,
		TargetQuantity:   dbRow.TargetQuantity,
		Status:           models.ProductionOrderStatus(dbRow.Status),
	}, nil
}

func (r *sqliteMRPRepository) GetOpenProductionOrders(ctx context.Context) ([]models.ProductionOrder, error) {
	type row struct {
		ID               string
		ProductModelCode string
		TargetQuantity   int
		Status           string
	}

	var rows []row
	err := r.db.WithContext(ctx).
		Table("production_orders").
		Select("id, product_model_code, target_quantity, status").
		Where("status IN ?", []string{string(models.StatusClearToBuild), string(models.StatusPartial), string(models.StatusShortage)}).
		Order("created_at DESC").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}

	orders := make([]models.ProductionOrder, 0, len(rows))
	for _, row := range rows {
		id, err := uuid.Parse(row.ID)
		if err != nil {
			continue
		}

		orders = append(orders, models.ProductionOrder{
			ID:               id,
			ProductModelCode: row.ProductModelCode,
			TargetQuantity:   row.TargetQuantity,
			Status:           models.ProductionOrderStatus(row.Status),
		})
	}

	return orders, nil
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
	if modelCode == "" {
		return []models.BomEntry{}, nil
	}

	type row struct {
		ID                      int
		ParentModelCode         string
		ComponentPartID         string
		RequiredQuantityPerUnit int
	}

	var rows []row
	err := r.db.WithContext(ctx).
		Table("bom_entries").
		Select("id, parent_model_code, component_part_id, required_quantity_per_unit").
		Where("parent_model_code = ?", modelCode).
		Order("id ASC").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}

	entries := make([]models.BomEntry, 0, len(rows))
	for _, row := range rows {
		partID, err := uuid.Parse(row.ComponentPartID)
		if err != nil {
			continue
		}

		entries = append(entries, models.BomEntry{
			ID:                      row.ID,
			ParentModelCode:         row.ParentModelCode,
			ComponentPartID:         partID,
			RequiredQuantityPerUnit: row.RequiredQuantityPerUnit,
		})
	}

	return entries, nil
}

func (r *sqliteMRPRepository) GetAllBOMs(ctx context.Context) ([]models.BomEntry, error) {
	return nil, nil
}

func (r *sqliteMRPRepository) GetWhereUsedByPartID(ctx context.Context, partID uuid.UUID) ([]models.BomEntry, error) {
	return nil, nil
}

func (r *sqliteMRPRepository) CreateShortageLog(ctx context.Context, log *models.ShortageLog) error {
	if log == nil {
		return fmt.Errorf("shortage log is nil")
	}

	type shortageRecord struct {
		ID                string `gorm:"column:id"`
		ProductionOrderID string `gorm:"column:production_order_id"`
		PartID            string `gorm:"column:part_id"`
		ShortageQty       int    `gorm:"column:shortage_qty"`
		ResolutionStatus  string `gorm:"column:resolution_status"`
	}

	rec := shortageRecord{
		ID:                log.ID.String(),
		ProductionOrderID: log.ProductionOrderID.String(),
		PartID:            log.PartID.String(),
		ShortageQty:       log.ShortageQty,
		ResolutionStatus:  log.ResolutionStatus,
	}

	return r.db.WithContext(ctx).Table("shortage_logs").Create(&rec).Error
}

func (r *sqliteMRPRepository) GetShortagesByOrderID(ctx context.Context, orderID uuid.UUID) ([]models.ShortageLog, error) {
	if orderID == uuid.Nil {
		return []models.ShortageLog{}, nil
	}

	type row struct {
		ID                string
		ProductionOrderID string
		PartID            string
		ShortageQty       int
		ResolutionStatus  string
	}

	var rows []row
	err := r.db.WithContext(ctx).
		Table("shortage_logs").
		Select("id, production_order_id, part_id, shortage_qty, resolution_status").
		Where("production_order_id = ?", orderID.String()).
		Order("id ASC").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}

	logs := make([]models.ShortageLog, 0, len(rows))
	for _, row := range rows {
		id, err := uuid.Parse(row.ID)
		if err != nil {
			continue
		}
		productionOrderID, err := uuid.Parse(row.ProductionOrderID)
		if err != nil {
			continue
		}
		partID, err := uuid.Parse(row.PartID)
		if err != nil {
			continue
		}

		logs = append(logs, models.ShortageLog{
			ID:                id,
			ProductionOrderID: productionOrderID,
			PartID:            partID,
			ShortageQty:       row.ShortageQty,
			ResolutionStatus:  row.ResolutionStatus,
		})
	}

	return logs, nil
}

func (r *sqliteMRPRepository) GetAggregatedShortages(ctx context.Context) ([]models.BOMExplosionResult, error) {
	type row struct {
		PartID           string
		TotalRequiredQty int
	}

	var rows []row
	err := r.db.WithContext(ctx).
		Table("shortage_logs").
		Select("part_id, SUM(shortage_qty) AS total_required_qty").
		Group("part_id").
		Order("part_id ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	results := make([]models.BOMExplosionResult, 0, len(rows))
	for _, row := range rows {
		partID, err := uuid.Parse(row.PartID)
		if err != nil {
			continue
		}

		results = append(results, models.BOMExplosionResult{
			PartID:           partID,
			TotalRequiredQty: row.TotalRequiredQty,
			AvailableQty:     0,
			IsShortage:       row.TotalRequiredQty > 0,
		})
	}

	return results, nil
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
