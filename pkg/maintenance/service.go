package maintenance

type Config struct {
	StatsSvc StatsSvc
}
type Service struct {
	cfg *Config
}

func NewService(cfg *Config) *Service {
	return &Service{
		cfg: cfg,
	}
}
