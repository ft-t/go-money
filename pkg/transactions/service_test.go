package transactions_test

import (
	"context"
	transactionsv1 "github.com/ft-t/go-money-pb/gen/gomoneypb/transactions/v1"
	"github.com/ft-t/go-money/pkg/configuration"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/ft-t/go-money/pkg/testingutils"
	"github.com/ft-t/go-money/pkg/transactions"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"os"
	"testing"
)

var gormDB *gorm.DB
var cfg *configuration.Configuration

func TestMain(m *testing.M) {
	cfg = configuration.GetConfiguration()
	gormDB = database.GetDb(database.DbTypeMaster)

	os.Exit(m.Run())
}

func TestCreateWithdrawal(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		assert.NoError(t, testingutils.FlushAllTables(cfg.Db))
		srv := transactions.NewService()

		resp, err := srv.Create(context.TODO(), &transactionsv1.CreateTransactionRequest{
			Notes:           "",
			Extra:           nil,
			LabelIds:        nil,
			TransactionDate: nil,
			Transaction:     &transactionsv1.CreateTransactionRequest_Withdrawal{},
		})
		assert.NoError(t, err)
		assert.NotNil(t, resp)
	})
}
