package accounts_test

import (
	"context"
	"math"
	"os"
	"testing"
	"time"

	accountsv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/accounts/v1"
	v1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/v1"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/cockroachdb/errors"
	"github.com/ft-t/go-money/pkg/accounts"
	"github.com/ft-t/go-money/pkg/configuration"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/ft-t/go-money/pkg/testingutils"
	"github.com/golang/mock/gomock"
	"github.com/samber/lo"
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

func TestService_GetAccountByID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

		acc := &database.Account{
			Extra: map[string]string{},
		}
		assert.NoError(t, gormDB.Create(&acc).Error)

		srv := accounts.NewService(&accounts.ServiceConfig{})

		resp, err := srv.GetAccountByID(context.TODO(), acc.ID)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
	})

	t.Run("not found", func(t *testing.T) {
		assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

		srv := accounts.NewService(&accounts.ServiceConfig{})

		resp, err := srv.GetAccountByID(context.TODO(), int32(math.MaxInt32))
		assert.ErrorContains(t, err, "failed to fetch account by id")
		assert.Nil(t, resp)
	})
}

func TestDelete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

		defaultAcc := &database.Account{
			Extra: map[string]string{},
			Flags: database.AccountFlagIsDefault,
			Type:  v1.AccountType_ACCOUNT_TYPE_ASSET,
		}
		assert.NoError(t, gormDB.Create(defaultAcc).Error)

		acc := &database.Account{
			Extra: map[string]string{},
			Type:  v1.AccountType_ACCOUNT_TYPE_ASSET,
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
		assert.NoError(t, gormDB.Unscoped().Where("id = ?", acc.ID).First(&rec).Error)

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

	t.Run("db error", func(t *testing.T) {
		mockGorm, _, sql := testingutils.GormMock()

		srv := accounts.NewService(&accounts.ServiceConfig{})

		rows := sqlmock.NewRows([]string{"id", "name"}).
			AddRow(999, "some-account")

		sql.ExpectBegin()
		sql.ExpectQuery(`SELECT \* FROM .*`).
			WillReturnRows(rows)

		sql.ExpectExec("UPDATE .*").
			WillReturnError(errors.New("failed to delete account"))

		sql.ExpectRollback()

		ctx := database.WithContext(context.TODO(), mockGorm)

		resp, err := srv.Delete(ctx, &accountsv1.DeleteAccountRequest{
			Id: 999,
		})

		assert.ErrorContains(t, err, "failed to delete account")
		assert.Nil(t, resp)
	})

	t.Run("db error - fail commit", func(t *testing.T) {
		mockGorm, _, sql := testingutils.GormMock()

		srv := accounts.NewService(&accounts.ServiceConfig{})

		rows := sqlmock.NewRows([]string{"id", "name"}).
			AddRow(999, "some-account")

		sql.ExpectBegin()
		sql.ExpectQuery(`SELECT \* FROM .*`).
			WillReturnRows(rows)

		sql.ExpectExec("UPDATE .*").
			WillReturnResult(sqlmock.NewResult(1, 1))
		sql.ExpectQuery("select count.*").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		sql.ExpectCommit().WillReturnError(errors.New("failed to commit transaction"))

		ctx := database.WithContext(context.TODO(), mockGorm)

		resp, err := srv.Delete(ctx, &accountsv1.DeleteAccountRequest{
			Id: 999,
		})

		assert.ErrorContains(t, err, "failed to commit transaction")
		assert.Nil(t, resp)
	})

	t.Run("cannot remove default", func(t *testing.T) {
		assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

		defaultAcc := &database.Account{
			Extra: map[string]string{},
			Flags: database.AccountFlagIsDefault,
			Type:  v1.AccountType_ACCOUNT_TYPE_ASSET,
		}
		assert.NoError(t, gormDB.Create(defaultAcc).Error)

		srv := accounts.NewService(&accounts.ServiceConfig{})

		resp, err := srv.Delete(context.TODO(), &accountsv1.DeleteAccountRequest{
			Id: defaultAcc.ID,
		})

		assert.ErrorContains(t, err, "at least one default account is required")
		assert.Nil(t, resp)
	})
}

