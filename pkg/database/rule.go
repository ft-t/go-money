package database

import (
	"gorm.io/gorm"
	"time"
)

type RuleInterpreter string

type Rule struct {
	ID          int32
	Script      string
	Interpreter RuleInterpreter
	SortOrder   int32
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Enabled     bool
	IsFinalRule bool
	DeletedAt   gorm.DeletedAt
}
