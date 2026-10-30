package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"flowledger/internal/adapter/httpapi"
	"flowledger/internal/adapter/sqlite"
	"flowledger/internal/usecase"
	"flowledger/web"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// 1. Parse configuration from flags and environment
	var (
		portFlag = flag.Int("port", 8080, "HTTP server listening port")
		dbFlag   = flag.String("db", "flowledger.db", "SQLite database file path")
		envFlag  = flag.String("env", "development", "Application environment (development, demo, production)")
		seedFlag = flag.Bool("seed", false, "Seed database with realistic Indonesian financial records")
		demoFlag = flag.Bool("demo", false, "Enable demo mode with automatic initial data seeding")
	)
	flag.Parse()

	// Environment variable overrides
	if p := os.Getenv("PORT"); p != "" {
		var envPort int
		if _, err := fmt.Sscanf(p, "%d", &envPort); err == nil && envPort > 0 {
			*portFlag = envPort
		}
	}
	if d := os.Getenv("DB_PATH"); d != "" {
		*dbFlag = d
	}
	if e := os.Getenv("APP_ENV"); e != "" {
		*envFlag = e
	}

	// 2. Configure structured observability with log/slog
	var logHandler slog.Handler
	if *envFlag == "production" || *envFlag == "demo" {
		logHandler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	} else {
		logHandler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	}
	logger := slog.New(logHandler)
	slog.SetDefault(logger)

	logger.Info("starting FlowLedger engine",
		slog.String("version", "v2.0"),
		slog.String("environment", *envFlag),
		slog.Int("port", *portFlag),
		slog.String("db_path", *dbFlag),
	)

	// 3. Initialize SQLite with mandatory Pragmas and auto-migrations
	db, err := sqlite.Open(*dbFlag)
	if err != nil {
		return fmt.Errorf("initializing sqlite: %w", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			logger.Error("error closing database", slog.String("error", err.Error()))
		}
	}()

	// 4. Seed data if requested or in demo mode
	if *seedFlag || *demoFlag {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := sqlite.SeedRealisticIndonesianData(ctx, db); err != nil {
			logger.Warn("data seeding warning", slog.String("error", err.Error()))
		} else {
			logger.Info("realistic Indonesian financial data successfully verified/seeded")
		}
	}

	// 5. Wire Clean Architecture dependencies
	catRepo := sqlite.NewCategoryRepository(db)
	txRepo := sqlite.NewTransactionRepository(db)
	budgetRepo := sqlite.NewBudgetRepository(db)
	idempotencyRepo := sqlite.NewIdempotencyRepository(db)
	reportRepo := sqlite.NewReportRepository(db)

	catUsecase := usecase.NewCategoryUsecase(catRepo)
	txUsecase := usecase.NewTransactionUsecase(txRepo, catRepo)
	budgetUsecase := usecase.NewBudgetUsecase(budgetRepo, catRepo)
	reportUsecase := usecase.NewReportUsecase(reportRepo, budgetRepo)

	router := httpapi.NewRouter(httpapi.RouterConfig{
		Logger:          logger,
		DB:              db,
		CategoryUsecase: catUsecase,
		TxUsecase:       txUsecase,
		BudgetUsecase:   budgetUsecase,
		ReportUsecase:   reportUsecase,
		IdempotencyRepo: idempotencyRepo,
		WebHandler:      web.Handler(),
	})

	// 6. Hardened HTTP server configuration
	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", *portFlag),
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// 7. Graceful lifecycle management with OS signal trapping
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	serverErrChan := make(chan error, 1)
	go func() {
		logger.Info("HTTP server listening", slog.String("addr", server.Addr))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrChan <- err
		}
	}()

	select {
	case err := <-serverErrChan:
		return fmt.Errorf("server error: %w", err)
	case <-ctx.Done():
		logger.Info("shutdown signal received, commencing graceful teardown")
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server graceful shutdown failed: %w", err)
	}

	logger.Info("FlowLedger engine stopped cleanly")
	return nil
}

