package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/ramindesu/Card-Number-To-IBAN/backend/internal/cardtoiban/client/zarinhub"
	cardhandler "github.com/ramindesu/Card-Number-To-IBAN/backend/internal/cardtoiban/handler"
	"github.com/ramindesu/Card-Number-To-IBAN/backend/internal/cardtoiban/model"
	cardqueue "github.com/ramindesu/Card-Number-To-IBAN/backend/internal/cardtoiban/queue"
	cardpostgres "github.com/ramindesu/Card-Number-To-IBAN/backend/internal/cardtoiban/repository/postgres"
	"github.com/ramindesu/Card-Number-To-IBAN/backend/internal/cardtoiban/service"
	"github.com/ramindesu/Card-Number-To-IBAN/backend/internal/platform/config"
	"github.com/ramindesu/Card-Number-To-IBAN/backend/internal/platform/database"
	"github.com/ramindesu/Card-Number-To-IBAN/backend/internal/platform/httpserver"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if err := run(logger); err != nil {
		logger.Error("API stopped", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	databaseConfig, err := config.LoadDatabase()
	if err != nil {
		return fmt.Errorf("load database configuration: %w", err)
	}

	zarinHubConfig, err := config.LoadZarinHub()
	if err != nil {
		return fmt.Errorf("load ZarinHub configuration: %w", err)
	}

	cardProtectionConfig, err := config.LoadCardProtection()
	if err != nil {
		return fmt.Errorf("load card protection configuration: %w", err)
	}

	cardQueueConfig, err := config.LoadCardQueue()
	if err != nil {
		return fmt.Errorf("load Card-to-IBAN queue configuration: %w", err)
	}

	httpConfig, err := config.LoadHTTPServer()
	if err != nil {
		return fmt.Errorf("load HTTP configuration: %w", err)
	}

	pool, err := database.NewPostgresPool(ctx, databaseConfig)
	if err != nil {
		return err
	}
	defer pool.Close()

	providerClient, err := zarinhub.NewCardToIBANClient(zarinhub.Config{
		BaseURL:        zarinHubConfig.BaseURL,
		Username:       zarinHubConfig.Username,
		Password:       zarinHubConfig.Password,
		RequestTimeout: zarinHubConfig.RequestTimeout,
	})
	if err != nil {
		return fmt.Errorf("create ZarinHub client: %w", err)
	}

	queueExecutor, err := cardqueue.NewBoundedExecutor(cardqueue.Config{
		Capacity:          cardQueueConfig.Capacity,
		Workers:           cardQueueConfig.Workers,
		RequestsPerSecond: cardQueueConfig.RequestsPerSecond,
		Burst:             cardQueueConfig.Burst,
	})
	if err != nil {
		return fmt.Errorf("create Card-to-IBAN queue: %w", err)
	}
	defer queueExecutor.Close()

	repository := cardpostgres.NewCardToIBANRepository(pool)
	applicationService, err := service.NewCardToIBANService(service.CardToIBANServiceConfig{
		Repository:     repository,
		Client:         providerClient,
		FingerprintKey: cardProtectionConfig.FingerprintKey,
		Executor:       queueExecutor,
		MaxAttempts:    cardQueueConfig.MaxAttempts,
		RetryBaseDelay: cardQueueConfig.RetryBaseDelay,
		HandlingType: model.CardToIBANRequestHandlingType(
			zarinHubConfig.RequestHandlingType,
		),
	})
	if err != nil {
		return fmt.Errorf("create Card-to-IBAN service: %w", err)
	}

	httpHandler, err := cardhandler.NewCardToIBANHandler(applicationService)
	if err != nil {
		return fmt.Errorf("create Card-to-IBAN handler: %w", err)
	}

	router := newRouter(httpHandler)
	server, err := httpserver.New(httpConfig, router)
	if err != nil {
		return fmt.Errorf("create HTTP server: %w", err)
	}
	logger.Info("HTTP server starting", "address", httpConfig.Address())

	return server.Run(ctx)
}

func newRouter(cardToIBANHandler *cardhandler.CardToIBANHandler) http.Handler {
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.Recoverer)
	router.Get("/healthz", httpserver.Health)
	router.Route("/api/v1", cardToIBANHandler.RegisterRoutes)
	return router
}
