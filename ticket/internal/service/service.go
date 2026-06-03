package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"ticket/internal/model"

	"github.com/rs/zerolog"
)

var (
	ErrInvalidInput = errors.New("invalid input")
	ErrNotFound     = errors.New("not found")
	ErrInvalidStep  = errors.New("invalid decision step")
)

type Repository interface {
	CreateTicket(ctx context.Context, ticket *model.Ticket) error
	GetTicket(ctx context.Context, id string) (*model.Ticket, error)
	GetTickets(ctx context.Context) ([]model.Ticket, error)
	UpdateTicket(ctx context.Context, ticket *model.Ticket) error
	DeleteTicket(ctx context.Context, id string) error
	Ping(ctx context.Context) error
}

type EventProducer interface {
	Publish(ctx context.Context, event Event) error
}

type Metrics interface {
	IncTicketEvent(event string)
}

type Event struct {
	ID        string         `json:"id"`
	Type      string         `json:"type"`
	TicketID  string         `json:"ticket_id"`
	CreatedAt time.Time      `json:"created_at"`
	Payload   map[string]any `json:"payload,omitempty"`
}

type TicketService struct {
	repo     Repository
	producer EventProducer
	metrics  Metrics
	logger   zerolog.Logger
	tree     map[string]model.Node
}

func NewTicketService(repo Repository, producer EventProducer, metrics Metrics, logger zerolog.Logger) *TicketService {
	return &TicketService{
		repo:     repo,
		producer: producer,
		metrics:  metrics,
		logger:   logger,
		tree:     defaultDecisionTree(),
	}
}

