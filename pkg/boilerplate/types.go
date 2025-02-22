package boilerplate

type DbConfig struct {
	Host                     string `json:"Host"`
	Port                     int    `json:"Port"`
	Db                       string `json:"Db"`
	User                     string `json:"User"`
	Password                 string `json:"Password"`
	MaxIdleConnections       int    `json:"MaxIdleConnections"`
	MaxConnectionLifetimeSec int    `json:"MaxConnectionLifetimeSec"`
	MaxOpenConnections       int    `json:"MaxOpenConnections"`
	MaxConnectionIdleSec     int    `json:"MaxConnectionIdleSec"`
}
