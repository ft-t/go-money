package accounts_test

import (
	"context"
	accountsv1 "github.com/ft-t/go-money-pb/gen/gomoneypb/accounts/v1"
	v1 "github.com/ft-t/go-money-pb/gen/gomoneypb/v1"
	"github.com/ft-t/go-money/pkg/accounts"
	"github.com/ft-t/go-money/pkg/configuration"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/ft-t/go-money/pkg/testingutils"
	"github.com/golang/mock/gomock"
	"github.com/shopspring/decimal"
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

func TestCreateAccount(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

		mapper := NewMockMapperSvc(gomock.NewController(t))

		srv := accounts.NewService(&accounts.ServiceConfig{
			MapperSvc: mapper,
		})

		mapper.EXPECT().MapAccount(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, account *database.Account) *v1.Account {
				assert.NotEmpty(t, account.ID)

				return &v1.Account{
					Id: account.ID,
				}
			})

		resp, err := srv.Create(context.TODO(), &accountsv1.CreateAccountRequest{
			Account: &v1.Account{
				Name:            "some-account",
				Currency:        "USD",
				CurrencyBalance: "100.001234122",
				Extra: map[string]string{
					"a": "b",
				},
				Type: v1.AccountType_ACCOUNT_TYPE_REGULAR,
				Note: "some note",
			},
		})

		assert.NoError(t, err)
		assert.NotNil(t, resp)

		var rec database.Account
		assert.NoError(t, gormDB.First(&rec).Error)

		assert.EqualValues(t, rec.ID, resp.Account.Id)
		assert.EqualValues(t, "some-account", rec.Name)
		assert.EqualValues(t, "USD", rec.Currency)
		assert.EqualValues(t, "100.001234122000", rec.CurrentBalance.StringFixed(12))
		assert.EqualValues(t, map[string]string{"a": "b"}, rec.Extra)
		assert.EqualValues(t, v1.AccountType_ACCOUNT_TYPE_REGULAR, rec.Type)
		assert.EqualValues(t, "some note", rec.Note)
	})

	t.Run("invalid balance format", func(t *testing.T) {
		assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

		srv := accounts.NewService(&accounts.ServiceConfig{})

		resp, err := srv.Create(context.TODO(), &accountsv1.CreateAccountRequest{
			Account: &v1.Account{
				Name:            "some-account",
				Currency:        "USD",
				CurrencyBalance: "100-00",
				Extra: map[string]string{
					"a": "b",
				},
				Type: v1.AccountType_ACCOUNT_TYPE_REGULAR,
				Note: "some note",
			},
		})

		assert.Nil(t, resp)
		assert.ErrorContains(t, err, "can't convert 100-00 to decimal")
	})
}

func TestUpdate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mapper := NewMockMapperSvc(gomock.NewController(t))
		srv := accounts.NewService(&accounts.ServiceConfig{
			MapperSvc: mapper,
		})

		acc := database.Account{
			Currency:       "USD",
			Name:           "xxx",
			Extra:          map[string]string{},
			CurrentBalance: decimal.RequireFromString("111"),
		}
		assert.NoError(t, gormDB.Create(&acc).Error)

		mapper.EXPECT().MapAccount(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, account *database.Account) *v1.Account {
				assert.NotEmpty(t, account.ID)
				assert.EqualValues(t, "yy", account.Name)

				return &v1.Account{
					Id: account.ID,
				}
			})

		resp, err := srv.Update(context.TODO(), &accountsv1.UpdateAccountRequest{
			Id:   acc.ID,
			Name: "yy",
			Extra: map[string]string{
				"updated": "b",
			},
			Type: v1.AccountType_ACCOUNT_TYPE_REGULAR,
			Note: "updated note",
		})

		assert.NoError(t, err)
		assert.NotNil(t, resp)

		var rec database.Account
		assert.NoError(t, gormDB.First(&rec).Error)

		assert.EqualValues(t, rec.ID, resp.Account.Id)
		assert.EqualValues(t, "yy", rec.Name)
		assert.EqualValues(t, "USD", rec.Currency)
		assert.EqualValues(t, map[string]string{"updated": "b"}, rec.Extra)
		assert.EqualValues(t, v1.AccountType_ACCOUNT_TYPE_REGULAR, rec.Type)
		assert.EqualValues(t, "updated note", rec.Note)
		assert.EqualValues(t, "111", rec.CurrentBalance.String())
	})

	t.Run("success empty extra", func(t *testing.T) {
		assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

		mapper := NewMockMapperSvc(gomock.NewController(t))
		srv := accounts.NewService(&accounts.ServiceConfig{
			MapperSvc: mapper,
		})

		acc := database.Account{
			Currency:       "USD",
			Name:           "xxx",
			Extra:          map[string]string{},
			CurrentBalance: decimal.RequireFromString("111"),
		}
		assert.NoError(t, gormDB.Create(&acc).Error)

		mapper.EXPECT().MapAccount(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, account *database.Account) *v1.Account {
				assert.NotEmpty(t, account.ID)
				assert.EqualValues(t, "yy", account.Name)

				return &v1.Account{
					Id: account.ID,
				}
			})

		resp, err := srv.Update(context.TODO(), &accountsv1.UpdateAccountRequest{
			Id:    acc.ID,
			Name:  "yy",
			Extra: nil,
			Type:  v1.AccountType_ACCOUNT_TYPE_REGULAR,
			Note:  "updated note",
		})

		assert.NoError(t, err)
		assert.NotNil(t, resp)

		var rec database.Account
		assert.NoError(t, gormDB.First(&rec).Error)

		assert.EqualValues(t, rec.ID, resp.Account.Id)
		assert.EqualValues(t, "yy", rec.Name)
		assert.EqualValues(t, "USD", rec.Currency)
		assert.NotNil(t, rec.Extra)
		assert.Len(t, rec.Extra, 0)
		assert.EqualValues(t, v1.AccountType_ACCOUNT_TYPE_REGULAR, rec.Type)
		assert.EqualValues(t, "updated note", rec.Note)
		assert.EqualValues(t, "111", rec.CurrentBalance.String())
	})
}
