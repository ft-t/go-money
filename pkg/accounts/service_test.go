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
	"github.com/samber/lo"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"os"
	"testing"
	"time"
)

var gormDB *gorm.DB
var cfg *configuration.Configuration

func TestMain(m *testing.M) {
	cfg = configuration.GetConfiguration()
	gormDB = database.GetDb(database.DbTypeMaster)

	os.Exit(m.Run())
}

func TestDelete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

		acc := &database.Account{
			Extra: map[string]string{},
		}
		assert.NoError(t, gormDB.Create(acc).Error)

		mapper := NewMockMapperSvc(gomock.NewController(t))
		mapper.EXPECT().MapAccount(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, account *database.Account) *v1.Account {
				assert.EqualValues(t, acc.ID, account.ID)

				return &v1.Account{
					Id: account.ID,
				}
			})

		srv := accounts.NewService(&accounts.ServiceConfig{
			MapperSvc: mapper,
		})

		resp, err := srv.Delete(context.TODO(), &accountsv1.DeleteAccountRequest{
			Id: acc.ID,
		})

		assert.NoError(t, err)
		assert.EqualValues(t, acc.ID, resp.Account.Id)

		var rec database.Account
		assert.NoError(t, gormDB.Unscoped().First(&rec).Error)

		assert.True(t, rec.DeletedAt.Valid)
		assert.NotEmpty(t, rec.DeletedAt.Time)
	})

	t.Run("acc not found", func(t *testing.T) {
		assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

		srv := accounts.NewService(&accounts.ServiceConfig{})

		resp, err := srv.Delete(context.TODO(), &accountsv1.DeleteAccountRequest{
			Id: 999,
		})

		assert.ErrorContains(t, err, "account not found")
		assert.Nil(t, resp)
	})
}

func TestList(t *testing.T) {
	assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

	acc := &database.Account{
		Extra:    map[string]string{},
		Position: 1,
	}
	assert.NoError(t, gormDB.Create(acc).Error)

	accDeleted := &database.Account{
		Extra: map[string]string{},
		DeletedAt: gorm.DeletedAt{
			Valid: true,
			Time:  time.Now().UTC(),
		},
		Position: 999,
	}
	assert.NoError(t, gormDB.Create(accDeleted).Error)

	mapper := NewMockMapperSvc(gomock.NewController(t))

	mapper.EXPECT().MapAccount(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, account *database.Account) *v1.Account {
			assert.EqualValues(t, accDeleted.ID, account.ID)

			return &v1.Account{
				Id: account.ID,
			}
		})

	mapper.EXPECT().MapAccount(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, account *database.Account) *v1.Account {
			assert.EqualValues(t, acc.ID, account.ID)

			return &v1.Account{
				Id: account.ID,
			}
		})

	srv := accounts.NewService(&accounts.ServiceConfig{
		MapperSvc: mapper,
	})

	resp, err := srv.List(context.TODO(), &accountsv1.ListAccountsRequest{})

	assert.NoError(t, err)
	assert.Len(t, resp.Accounts, 2)

	assert.EqualValues(t, accDeleted.ID, resp.Accounts[0].Account.Id)
	assert.EqualValues(t, acc.ID, resp.Accounts[1].Account.Id)
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
			Name:     "some-account",
			Currency: "USD",
			Extra: map[string]string{
				"a": "b",
			},
			Type:             v1.AccountType_ACCOUNT_TYPE_REGULAR,
			Note:             "some note",
			LiabilityPercent: nil,
			Iban:             "some-iban",
			AccountNumber:    "some-account-number",
		})

		assert.NoError(t, err)
		assert.NotNil(t, resp)

		var rec database.Account
		assert.NoError(t, gormDB.First(&rec).Error)

		assert.EqualValues(t, rec.ID, resp.Account.Id)
		assert.EqualValues(t, "some-account", rec.Name)
		assert.EqualValues(t, "USD", rec.Currency)
		assert.EqualValues(t, "0.000000000000", rec.CurrentBalance.StringFixed(12))
		assert.EqualValues(t, map[string]string{"a": "b"}, rec.Extra)
		assert.EqualValues(t, v1.AccountType_ACCOUNT_TYPE_REGULAR, rec.Type)
		assert.EqualValues(t, "some note", rec.Note)
	})

	t.Run("invalid liability format", func(t *testing.T) {
		assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

		srv := accounts.NewService(&accounts.ServiceConfig{})

		resp, err := srv.Create(context.TODO(), &accountsv1.CreateAccountRequest{
			Name:             "some-account",
			Currency:         "USD",
			LiabilityPercent: lo.ToPtr("100-00"),
			Extra: map[string]string{
				"a": "b",
			},
			Type: v1.AccountType_ACCOUNT_TYPE_REGULAR,
			Note: "some note",
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
			Type:             v1.AccountType_ACCOUNT_TYPE_REGULAR,
			Note:             "updated note",
			LiabilityPercent: lo.ToPtr("12"),
			Iban:             "iban",
			AccountNumber:    "num",
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
		assert.EqualValues(t, "12", rec.LiabilityPercent.Decimal.String())
		assert.EqualValues(t, "iban", rec.Iban)
		assert.EqualValues(t, "num", rec.AccountNumber)
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

	t.Run("invalid liability format", func(t *testing.T) {
		srv := accounts.NewService(&accounts.ServiceConfig{})

		acc := database.Account{
			Currency:       "USD",
			Name:           "xxx",
			Extra:          map[string]string{},
			CurrentBalance: decimal.RequireFromString("111"),
		}
		assert.NoError(t, gormDB.Create(&acc).Error)

		resp, err := srv.Update(context.TODO(), &accountsv1.UpdateAccountRequest{
			Id:   acc.ID,
			Name: "yy",
			Extra: map[string]string{
				"updated": "b",
			},
			Type:             v1.AccountType_ACCOUNT_TYPE_REGULAR,
			Note:             "updated note",
			LiabilityPercent: lo.ToPtr("-1-2"),
			Iban:             "iban",
			AccountNumber:    "num",
		})

		assert.Nil(t, resp)
		assert.ErrorContains(t, err, "can't convert -1-2 to decimal")
	})
}
