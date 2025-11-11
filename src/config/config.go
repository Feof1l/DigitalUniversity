package config

import (
	"os"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"

	"digitalUniversity/logger"
)

type Config struct {
	LogLevel logger.LogLevel `env:"LOG_LEVEL" envDefault:"1"`
	LogDir   string          `env:"LOG_DIR" envDefault:"./logs"`
	Database DatabaseConfig  `envPrefix:"DATABASE_"`
	MaxAPI   MaxConfig       `envPrefix:"MAX_"`
}

type MaxConfig struct {
	Token string `env:"TOKEN"`
}

type DatabaseConfig struct {
	URI string `env:"URI"`
}

func Load() (*Config, error) {
	if err := godotenv.Load(".env"); err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
	}

	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