func (s *TicketService) CreateTicket(ctx context.Context, req model.CreateTicketRequest) (*model.Ticket, error) {
	if strings.TrimSpace(req.Title) == "" {
		return nil, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}

	now := time.Now().UTC()
	ticket := &model.Ticket{
		ID:            newID(),
		Title:         strings.TrimSpace(req.Title),
		Description:   strings.TrimSpace(req.Description),
		Status:        model.StatusOpen,
		CurrentNodeID: "start",
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := s.repo.CreateTicket(ctx, ticket); err != nil {
		return nil, err
	}

	s.logger.Info().Str("ticket_id", ticket.ID).Msg("ticket created")
	s.inc("created")
	s.publish(ctx, "ticket.created", ticket.ID, map[string]any{"title": ticket.Title})
	return ticket, nil
}

func (s *TicketService) GetTicket(ctx context.Context, id string) (*model.Ticket, error) {
	ticket, err := s.repo.GetTicket(ctx, strings.TrimSpace(id))
	if err != nil {
		return nil, normalizeNotFound(err)
	}
	return ticket, nil
}

func (s *TicketService) GetTickets(ctx context.Context) ([]model.Ticket, error) {
	return s.repo.GetTickets(ctx)
}

func (s *TicketService) UpdateTicket(ctx context.Context, id string, req model.UpdateTicketRequest) (*model.Ticket, error) {
	ticket, err := s.GetTicket(ctx, id)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(req.Title) != "" {
		ticket.Title = strings.TrimSpace(req.Title)
	}
	ticket.Description = strings.TrimSpace(req.Description)
	if req.Status != "" {
		if !validStatus(req.Status) {
			return nil, fmt.Errorf("%w: unknown status", ErrInvalidInput)
		}
		ticket.Status = req.Status
	}
	ticket.UpdatedAt = time.Now().UTC()

	if err := s.repo.UpdateTicket(ctx, ticket); err != nil {
		return nil, normalizeNotFound(err)
	}

	s.logger.Info().Str("ticket_id", ticket.ID).Msg("ticket updated")
	s.inc("updated")
	s.publish(ctx, "ticket.updated", ticket.ID, map[string]any{"status": ticket.Status})
	return ticket, nil
}

func (s *TicketService) DeleteTicket(ctx context.Context, id string) error {
	if err := s.repo.DeleteTicket(ctx, strings.TrimSpace(id)); err != nil {
		return normalizeNotFound(err)
	}
	s.logger.Info().Str("ticket_id", id).Msg("ticket deleted")
	s.inc("deleted")
	s.publish(ctx, "ticket.deleted", id, nil)
	return nil
}

func (s *TicketService) GetCurrentDecision(ctx context.Context, ticketID string) (*model.DecisionResponse, error) {
	ticket, err := s.GetTicket(ctx, ticketID)
	if err != nil {
		return nil, err
	}
	node, ok := s.tree[ticket.CurrentNodeID]
	if !ok {
		return nil, fmt.Errorf("%w: current node not found", ErrInvalidStep)
	}
	return &model.DecisionResponse{Ticket: ticket, Node: node, Done: node.Solution != ""}, nil
}

func (s *TicketService) ChooseProblem(ctx context.Context, ticketID string, req model.ChooseProblemRequest) (*model.DecisionResponse, error) {
	ticket, err := s.GetTicket(ctx, ticketID)
	if err != nil {
		return nil, err
	}
	current, ok := s.tree[ticket.CurrentNodeID]
	if !ok {
		return nil, fmt.Errorf("%w: current node not found", ErrInvalidStep)
	}

	nextID := strings.TrimSpace(req.NextNodeID)
	if !canMove(current, nextID) {
		return nil, fmt.Errorf("%w: option is not available from current node", ErrInvalidStep)
	}

	next, ok := s.tree[nextID]
	if !ok {
		return nil, fmt.Errorf("%w: target node not found", ErrInvalidStep)
	}

	ticket.CurrentNodeID = next.ID
	ticket.Solution = next.Solution
	ticket.Status = model.StatusInProgress
	if next.Solution != "" {
		ticket.Status = model.StatusClosed
	}
	ticket.UpdatedAt = time.Now().UTC()

	if err := s.repo.UpdateTicket(ctx, ticket); err != nil {
		return nil, normalizeNotFound(err)
	}

	s.logger.Info().
		Str("ticket_id", ticket.ID).
		Str("node_id", next.ID).
		Bool("done", next.Solution != "").
		Msg("decision step selected")
	s.inc("decision_step")
	s.publish(ctx, "ticket.decision_step_selected", ticket.ID, map[string]any{
		"node_id": next.ID,
		"done":    next.Solution != "",
	})

	return &model.DecisionResponse{Ticket: ticket, Node: next, Done: next.Solution != ""}, nil
}

func (s *TicketService) GetDecisionNode(id string) (model.Node, error) {
	node, ok := s.tree[strings.TrimSpace(id)]
	if !ok {
		return model.Node{}, ErrNotFound
	}
	return node, nil
}

func (s *TicketService) Ping(ctx context.Context) error {
	return s.repo.Ping(ctx)
}

func (s *TicketService) publish(ctx context.Context, eventType, ticketID string, payload map[string]any) {
	if s.producer == nil {
		return
	}
	event := Event{
		ID:        newID(),
		Type:      eventType,
		TicketID:  ticketID,
		CreatedAt: time.Now().UTC(),
		Payload:   payload,
	}
	if err := s.producer.Publish(ctx, event); err != nil {
		s.logger.Error().Err(err).Str("ticket_id", ticketID).Str("event_type", eventType).Msg("failed to publish event")
	}
}

func (s *TicketService) inc(event string) {
	if s.metrics != nil {
		s.metrics.IncTicketEvent(event)
	}
}

func defaultDecisionTree() map[string]model.Node {
	return map[string]model.Node{
		"start": {
			ID:   "start",
			Text: "Какая проблема возникла?",
			Options: []model.Option{
				{Text: "Проблема с товаром", NextNodeID: "product_problem"},
				{Text: "Проблема с доставкой", NextNodeID: "delivery_problem"},
				{Text: "Проблема с оплатой", NextNodeID: "payment_problem"},
			},
		},
		"product_problem": {
			ID:   "product_problem",
			Text: "Что случилось с товаром?",
			Options: []model.Option{
				{Text: "Не читается штрихкод", NextNodeID: "barcode_solution"},
				{Text: "Товар поврежден", NextNodeID: "defect_solution"},
			},
		},
		"delivery_problem": {
			ID:   "delivery_problem",
			Text: "Что случилось с доставкой?",
			Options: []model.Option{
				{Text: "Заказ не найден", NextNodeID: "delivery_lookup_solution"},
				{Text: "Не выдает заказ", NextNodeID: "delivery_issue_solution"},
			},
		},
		"payment_problem": {
			ID:   "payment_problem",
			Text: "Что случилось с оплатой?",
			Options: []model.Option{
				{Text: "Деньги списались, заказ не создан", NextNodeID: "payment_missing_order_solution"},
				{Text: "Оплата не проходит", NextNodeID: "payment_declined_solution"},
			},
		},
		"barcode_solution": {
			ID:       "barcode_solution",
			Text:     "Штрихкод не читается",
			Solution: "Проверьте читаемость штрихкода, очистите поверхность и попробуйте повторное сканирование. Если не помогло, оформите ручную проверку товара.",
		},
		"defect_solution": {
			ID:       "defect_solution",
			Text:     "Товар поврежден",
			Solution: "Зафиксируйте повреждение фото, переведите тикет в обработку брака и передайте товар ответственному сотруднику.",
		},
		"delivery_lookup_solution": {
			ID:       "delivery_lookup_solution",
			Text:     "Заказ не найден",
			Solution: "Проверьте номер заказа и телефон клиента. Если данных нет в системе, создайте обращение на ручной поиск.",
		},
		"delivery_issue_solution": {
			ID:       "delivery_issue_solution",
			Text:     "Не выдает заказ",
			Solution: "Проверьте статус заказа, ПВЗ и блокировки выдачи. При активной блокировке передайте тикет ответственному за доставку.",
		},
		"payment_missing_order_solution": {
			ID:       "payment_missing_order_solution",
			Text:     "Деньги списались, заказ не создан",
			Solution: "Проверьте платеж по транзакции. Если платеж успешен, создайте заявку на сверку и сообщите клиенту срок возврата или восстановления заказа.",
		},
		"payment_declined_solution": {
			ID:       "payment_declined_solution",
			Text:     "Оплата не проходит",
			Solution: "Попросите клиента проверить лимиты карты и повторить оплату. Если ошибка повторяется, предложите другой способ оплаты.",
		},
	}
}

func canMove(from model.Node, nextID string) bool {
	for _, option := range from.Options {
		if option.NextNodeID == nextID {
			return true
		}
	}
	return false
}

func validStatus(status model.Status) bool {
	switch status {
	case model.StatusOpen, model.StatusInProgress, model.StatusClosed:
		return true
	default:
		return false
	}
}

func normalizeNotFound(err error) error {
	if errors.Is(err, ErrNotFound) {
		return err
	}
	if strings.Contains(err.Error(), "not found") {
		return ErrNotFound
	}
	return err
}

func newID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%s-%s-%s-%s-%s",
		hex.EncodeToString(b[0:4]),
		hex.EncodeToString(b[4:6]),
		hex.EncodeToString(b[6:8]),
		hex.EncodeToString(b[8:10]),
		hex.EncodeToString(b[10:16]),
	)
}