func TestGetAllAccounts(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

		acc := &database.Account{
			Extra:        map[string]string{},
			DisplayOrder: lo.ToPtr(int32(1)),
		}
		assert.NoError(t, gormDB.Create(acc).Error)

		srv := accounts.NewService(&accounts.ServiceConfig{})

		resp, err := srv.GetAllAccounts(context.TODO())

		assert.NoError(t, err)
		assert.Len(t, resp, 1)
	})

	t.Run("db fail", func(t *testing.T) {
		mockGorm, _, sql := testingutils.GormMock()

		srv := accounts.NewService(&accounts.ServiceConfig{})

		sql.ExpectQuery("SELECT \\* FROM `accounts`.*").
			WillReturnError(errors.New("failed to fetch accounts"))

		ctx := database.WithContext(context.TODO(), mockGorm)

		resp, err := srv.GetAllAccounts(ctx)

		assert.ErrorContains(t, err, "failed to fetch accounts")
		assert.Nil(t, resp)
	})
}

func TestList(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

		acc := &database.Account{
			Extra:        map[string]string{},
			DisplayOrder: lo.ToPtr(int32(1)),
		}
		assert.NoError(t, gormDB.Create(acc).Error)

		accDeleted := &database.Account{
			Extra: map[string]string{},
			DeletedAt: gorm.DeletedAt{
				Valid: true,
				Time:  time.Now().UTC(),
			},
			DisplayOrder: lo.ToPtr(int32(999)),
		}
		assert.NoError(t, gormDB.Create(accDeleted).Error)

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

		resp, err := srv.List(context.TODO(), &accountsv1.ListAccountsRequest{
			Ids: []int32{acc.ID, accDeleted.ID},
		})

		assert.NoError(t, err)
		assert.Len(t, resp.Accounts, 1)

		assert.EqualValues(t, acc.ID, resp.Accounts[0].Account.Id)
	})

	t.Run("include deleted", func(t *testing.T) {
		assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

		acc := &database.Account{
			Extra:        map[string]string{},
			DisplayOrder: lo.ToPtr(int32(1)),
		}
		assert.NoError(t, gormDB.Create(acc).Error)

		accDeleted := &database.Account{
			Extra: map[string]string{},
			DeletedAt: gorm.DeletedAt{
				Valid: true,
				Time:  time.Now().UTC(),
			},
			DisplayOrder: lo.ToPtr(int32(999)),
		}
		assert.NoError(t, gormDB.Create(accDeleted).Error)

		mapper := NewMockMapperSvc(gomock.NewController(t))

		mapper.EXPECT().MapAccount(gomock.Any(), gomock.Any()).
			Times(2).
			DoAndReturn(func(ctx context.Context, account *database.Account) *v1.Account {
				return &v1.Account{
					Id: account.ID,
				}
			})

		srv := accounts.NewService(&accounts.ServiceConfig{
			MapperSvc: mapper,
		})

		resp, err := srv.List(context.TODO(), &accountsv1.ListAccountsRequest{
			IncludeDeleted: true,
		})

		assert.NoError(t, err)
		assert.Len(t, resp.Accounts, 2)

		assert.EqualValues(t, acc.ID, resp.Accounts[0].Account.Id)
		assert.EqualValues(t, accDeleted.ID, resp.Accounts[1].Account.Id)
	})

	t.Run("db fail", func(t *testing.T) {
		mockGorm, _, sql := testingutils.GormMock()

		srv := accounts.NewService(&accounts.ServiceConfig{})

		sql.ExpectQuery("SELECT \\* FROM .*").
			WillReturnError(errors.New("failed to fetch accounts"))

		ctx := database.WithContext(context.TODO(), mockGorm)

		resp, err := srv.List(ctx, &accountsv1.ListAccountsRequest{})

		assert.ErrorContains(t, err, "failed to fetch accounts")
		assert.Nil(t, resp)
	})
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
			Type:             v1.AccountType_ACCOUNT_TYPE_ASSET,
			Note:             "some note",
			LiabilityPercent: nil,
			Iban:             "some-iban",
			AccountNumber:    "some-account-number",
			Flags:            database.AccountFlagIsDefault,
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
		assert.EqualValues(t, v1.AccountType_ACCOUNT_TYPE_ASSET, rec.Type)
		assert.EqualValues(t, "some note", rec.Note)
	})

	t.Run("no default account", func(t *testing.T) {
		assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

		srv := accounts.NewService(&accounts.ServiceConfig{})

		resp, err := srv.Create(context.TODO(), &accountsv1.CreateAccountRequest{
			Name:     "some-account",
			Currency: "USD",
			Extra: map[string]string{
				"a": "b",
			},
			Type:             v1.AccountType_ACCOUNT_TYPE_ASSET,
			Note:             "some note",
			LiabilityPercent: nil,
			Iban:             "some-iban",
			AccountNumber:    "some-account-number",
		})

		assert.ErrorContains(t, err, "at least one default account is required")
		assert.Nil(t, resp)
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
			Type: v1.AccountType_ACCOUNT_TYPE_ASSET,
			Note: "some note",
		})

		assert.Nil(t, resp)
		assert.ErrorContains(t, err, "can't convert 100-00 to decimal")
	})

	t.Run("db error", func(t *testing.T) {
		mockGorm, _, sql := testingutils.GormMock()

		srv := accounts.NewService(&accounts.ServiceConfig{})

		sql.ExpectBegin()
		sql.ExpectExec("INSERT INTO `accounts`.*").
			WillReturnError(errors.New("failed to create account"))
		sql.ExpectRollback()

		ctx := database.WithContext(context.TODO(), mockGorm)

		resp, err := srv.Create(ctx, &accountsv1.CreateAccountRequest{
			Name:     "some-account",
			Currency: "USD",
			Extra: map[string]string{
				"a": "b",
			},
			Type:             v1.AccountType_ACCOUNT_TYPE_ASSET,
			Note:             "some note",
			LiabilityPercent: nil,
			Iban:             "some-iban",
			AccountNumber:    "some-account-number",
			Flags:            database.AccountFlagIsDefault,
		})

		assert.ErrorContains(t, err, "failed to create account")
		assert.Nil(t, resp)
	})

	t.Run("db error - commit fail", func(t *testing.T) {
		mockGorm, _, sql := testingutils.GormMock()

		srv := accounts.NewService(&accounts.ServiceConfig{})

		sql.ExpectBegin()
		sql.ExpectQuery("INSERT INTO.*").
			WillReturnRows(sql.NewRows([]string{"id"}).AddRow(1))
		sql.ExpectExec("update accounts .*").WillReturnResult(sqlmock.NewResult(1, 1))
		sql.ExpectQuery("select count.*").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		sql.ExpectCommit().WillReturnError(errors.New("failed to commit transaction"))

		ctx := database.WithContext(context.TODO(), mockGorm)

		resp, err := srv.Create(ctx, &accountsv1.CreateAccountRequest{
			Name:     "some-account",
			Currency: "USD",
			Extra: map[string]string{
				"a": "b",
			},
			Type:             v1.AccountType_ACCOUNT_TYPE_ASSET,
			Note:             "some note",
			LiabilityPercent: nil,
			Iban:             "some-iban",
			AccountNumber:    "some-account-number",
			Flags:            database.AccountFlagIsDefault,
		})

		assert.ErrorContains(t, err, "failed to commit transaction")
		assert.Nil(t, resp)
	})
}

