package main

import (
	"github.com/iamvkosarev/book-shelf/config"
	"github.com/iamvkosarev/book-shelf/internal/app"
	"log"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("loading config error: %s\n", err)
	}
	if err = app.Run(cfg); err != nil {
		log.Fatalf("app error: %s\n", err)
	}
}
