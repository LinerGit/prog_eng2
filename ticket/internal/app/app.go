package app

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"ticket/internal/config"
	"ticket/internal/metrics"
	ticketrabbitmq "ticket/internal/rabbitmq"
	"ticket/internal/repository"
	"ticket/internal/service"
	transporthttp "ticket/internal/transport/http"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/rs/zerolog"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type App struct {
	cfg    *config.Config
	logger zerolog.Logger
}

func New(logger zerolog.Logger, cfg *config.Config) *App {
	return &App{
		cfg:    cfg,
		logger: logger,
	}
}

func (a *App) Run(ctx context.Context) error {
	if a.cfg.RunMigrations {
		if err := a.runMigrations(ctx); err != nil {
			return err
		}
	}

	db, err := gorm.Open(postgres.Open(a.cfg.DbDsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to unwrap database: %w", err)
	}
	sqlDB.SetMaxOpenConns(a.cfg.DBMaxOpenConns)
	sqlDB.SetMaxIdleConns(a.cfg.DBMaxIdleConns)
	sqlDB.SetConnMaxLifetime(a.cfg.DBConnMaxLifetime)
	defer sqlDB.Close()

	m := metrics.New(a.cfg.ServiceName)
	repo := repository.NewRepository(db, m)

	var producer service.EventProducer = ticketrabbitmq.NoopProducer{}
	var closeProducer func() error
	if a.cfg.RabbitMQEnabled {
		rabbitProducer, err := ticketrabbitmq.NewProducer(ticketrabbitmq.Config{
			Host:     a.cfg.RabbitMQHost,
			Port:     a.cfg.RabbitMQPort,
			User:     a.cfg.RabbitMQUser,
			Password: a.cfg.RabbitMQPassword,
			Exchange: a.cfg.RabbitMQExchange,
		}, m, a.logger)
		if err != nil {
			return err
		}
		producer = rabbitProducer
		closeProducer = rabbitProducer.Close
		a.logger.Info().
			Str("host", a.cfg.RabbitMQHost).
			Int("port", a.cfg.RabbitMQPort).
			Str("exchange", a.cfg.RabbitMQExchange).
			Msg("rabbitmq producer enabled")
	} else {
		a.logger.Warn().Msg("rabbitmq producer disabled")
	}
	if closeProducer != nil {
		defer func() {
			if err := closeProducer(); err != nil {
				a.logger.Error().Err(err).Msg("failed to close rabbitmq producer")
			}
		}()
	}

	ticketService := service.NewTicketService(repo, producer, m, a.logger)
	router := transporthttp.NewRouter(ticketService, m, a.logger)

	server := &http.Server{
		Addr:              fmt.Sprintf("%s:%d", a.cfg.Host, a.cfg.Port),
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		a.logger.Info().Str("addr", server.Addr).Msg("http server started")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		a.logger.Info().Msg("http server shutting down")
		return server.Shutdown(shutdownCtx)
	case err := <-errCh:
		return err
	}
}

func (a *App) runMigrations(ctx context.Context) error {
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}

	db, err := sql.Open("pgx", a.cfg.DbDsn)
	if err != nil {
		return fmt.Errorf("failed to connect database for migrations: %w", err)
	}
	defer db.Close()

	if err := goose.UpContext(ctx, db, a.cfg.MigrationsDir); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}
	a.logger.Info().Str("dir", a.cfg.MigrationsDir).Msg("migrations applied")
	return nil
}
