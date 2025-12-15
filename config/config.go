package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"time"
)

type Http struct {
	Port string `yaml:"port"`
}

type App struct {
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout"`
	LogMode         string        `yaml:"log_mode" default:"debug"`
}

type Config struct {
	Http Http `yaml:"http"`
	App  App  `yaml:"app"`
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
