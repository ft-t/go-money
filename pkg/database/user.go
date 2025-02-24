package database

import (
	"gorm.io/gorm"
	"time"
)

type User struct {
	ID       int32
	Login    string
	Password string

	CreatedAt time.Time
	DeletedAt gorm.DeletedAt
}
