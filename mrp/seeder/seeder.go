package seeder

import (
	"context"
	"fmt"
	"log"
	"time"

	"zeus-mrp-service/internal/models"
	"zeus-mrp-service/internal/repository"

	"github.com/google/uuid"
)

func SeedAll(ctx context.Context, repo repository.DbRepository) error {
	if repo == nil {
		return fmt.Errorf("repository is required")
	}

	log.Println("Starting MRP seeder...")

	openOrders, err := repo.GetOpenProductionOrders(ctx)
	if err != nil {
		return err
	}
	allBOMs, err := repo.GetAllBOMs(ctx)
	if err != nil {
		return err
	}
	if len(openOrders) > 0 || len(allBOMs) > 0 {
		log.Println("MRP seed data already exists, skipping.")
		return nil
	}

	now := time.Now().UTC()
	alphaModel := "WORKSTATION-ALPHA"
	betaModel := "WORKSTATION-BETA"

	alphaPartCPU := uuid.NewSHA1(uuid.NameSpaceURL, []byte("mrp:part:cpu"))
	alphaPartPSU := uuid.NewSHA1(uuid.NameSpaceURL, []byte("mrp:part:psu"))
	alphaPartRAM := uuid.NewSHA1(uuid.NameSpaceURL, []byte("mrp:part:ram"))
	betaPartSSD := uuid.NewSHA1(uuid.NameSpaceURL, []byte("mrp:part:ssd"))
	betaPartFAN := uuid.NewSHA1(uuid.NameSpaceURL, []byte("mrp:part:fan"))

	bomEntries := []models.BomEntry{
		{ParentModelCode: alphaModel, ComponentPartID: alphaPartCPU, RequiredQuantityPerUnit: 1},
		{ParentModelCode: alphaModel, ComponentPartID: alphaPartPSU, RequiredQuantityPerUnit: 1},
		{ParentModelCode: alphaModel, ComponentPartID: alphaPartRAM, RequiredQuantityPerUnit: 2},
		{ParentModelCode: betaModel, ComponentPartID: alphaPartCPU, RequiredQuantityPerUnit: 1},
		{ParentModelCode: betaModel, ComponentPartID: betaPartSSD, RequiredQuantityPerUnit: 1},
		{ParentModelCode: betaModel, ComponentPartID: betaPartFAN, RequiredQuantityPerUnit: 2},
	}
	if err := repo.CreateBOMEntries(ctx, bomEntries); err != nil {
		return fmt.Errorf("seed bom entries: %w", err)
	}

	orderAlphaID := uuid.NewSHA1(uuid.NameSpaceURL, []byte("mrp:order:alpha"))
	orderBetaID := uuid.NewSHA1(uuid.NameSpaceURL, []byte("mrp:order:beta"))
	productionOrders := []models.ProductionOrder{
		{
			ID:               orderAlphaID,
			ProductModelCode: alphaModel,
			TargetQuantity:   10,
			Status:           models.StatusClearToBuild,
			ScheduledAt:      now.Add(72 * time.Hour),
			CreatedAt:        now,
		},
		{
			ID:               orderBetaID,
			ProductModelCode: betaModel,
			TargetQuantity:   6,
			Status:           models.StatusShortage,
			ScheduledAt:      now.Add(96 * time.Hour),
			CreatedAt:        now,
		},
	}
	for i := range productionOrders {
		order := productionOrders[i]
		if err := repo.CreateProductionOrder(ctx, &order); err != nil {
			return fmt.Errorf("seed production order %s: %w", order.ProductModelCode, err)
		}
	}

	shortageLogs := []models.ShortageLog{
		{
			ID:                uuid.NewSHA1(uuid.NameSpaceURL, []byte("mrp:shortage:beta:ssd")),
			ProductionOrderID: orderBetaID,
			PartID:            betaPartSSD,
			ShortageQty:       6,
			ResolutionStatus:  "EMITTED",
		},
		{
			ID:                uuid.NewSHA1(uuid.NameSpaceURL, []byte("mrp:shortage:beta:fan")),
			ProductionOrderID: orderBetaID,
			PartID:            betaPartFAN,
			ShortageQty:       4,
			ResolutionStatus:  "EMITTED",
		},
	}
	for i := range shortageLogs {
		entry := shortageLogs[i]
		if err := repo.CreateShortageLog(ctx, &entry); err != nil {
			return fmt.Errorf("seed shortage log %s: %w", entry.ID.String(), err)
		}
	}

	log.Println("MRP seeder finished successfully.")
	return nil
}
