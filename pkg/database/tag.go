package database

import (
	"gorm.io/gorm"
	"time"
)

type Tag struct {
	ID        int32
	Name      string
	Color     string
	Icon      string
	DeletedAt gorm.DeletedAt
	CreatedAt time.Time
}

func (Tag) TableName() string {
	return "tags"
}
