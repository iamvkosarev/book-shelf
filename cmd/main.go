package main

import (
	"github.com/iamvkosarev/book-shelf/config"
	"github.com/iamvkosarev/book-shelf/internal/app"
	"log"
)

const (
	cfgPath = "./config/config.yaml"
)

func main() {
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		log.Fatalf("loading config error: %s\n", err)
	}
	if err = app.Run(cfg); err != nil {
		log.Fatalf("app error: %s\n", err)
	}
}
