package database

import "gorm.io/gorm"

type Tag struct {
	ID        int32
	Name      string
	Color     string
	Icon      string
	DeletedAt gorm.DeletedAt
}

func (Tag) TableName() string {
	return "tags"
}
