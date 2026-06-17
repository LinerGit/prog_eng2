package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"ticket/internal/model"
	"ticket/internal/repository/mapper"
	repositorymodel "ticket/internal/repository/model"

	"gorm.io/gorm"
)

var ErrNotFound = errors.New("ticket not found")

type Metrics interface {
	ObserveDB(operation string, startedAt time.Time, err error)
}

type Repository struct {
	db      *gorm.DB
	metrics Metrics
}

func NewRepository(db *gorm.DB, metrics Metrics) *Repository {
	return &Repository{db: db, metrics: metrics}
}

func New(db *gorm.DB, metrics Metrics) *Repository {
	return NewRepository(db, metrics)
}

func (r *Repository) CreateTicket(ctx context.Context, ticket *model.Ticket) error {
	startedAt := time.Now()
	err := r.db.WithContext(ctx).Create(mapper.ToRepo(ticket)).Error
	r.observe("create_ticket", startedAt, err)
	if err != nil {
		return fmt.Errorf("failed to create ticket: %w", err)
	}
	return nil
}

func (r *Repository) UpdateTicket(ctx context.Context, ticket *model.Ticket) error {
	startedAt := time.Now()
	result := r.db.WithContext(ctx).
		Model(&repositorymodel.Ticket{}).
		Where("id = ? AND deleted_at IS NULL", ticket.ID).
		Updates(map[string]any{
			"title":           ticket.Title,
			"description":     ticket.Description,
			"status":          string(ticket.Status),
			"current_node_id": ticket.CurrentNodeID,
			"solution":        ticket.Solution,
			"updated_at":      time.Now().UTC(),
		})
	r.observe("update_ticket", startedAt, result.Error)
	if result.Error != nil {
		return fmt.Errorf("failed to update ticket: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) DeleteTicket(ctx context.Context, id string) error {
	now := time.Now().UTC()
	startedAt := time.Now()
	result := r.db.WithContext(ctx).
		Model(&repositorymodel.Ticket{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Updates(map[string]any{
			"deleted_at": now,
			"updated_at": now,
		})
	r.observe("delete_ticket", startedAt, result.Error)
	if result.Error != nil {
		return fmt.Errorf("failed to delete ticket: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) GetTicket(ctx context.Context, id string) (*model.Ticket, error) {
	ticket := &repositorymodel.Ticket{}
	startedAt := time.Now()
	err := r.db.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", id).
		First(ticket).
		Error
	r.observe("get_ticket", startedAt, err)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get ticket: %w", err)
	}
	return mapper.ToDomain(ticket), nil
}

func (r *Repository) GetTickets(ctx context.Context) ([]model.Ticket, error) {
	var tickets []repositorymodel.Ticket
	startedAt := time.Now()
	err := r.db.WithContext(ctx).
		Where("deleted_at IS NULL").
		Order("created_at DESC").
		Find(&tickets).
		Error
	r.observe("get_tickets", startedAt, err)
	if err != nil {
		return nil, fmt.Errorf("failed to get tickets: %w", err)
	}
	return mapper.ToDomainList(tickets), nil
}

func (r *Repository) Ping(ctx context.Context) error {
	sqlDB, err := r.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.PingContext(ctx)
}

func (r *Repository) observe(operation string, startedAt time.Time, err error) {
	if r.metrics != nil {
		r.metrics.ObserveDB(operation, startedAt, err)
	}
}
