package auth_test

import (
	"context"
	"github.com/ft-t/go-money/pkg/auth"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestJwtToken(t *testing.T) {
	t.Run("success", func(t *testing.T) {
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
}