func TestCreateBulk(t *testing.T) {
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
			}).Times(2)

		resp, err := srv.CreateBulk(context.TODO(), &accountsv1.CreateAccountsBulkRequest{
			Accounts: []*accountsv1.CreateAccountRequest{
				{
					Name:     "some-account1",
					Currency: "USD",
					Extra: map[string]string{
						"a": "b",
					},
					Type:             v1.AccountType_ACCOUNT_TYPE_ASSET,
					Note:             "some note",
					LiabilityPercent: nil,
					Iban:             "some-iban",
					AccountNumber:    "some-account-number",
					Flags:            database.AccountFlagIsDefault,
				},
				{
					Name:     "some-account2",
					Currency: "PLN",
					Extra: map[string]string{
						"a": "b",
					},
					Type:             v1.AccountType_ACCOUNT_TYPE_ASSET,
					Note:             "some note",
					LiabilityPercent: nil,
					Iban:             "some-iban",
					AccountNumber:    "some-account-number",
				},
			},
		})

		assert.NoError(t, err)
		assert.NotNil(t, resp)

		var rec []database.Account
		assert.NoError(t, gormDB.Find(&rec).Error)

		assert.Len(t, rec, 2)
		assert.Len(t, resp.Messages, 2)

		assert.EqualValues(t, 2, resp.CreatedCount)
		assert.EqualValues(t, 0, resp.DuplicateCount)

		assert.Contains(t, resp.Messages[0], "created successfully")
		assert.Contains(t, resp.Messages[1], "created successfully")
	})

	t.Run("one duplicate", func(t *testing.T) {
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
			}).Times(1)

		resp, err := srv.CreateBulk(context.TODO(), &accountsv1.CreateAccountsBulkRequest{
			Accounts: []*accountsv1.CreateAccountRequest{
				{
					Name:             "some-account",
					Currency:         "USD",
					Type:             v1.AccountType_ACCOUNT_TYPE_ASSET,
					Note:             "some note",
					LiabilityPercent: nil,
					Iban:             "some-iban",
					AccountNumber:    "some-account-number",
					Flags:            database.AccountFlagIsDefault,
				},
				{
					Name:     "some-account",
					Currency: "USD",
					Extra: map[string]string{
						"a": "b",
					},
					Type:             v1.AccountType_ACCOUNT_TYPE_ASSET,
					Note:             "some note",
					LiabilityPercent: nil,
					Iban:             "some-iban",
					AccountNumber:    "some-account-number",
				},
			},
		})

		assert.NoError(t, err)
		assert.NotNil(t, resp)

		var rec []database.Account
		assert.NoError(t, gormDB.Find(&rec).Error)

		assert.Len(t, rec, 1)
		assert.Len(t, resp.Messages, 2)

		assert.EqualValues(t, 1, resp.CreatedCount)
		assert.EqualValues(t, 1, resp.DuplicateCount)

		assert.Contains(t, resp.Messages[0], "created successfully")
		assert.Contains(t, resp.Messages[1], "account with name 'some-account', type 'ACCOUNT_TYPE_ASSET', and currency 'USD' already exists")
	})

	t.Run("one duplicate", func(t *testing.T) {
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
			}).Times(1)

		resp, err := srv.CreateBulk(context.TODO(), &accountsv1.CreateAccountsBulkRequest{
			Accounts: []*accountsv1.CreateAccountRequest{
				{
					Name:             "some-account",
					Currency:         "USD",
					Type:             v1.AccountType_ACCOUNT_TYPE_ASSET,
					Note:             "some note",
					LiabilityPercent: nil,
					Iban:             "some-iban",
					AccountNumber:    "some-account-number",
					Flags:            database.AccountFlagIsDefault,
				},
			},
		})
		assert.NoError(t, err)
		assert.NotNil(t, resp)

		var rec []database.Account
		assert.NoError(t, gormDB.Find(&rec).Error)

		assert.Len(t, rec, 1)
		assert.Len(t, resp.Messages, 1)

		assert.EqualValues(t, 1, resp.CreatedCount)
		assert.EqualValues(t, 0, resp.DuplicateCount)

		assert.Contains(t, resp.Messages[0], "created successfully")

		// second call with the same account
		resp, err = srv.CreateBulk(context.TODO(), &accountsv1.CreateAccountsBulkRequest{
			Accounts: []*accountsv1.CreateAccountRequest{
				{
					Name:             "some-account",
					Currency:         "USD",
					Type:             v1.AccountType_ACCOUNT_TYPE_ASSET,
					Note:             "some note",
					LiabilityPercent: nil,
					Iban:             "some-iban",
					AccountNumber:    "some-account-number",
				},
			},
		})
		assert.NoError(t, err)
		assert.NotNil(t, resp)

		assert.NoError(t, gormDB.Find(&rec).Error)

		assert.Len(t, rec, 1)
		assert.Len(t, resp.Messages, 1)

		assert.Contains(t, resp.Messages[0], "account with name 'some-account', type 'ACCOUNT_TYPE_ASSET', and currency 'USD' already exists")

		assert.EqualValues(t, 0, resp.CreatedCount)
		assert.EqualValues(t, 1, resp.DuplicateCount)
	})

	t.Run("db error", func(t *testing.T) {
		mockGorm, _, sql := testingutils.GormMock()

		srv := accounts.NewService(&accounts.ServiceConfig{})

		sql.ExpectBegin()
		sql.ExpectQuery("SELECT .*").
			WillReturnError(errors.New("failed to create account"))

		sql.ExpectRollback()

		ctx := database.WithContext(context.TODO(), mockGorm)

		resp, err := srv.CreateBulk(ctx, &accountsv1.CreateAccountsBulkRequest{
			Accounts: []*accountsv1.CreateAccountRequest{
				{
					Name:     "some-account",
					Currency: "USD",
					Type:     v1.AccountType_ACCOUNT_TYPE_ASSET,
					Note:     "some note",
				},
			},
		})

		assert.ErrorContains(t, err, "failed to create account")
		assert.Nil(t, resp)
	})

	t.Run("no default account", func(t *testing.T) {
		assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

		srv := accounts.NewService(&accounts.ServiceConfig{})

		resp, err := srv.CreateBulk(context.TODO(), &accountsv1.CreateAccountsBulkRequest{
			Accounts: []*accountsv1.CreateAccountRequest{
				{
					Name:     "some-account",
					Currency: "USD",
					Type:     v1.AccountType_ACCOUNT_TYPE_ASSET,
					Note:     "some note",
				},
			},
		})

		assert.ErrorContains(t, err, "at least one default account is required")
		assert.Nil(t, resp)
	})
}

