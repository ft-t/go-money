package mappers_test

import (
	"context"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/ft-t/go-money/pkg/mappers"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestMapTag(t *testing.T) {
	m := mappers.NewMapper(&mappers.MapperConfig{})

	resp := m.MapTag(context.TODO(), &database.Tag{
		ID:        123,
		Name:      "test",
		Color:     "color",
		Icon:      "icon",
		CreatedAt: time.Time{},
	})

	assert.EqualValues(t, 123, resp.Id)
	assert.Equal(t, "test", resp.Name)
	assert.Equal(t, "color", resp.Color)
	assert.Equal(t, "icon", resp.Icon)
}
