package seeder

import (
	"github.com/brianvoe/gofakeit/v6"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"zeus-scm-service/internal/models"
)

func seedSuppliers(db *gorm.DB, count int) []models.Supplier {
	var suppliers []models.Supplier
	for i := 0; i < count; i++ {
		s := models.Supplier{
			ID:           uuid.New(),
			Name:         gofakeit.Company(),
			Contact:      gofakeit.Email(),
			Tier:         models.SupplierTierQualified,
			LeadTimeDays: gofakeit.Number(5, 30),
			QualityScore: gofakeit.Float64Range(80.0, 99.9),
			OnTimeRate:   gofakeit.Float64Range(80.0, 99.9),
		}
		db.Create(&s)
		suppliers = append(suppliers, s)
	}
	return suppliers
}
