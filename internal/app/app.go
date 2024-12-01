package app

import (
	"context"
	"effective_mobile_tz/config"
	v1 "effective_mobile_tz/internal/controller/http/v1"
	"effective_mobile_tz/internal/repository"
	"effective_mobile_tz/internal/service"
	"effective_mobile_tz/pkg/validator"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"os/signal"
)

// @title Song Library Service
// @version 1.0
// @description This is a service for managing songs in the library.
// @host localhost:8080
// @BasePath /api/v1
// @schemes http

func Run(configPath string) {
	ctx := context.Background()

	// Configurations set up
	cfg, err := config.NewConfig(configPath)
	if err != nil {
		log.Fatalf("Config error: %s", err)
	}

	// Logger
	SetupLogrus(cfg.Log.Level)

	// Connecting to Postgres
	log.Info("Connecting postgres...")
	pg, err := pgx.Connect(ctx, cfg.PG.URL)
	if err != nil {
		log.Fatal(fmt.Errorf("error connecting postgres: %w", err))
	}
	defer pg.Close(ctx)

	// Running Migrations
	log.Info("Running migrations...")
	err = RunMigrations(cfg.PG.URL, cfg.PG.MigrationPath)
	if err != nil {
		log.Debug(fmt.Errorf("error running migrations: %w", err))
	}

	// Repositories
	log.Info("Initializing repositories...")
	repositories := repository.NewRepository(pg)

	// Service
	log.Info("Initializing services")
	dependencies := service.Dependencies{
		Repository:     repositories,
		ExternalApiURL: cfg.ExternalAPI.URL,
	}
	services := service.NewService(dependencies)

	// Handler
	log.Info("Initializing handlers and routes...")
	handler := echo.New()
	handler.Validator = validator.NewCustomValidator()
	v1.NewRouter(handler, services)

	// HTTP server
	log.Info("Starting http server...")
	log.Debugf("Server port: %s", cfg.HTTP.Port)

	httpServer := &http.Server{
		Addr:    cfg.HTTP.Port,
		Handler: handler,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, os.Kill)

	done := make(chan struct{})

	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("error starting server: %v", err)
		}
	}()

	// Graceful Shutdown
	go func() {
		<-stop
		ctx, cancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
		defer cancel()

		if err := httpServer.Shutdown(ctx); err != nil {
			log.Fatalf("error shutdown: %v", err)
		}

		close(done)
	}()

	<-done
	log.Info("Server stopped gracefully.")
}
