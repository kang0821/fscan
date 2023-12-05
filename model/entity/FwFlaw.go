package entity

type FwFlaw struct {
	ID      int32 `gorm:"primaryKey"`
	CODE    string
	NAME    string
	CONFIG  string
	DELETED bool
}
