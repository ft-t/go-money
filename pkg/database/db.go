package database

import (
	"context"
	"github.com/rs/zerolog/log"
	"go-money/pkg/boilerplate"
	"go-money/pkg/configuration"
	"go-money/pkg/testing"
	"gorm.io/gorm"
)

type dbKey struct {
}
type DbType int

const (
	DbTypeMaster   = DbType(1)
	DbTypeReadonly = DbType(2)
)

var masterGormDb *gorm.DB
var readonlyGormDb *gorm.DB

func init() {
	config := configuration.GetConfiguration()

	log.Info().Msg("setup postgres database")

	if boilerplate.GetCurrentEnvironment() == boilerplate.Ci {
		if err := testing.EnsurePostgresDbExists(config.Db); err != nil {
			panic(err)
		}
	}

	mainDb, err := boilerplate.GetGormConnection(config.Db)

	if err != nil {
		panic(err)
	}

	masterGormDb = mainDb
	readonlyGormDb = masterGormDb

	//migrations := GetMigrations()
	//if len(migrations) > 0 {
	//	log.Info().Msg("[Db] start migrations")
	//	if err = gormigrate.New(mainDb, gormigrate.DefaultOptions, migrations).Migrate(); err != nil {
	//		panic(err)
	//	}
	//}
}

func GetDb(t DbType) *gorm.DB {
	switch t {
	case DbTypeMaster:
		return masterGormDb
	case DbTypeReadonly:
		return readonlyGormDb
	default:
		return masterGormDb
	}
}

func GetDbWithContext(ctx context.Context, t DbType) *gorm.DB {
	return GetDb(t).WithContext(ctx)
}

func WithContext(ctx context.Context, db *gorm.DB) context.Context {
	return context.WithValue(ctx, dbKey{}, db.WithContext(ctx))
}

func WithContextType(ctx context.Context, dbType DbType) context.Context {
	return context.WithValue(ctx, dbKey{}, GetDbWithContext(ctx, dbType))
}

func FromContext(ctx context.Context, defaultInstance *gorm.DB) *gorm.DB {
	v := ctx.Value(dbKey{})

	if v != nil {
		return v.(*gorm.DB)
	}

	return defaultInstance.WithContext(ctx)
}
