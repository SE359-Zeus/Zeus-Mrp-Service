package seeder

import (
	"fmt"
	"time"
	"gorm.io/gorm"
	"zeus-scm-service/internal/models"
)

func seedProductModels(db *gorm.DB, installs map[string][]PartInstallationData, catMap map[string]models.PartCatalog) []models.ProductModel {
	baseModels := []models.ProductModel{
		{ModelCode: "82SN003JVN", ModelName: "IdeaPad 5 Pro 16ARH7"},
		{ModelCode: "83LY00HQVN", ModelName: "Legion 5 15IRX10"},
	}
	
	newModels := []models.ProductModel{
		{ModelCode: "21CB000QUS", ModelName: "ThinkPad X1 Carbon Gen 11"},
		{ModelCode: "82A3000GUS", ModelName: "Yoga Slim 7i"},
		{ModelCode: "82WQ002RUS", ModelName: "Legion Pro 7i"},
	}

	allModels := append(baseModels, newModels...)

	for _, m := range allModels {
		desc := "Seeded product model"
		m.Description = &desc
		m.CreatedAt = time.Now()
		m.UpdatedAt = time.Now()
		db.Create(&m)

		bomList, exists := installs[m.ModelCode]
		if !exists {
			bomList = installs["82SN003JVN"] 
		}

		for _, item := range bomList {
			key := fmt.Sprintf("%s|%s", item.PartNumber, item.MfgNumber)
			if cat, ok := catMap[key]; ok {
				db.FirstOrCreate(&models.PartsByModel{
					PartCatalogID:    cat.ID,
					ProductModelCode: m.ModelCode,
					Quantity:         int32(item.Quantity),
				}, models.PartsByModel{PartCatalogID: cat.ID, ProductModelCode: m.ModelCode})
			}
		}
	}
	return allModels
}
