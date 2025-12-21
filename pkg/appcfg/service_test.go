package appcfg_test

import (
	"context"
	"os"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	configurationv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/configuration/v1"
	"github.com/cockroachdb/errors"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"github.com/ft-t/go-money/pkg/appcfg"
	"github.com/ft-t/go-money/pkg/configuration"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/ft-t/go-money/pkg/testingutils"
)

var gormDB *gorm.DB
var cfg *configuration.Configuration

func TestMain(m *testing.M) {
	cfg = configuration.GetConfiguration()
	gormDB = database.GetDb(database.DbTypeMaster)

	os.Exit(m.Run())
}

func TestGetConfiguration(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		userSvc := NewMockUserSvc(gomock.NewController(t))
		userSvc.EXPECT().ShouldCreateAdmin(gomock.Any()).
			Return(true, nil)

		srv := appcfg.NewService(&appcfg.ServiceConfig{
			UserSvc: userSvc,
			AppCfg: &configuration.Configuration{
				CurrencyConfig: configuration.CurrencyConfig{
					BaseCurrency: "USD",
				},
			},
		})

		resp, err := srv.GetConfiguration(context.TODO(), &configurationv1.GetConfigurationRequest{})
		assert.NoError(t, err)
		assert.True(t, resp.ShouldCreateAdmin)
		assert.Equal(t, "USD", resp.BaseCurrency)
	})

	t.Run("fail", func(t *testing.T) {
		userSvc := NewMockUserSvc(gomock.NewController(t))
		userSvc.EXPECT().ShouldCreateAdmin(gomock.Any()).
			Return(false, errors.New("some error"))

		srv := appcfg.NewService(&appcfg.ServiceConfig{
			UserSvc: userSvc,
			AppCfg:  &configuration.Configuration{},
		})

		resp, err := srv.GetConfiguration(context.TODO(), &configurationv1.GetConfigurationRequest{})
		assert.ErrorContains(t, err, "some error")
		assert.Nil(t, resp)
	})
}

func TestGetConfigsByKeys_Success(t *testing.T) {
	assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

	srv := appcfg.NewService(&appcfg.ServiceConfig{
		AppCfg: cfg,
	})

	assert.NoError(t, gormDB.Create(&database.AppConfig{
		ID:    "test_key_1",
		Value: `{"foo":"bar"}`,
	}).Error)
	assert.NoError(t, gormDB.Create(&database.AppConfig{
		ID:    "test_key_2",
		Value: `{"baz":"qux"}`,
	}).Error)

	resp, err := srv.GetConfigsByKeys(context.TODO(), &configurationv1.GetConfigsByKeysRequest{
		Keys: []string{"test_key_1", "test_key_2", "non_existent"},
	})
	assert.NoError(t, err)
	assert.Len(t, resp.Configs, 2)
	assert.Equal(t, `{"foo":"bar"}`, resp.Configs["test_key_1"])
	assert.Equal(t, `{"baz":"qux"}`, resp.Configs["test_key_2"])
	assert.Empty(t, resp.Configs["non_existent"])
}

func TestGetConfigsByKeys_EmptyKeys(t *testing.T) {
	srv := appcfg.NewService(&appcfg.ServiceConfig{
		AppCfg: cfg,
	})

	resp, err := srv.GetConfigsByKeys(context.TODO(), &configurationv1.GetConfigsByKeysRequest{
		Keys: []string{},
	})
	assert.NoError(t, err)
	assert.Empty(t, resp.Configs)
}

func TestSetConfigByKey_CreateNew(t *testing.T) {
	assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

	srv := appcfg.NewService(&appcfg.ServiceConfig{
		AppCfg: cfg,
	})

	resp, err := srv.SetConfigByKey(context.TODO(), &configurationv1.SetConfigByKeyRequest{
		Key:   "new_key",
		Value: `{"snippets":[]}`,
	})
	assert.NoError(t, err)
	assert.Equal(t, "new_key", resp.Key)

	var stored database.AppConfig
	assert.NoError(t, gormDB.Where("id = ?", "new_key").First(&stored).Error)
	assert.Equal(t, `{"snippets":[]}`, stored.Value)
	assert.NotZero(t, stored.CreatedAt)
	assert.NotZero(t, stored.UpdatedAt)
}

func TestSetConfigByKey_UpdateExisting(t *testing.T) {
	assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

	srv := appcfg.NewService(&appcfg.ServiceConfig{
		AppCfg: cfg,
	})

	assert.NoError(t, gormDB.Create(&database.AppConfig{
		ID:    "existing_key",
		Value: `{"old":"value"}`,
	}).Error)

	resp, err := srv.SetConfigByKey(context.TODO(), &configurationv1.SetConfigByKeyRequest{
		Key:   "existing_key",
		Value: `{"new":"value"}`,
	})
	assert.NoError(t, err)
	assert.Equal(t, "existing_key", resp.Key)

	var stored database.AppConfig
	assert.NoError(t, gormDB.Where("id = ?", "existing_key").First(&stored).Error)
	assert.Equal(t, `{"new":"value"}`, stored.Value)
}

