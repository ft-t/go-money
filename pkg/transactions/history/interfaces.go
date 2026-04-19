package history

import (
	"context"

	"github.com/ft-t/go-money/pkg/database"
	"gorm.io/gorm"
)

//go:generate mockgen -destination interfaces_mocks_test.go -package history_test -source=interfaces.go

type Recorder interface {
	Record(ctx context.Context, tx *gorm.DB, req RecordRequest) error
}

type Reader interface {
	List(ctx context.Context, transactionID int64) ([]*database.TransactionHistory, error)
}
