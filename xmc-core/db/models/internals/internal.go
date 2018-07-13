package internals

import "time"

type Internal struct {
	ID                     uint `gorm:"primary_key;auto_increment:false;default:0"`
	LastPageEvent          time.Time
	LastPageChildrenUpdate time.Time
}

func (Internal) TableName() string {
	return "internals"
}
