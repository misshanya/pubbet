package config

import "github.com/ilyakaznacheev/cleanenv"

type Config struct {
	ServerAddress string `env:"PUBBET_SERVER_ADDRESS" env-default:":5000"`
}

func New() (*Config, error) {
	var cfg Config

	// Read .env file
	// If failed to read file, will try ReadEnv
	if err := cleanenv.ReadConfig(".env", &cfg); err == nil {
		return &cfg, nil
	}

	// Read env
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
