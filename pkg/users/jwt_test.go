package users_test

import (
	"context"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/ft-t/go-money/pkg/jwt"
	"github.com/ft-t/go-money/pkg/users"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestJwtToken(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		keyGen := jwt.NewKeyGenerator()
		key := keyGen.Generate()

		jwtGenerator, err := users.NewJwtGenerator(string(keyGen.Serialize(key)))
		assert.NoError(t, err)

		resp, err := jwtGenerator.GenerateToken(context.TODO(), &database.User{})

		assert.NoError(t, err)
		assert.NotEmpty(t, resp)
	})

	t.Run("invalid private key", func(t *testing.T) {
		jwtGenerator, err := users.NewJwtGenerator("not-a-key")
		assert.ErrorContains(t, err, "failed to decode PEM block")
		assert.Nil(t, jwtGenerator)
	})
}
