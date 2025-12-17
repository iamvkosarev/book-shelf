package app

import (
	"context"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/iamvkosarev/book-shelf/config"
	"github.com/iamvkosarev/book-shelf/internal/handler"
	"github.com/iamvkosarev/book-shelf/internal/router"
	"github.com/iamvkosarev/book-shelf/internal/storage/postgres"
	"github.com/iamvkosarev/book-shelf/internal/usecase"
	"github.com/iamvkosarev/book-shelf/pkg/logs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func Run(cfg *config.Config) error {
	var joinedErrors error
	log, err := logs.NewSlogLogger(cfg.App.LogMode, os.Stdout)
	if err != nil {
		joinedErrors = errors.Join(joinedErrors, fmt.Errorf("failed to initialize logger: %w", err))
		return err
	}
	slog.SetDefault(log)
	newRouter := mux.NewRouter()

	address := fmt.Sprintf(":%s", cfg.Http.Port)
	server := &http.Server{
		Handler:      newRouter,
		Addr:         address,
		ReadTimeout:  cfg.Http.ReadTimeout * time.Second,
		WriteTimeout: cfg.Http.WriteTimeout * time.Second,
	}
	ctx, cancel := context.WithCancel(context.Background())

	pool, err := postgres.NewPool(ctx, cfg.Database)
	if err != nil {
		joinedErrors = errors.Join(joinedErrors, fmt.Errorf("failed to initialize postgres pool: %w", err))
		cancel()
	} else {
		log.Info("start connect postgres")
	}
	personsStorage := postgres.NewPersonsStorage(pool)
	personsUsecase := usecase.NewPersonsUsecase(personsStorage)
	personsHandler := handler.NewPersonHandler(personsUsecase)

	authorsStorage := postgres.NewAuthorsStorage(pool)
	authorsUsecase := usecase.NewAuthorsUsecase(authorsStorage, personsUsecase)
	authorsHandler := handler.NewAuthorHandler(authorsUsecase)

	publishersStorage := postgres.NewPublishersStorage(pool)
	publishersUsecase := usecase.NewPublishersUsecase(publishersStorage)
	publisherHandler := handler.NewPublisherHandler(publishersUsecase)

	router.Setup(
		newRouter, cfg.Router, router.Deps{
			PublishersHandler: publisherHandler,
			AuthorsHandler:    authorsHandler,
			PersonsHandler:    personsHandler,
		},
	)

	go func() {
		log.Info("start server", slog.String("address", address))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			joinedErrors = errors.Join(joinedErrors, fmt.Errorf("server err: %w", err))
			cancel()
		}
	}()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)

	log.Info("start app")
	select {
	case <-ctx.Done():
	case <-signalChan:
		cancel()
	}

	wg := sync.WaitGroup{}
	wgChan := make(chan struct{})

	stopCtx, termCancel := context.WithTimeout(context.Background(), cfg.App.ShutdownTimeout)
	defer termCancel()

	log.Info("start shutdown")
	wg.Add(1)
	go func() {
		defer wg.Done()
		if shutdownErr := server.Shutdown(stopCtx); err != nil {
			err = errors.Join(err, shutdownErr)
		}
		log.Info("stop server")
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if pool == nil {
			return
		}
		pool.Close()
		log.Info("stop connect postgres")
	}()

	go func() {
		defer close(wgChan)
		wg.Wait()
	}()

	select {
	case <-wgChan:
		log.Info("stop shutdown")
	case <-stopCtx.Done():
		log.Info("finish context: timeout")
	}

	log.Info("stop app")
	return joinedErrors
}
