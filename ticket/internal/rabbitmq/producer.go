package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"ticket/internal/service"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog"
)

type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	Exchange string
}

type Metrics interface {
	IncRabbitMQEvent(exchange, result string)
}

type Producer struct {
	conn     *amqp.Connection
	channel  *amqp.Channel
	exchange string
	metrics  Metrics
	logger   zerolog.Logger
}

func NewProducer(cfg Config, metrics Metrics, logger zerolog.Logger) (*Producer, error) {
	uri := fmt.Sprintf("amqp://%s:%s@%s:%d/", cfg.User, cfg.Password, cfg.Host, cfg.Port)
	conn, err := amqp.Dial(uri)
	if err != nil {
		return nil, fmt.Errorf("failed to connect rabbitmq: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("failed to open rabbitmq channel: %w", err)
	}

	if err := channel.ExchangeDeclare(
		cfg.Exchange,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		_ = channel.Close()
		_ = conn.Close()
		return nil, fmt.Errorf("failed to declare rabbitmq exchange: %w", err)
	}

	return &Producer{
		conn:     conn,
		channel:  channel,
		exchange: cfg.Exchange,
		metrics:  metrics,
		logger:   logger,
	}, nil
}

func (p *Producer) Publish(ctx context.Context, event service.Event) error {
	payload, err := json.Marshal(event)
	if err != nil {
		p.inc("failed")
		return err
	}

	err = p.channel.PublishWithContext(
		ctx,
		p.exchange,
		event.Type,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			MessageId:    event.ID,
			Type:         event.Type,
			Timestamp:    time.Now().UTC(),
			Body:         payload,
			Headers: amqp.Table{
				"ticket_id": event.TicketID,
			},
		},
	)
	if err != nil {
		p.inc("failed")
		return fmt.Errorf("failed to publish rabbitmq event: %w", err)
	}

	p.logger.Debug().Str("event_id", event.ID).Str("event_type", event.Type).Msg("rabbitmq event published")
	p.inc("published")
	return nil
}

func (p *Producer) Close() error {
	if err := p.channel.Close(); err != nil {
		_ = p.conn.Close()
		return err
	}
	return p.conn.Close()
}

func (p *Producer) inc(result string) {
	if p.metrics != nil {
		p.metrics.IncRabbitMQEvent(p.exchange, result)
	}
}

type NoopProducer struct{}

func (NoopProducer) Publish(context.Context, service.Event) error {
	return nil
}
