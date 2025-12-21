package database

import (
	"time"

	"gorm.io/gorm"
)

type AppConfig struct {
	ID        string
	Value     string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt
}

func (AppConfig) TableName() string {
	return "app_configs"
}
