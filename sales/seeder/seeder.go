package seeder

import (
	"context"
	"fmt"
	"log"
	"time"

	"zeus-sales-service/internal/models"
	"zeus-sales-service/internal/repository"
	"zeus-sales-service/internal/service"
)

func SeedAll(ctx context.Context, sqliteRepo repository.DbRepository, cacheRepo repository.CacheRepository) error {
	if sqliteRepo == nil || cacheRepo == nil {
		return fmt.Errorf("sqlite and cache repositories are required")
	}

	log.Println("Starting Sales Seeder...")
	services := service.NewServices(sqliteRepo, cacheRepo)

	clients, err := sqliteRepo.ListClients(ctx)
	if err != nil {
		return err
	}
	orders, err := sqliteRepo.ListOrders(ctx)
	if err != nil {
		return err
	}
	if len(clients) > 0 || len(orders) > 0 {
		log.Println("Sales data already exists, skipping seed.")
		return nil
	}

	for sku, quantity := range map[string]int{
		"SKU-100": 200,
		"SKU-200": 150,
		"SKU-300": 100,
	} {
		if err := cacheRepo.SetATP(ctx, sku, quantity); err != nil {
			return fmt.Errorf("set atp for %s: %w", sku, err)
		}
	}

	seedOrders := []models.CreateOrderRequest{
		{
			ClientName:         "TechCorp Solutions",
			ClientTier:         models.ClientTierB2B,
			DestinationAddress: "123 Innovation Drive, Silicon Valley, CA 94043",
			RequiredDate:       time.Now().Add(72 * time.Hour).UTC(),
			Items: []models.OrderItemRequest{
				{SKU: "SKU-100", RequestedQty: 5, UnitPrice: 12.5},
				{SKU: "SKU-200", RequestedQty: 2, UnitPrice: 9.0},
			},
		},
		{
			ClientName:         "TechCorp Solutions",
			ClientTier:         models.ClientTierB2B,
			DestinationAddress: "123 Innovation Drive, Silicon Valley, CA 94043",
			RequiredDate:       time.Now().Add(96 * time.Hour).UTC(),
			Items: []models.OrderItemRequest{
				{SKU: "SKU-200", RequestedQty: 3, UnitPrice: 9.0},
			},
		},
		{
			ClientName:         "Bright Retail",
			ClientTier:         models.ClientTierB2C,
			DestinationAddress: "55 Market St, San Jose, CA 95113",
			RequiredDate:       time.Now().Add(48 * time.Hour).UTC(),
			Items: []models.OrderItemRequest{
				{SKU: "SKU-300", RequestedQty: 1, UnitPrice: 49.0},
			},
		},
	}

	for _, req := range seedOrders {
		if _, err := services.Orders.CreateOrder(ctx, req); err != nil {
			return fmt.Errorf("seed order for %s: %w", req.ClientName, err)
		}
	}

	log.Println("Sales Seeder finished successfully.")
	return nil
}
