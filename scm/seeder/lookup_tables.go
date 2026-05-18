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
}
