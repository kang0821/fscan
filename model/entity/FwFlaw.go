package entity

type FwFlaw struct {
	ID     int32 `gorm:"primaryKey"`
	NAME   string
	CONFIG string
}
