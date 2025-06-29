package database

import (
	"gorm.io/gorm"
	"time"
)

type RuleInterpreter int32

type Rule struct {
	ID              int32
	Title           string
	Script          string
	InterpreterType RuleInterpreter
	SortOrder       int32
	CreatedAt       time.Time
	UpdatedAt       time.Time
	Enabled         bool
	IsFinalRule     bool
	DeletedAt       gorm.DeletedAt
	GroupName       string
}
