package mappers_test

import (
	"context"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/ft-t/go-money/pkg/mappers"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"testing"
	"time"
)

func TestMapCategory(t *testing.T) {
	input := &database.Category{
		ID:   1,
		Name: "Test Category",
		DeletedAt: gorm.DeletedAt{
			Time:  time.Now().UTC(),
			Valid: true,
		},
	}

	mapper := mappers.NewMapper(&mappers.MapperConfig{})

	mapped := mapper.MapCategory(context.TODO(), input)

	assert.EqualValues(t, input.ID, mapped.Id)
	assert.EqualValues(t, input.Name, mapped.Name)
	assert.EqualValues(t, input.DeletedAt.Time, mapped.DeletedAt.AsTime())
}
