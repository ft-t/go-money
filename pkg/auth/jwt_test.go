package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/ft-t/go-money/pkg/auth"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
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

	t.Run("validate token twice - cache hit", func(t *testing.T) {
		keyGen := auth.NewKeyGenerator()
		key := keyGen.Generate()

		jwtGenerator, err := auth.NewService(string(keyGen.Serialize(key)), 5*time.Minute)
		assert.NoError(t, err)

		resp, err := jwtGenerator.GenerateToken(context.TODO(), &database.User{
			ID:    2,
			Login: "cachetest",
		})
		assert.NoError(t, err)

		claims1, err := jwtGenerator.ValidateToken(context.TODO(), resp)
		assert.NoError(t, err)
		assert.NotNil(t, claims1)

		claims2, err := jwtGenerator.ValidateToken(context.TODO(), resp)
		assert.NoError(t, err)
		assert.NotNil(t, claims2)
		assert.Equal(t, claims1.ID, claims2.ID)
	})
}

func TestJwtToken_Failure(t *testing.T) {
	t.Run("invalid private key - not PEM", func(t *testing.T) {
		jwtGenerator, err := auth.NewService("not-a-key", 5*time.Minute)
		assert.ErrorContains(t, err, "failed to decode PEM block")
		assert.Nil(t, jwtGenerator)
	})

	t.Run("invalid private key - valid PEM but invalid PKCS1", func(t *testing.T) {
		invalidPEM := `-----BEGIN RSA PRIVATE KEY-----
aW52YWxpZCBrZXkgZGF0YQ==
-----END RSA PRIVATE KEY-----`
		jwtGenerator, err := auth.NewService(invalidPEM, 5*time.Minute)
		assert.ErrorContains(t, err, "failed to parse private key")
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

	t.Run("unexpected signing method", func(t *testing.T) {
		keyGen := auth.NewKeyGenerator()
		key := keyGen.Generate()

		jwtGenerator, err := auth.NewService(string(keyGen.Serialize(key)), 5*time.Minute)
		assert.NoError(t, err)

		hmacToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"user_id": 123,
			"exp":     time.Now().Add(time.Hour).Unix(),
		})
		tokenStr, err := hmacToken.SignedString([]byte("secret"))
		assert.NoError(t, err)

		claims, err := jwtGenerator.ValidateToken(context.TODO(), tokenStr)
		assert.Nil(t, claims)
		assert.ErrorContains(t, err, "unexpected signing method")
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

	serviceTTL := 5 * time.Minute
	requestTTL := 24 * time.Hour

	jwtGenerator, err := auth.NewService(string(keyGen.Serialize(key)), serviceTTL)
	assert.NoError(t, err)

	beforeGeneration := time.Now().UTC()

	claims, token, err := jwtGenerator.CreateServiceToken(context.TODO(), &auth.GenerateTokenRequest{
		TTL: requestTTL,
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

	expectedExpiresAt := beforeGeneration.Add(requestTTL)
	actualExpiresAt := claims.ExpiresAt.Time

	assert.WithinDuration(t, expectedExpiresAt, actualExpiresAt, 5*time.Second)
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
