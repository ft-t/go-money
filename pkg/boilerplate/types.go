package boilerplate

type DbConfig struct {
	Host                     string `env:"HOST, default=localhost"`
	Port                     int    `env:"PORT, default=5432"`
	Db                       string `env:"DB, default=money"`
	User                     string `env:"USER, default=money"`
	Password                 string `env:"PASSWORD, default=postgres"`
	MaxIdleConnections       int    `env:"MAX_IDLE_CONNECTIONS"`
	MaxConnectionLifetimeSec int    `env:"MAX_CONNECTION_LIFETIME_SEC"`
	MaxOpenConnections       int    `env:"MAX_OPEN_CONNECTIONS"`
	MaxConnectionIdleSec     int    `env:"MAX_CONNECTION_IDLE_SEC"`
}
