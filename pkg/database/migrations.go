package database

import (
	"fmt"

	"github.com/ft-t/go-money/pkg/boilerplate"
	"github.com/ft-t/go-money/pkg/configuration"
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func getMigrations(cfg *configuration.Configuration) []*gormigrate.Migration {
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
		{
			ID: "2025-03-03-FirstTransactionAt",
			Migrate: func(db *gorm.DB) error {
				return boilerplate.ExecuteSql(db,
					`alter table accounts add column if not exists first_transaction_at timestamp;`,
				)
			},
		},
		{
			ID: "2025-06-16-AddTitleToTransaction",
			Migrate: func(db *gorm.DB) error {
				return boilerplate.ExecuteSql(db,
					`alter table transactions add column if not exists title text;`,
					`alter table transactions
    alter column source_amount drop not null;

alter table transactions
    alter column destination_amount drop not null;
`,
				)
			},
		},
		{
			ID: "2025-06-22-AddImportDeduplication",
			Migrate: func(db *gorm.DB) error {
				return boilerplate.ExecuteSql(db,
					`create table if not exists import_deduplication
(
    import_source integer   not null,
    key           text      not null,
    created_at    timestamp not null,
    transaction_id bigint    not null,
    constraint import_deduplication_pk
        primary key (import_source, key)
);`,
					`alter table transactions add column if not exists reference_number text;`,
					`alter table transactions add column if not exists internal_reference_number text;`,
				)
			},
		},
		{
			ID: "2025-06-24-RenameToDisplayOrder",
			Migrate: func(db *gorm.DB) error {
				return boilerplate.ExecuteSql(db,
					`alter table accounts rename column position to display_order;`)
			},
		},
		{
			ID: "2025-06-25-AddTags",
			Migrate: func(db *gorm.DB) error {
				return boilerplate.ExecuteSql(db,
					`create table if not exists tags
(
    id         serial
        constraint tags_pk
            primary key,
    name       text      not null,
    color      text,
    icon       text,
    created_at timestamp not null,
    deleted_at timestamp
);`,
					`create unique index ix_tag_name on tags (name) where (deleted_at is null)`,
					`alter table transactions rename column label_ids to tag_ids;`,
				)
			},
		},
		{
			ID: "2025-06-26-AmountsInBaseCurrency",
			Migrate: func(db *gorm.DB) error {
				return boilerplate.ExecuteSql(db,
					`alter table transactions
    add column if not exists destination_amount_in_base_currency numeric;

alter table transactions
    add column if not exists source_amount_in_base_currency numeric;
`)
			},
		},
		{
			ID: "2025-06-28-AddRules",
			Migrate: func(db *gorm.DB) error {
				return boilerplate.ExecuteSql(db,
					`create table if not exists rules
(
    id               serial,
    script           text      not null,
    interpreter_type integer   not null,
    sort_order       integer   not null,
    created_at       timestamp not null,
    updated_at       timestamp not null,
    enabled          bool      not null,
    is_final_rule    bool      not null,
    deleted_at       timestamp,
    group_name       text      not null
);
`,
				)
			},
		},
		{
			ID: "2025-06-29-AddRuleTitle",
			Migrate: func(db *gorm.DB) error {
				return boilerplate.ExecuteSql(db,
					`alter table rules add column if not exists title text;`,
				)
			},
		},
		{
			ID: "2025-07-12-AddTxIndexes",
			Migrate: func(db *gorm.DB) error {
				return boilerplate.ExecuteSql(db,
					`create index if not exists ix_source_dest_tx on transactions (source_account_id, destination_account_id, transaction_date_only) include (source_amount, destination_amount);
							create index if not exists ix_source_tx on transactions (source_account_id, transaction_date_only) include (source_amount, destination_amount);
							create index if not exists ix_dest_tx on transactions (destination_account_id, transaction_date_only) include (source_amount, destination_amount);
							create index if not exists ix_latest_stat on daily_stat(account_id, date desc);`,
				)
			},
		},
		{
			ID: "2025-07-13-AddCategories",
			Migrate: func(db *gorm.DB) error {
				return boilerplate.ExecuteSql(db,
					`create table if not exists categories
(
    id         serial
        constraint categories_pk
            primary key,
    name       text      not null,
    created_at timestamp not null,
    deleted_at timestamp
);
create unique index if not exists categories__uindex
    on categories (name) where (deleted_at is null);
alter table transactions
    add column if not exists category_id int null;
`)
			},
		},
		{
			ID: "2025-07-25-ActivateCurrencies",
			Migrate: func(db *gorm.DB) error {
				return boilerplate.ExecuteSql(db,
					`with currencies as (select currency
                    from accounts
                    union
                    select source_currency
                    from transactions
                    where coalesce(source_currency, '') != ''
                    union
                    select destination_currency
                    from transactions
                    where coalesce(destination_currency, '') != '')
update currencies
set is_active = true
where id in (select * from currencies)`,
				)
			},
		},
		{
			ID: "2025-07-26-AddScheduleRules",
			Migrate: func(db *gorm.DB) error {
				return boilerplate.ExecuteSql(db,
					`create table if not exists schedule_rules
(
    id               serial,
    script           text      not null,
    interpreter_type integer   not null,
    created_at       timestamp not null,
    updated_at       timestamp not null,
    enabled          bool      not null,
    deleted_at       timestamp,
    group_name       text      not null,
	last_run_at     timestamp,
	cron_expression text not null,
	title text not null
);
`,
				)
			},
		},
		{
			ID: "2025-08-10-AddFxSourceCurrency",
			Migrate: func(db *gorm.DB) error {
				queries := []string{
					`alter table transactions add column if not exists fx_source_currency text;`,
					`alter table transactions add column if not exists fx_source_amount DECIMAL;`,
					`begin ;
update transactions set fx_source_currency = destination_currency where transaction_type = 3 and destination_currency != '';
update transactions set destination_currency = '' where transaction_type = 3;

update transactions set fx_source_amount = -abs(destination_amount) where transaction_type = 3 and destination_amount != 0;
update transactions set destination_amount = 0 where transaction_type = 3;
update accounts set type = 1 where type in (2,3);
commit ;
`,
				}

				queries = append(queries, generateDefaultAccounts(cfg)...)

				for _, q := range queries {
					if err := db.Exec(q).Error; err != nil {
						return err
					}
				}

				return nil
			},
		},
		{
			ID: "2025-08-14-AddDoubleEntry",
			Migrate: func(db *gorm.DB) error {
				return boilerplate.ExecuteSql(db,
					`create table if not exists double_entries
(
    id                      bigserial
        constraint double_entries_pk
            primary key,
    transaction_id          bigint    not null,
    is_debit                boolean   not null,
    amount_in_base_currency decimal,
    base_currency           text,
    account_id              integer   not null,
    created_at              timestamp not null,
    deleted_at              timestamp
);

create index if not exists ix_transaction on double_entries (transaction_id);
create unique index if not exists ix_uniq_record on double_entries (transaction_id, is_debit) where (deleted_at is null);
`)
			},
		},
		{
			ID: "2025-08-19-AddDefaultCurrency",
			Migrate: func(db *gorm.DB) error {
				return boilerplate.ExecuteSql(db,
					fmt.Sprintf("insert into currencies(id, rate, is_active, decimal_places, updated_at) select '%v', 1, true, 2, now() on conflict do nothing;", cfg.CurrencyConfig.BaseCurrency))
			},
		},
		{
			ID: "2025-09-26-AddTransactionSoftDelete",
			Migrate: func(db *gorm.DB) error {
				return boilerplate.ExecuteSql(db,
					`alter table transactions add column if not exists deleted_at timestamp;`)
			},
		},
		{
			ID: "2025-10-09-AddTransactionDateToDoubleEntry",
			Migrate: func(db *gorm.DB) error {
				return boilerplate.ExecuteSql(db,
					`alter table double_entries add column if not exists transaction_date timestamp;`,
					`update double_entries de
					 set transaction_date = t.transaction_date_time
					 from transactions t
					 where de.transaction_id = t.id
					 and de.transaction_date is null;`,
					`create index if not exists ix_double_entries_transaction_date on double_entries(account_id, transaction_date) where deleted_at is null;`,
				)
			},
		},
	}
}
