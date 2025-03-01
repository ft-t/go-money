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
		{
			ID: "2025-02-25-InitialAccounts",
			Migrate: func(db *gorm.DB) error {
				return boilerplate.ExecuteSql(db, `create table accounts
						(
							id              serial             not null
								constraint accounts_pk
									primary key,
							name            text,
							current_balance decimal            not null,
							currency        text               not null,
							extra           jsonb default '{}' not null,
							flags           bigint             not null,
							last_updated_at timestamp          not null,
							created_at      timestamp          not null,
							deleted_at      timestamp,
							type            int,
							note            text               not null
						);
				`)
			},
		},
		{
			ID: "2025-02-25-MoreAccountFields",
			Migrate: func(db *gorm.DB) error {
				return boilerplate.ExecuteSql(db,
					`alter table accounts add column if not exists account_number text not null;`,
					`alter table accounts add column if not exists iban text not null;`,
					`alter table accounts add column if not exists liability_percent decimal;`,
					`alter table accounts add column if not exists position int;`,
				)
			},
		},
		{
			ID: "2025-02-25-AddCurrencyTable",
			Migrate: func(db *gorm.DB) error {
				return boilerplate.ExecuteSql(db,
					`create table currencies
					(
						id             text                  not null
							constraint currencies_pk
								primary key,
						rate           decimal default 1     not null,
						is_active      bool    default false not null,
						decimal_places integer default 2     not null,
						updated_at     timestamp             not null,
						deleted_at     timestamp
					);
					`,
				)
			},
		},
		{
			ID: "2025-03-01-AddTransactionTable",
			Migrate: func(db *gorm.DB) error {
				return boilerplate.ExecuteSql(db,
					`CREATE TABLE IF NOT EXISTS transactions (
								id BIGSERIAL PRIMARY KEY,
								
								source_amount DECIMAL NOT NULL,
								source_currency TEXT NOT NULL,
							
								destination_amount DECIMAL NOT NULL,
								destination_currency TEXT NOT NULL,
							
								source_account_id INT,
								destination_account_id INT,
							
								label_ids INTEGER[],
							
								created_at TIMESTAMP NOT NULL DEFAULT NOW(),
								updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
							
								notes TEXT,
								extra JSONB default '{}' not null,
							
								transaction_date_time TIMESTAMP NOT NULL,
								transaction_date_only DATE NOT NULL,
							
								transaction_type INT NOT NULL,
								flags BIGINT NOT NULL,
							
								voided_by_transaction_id BIGINT
							);`,
					`create table if not exists daily_stat
(
    account_id int  not null,
    date       date not null,
    amount     decimal,
    constraint daily_stat_pk
        primary key (account_id, date)
);`,
					`
create index if not exists daily_stat_account_id_index on public.daily_stat (account_id);
`,
				)
			},
		},
	}
}