func TestUpdate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
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
			Type:           v1.AccountType_ACCOUNT_TYPE_ASSET,
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
			Type:             v1.AccountType_ACCOUNT_TYPE_ASSET,
			Note:             "updated note",
			LiabilityPercent: lo.ToPtr("12"),
			Iban:             "iban",
			AccountNumber:    "num",
			Flags:            database.AccountFlagIsDefault,
		})

		assert.NoError(t, err)
		assert.NotNil(t, resp)

		var rec database.Account
		assert.NoError(t, gormDB.First(&rec).Error)

		assert.EqualValues(t, rec.ID, resp.Account.Id)
		assert.EqualValues(t, "yy", rec.Name)
		assert.EqualValues(t, "USD", rec.Currency)
		assert.EqualValues(t, map[string]string{"updated": "b"}, rec.Extra)
		assert.EqualValues(t, v1.AccountType_ACCOUNT_TYPE_ASSET, rec.Type)
		assert.EqualValues(t, "updated note", rec.Note)
		assert.EqualValues(t, "111", rec.CurrentBalance.String())
		assert.EqualValues(t, "12", rec.LiabilityPercent.Decimal.String())
		assert.EqualValues(t, "iban", rec.Iban)
		assert.EqualValues(t, "num", rec.AccountNumber)
	})

	t.Run("update non existing", func(t *testing.T) {
		assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

		srv := accounts.NewService(&accounts.ServiceConfig{})

		resp, err := srv.Update(context.TODO(), &accountsv1.UpdateAccountRequest{
			Id:   -100,
			Name: "yy",
			Extra: map[string]string{
				"updated": "b",
			},
			Type:             v1.AccountType_ACCOUNT_TYPE_ASSET,
			Note:             "updated note",
			LiabilityPercent: lo.ToPtr("12"),
			Iban:             "iban",
			AccountNumber:    "num",
		})

		assert.ErrorContains(t, err, "record not found")
		assert.Nil(t, resp)
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
			Type:  v1.AccountType_ACCOUNT_TYPE_ASSET,
			Note:  "updated note",
			Flags: database.AccountFlagIsDefault,
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
		assert.EqualValues(t, v1.AccountType_ACCOUNT_TYPE_ASSET, rec.Type)
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
			Type:             v1.AccountType_ACCOUNT_TYPE_ASSET,
			Note:             "updated note",
			LiabilityPercent: lo.ToPtr("-1-2"),
			Iban:             "iban",
			AccountNumber:    "num",
		})

		assert.Nil(t, resp)
		assert.ErrorContains(t, err, "can't convert -1-2 to decimal")
	})

	t.Run("removing flag from default account", func(t *testing.T) {
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
			Type:           v1.AccountType_ACCOUNT_TYPE_ASSET,
			Flags:          database.AccountFlagIsDefault,
		}
		assert.NoError(t, gormDB.Create(&acc).Error)

		resp, err := srv.Update(context.TODO(), &accountsv1.UpdateAccountRequest{
			Id:    acc.ID,
			Name:  "yy",
			Type:  v1.AccountType_ACCOUNT_TYPE_ASSET,
			Note:  "updated note",
			Flags: 0, // removing default flag
		})
		assert.Nil(t, resp)
		assert.ErrorContains(t, err, "at least one default account is required")
	})

	t.Run("db error - save", func(t *testing.T) {
		assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

		mockGorm, _, sql := testingutils.GormMock()

		srv := accounts.NewService(&accounts.ServiceConfig{})

		acc := database.Account{
			Currency:       "USD",
			Name:           "xxx",
			Extra:          map[string]string{},
			CurrentBalance: decimal.RequireFromString("111"),
			Type:           v1.AccountType_ACCOUNT_TYPE_ASSET,
			Flags:          database.AccountFlagIsDefault,
		}
		assert.NoError(t, gormDB.Create(&acc).Error)

		sql.ExpectBegin()
		sql.ExpectQuery("SELECT .*").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(acc.ID, acc.Name))
		sql.ExpectQuery("UPDATE .*").
			WillReturnError(errors.New("failed to update account"))
		sql.ExpectRollback()

		ctx := database.WithContext(context.TODO(), mockGorm)

		resp, err := srv.Update(ctx, &accountsv1.UpdateAccountRequest{
			Id:    acc.ID,
			Name:  "yy",
			Type:  v1.AccountType_ACCOUNT_TYPE_ASSET,
			Note:  "updated note",
			Flags: database.AccountFlagIsDefault,
		})

		assert.ErrorContains(t, err, "failed to update account")
		assert.Nil(t, resp)
	})

	t.Run("db error - commit", func(t *testing.T) {
		assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

		mockGorm, _, sql := testingutils.GormMock()

		srv := accounts.NewService(&accounts.ServiceConfig{})

		acc := database.Account{
			Currency:       "USD",
			Name:           "xxx",
			Extra:          map[string]string{},
			CurrentBalance: decimal.RequireFromString("111"),
			Type:           v1.AccountType_ACCOUNT_TYPE_ASSET,
			Flags:          database.AccountFlagIsDefault,
		}
		assert.NoError(t, gormDB.Create(&acc).Error)

		sql.ExpectBegin()
		sql.ExpectQuery("SELECT .*").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(acc.ID, acc.Name))
		sql.ExpectExec("UPDATE .*").
			WillReturnResult(sqlmock.NewResult(1, 1))
		sql.ExpectExec("update .*").
			WillReturnResult(sqlmock.NewResult(1, 1))
		sql.ExpectQuery("select count.*").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
		sql.ExpectCommit().WillReturnError(errors.New("failed to commit transaction"))

		ctx := database.WithContext(context.TODO(), mockGorm)

		resp, err := srv.Update(ctx, &accountsv1.UpdateAccountRequest{
			Id:    acc.ID,
			Name:  "yy",
			Type:  v1.AccountType_ACCOUNT_TYPE_ASSET,
			Note:  "updated note",
			Flags: database.AccountFlagIsDefault,
		})

		assert.ErrorContains(t, err, "failed to commit transaction")
		assert.Nil(t, resp)
	})
}

