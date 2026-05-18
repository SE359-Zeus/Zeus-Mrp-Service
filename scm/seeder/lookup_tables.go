package seeder

import (
	"time"
	"gorm.io/gorm"
	"zeus-scm-service/internal/models"
)

func seedLookupTables(db *gorm.DB) {
	conditions := []string{"OPERATIONAL", "DEGRADED", "DAMAGED", "SCRAPPED", "LOST_STOLEN"}
	for i, name := range conditions {
		db.FirstOrCreate(&models.PartCondition{ID: int32(i + 1), Name: name}, models.PartCondition{Name: name})
	}

	statuses := []string{"Pending", "Production", "Discontinued"}
	for i, name := range statuses {
		db.FirstOrCreate(&models.PartMfgStatus{ID: int32(i + 1), Name: name, CreatedAt: time.Now(), UpdatedAt: time.Now()}, models.PartMfgStatus{Name: name})
	}

	poStates := []string{"Draft", "Approved", "In Transit", "Received"}
	for i, name := range poStates {
		db.FirstOrCreate(&models.PurchaseOrderState{ID: int32(i + 1), Name: name}, models.PurchaseOrderState{Name: name})
	}

	grStates := []string{"Pending", "Inspected", "Complete", "Discrepancy"}
	for i, name := range grStates {
		db.FirstOrCreate(&models.GoodsReceiptState{ID: int32(i + 1), Name: name}, models.GoodsReceiptState{Name: name})
	}

	stockStates := []string{"In Stock", "Low Stock", "Out of Stock", "Discontinued"}
	for i, name := range stockStates {
		db.FirstOrCreate(&models.ComponentStockState{ID: int32(i + 1), Name: name}, models.ComponentStockState{Name: name})
	}

	shipmentStates := []string{"Scheduled", "In Transit", "Delivered", "Delayed"}
	for i, name := range shipmentStates {
		db.FirstOrCreate(&models.ShipmentState{ID: int32(i + 1), Name: name}, models.ShipmentState{Name: name})
	}
}
