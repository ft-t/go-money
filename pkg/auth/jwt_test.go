package auth_test

import (
	"context"
	"github.com/ft-t/go-money/pkg/auth"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestJwtToken_Success(t *testing.T) {
	t.Run("generate and validate web token", func(t *testing.T) {
		keyGen := auth.NewKeyGenerator()
		key := keyGen.Generate()

		assert.NoError(t, keyGen.Save(key, "test_key.pem"))

		jwtGenerator, err := auth.NewService(string(keyGen.Serialize(key)), 5*time.Minute)
		assert.NoError(t, err)

		resp, err := jwtGenerator.GenerateToken(context.TODO(), &database.User{
			ID:    1,
			Login: "abcd",
		})

		assert.NoError(t, err)
		assert.NotEmpty(t, resp)

		claims, err := jwtGenerator.ValidateToken(context.TODO(), resp)
		assert.NoError(t, err)
		assert.NotNil(t, claims)

		assert.EqualValues(t, 1, claims.UserID)
	})
}

func TestJwtToken_Failure(t *testing.T) {

	t.Run("invalid private key", func(t *testing.T) {
		jwtGenerator, err := auth.NewService("not-a-key", 5*time.Minute)
		assert.ErrorContains(t, err, "failed to decode PEM block")
		assert.Nil(t, jwtGenerator)
	})

	t.Run("parse jwt with invalid key", func(t *testing.T) {
		keyGen := auth.NewKeyGenerator()
		key := keyGen.Generate()

		jwtGenerator1, err := auth.NewService(string(keyGen.Serialize(key)), 5*time.Minute)
		assert.NoError(t, err)

		resp, err := jwtGenerator1.GenerateToken(context.TODO(), &database.User{})

		assert.NoError(t, err)
		assert.NotEmpty(t, resp)

		jwtGenerator2, err := auth.NewService(string(keyGen.Serialize(keyGen.Generate())), 5*time.Minute)
		assert.NoError(t, err)

		claims, err := jwtGenerator2.ValidateToken(context.TODO(), resp)
		assert.Nil(t, claims)
		assert.ErrorContains(t, err, "failed to parse token: token signature is invalid: crypto/rsa: verification error")
	})

	t.Run("parse expired token", func(t *testing.T) {
		keyGen := auth.NewKeyGenerator()
		key := keyGen.Generate()

		jwtGenerator1, err := auth.NewService(string(keyGen.Serialize(key)), -5*time.Minute)
		assert.NoError(t, err)

		resp, err := jwtGenerator1.GenerateToken(context.TODO(), &database.User{})

		assert.NoError(t, err)
		assert.NotEmpty(t, resp)

		claims, err := jwtGenerator1.ValidateToken(context.TODO(), resp)
		assert.Nil(t, claims)
		assert.ErrorContains(t, err, "failed to parse token: token has invalid claims: token is expired")
	})

	t.Run("invalid claims type or invalid token", func(t *testing.T) {
		keyGen := auth.NewKeyGenerator()
		key := keyGen.Generate()
		jwtGenerator, err := auth.NewService(string(keyGen.Serialize(key)), 5*time.Minute)
		assert.NoError(t, err)

		t.Run("nil token", func(t *testing.T) {
			resp, claimsErr := jwtGenerator.CheckClaims(nil)
			assert.Nil(t, resp)
			assert.ErrorContains(t, claimsErr, "token is nil")
		})

		t.Run("claims wrong type", func(t *testing.T) {
			resp, claimsErr := jwtGenerator.CheckClaims(&jwt.Token{
				Claims: jwt.MapClaims{},
			})
			assert.Nil(t, resp)
			assert.ErrorContains(t, claimsErr, "invalid token claims")
		})

		t.Run("not valid", func(t *testing.T) {
			resp, claimsErr := jwtGenerator.CheckClaims(&jwt.Token{
				Claims: &auth.JwtClaims{},
				Valid:  false,
			})
			assert.Nil(t, resp)
			assert.ErrorContains(t, claimsErr, "token is not valid")
		})
	})
}

func TestCreateServiceToken_Success(t *testing.T) {
	keyGen := auth.NewKeyGenerator()
	key := keyGen.Generate()

	jwtGenerator, err := auth.NewService(string(keyGen.Serialize(key)), 5*time.Minute)
	assert.NoError(t, err)

	claims, token, err := jwtGenerator.CreateServiceToken(context.TODO(), &auth.GenerateTokenRequest{
		TTL: 24 * time.Hour,
		User: &database.User{
			ID:    123,
			Login: "testuser",
		},
	})

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.NotNil(t, claims)
	assert.EqualValues(t, 123, claims.UserID)
	assert.Equal(t, auth.ServiceTokenType, claims.TokenType)
	assert.NotEmpty(t, claims.ID)
}

func TestCreateServiceToken_Failure(t *testing.T) {
	t.Run("user is nil", func(t *testing.T) {
		keyGen := auth.NewKeyGenerator()
		key := keyGen.Generate()

		jwtGenerator, err := auth.NewService(string(keyGen.Serialize(key)), 5*time.Minute)
		assert.NoError(t, err)

		claims, token, err := jwtGenerator.CreateServiceToken(context.TODO(), &auth.GenerateTokenRequest{
			TTL:  24 * time.Hour,
			User: nil,
		})

		assert.ErrorContains(t, err, "user is required to generate service token")
		assert.Empty(t, token)
		assert.Nil(t, claims)
	})

	t.Run("ttl is zero", func(t *testing.T) {
		keyGen := auth.NewKeyGenerator()
		key := keyGen.Generate()

		jwtGenerator, err := auth.NewService(string(keyGen.Serialize(key)), 5*time.Minute)
		assert.NoError(t, err)

		claims, token, err := jwtGenerator.CreateServiceToken(context.TODO(), &auth.GenerateTokenRequest{
			TTL: 0,
			User: &database.User{
				ID:    123,
				Login: "testuser",
			},
		})

		assert.ErrorContains(t, err, "ttl is required to generate service token")
		assert.Empty(t, token)
		assert.Nil(t, claims)
	})
}
