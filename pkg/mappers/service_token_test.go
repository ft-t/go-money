package mappers_test

import (
	"context"
	"testing"
	"time"

	"github.com/ft-t/go-money/pkg/database"
	"github.com/ft-t/go-money/pkg/mappers"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestMapper_MapServiceToken_Success(t *testing.T) {
	type tc struct {
		name      string
		token     *database.ServiceToken
		hasDelete bool
	}

	createdAt := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	expiresAt := time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC)
	deletedAt := time.Date(2024, 6, 15, 10, 0, 0, 0, time.UTC)

	cases := []tc{
		{
			name: "active token without deleted_at",
			token: &database.ServiceToken{
				ID:        "token-id-123",
				Name:      "API Token",
				CreatedAt: createdAt,
				ExpiresAt: expiresAt,
			},
			hasDelete: false,
		},
		{
			name: "revoked token with deleted_at",
			token: &database.ServiceToken{
				ID:        "token-id-456",
				Name:      "Revoked Token",
				CreatedAt: createdAt,
				ExpiresAt: expiresAt,
				DeletedAt: gorm.DeletedAt{Time: deletedAt, Valid: true},
			},
			hasDelete: true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			decimalSvc := NewMockDecimalSvc(gomock.NewController(t))
			mapper := mappers.NewMapper(&mappers.MapperConfig{
				DecimalSvc: decimalSvc,
			})

			result := mapper.MapServiceToken(context.TODO(), c.token)

			assert.Equal(t, c.token.ID, result.Id)
			assert.Equal(t, c.token.Name, result.Name)
			assert.Equal(t, c.token.CreatedAt, result.CreatedAt.AsTime())
			assert.Equal(t, c.token.ExpiresAt, result.ExpiresAt.AsTime())

			if c.hasDelete {
				assert.NotNil(t, result.DeletedAt)
				assert.Equal(t, c.token.DeletedAt.Time, result.DeletedAt.AsTime())
			} else {
				assert.Nil(t, result.DeletedAt)
			}
		})
	}
}
