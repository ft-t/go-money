package database

import (
	importv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/import/v1"
	"time"
)

type ImportDeduplication struct {
	ImportSource  importv1.ImportSource
	Key           string
	CreatedAt     time.Time
	TransactionID int64
}

func (*ImportDeduplication) TableName() string {
	return "import_deduplication"
}
