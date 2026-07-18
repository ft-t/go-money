package database

import (
	importv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/import/v1"
	"time"
)

type ImportIgnoredTransaction struct {
	ImportSource importv1.ImportSource
	RefKey       string
	Reason       *string
	CreatedAt    time.Time
}

func (*ImportIgnoredTransaction) TableName() string {
	return "import_ignored_transactions"
}
