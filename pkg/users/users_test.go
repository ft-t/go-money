package users_test

import (
	usersv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/users/v1"
	"context"
	"github.com/cockroachdb/errors"
	"github.com/ft-t/go-money/pkg/configuration"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/ft-t/go-money/pkg/testingutils"
	"github.com/ft-t/go-money/pkg/users"
	"github.com/golang/mock/gomock"
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

func TestCreateUserAndLogin(t *testing.T) {
	assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

	srv := users.NewService(&users.ServiceConfig{})

	password := "qwerty"
	login := "somelogin"

	resp, err := srv.Create(context.TODO(), &usersv1.CreateRequest{
		Login:    login,
		Password: password,
	})
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	t.Run("fail register 2 account", func(t *testing.T) {
		resp2, err2 := srv.Create(context.TODO(), &usersv1.CreateRequest{
			Login:    "yyy",
			Password: "yyyy",
		})
		assert.ErrorContains(t, err2, "admin already exists")
		assert.Nil(t, resp2)
	})

	t.Run("success login", func(t *testing.T) {
		jwtSvc := NewMockJwtSvc(gomock.NewController(t))
		jwtSvc.EXPECT().GenerateToken(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, user *database.User) (string, error) {
				assert.EqualValues(t, resp.Id, user.ID)

				return "sometoken", nil
			})

		loginSvc := users.NewService(&users.ServiceConfig{
			JwtSvc: jwtSvc,
		})

		lResp, lErr := loginSvc.Login(context.TODO(), &usersv1.LoginRequest{
			Login:    login,
			Password: password,
		})

		assert.NoError(t, lErr)
		assert.EqualValues(t, "sometoken", lResp.Token)
	})

	t.Run("token err", func(t *testing.T) {
		jwtSvc := NewMockJwtSvc(gomock.NewController(t))
		jwtSvc.EXPECT().GenerateToken(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, user *database.User) (string, error) {
				assert.EqualValues(t, resp.Id, user.ID)

				return "", errors.New("token err")
			})

		loginSvc := users.NewService(&users.ServiceConfig{
			JwtSvc: jwtSvc,
		})

		lResp, lErr := loginSvc.Login(context.TODO(), &usersv1.LoginRequest{
			Login:    login,
			Password: password,
		})

		assert.ErrorContains(t, lErr, "token err")
		assert.Nil(t, lResp)
	})

	t.Run("invalid password", func(t *testing.T) {
		loginSvc := users.NewService(&users.ServiceConfig{})

		lResp, lErr := loginSvc.Login(context.TODO(), &usersv1.LoginRequest{
			Login:    login,
			Password: "xxx",
		})

		assert.ErrorContains(t, lErr, "password is invalid")
		assert.Nil(t, lResp)
	})

	t.Run("login not set", func(t *testing.T) {
		loginSvc := users.NewService(&users.ServiceConfig{})

		lResp, lErr := loginSvc.Login(context.TODO(), &usersv1.LoginRequest{
			Login:    "",
			Password: "xxx",
		})

		assert.ErrorContains(t, lErr, "login is required")
		assert.Nil(t, lResp)
	})

	t.Run("user not found", func(t *testing.T) {
		loginSvc := users.NewService(&users.ServiceConfig{})

		lResp, lErr := loginSvc.Login(context.TODO(), &usersv1.LoginRequest{
			Login:    "xxx",
			Password: "xxx",
		})

		assert.ErrorContains(t, lErr, "user not found")
		assert.Nil(t, lResp)
	})
}