func TestSetConfigByKey_RestoreSoftDeleted(t *testing.T) {
	assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

	srv := appcfg.NewService(&appcfg.ServiceConfig{
		AppCfg: cfg,
	})

	assert.NoError(t, gormDB.Create(&database.AppConfig{
		ID:    "deleted_key",
		Value: `{"old":"value"}`,
	}).Error)
	assert.NoError(t, gormDB.Delete(&database.AppConfig{}, "id = ?", "deleted_key").Error)

	resp, err := srv.SetConfigByKey(context.TODO(), &configurationv1.SetConfigByKeyRequest{
		Key:   "deleted_key",
		Value: `{"restored":"value"}`,
	})
	assert.NoError(t, err)
	assert.Equal(t, "deleted_key", resp.Key)

	var stored database.AppConfig
	assert.NoError(t, gormDB.Where("id = ?", "deleted_key").First(&stored).Error)
	assert.Equal(t, `{"restored":"value"}`, stored.Value)
	assert.False(t, stored.DeletedAt.Valid)
}

func TestSetConfigByKey_FullRoundTrip(t *testing.T) {
	assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

	srv := appcfg.NewService(&appcfg.ServiceConfig{
		AppCfg: cfg,
	})

	_, err := srv.SetConfigByKey(context.TODO(), &configurationv1.SetConfigByKeyRequest{
		Key:   "roundtrip_key",
		Value: `{"version":1}`,
	})
	assert.NoError(t, err)

	assert.NoError(t, gormDB.Delete(&database.AppConfig{}, "id = ?", "roundtrip_key").Error)

	var deleted database.AppConfig
	err = gormDB.Where("id = ?", "roundtrip_key").First(&deleted).Error
	assert.Error(t, err)

	_, err = srv.SetConfigByKey(context.TODO(), &configurationv1.SetConfigByKeyRequest{
		Key:   "roundtrip_key",
		Value: `{"version":2}`,
	})
	assert.NoError(t, err)

	var stored database.AppConfig
	assert.NoError(t, gormDB.Where("id = ?", "roundtrip_key").First(&stored).Error)
	assert.Equal(t, `{"version":2}`, stored.Value)
	assert.False(t, stored.DeletedAt.Valid)

	var count int64
	assert.NoError(t, gormDB.Unscoped().Model(&database.AppConfig{}).Where("id = ?", "roundtrip_key").Count(&count).Error)
	assert.Equal(t, int64(2), count)

	var activeCount int64
	assert.NoError(t, gormDB.Model(&database.AppConfig{}).Where("id = ?", "roundtrip_key").Count(&activeCount).Error)
	assert.Equal(t, int64(1), activeCount)
}

func TestGetConfigsByKeys_DbError(t *testing.T) {
	mockGorm, _, sql := testingutils.GormMock()

	srv := appcfg.NewService(&appcfg.ServiceConfig{
		AppCfg: cfg,
	})

	ctx := database.WithContext(context.TODO(), mockGorm)

	sql.ExpectQuery("SELECT .*").WillReturnError(errors.New("db error"))

	resp, err := srv.GetConfigsByKeys(ctx, &configurationv1.GetConfigsByKeysRequest{
		Keys: []string{"test_key"},
	})

	assert.ErrorContains(t, err, "db error")
	assert.Nil(t, resp)
}

func TestSetConfigByKey_DeleteError(t *testing.T) {
	mockGorm, _, sql := testingutils.GormMock()

	srv := appcfg.NewService(&appcfg.ServiceConfig{
		AppCfg: cfg,
	})

	ctx := database.WithContext(context.TODO(), mockGorm)

	sql.ExpectBegin()
	sql.ExpectExec("UPDATE .*").WillReturnError(errors.New("delete error"))
	sql.ExpectRollback()

	resp, err := srv.SetConfigByKey(ctx, &configurationv1.SetConfigByKeyRequest{
		Key:   "test_key",
		Value: `{"test":"value"}`,
	})

	assert.ErrorContains(t, err, "delete error")
	assert.Nil(t, resp)
}

func TestSetConfigByKey_CreateError(t *testing.T) {
	mockGorm, _, sql := testingutils.GormMock()

	srv := appcfg.NewService(&appcfg.ServiceConfig{
		AppCfg: cfg,
	})

	ctx := database.WithContext(context.TODO(), mockGorm)

	sql.ExpectBegin()
	sql.ExpectExec("UPDATE .*").WillReturnResult(sqlmock.NewResult(0, 0))
	sql.ExpectExec("INSERT .*").WillReturnError(errors.New("create error"))
	sql.ExpectRollback()

	resp, err := srv.SetConfigByKey(ctx, &configurationv1.SetConfigByKeyRequest{
		Key:   "test_key",
		Value: `{"test":"value"}`,
	})

	assert.ErrorContains(t, err, "create error")
	assert.Nil(t, resp)
}

func TestSetConfigByKey_CommitError(t *testing.T) {
	mockGorm, _, sql := testingutils.GormMock()

	srv := appcfg.NewService(&appcfg.ServiceConfig{
		AppCfg: cfg,
	})

	ctx := database.WithContext(context.TODO(), mockGorm)

	sql.ExpectBegin()
	sql.ExpectExec("UPDATE .*").WillReturnResult(sqlmock.NewResult(0, 0))
	sql.ExpectExec("INSERT .*").WillReturnResult(sqlmock.NewResult(1, 1))
	sql.ExpectCommit().WillReturnError(errors.New("commit error"))
	sql.ExpectRollback()

	resp, err := srv.SetConfigByKey(ctx, &configurationv1.SetConfigByKeyRequest{
		Key:   "test_key",
		Value: `{"test":"value"}`,
	})

	assert.ErrorContains(t, err, "commit error")
	assert.Nil(t, resp)
}
