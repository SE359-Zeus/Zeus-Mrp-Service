package models

type ShipmentState struct {
	ID   int32  `gorm:"primaryKey;autoIncrement:false"`
	Name string `gorm:"type:varchar;not null;uniqueIndex"`
}
