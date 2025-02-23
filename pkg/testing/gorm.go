package testing

import (
	"context"
	"fmt"
	"github.com/ft-t/go-money/pkg/boilerplate"
	"github.com/jackc/pgx/v5"
	"github.com/juju/fslock"
	"os"
	"path"
)

func getLock() *fslock.Lock {
	return fslock.New(path.Join(os.TempDir(), "go_fs_lock"))
}

func EnsurePostgresDbExists(config boilerplate.DbConfig) error {
	lock := getLock()

	if err := lock.Lock(); err != nil {
		return err
	}

	defer func() {
		_ = lock.Unlock()
	}()

	oldDbName := config.Db

	config.Db = "postgres"
	rawStr, _ := boilerplate.GetDbConnectionString(config)

	conn, err := pgx.Connect(context.Background(), rawStr)
	config.Db = oldDbName

	if err != nil {
		return err
	}

	defer conn.Close(context.Background())

	r, err := conn.Query(context.Background(), fmt.Sprintf("SELECT * FROM pg_database WHERE datname='%v'", oldDbName))

	if err != nil {
		return err
	}

	r.Next()

	exists := true

	if len(r.RawValues()) == 0 {
		exists = false
	}

	if !exists {
		_, err = conn.Exec(context.Background(), fmt.Sprintf("CREATE DATABASE %v;", oldDbName))

		if err != nil {
			return err
		}
	}

	return nil
}
