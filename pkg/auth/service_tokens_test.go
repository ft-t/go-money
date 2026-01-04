package auth_test

import (
	"context"
	"os"
	"testing"
	"time"

	configurationv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/configuration/v1"
	gomoneypbv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/v1"
	"github.com/golang-jwt/jwt/v5"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"

	"github.com/ft-t/go-money/pkg/auth"
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

func TestServiceTokenService_CreateServiceToken_Success(t *testing.T) {
	assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

	ctrl := gomock.NewController(t)
	mockJwtSvc := NewMockJwtSvc(ctrl)
	mockMapper := NewMockServiceTokenMapper(ctrl)

	user := &database.User{
		Login:    "testuser",
		Password: "hash",
	}
	assert.NoError(t, gormDB.Create(user).Error)

	expiresAt := time.Now().UTC().Add(24 * time.Hour)
	tokenID := "test-token-id-123"

	mockJwtSvc.EXPECT().CreateServiceToken(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, req *auth.GenerateTokenRequest) (*auth.JwtClaims, string, error) {
			assert.Equal(t, user.ID, req.User.ID)
			assert.True(t, req.TTL > 0)

			return &auth.JwtClaims{
				RegisteredClaims: &jwt.RegisteredClaims{
					ID: tokenID,
				},
				UserID:    user.ID,
				TokenType: auth.ServiceTokenType,
			}, "jwt.token.string", nil
		},
	)

	mockMapper.EXPECT().MapServiceToken(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, token *database.ServiceToken) *gomoneypbv1.ServiceToken {
			assert.Equal(t, tokenID, token.ID)
			assert.Equal(t, "Test Token", token.Name)

			return &gomoneypbv1.ServiceToken{
				Id:        token.ID,
				Name:      token.Name,
				CreatedAt: timestamppb.New(token.CreatedAt),
				ExpiresAt: timestamppb.New(token.ExpiresAt),
			}
		},
	)

	svc := auth.NewServiceTokenService(mockJwtSvc, mockMapper)

	resp, err := svc.CreateServiceToken(context.TODO(), &auth.CreateServiceTokenRequest{
		Req: &configurationv1.CreateServiceTokenRequest{
			Name:      "Test Token",
			ExpiresAt: timestamppb.New(expiresAt),
		},
		CurrentUserID: user.ID,
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, tokenID, resp.ServiceToken.Id)
	assert.Equal(t, "Test Token", resp.ServiceToken.Name)
	assert.Equal(t, "jwt.token.string", resp.Token)

	var savedToken database.ServiceToken
	assert.NoError(t, gormDB.Where("id = ?", tokenID).First(&savedToken).Error)
	assert.Equal(t, "Test Token", savedToken.Name)
}

func TestServiceTokenService_CreateServiceToken_Failure(t *testing.T) {
	t.Run("missing expiresAt", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockJwtSvc := NewMockJwtSvc(ctrl)
		mockMapper := NewMockServiceTokenMapper(ctrl)

		svc := auth.NewServiceTokenService(mockJwtSvc, mockMapper)

		resp, err := svc.CreateServiceToken(context.TODO(), &auth.CreateServiceTokenRequest{
			Req: &configurationv1.CreateServiceTokenRequest{
				Name: "Test Token",
			},
			CurrentUserID: 1,
		})

		assert.ErrorContains(t, err, "expiresAt is required")
		assert.Nil(t, resp)
	})

	t.Run("user not found", func(t *testing.T) {
		assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

		ctrl := gomock.NewController(t)
		mockJwtSvc := NewMockJwtSvc(ctrl)
		mockMapper := NewMockServiceTokenMapper(ctrl)

		svc := auth.NewServiceTokenService(mockJwtSvc, mockMapper)

		resp, err := svc.CreateServiceToken(context.TODO(), &auth.CreateServiceTokenRequest{
			Req: &configurationv1.CreateServiceTokenRequest{
				Name:      "Test Token",
				ExpiresAt: timestamppb.New(time.Now().Add(time.Hour)),
			},
			CurrentUserID: 99999,
		})

		assert.ErrorContains(t, err, "record not found")
		assert.Nil(t, resp)
	})
}

func TestServiceTokenService_GetServiceTokens_Success(t *testing.T) {
	assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

	ctrl := gomock.NewController(t)
	mockJwtSvc := NewMockJwtSvc(ctrl)
	mockMapper := NewMockServiceTokenMapper(ctrl)

	token1 := &database.ServiceToken{
		ID:        "token-1",
		Name:      "Token One",
		ExpiresAt: time.Now().UTC().Add(24 * time.Hour),
		CreatedAt: time.Now().UTC(),
	}
	token2 := &database.ServiceToken{
		ID:        "token-2",
		Name:      "Token Two",
		ExpiresAt: time.Now().UTC().Add(48 * time.Hour),
		CreatedAt: time.Now().UTC(),
	}
	assert.NoError(t, gormDB.Create(token1).Error)
	assert.NoError(t, gormDB.Create(token2).Error)

	svc := auth.NewServiceTokenService(mockJwtSvc, mockMapper)

	t.Run("get all tokens", func(t *testing.T) {
		mockMapper.EXPECT().MapServiceToken(gomock.Any(), gomock.Any()).Times(2).DoAndReturn(
			func(_ context.Context, token *database.ServiceToken) *gomoneypbv1.ServiceToken {
				return &gomoneypbv1.ServiceToken{
					Id:   token.ID,
					Name: token.Name,
				}
			},
		)

		resp, err := svc.GetServiceTokens(context.TODO(), &configurationv1.GetServiceTokensRequest{})

		assert.NoError(t, err)
		assert.Len(t, resp.ServiceTokens, 2)
	})

	t.Run("get specific token by id", func(t *testing.T) {
		mockMapper.EXPECT().MapServiceToken(gomock.Any(), gomock.Any()).DoAndReturn(
			func(_ context.Context, token *database.ServiceToken) *gomoneypbv1.ServiceToken {
				return &gomoneypbv1.ServiceToken{
					Id:   token.ID,
					Name: token.Name,
				}
			},
		)

		resp, err := svc.GetServiceTokens(context.TODO(), &configurationv1.GetServiceTokensRequest{
			Ids: []string{"token-1"},
		})

		assert.NoError(t, err)
		assert.Len(t, resp.ServiceTokens, 1)
		assert.Equal(t, "token-1", resp.ServiceTokens[0].Id)
		assert.Equal(t, "Token One", resp.ServiceTokens[0].Name)
	})
}

