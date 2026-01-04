package database

import "time"

type JtiRevocation struct {
	ID        string    `gorm:"type:text;primaryKey"`
	ExpiresAt time.Time `gorm:"type:timestamp;not null"`
}

func (JtiRevocation) TableName() string {
	return "jti_revocations"
}