func TestEnsureDefaultExists(t *testing.T) {
	t.Run("current is default", func(t *testing.T) {
		assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

		prevDefault := &database.Account{
			Extra: map[string]string{},
			Flags: database.AccountFlagIsDefault,
			Type:  v1.AccountType_ACCOUNT_TYPE_ASSET,
		}
		assert.NoError(t, gormDB.Create(prevDefault).Error)

		defaultOtherType := &database.Account{
			Extra: map[string]string{},
			Flags: database.AccountFlagIsDefault,
			Type:  v1.AccountType_ACCOUNT_TYPE_INCOME,
		}
		assert.NoError(t, gormDB.Create(defaultOtherType).Error)

		acc := &database.Account{
			Extra: map[string]string{},
			Flags: database.AccountFlagIsDefault,
			Type:  v1.AccountType_ACCOUNT_TYPE_ASSET,
		}
		assert.NoError(t, gormDB.Create(acc).Error)

		srv := accounts.NewService(&accounts.ServiceConfig{})

		err := srv.EnsureDefaultExists(context.TODO(), gormDB, acc)
		assert.NoError(t, err)

		var updatedPrevDefault database.Account
		assert.NoError(t, gormDB.First(&updatedPrevDefault, "id = ?", prevDefault.ID).Error)
		assert.EqualValues(t, 0, updatedPrevDefault.Flags)

		var updatedDefaultOtherType database.Account
		assert.NoError(t, gormDB.First(&updatedDefaultOtherType, "id = ?", defaultOtherType.ID).Error)
		assert.EqualValues(t, database.AccountFlagIsDefault, updatedDefaultOtherType.Flags)
	})

	t.Run("current is not default", func(t *testing.T) {
		assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

		prevDefault := &database.Account{
			Extra: map[string]string{},
			Flags: database.AccountFlagIsDefault,
			Type:  v1.AccountType_ACCOUNT_TYPE_ASSET,
		}
		assert.NoError(t, gormDB.Create(prevDefault).Error)

		acc := &database.Account{
			Extra: map[string]string{},
			Type:  v1.AccountType_ACCOUNT_TYPE_ASSET,
		}
		assert.NoError(t, gormDB.Create(acc).Error)

		srv := accounts.NewService(&accounts.ServiceConfig{})

		err := srv.EnsureDefaultExists(context.TODO(), gormDB, acc)
		assert.NoError(t, err)

		var updatedPrevDefault database.Account
		assert.NoError(t, gormDB.First(&updatedPrevDefault, "id = ?", prevDefault.ID).Error)
		assert.EqualValues(t, database.AccountFlagIsDefault, updatedPrevDefault.Flags)

		var updatedAcc database.Account
		assert.NoError(t, gormDB.First(&updatedAcc, "id = ?", acc.ID).Error)
		assert.EqualValues(t, 0, updatedAcc.Flags)
	})

	t.Run("no default exists", func(t *testing.T) {
		assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

		acc := &database.Account{
			Extra: map[string]string{},
			Type:  v1.AccountType_ACCOUNT_TYPE_ASSET,
		}
		assert.NoError(t, gormDB.Create(acc).Error)

		srv := accounts.NewService(&accounts.ServiceConfig{})

		err := srv.EnsureDefaultExists(context.TODO(), gormDB, acc)
		assert.ErrorContains(t, err, "at least one default account is required")
	})

	t.Run("db fail - current default", func(t *testing.T) {
		assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

		acc := &database.Account{
			Extra: map[string]string{},
			Type:  v1.AccountType_ACCOUNT_TYPE_ASSET,
			Flags: database.AccountFlagIsDefault,
		}
		assert.NoError(t, gormDB.Create(acc).Error)

		mockGorm, _, sql := testingutils.GormMock()

		srv := accounts.NewService(&accounts.ServiceConfig{})

		sql.ExpectExec("update accounts .*").
			WillReturnError(errors.New("failed to update account"))

		ctx := database.WithContext(context.TODO(), mockGorm)

		err := srv.EnsureDefaultExists(ctx, mockGorm, acc)
		assert.ErrorContains(t, err, "failed to update account")
	})

	t.Run("db fail - current not default", func(t *testing.T) {
		assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

		acc := &database.Account{
			Extra: map[string]string{},
			Type:  v1.AccountType_ACCOUNT_TYPE_ASSET,
		}
		assert.NoError(t, gormDB.Create(acc).Error)

		mockGorm, _, sql := testingutils.GormMock()

		srv := accounts.NewService(&accounts.ServiceConfig{})

		sql.ExpectExec("update accounts .*").
			WillReturnError(errors.New("failed to update account"))

		ctx := database.WithContext(context.TODO(), mockGorm)

		err := srv.EnsureDefaultExists(ctx, mockGorm, acc)
		assert.ErrorContains(t, err, "failed to update account")
	})
}
