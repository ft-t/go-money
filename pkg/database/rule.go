package database

import (
	gomoneypbv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/v1"
	"gorm.io/gorm"
	"time"
)

type Rule struct {
	ID              int32
	Title           string
	Script          string
	InterpreterType gomoneypbv1.RuleInterpreterType
	SortOrder       int32
	CreatedAt       time.Time
	UpdatedAt       time.Time
	Enabled         bool
	IsFinalRule     bool
	DeletedAt       gorm.DeletedAt
	GroupName       string
}
