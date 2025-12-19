package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"time"
)

type Database struct {
	URL string `env:"DB_URL"`
}

type Http struct {
	Port         string        `env:"HTTP_PORT" env-default:"8081"`
	ReadTimeout  time.Duration `env:"HTTP_READ_TIMEOUT" env-default:"5s"`
	WriteTimeout time.Duration `env:"HTTP_WRITE_TIMEOUT" env-default:"5s"`
}

type Router struct {
	APITimeout time.Duration `env:"API_TIMEOUT" env-default:"5s"`
}

type App struct {
	ShutdownTimeout time.Duration `env:"SHUTDOWN_TIMEOUT" env-default:"10s"`
	LogMode         string        `env:"LOG_MODE" env-default:"debug"` // debug, dev or prod
}

type Authorization struct {
	PrivateKey string        `env:"PRIVATE_KEY"`
	PublicKey  string        `env:"PUBLIC_KEY"`
	TokenTTL   time.Duration `env:"TOKEN_TTL"`
}

type Config struct {
	Authorization Authorization
	Http          Http
	Router        Router
	Database      Database
	App           App
}

func LoadConfig() (*Config, error) {
	var cfg Config
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
