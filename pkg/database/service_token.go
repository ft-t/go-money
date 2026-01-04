package database

import (
	"time"

	"gorm.io/gorm"
)

type ServiceToken struct {
	ID string `gorm:"type:uuid;primaryKey"`

	Name      string         `gorm:"type:text;not null"`
	ExpiresAt time.Time      `gorm:"type:timestamp;not null"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
	CreatedAt time.Time      `gorm:"type:timestamp;not null"`
}
