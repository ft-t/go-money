package configuration

import "github.com/ft-t/go-money/pkg/boilerplate"

type Configuration struct {
	Db                   boilerplate.DbConfig `env:", prefix=DB_"`
	ReadOnlyDb           boilerplate.DbConfig `env:", prefix=READONLY_DB_"`
	GrpcPort             int                  `env:"GRPC_PORT, default=52055"`
	JwtPrivateKey        string               `env:"JWT_PRIVATE_KEY"`
	ExchangeRatesUrl     string               `env:"EXCHANGE_RATES_URL, default=http://go-money-exchange-rates.s3-website.eu-north-1.amazonaws.com/latest.json"`
	StaticFilesDirectory string               `env:"STATIC_FILES_DIRECTORY"`
	CurrencyConfig       CurrencyConfig       `env:", prefix=CURRENCY_CONFIG_"`
	GrafanaConfig        GrafanaConfig        `env:", prefix=GRAFANA_CONFIG_"`
}

type CurrencyConfig struct {
	UpdateTransactionAmountInBaseCurrency bool   `env:"UPDATE_TRANSACTION_AMOUNT_IN_BASE_CURRENCY, default=false"`
	BaseCurrency                          string `env:"BASE_CURRENCY, default=USD"`
}

type GrafanaConfig struct {
	Url string `env:"URL"`
}