func TestServiceTokenService_RevokeServiceToken_Success(t *testing.T) {
	assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

	ctrl := gomock.NewController(t)
	mockJwtSvc := NewMockJwtSvc(ctrl)
	mockMapper := NewMockServiceTokenMapper(ctrl)

	expiresAt := time.Now().UTC().Add(24 * time.Hour)
	token := &database.ServiceToken{
		ID:        "token-to-revoke",
		Name:      "Token To Revoke",
		ExpiresAt: expiresAt,
		CreatedAt: time.Now().UTC(),
	}
	assert.NoError(t, gormDB.Create(token).Error)

	mockJwtSvc.EXPECT().RevokeServiceToken(gomock.Any(), "token-to-revoke", gomock.Any()).DoAndReturn(
		func(_ context.Context, jti string, originalExpiresAt time.Time) error {
			assert.Equal(t, "token-to-revoke", jti)
			assert.True(t, originalExpiresAt.Sub(expiresAt) < time.Second)
			return nil
		},
	)

	svc := auth.NewServiceTokenService(mockJwtSvc, mockMapper)

	resp, err := svc.RevokeServiceToken(context.TODO(), &configurationv1.RevokeServiceTokenRequest{
		Id: "token-to-revoke",
	})

	assert.NoError(t, err)
	assert.Equal(t, "token-to-revoke", resp.Id)

	var revokedToken database.ServiceToken
	err = gormDB.Unscoped().Where("id = ?", "token-to-revoke").First(&revokedToken).Error
	assert.NoError(t, err)
	assert.True(t, revokedToken.DeletedAt.Valid)
}

func TestServiceTokenService_RevokeServiceToken_Failure(t *testing.T) {
	t.Run("token not found", func(t *testing.T) {
		assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

		ctrl := gomock.NewController(t)
		mockJwtSvc := NewMockJwtSvc(ctrl)
		mockMapper := NewMockServiceTokenMapper(ctrl)

		svc := auth.NewServiceTokenService(mockJwtSvc, mockMapper)

		resp, err := svc.RevokeServiceToken(context.TODO(), &configurationv1.RevokeServiceTokenRequest{
			Id: "nonexistent-token",
		})

		assert.ErrorContains(t, err, "failed to find service token")
		assert.Nil(t, resp)
	})
}

func TestJwtService_RevokeServiceToken_Success(t *testing.T) {
	assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

	keyGen := auth.NewKeyGenerator()
	key := keyGen.Generate()

	jwtSvc, err := auth.NewService(string(keyGen.Serialize(key)), 5*time.Minute)
	assert.NoError(t, err)

	jti := "test-jti-to-revoke"
	expiresAt := time.Now().UTC().Add(24 * time.Hour)

	err = jwtSvc.RevokeServiceToken(context.TODO(), jti, expiresAt)
	assert.NoError(t, err)

	var revocation database.JtiRevocation
	err = gormDB.Where("id = ?", jti).First(&revocation).Error
	assert.NoError(t, err)
	assert.Equal(t, jti, revocation.ID)
	assert.True(t, revocation.ExpiresAt.After(expiresAt))
}

func TestJwtService_ValidateServiceToken_WithRevocation(t *testing.T) {
	assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

	keyGen := auth.NewKeyGenerator()
	key := keyGen.Generate()

	jwtSvc, err := auth.NewService(string(keyGen.Serialize(key)), 5*time.Minute)
	assert.NoError(t, err)

	t.Run("valid service token not revoked", func(t *testing.T) {
		claims, tokenStr, err := jwtSvc.CreateServiceToken(context.TODO(), &auth.GenerateTokenRequest{
			TTL: 24 * time.Hour,
			User: &database.User{
				ID:    1,
				Login: "testuser",
			},
		})
		assert.NoError(t, err)
		assert.NotEmpty(t, tokenStr)

		validatedClaims, err := jwtSvc.ValidateToken(context.TODO(), tokenStr)
		assert.NoError(t, err)
		assert.NotNil(t, validatedClaims)
		assert.Equal(t, claims.ID, validatedClaims.ID)
		assert.Equal(t, auth.ServiceTokenType, validatedClaims.TokenType)
	})

	t.Run("revoked service token", func(t *testing.T) {
		claims, tokenStr, err := jwtSvc.CreateServiceToken(context.TODO(), &auth.GenerateTokenRequest{
			TTL: 24 * time.Hour,
			User: &database.User{
				ID:    2,
				Login: "testuser2",
			},
		})
		assert.NoError(t, err)

		err = gormDB.Create(&database.JtiRevocation{
			ID:        claims.ID,
			ExpiresAt: time.Now().UTC().Add(7 * 24 * time.Hour),
		}).Error
		assert.NoError(t, err)

		validatedClaims, err := jwtSvc.ValidateToken(context.TODO(), tokenStr)
		assert.Nil(t, validatedClaims)
		assert.ErrorContains(t, err, "token has been revoked")
	})
}
