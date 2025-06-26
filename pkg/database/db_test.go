package database_test

import (
	"context"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDb(t *testing.T) {
	assert.NotNil(t, database.GetDb(database.DbTypeMaster))
	assert.NotNil(t, database.GetDb(database.DbTypeReadonly))
}

func TestContext(t *testing.T) {
	ctx := database.WithContext(context.TODO(), database.GetDb(database.DbTypeMaster))
	assert.NotNil(t, ctx)

	db := database.FromContext(ctx, database.GetDb(database.DbTypeReadonly))
	assert.NotNil(t, db)
}

func TestWithContextType(t *testing.T) {
	ctx := database.WithContextType(context.TODO(), database.DbTypeMaster)
	assert.NotNil(t, ctx)
}

func TestFromContext(t *testing.T) {
	db := database.GetDbWithContext(context.TODO(), database.DbTypeMaster)

	assert.NotNil(t, database.FromContext(context.TODO(), db))
}
