package seeder

import (
	"time"
	"github.com/google/uuid"
	"github.com/brianvoe/gofakeit/v6"
	"gorm.io/gorm"
	"zeus-scm-service/internal/models"
)

func seedProductsAndParts(db *gorm.DB, modelsList []models.ProductModel, catMap map[string]models.PartCatalog) {
	for _, pm := range modelsList {
		for i := 0; i < 2; i++ {
			prod := models.Product{
				ID:               uuid.New(),
				ProductModelCode: pm.ModelCode,
				CustomerID:       uuid.New(), // Random UUID for customer
				ProductName:      pm.ModelName + " Build " + gofakeit.LetterN(3),
				SerialNumber:     "SN-" + gofakeit.LetterN(8),
				CreatedAt:        time.Now(),
				UpdatedAt:        time.Now(),
			}
			db.Create(&prod)

			var boms []models.PartsByModel
			db.Where("product_model_code = ?", pm.ModelCode).Find(&boms)

			for _, bom := range boms {
				for q := int32(0); q < bom.Quantity; q++ {
					pid := prod.ID
					p := models.Part{
						ID:               uuid.New(),
						PartCatalogID:    bom.PartCatalogID,
						ProductID:        &pid,
						SerialNumber:     "PART-" + gofakeit.LetterN(10),
						PartConditionID:  1,
						ManufacturedDate: time.Now().AddDate(0, -gofakeit.Number(1, 12), 0),
						InstallationDate: &prod.CreatedAt,
						CreatedAt:        time.Now(),
						UpdatedAt:        time.Now(),
					}
					db.Create(&p)
				}
			}
		}
	}
}
