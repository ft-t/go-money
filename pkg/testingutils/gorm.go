package testingutils

import (
	"context"
	"fmt"
	"github.com/ft-t/go-money/pkg/boilerplate"
	"github.com/jackc/pgx/v5"
	"github.com/juju/fslock"
	"github.com/rs/zerolog/log"
	"os"
	"path"
	"strings"
)

func getLock() *fslock.Lock {
	return fslock.New(path.Join(os.TempDir(), "go_fs_lock"))
}

func FlushAllTables(config boilerplate.DbConfig) error {
	return flushInternal(config, nil)
}

func ensureThatItsLocal(config boilerplate.DbConfig) {
	allowedHosts := []string{"localhost", "127.0.0.1", "postgres"}

	for _, h := range allowedHosts {
		if strings.EqualFold(h, config.Host) {
			return
		}
	}

	if strings.HasPrefix(config.Host, "192.168.") { // home network
		return
	}

	panic("host is not allowed, please check config for postgres")
}

func flushInternal(config boilerplate.DbConfig, tables []string) error {
	ensureThatItsLocal(config)

	rawStr, _ := boilerplate.GetDbConnectionString(config)

	conn, err := pgx.Connect(context.Background(), rawStr)

	if err != nil {
		return err
	}

	defer func() {
		_ = conn.Close(context.Background())
	}()

	res, err := conn.Query(context.Background(), "SELECT table_schema, table_name FROM information_schema.tables where table_schema != 'pg_catalog' and table_schema != 'information_schema';")

	if err != nil {
		return err
	}

	existing := map[string]bool{}

	for res.Next() {
		values := res.RawValues()

		resultPath := fmt.Sprintf("%v.%v", string(values[0]), string(values[1]))
		existing[resultPath] = true
	}

	builder := strings.Builder{}

	builder.WriteString("BEGIN TRANSACTION; ")

	if tables != nil {
		for _, table := range tables {
			if strings.ToLower(table) == "public.migrations" {
				continue
			}

			if _, ok := existing[strings.ToLower(table)]; ok {

				builder.WriteString(fmt.Sprintf(" truncate table %v CASCADE; ", strings.ToLower(table)))
			} else {
				log.Warn().Msgf("table %v does not exists", table)
			}
		}
	} else {
		for name := range existing {
			if strings.ToLower(name) == "public.migrations" {
				continue
			}
			builder.WriteString(fmt.Sprintf(" truncate table %v CASCADE; ", strings.ToLower(name)))
		}
	}

	builder.WriteString(" COMMIT;")

	_, err = conn.Exec(context.Background(), builder.String())

	if err != nil {
		return err
	}

	return nil
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

	defer func() {
		_ = conn.Close(context.Background())
	}()

	r, err := conn.Query(context.Background(), fmt.Sprintf("SELECT * FROM pg_database WHERE datname='%v'", oldDbName))

	if err != nil {
		return err
	}

	r.Next()

	exists := len(r.RawValues()) == 0

	if !exists {
		_, err = conn.Exec(context.Background(), fmt.Sprintf("CREATE DATABASE %v;", oldDbName))

		if err != nil {
			return err
		}
	}

	return nil
}
