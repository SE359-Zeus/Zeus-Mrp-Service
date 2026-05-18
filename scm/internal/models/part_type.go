package models

type PartType struct {
	ID           int32   `gorm:"primaryKey;autoIncrement:false"`
	PartTypeName string  `gorm:"type:varchar;uniqueIndex;not null"`
	Description  *string `gorm:"type:text"`
}
