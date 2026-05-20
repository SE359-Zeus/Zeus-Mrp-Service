package seeder

import (
	"fmt"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
	"zeus-scm-service/internal/models"
)

func seedCatalogs(db *gorm.DB, data *PartsFile) (map[string]int32, map[string]models.PartCatalog) {
	typeMap := make(map[string]int32)
	catMap := make(map[string]models.PartCatalog)

	for i, pt := range data.PartTypes {
		id := int32(i + 1)
		desc := pt.Description
		db.FirstOrCreate(&models.PartType{ID: id, PartTypeName: pt.CommodityType, Description: &desc}, models.PartType{PartTypeName: pt.CommodityType})
		typeMap[pt.CommodityType] = id
	}

	for _, pc := range data.PartCatalogs {
		tid, ok := typeMap[pc.CommodityType]
		if !ok {
			continue
		}
		desc := pc.Description
		cat := models.PartCatalog{
			ID:            uuid.New(),
			PartNumber:    pc.PartNumber,
			PartTypesID:   tid,
			MfgNumber:     pc.MfgNumber,
			Description:   &desc,
			PartMfgStatus: 1, // pending
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}
		db.Create(&cat)
		key := fmt.Sprintf("%s|%s", pc.PartNumber, pc.MfgNumber)
		catMap[key] = cat
	}
	return typeMap, catMap
}
