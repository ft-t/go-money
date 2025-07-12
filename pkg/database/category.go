package database

import (
	"gorm.io/gorm"
	"time"
)

type Category struct {
	ID        int32 `gorm:"primaryKey"`
	Name      string
	CreatedAt time.Time

	DeletedAt gorm.DeletedAt
}

func (*Category) TableName() string {
	return "categories"
}
