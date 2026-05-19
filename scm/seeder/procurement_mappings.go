package seeder

import (
	"fmt"
	"math"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"zeus-scm-service/internal/models"
)

func seedSkuMappings(db *gorm.DB, suppliers []models.Supplier, catalogs []PartCatalogData) map[uuid.UUID][]models.SkuMapping {
	mappingsBySupplier := make(map[uuid.UUID][]models.SkuMapping)
	if len(suppliers) == 0 {
		return mappingsBySupplier
	}

	for i, catalog := range catalogs {
		primary := suppliers[i%len(suppliers)]
		secondary := suppliers[(i+1)%len(suppliers)]
		seedSkuMapping(db, mappingsBySupplier, primary, catalog, i, true)
		if secondary.ID != primary.ID {
			seedSkuMapping(db, mappingsBySupplier, secondary, catalog, i, false)
		}
	}

	return mappingsBySupplier
}

func seedSkuMapping(db *gorm.DB, mappingsBySupplier map[uuid.UUID][]models.SkuMapping, supplier models.Supplier, catalog PartCatalogData, catalogIndex int, primary bool) {
	description := catalog.Description
	if description == "" {
		description = catalog.PartNumber
	}

	basePrice := 18.50 + float64(catalogIndex)*4.75
	if !primary {
		basePrice += 6.25
	}

	mapping := models.SkuMapping{
		ID:           uuid.NewMD5(uuid.Nil, []byte(fmt.Sprintf("sku-mapping|%s|%s", supplier.ID.String(), catalog.PartNumber))),
		SupplierID:   supplier.ID,
		SKU:          catalog.PartNumber,
		Name:         description,
		UnitPrice:    math.Round(basePrice*100) / 100,
		LeadTimeDays: supplier.LeadTimeDays + 2 + catalogIndex%4,
		MinOrderQty:  10 + (catalogIndex%5)*5,
	}

	db.FirstOrCreate(&mapping, models.SkuMapping{ID: mapping.ID})
	mappingsBySupplier[supplier.ID] = append(mappingsBySupplier[supplier.ID], mapping)
}
