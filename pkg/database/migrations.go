package database

import (
	"github.com/ft-t/go-money/pkg/boilerplate"
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func getMigrations() []*gormigrate.Migration {
	return []*gormigrate.Migration{
		{
			ID: "2025-02-24-InitialUsers",
			Migrate: func(db *gorm.DB) error {
				return boilerplate.ExecuteSql(db,
					`create table if not exists users
								(
									id         serial    not null
										constraint users_pk
											primary key,
									login      text      not null,
									password   text      not null,
									created_at timestamp not null,
									deleted_at timestamp
								);`,
					`create unique index if not exists users_login_uindex
								on public.users (login)
								where deleted_at is null;`,
				)
			},
		},
	}
}
