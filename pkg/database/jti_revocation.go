package database

import "time"

type JtiRevocation struct {
	ID        string    `gorm:"type:uuid;primaryKey"`
	ExpiresAt time.Time `gorm:"type:timestamp;not null"`
}
