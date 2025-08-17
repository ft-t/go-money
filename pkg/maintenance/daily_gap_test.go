package maintenance_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/cockroachdb/errors"
	"github.com/ft-t/go-money/pkg/configuration"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/ft-t/go-money/pkg/maintenance"
	"github.com/ft-t/go-money/pkg/testingutils"
	"github.com/ft-t/go-money/pkg/transactions"
	"github.com/golang/mock/gomock"
	"github.com/jinzhu/now"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

var gormDB *gorm.DB
var cfg *configuration.Configuration

func TestMain(m *testing.M) {
	cfg = configuration.GetConfiguration()
	gormDB = database.GetDb(database.DbTypeMaster)

	os.Exit(m.Run())
}

func TestDailyGap(t *testing.T) {
	t.Run("one day ago. daily stat exists", func(t *testing.T) {
		days := []int{1, 7}

		for _, day := range days {
			t.Run(fmt.Sprintf("with %v days", day), func(t *testing.T) {

				assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

				statSvc := NewMockStatsSvc(gomock.NewController(t))
				srv := maintenance.NewService(&maintenance.Config{
					StatsSvc: statSvc,
				})

				account := &database.Account{
					Extra: make(map[string]string),
				}
				assert.NoError(t, gormDB.Create(account).Error)

				lastDate := time.Now().AddDate(0, 0, -day)
				assert.NoError(t, gormDB.Create(&database.DailyStat{
					AccountID: account.ID,
					Date:      lastDate,
					Amount:    decimal.NewFromInt(200),
				}).Error)

				statSvc.EXPECT().CalculateDailyStat(gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, db *gorm.DB, request transactions.CalculateDailyStatRequest) error {
						assert.EqualValues(t, account.ID, request.AccountID)
						assert.EqualValues(t, now.With(lastDate.UTC()).BeginningOfDay(), request.StartDate)
						return nil
					})

				assert.NoError(t, srv.FixDailyGaps(context.TODO()))
			})
		}
	})

	t.Run("no latest daily stat. latest transaction exists", func(t *testing.T) {
		assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

		statSvc := NewMockStatsSvc(gomock.NewController(t))
		srv := maintenance.NewService(&maintenance.Config{
			StatsSvc: statSvc,
		})

		account := &database.Account{
			Extra: make(map[string]string),
		}
		assert.NoError(t, gormDB.Create(account).Error)

		lastDate := time.Now().AddDate(0, 0, -1).UTC()

		assert.NoError(t, gormDB.Create(&database.Transaction{
			SourceAccountID:     account.ID,
			TransactionDateOnly: lastDate,
			TransactionDateTime: lastDate,
			Extra:               make(map[string]string),
		}).Error)

		statSvc.EXPECT().CalculateDailyStat(gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, db *gorm.DB, request transactions.CalculateDailyStatRequest) error {
				assert.EqualValues(t, account.ID, request.AccountID)
				assert.WithinDuration(t, lastDate, request.StartDate, time.Second)
				return nil
			})

		assert.NoError(t, srv.FixDailyGaps(context.TODO()))
	})

	t.Run("no latest daily stat. no latest transaction. account crated", func(t *testing.T) {
		assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

		statSvc := NewMockStatsSvc(gomock.NewController(t))
		srv := maintenance.NewService(&maintenance.Config{
			StatsSvc: statSvc,
		})

		createdAt := time.Now().AddDate(0, 0, -55).UTC()
		account := &database.Account{
			Extra:     make(map[string]string),
			CreatedAt: createdAt,
		}
		assert.NoError(t, gormDB.Create(account).Error)

		statSvc.EXPECT().CalculateDailyStat(gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, db *gorm.DB, request transactions.CalculateDailyStatRequest) error {
				assert.EqualValues(t, account.ID, request.AccountID)
				assert.WithinDuration(t, createdAt, request.StartDate, time.Second)
				return nil
			})

		assert.NoError(t, srv.FixDailyGaps(context.TODO()))
	})

	t.Run("calculate fail", func(t *testing.T) {
		assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

		statSvc := NewMockStatsSvc(gomock.NewController(t))
		srv := maintenance.NewService(&maintenance.Config{
			StatsSvc: statSvc,
		})

		createdAt := time.Now().AddDate(0, 0, -55).UTC()
		account := &database.Account{
			Extra:     make(map[string]string),
			CreatedAt: createdAt,
		}
		assert.NoError(t, gormDB.Create(account).Error)

		statSvc.EXPECT().CalculateDailyStat(gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, db *gorm.DB, request transactions.CalculateDailyStatRequest) error {
				return errors.New("failed to calculate daily stat")
			})

		assert.ErrorContains(t, srv.FixDailyGaps(context.TODO()), "failed to calculate daily stat")
	})

	t.Run("db error on account list", func(t *testing.T) {
		statSvc := NewMockStatsSvc(gomock.NewController(t))
		srv := maintenance.NewService(&maintenance.Config{
			StatsSvc: statSvc,
		})

		mockGorm, _, sql := testingutils.GormMock()
		sql.ExpectBegin()
		sql.ExpectQuery("SELECT .*").WillReturnError(errors.New("failed to list accounts"))
		sql.ExpectRollback()

		ctx := database.WithContext(context.TODO(), mockGorm)
		err := srv.FixDailyGaps(ctx)
		assert.ErrorContains(t, err, "failed to list accounts")
	})

	t.Run("db error on get daily stats", func(t *testing.T) {
		statSvc := NewMockStatsSvc(gomock.NewController(t))
		srv := maintenance.NewService(&maintenance.Config{
			StatsSvc: statSvc,
		})

		mockGorm, _, sql := testingutils.GormMock()

		rows := sqlmock.NewRows([]string{"id", "name"}).
			AddRow(1, "account")

		sql.ExpectBegin()
		sql.ExpectQuery("SELECT .*").
			WillReturnRows(rows)

		sql.ExpectQuery("SELECT \\* FROM \"daily_stat\" WHERE .*").
			WillReturnError(errors.New("failed to get daily stats"))
		sql.ExpectRollback()

		ctx := database.WithContext(context.TODO(), mockGorm)
		err := srv.FixDailyGaps(ctx)
		assert.ErrorContains(t, err, "failed to get daily stats")
	})

	t.Run("db error on transaction lookup", func(t *testing.T) {
		statSvc := NewMockStatsSvc(gomock.NewController(t))
		srv := maintenance.NewService(&maintenance.Config{
			StatsSvc: statSvc,
		})

		mockGorm, _, sql := testingutils.GormMock()

		rows := sqlmock.NewRows([]string{"id", "name"}).
			AddRow(1, "account")

		sql.ExpectBegin()
		sql.ExpectQuery("SELECT .*").
			WillReturnRows(rows)
		sql.ExpectQuery("SELECT .*").
			WillReturnRows(sqlmock.NewRows([]string{"account_id"})) //empty set

		sql.ExpectQuery("SELECT \\* FROM \"transactions\" WHERE .*").
			WillReturnError(errors.New("failed to lookup transactions"))
		sql.ExpectRollback()

		ctx := database.WithContext(context.TODO(), mockGorm)
		err := srv.FixDailyGaps(ctx)
		assert.ErrorContains(t, err, "failed to lookup transactions")
	})

	t.Run("db error on commit", func(t *testing.T) {
		statSvc := NewMockStatsSvc(gomock.NewController(t))
		srv := maintenance.NewService(&maintenance.Config{
			StatsSvc: statSvc,
		})

		mockGorm, _, sql := testingutils.GormMock()

		rows := sqlmock.NewRows([]string{"id", "name"})

		sql.ExpectBegin()
		sql.ExpectQuery("SELECT .*").
			WillReturnRows(rows)

		sql.ExpectCommit().WillReturnError(errors.New("failed to commit transaction"))
		sql.ExpectRollback()

		ctx := database.WithContext(context.TODO(), mockGorm)
		err := srv.FixDailyGaps(ctx)
		assert.ErrorContains(t, err, "failed to commit transaction")
	})
}
