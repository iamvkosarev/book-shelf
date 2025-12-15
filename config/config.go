package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"time"
)

type Database struct {
	URL string `env:"DB_URL"`
}

type Http struct {
	Port string `env:"HTTP_PORT"`
}

type App struct {
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout"`
	LogMode         string        `yaml:"log_mode" default:"debug"`
}

type Config struct {
	Http     Http `yaml:"http"`
	Database Database
	App      App `yaml:"app"`
}

func LoadConfig(cfgPath string) (*Config, error) {
	var cfg Config
	if err := cleanenv.ReadConfig(cfgPath, &cfg); err != nil {
		return nil, err
	}
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
