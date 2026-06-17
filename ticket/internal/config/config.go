package config

import (
	"log"
	"strings"
	"time"

	"github.com/caarlos0/env/v10"
	"github.com/joho/godotenv"
)

type Config struct {
	ServiceName string `env:"SERVICE_NAME" envDefault:"ticket-service"`
	AppEnv      string `env:"APP_ENV" envDefault:"development"`
	Host        string `env:"HTTP_HOST" envDefault:"0.0.0.0"`
	Port        int    `env:"HTTP_PORT" envDefault:"8080"`
	LogLevel    string `env:"LOG_LEVEL" envDefault:"info"`

	DbDsn             string        `env:"DB_DSN,required"`
	DBMaxOpenConns    int           `env:"DB_MAX_OPEN_CONNS" envDefault:"25"`
	DBMaxIdleConns    int           `env:"DB_MAX_IDLE_CONNS" envDefault:"10"`
	DBConnMaxLifetime time.Duration `env:"DB_CONN_MAX_LIFETIME" envDefault:"30m"`
	RunMigrations     bool          `env:"RUN_MIGRATIONS" envDefault:"true"`
	MigrationsDir     string        `env:"MIGRATIONS_DIR" envDefault:"migrations"`
	AuthServiceURL    string        `env:"AUTH_SERVICE_URL" envDefault:"http://users-service:8080"`

	RabbitMQEnabled  bool   `env:"RABBITMQ_ENABLED" envDefault:"true"`
	RabbitMQHost     string `env:"RABBITMQ_HOST" envDefault:"rabbitmq"`
	RabbitMQPort     int    `env:"RABBITMQ_PORT" envDefault:"5672"`
	RabbitMQUser     string `env:"RABBITMQ_USER" envDefault:"guest"`
	RabbitMQPassword string `env:"RABBITMQ_PASSWORD" envDefault:"guest"`
	RabbitMQExchange string `env:"RABBITMQ_EXCHANGE" envDefault:"warehouse"`
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}

	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}

	if strings.TrimSpace(cfg.RabbitMQHost) == "" {
		cfg.RabbitMQEnabled = false
	}

	return cfg, nil
}
