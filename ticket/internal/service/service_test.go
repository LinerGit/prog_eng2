package service

import (
	"context"
	"errors"
	"testing"

	"ticket/internal/model"

	"github.com/rs/zerolog"
)

func TestCreateTicketStartsDecisionTree(t *testing.T) {
	repo := newMemoryRepo()
	producer := &memoryProducer{}
	svc := NewTicketService(repo, producer, nil, zerolog.Nop())

	ticket, err := svc.CreateTicket(context.Background(), model.CreateTicketRequest{
		Title:       "problem",
		Description: "details",
	})
	if err != nil {
		t.Fatalf("CreateTicket() error = %v", err)
	}

	if ticket.ID == "" {
		t.Fatal("expected generated ticket id")
	}
	if ticket.Status != model.StatusOpen {
		t.Fatalf("status = %q, want %q", ticket.Status, model.StatusOpen)
	}
	if ticket.CurrentNodeID != "start" {
		t.Fatalf("current node = %q, want start", ticket.CurrentNodeID)
	}
	if len(producer.events) != 1 || producer.events[0].Type != "ticket.created" {
		t.Fatalf("unexpected producer events: %+v", producer.events)
	}
}

func TestChooseProblemReturnsSolutionAndClosesTicket(t *testing.T) {
	repo := newMemoryRepo()
	svc := NewTicketService(repo, &memoryProducer{}, nil, zerolog.Nop())

	ticket, err := svc.CreateTicket(context.Background(), model.CreateTicketRequest{Title: "problem"})
	if err != nil {
		t.Fatalf("CreateTicket() error = %v", err)
	}

	if _, err := svc.ChooseProblem(context.Background(), ticket.ID, model.ChooseProblemRequest{NextNodeID: "product_problem"}); err != nil {
		t.Fatalf("ChooseProblem(product_problem) error = %v", err)
	}

	resp, err := svc.ChooseProblem(context.Background(), ticket.ID, model.ChooseProblemRequest{NextNodeID: "barcode_solution"})
	if err != nil {
		t.Fatalf("ChooseProblem(barcode_solution) error = %v", err)
	}

	if !resp.Done {
		t.Fatal("expected decision flow to be done")
	}
	if resp.Ticket.Status != model.StatusClosed {
		t.Fatalf("status = %q, want %q", resp.Ticket.Status, model.StatusClosed)
	}
	if resp.Ticket.Solution == "" {
		t.Fatal("expected solution")
	}
}

func TestChooseProblemRejectsUnavailableOption(t *testing.T) {
	repo := newMemoryRepo()
	svc := NewTicketService(repo, nil, nil, zerolog.Nop())

	ticket, err := svc.CreateTicket(context.Background(), model.CreateTicketRequest{Title: "problem"})
	if err != nil {
		t.Fatalf("CreateTicket() error = %v", err)
	}

	_, err = svc.ChooseProblem(context.Background(), ticket.ID, model.ChooseProblemRequest{NextNodeID: "barcode_solution"})
	if !errors.Is(err, ErrInvalidStep) {
		t.Fatalf("error = %v, want ErrInvalidStep", err)
	}
}

type memoryRepo struct {
	tickets map[string]*model.Ticket
}

func newMemoryRepo() *memoryRepo {
	return &memoryRepo{tickets: make(map[string]*model.Ticket)}
}

func (r *memoryRepo) CreateTicket(_ context.Context, ticket *model.Ticket) error {
	r.tickets[ticket.ID] = cloneTicket(ticket)
	return nil
}

func (r *memoryRepo) GetTicket(_ context.Context, id string) (*model.Ticket, error) {
	ticket, ok := r.tickets[id]
	if !ok {
		return nil, ErrNotFound
	}
	return cloneTicket(ticket), nil
}

func (r *memoryRepo) GetTickets(context.Context) ([]model.Ticket, error) {
	tickets := make([]model.Ticket, 0, len(r.tickets))
	for _, ticket := range r.tickets {
		tickets = append(tickets, *cloneTicket(ticket))
	}
	return tickets, nil
}

func (r *memoryRepo) UpdateTicket(_ context.Context, ticket *model.Ticket) error {
	if _, ok := r.tickets[ticket.ID]; !ok {
		return ErrNotFound
	}
	r.tickets[ticket.ID] = cloneTicket(ticket)
	return nil
}

func (r *memoryRepo) DeleteTicket(_ context.Context, id string) error {
	if _, ok := r.tickets[id]; !ok {
		return ErrNotFound
	}
	delete(r.tickets, id)
	return nil
}

func (r *memoryRepo) Ping(context.Context) error {
	return nil
}

type memoryProducer struct {
	events []Event
}

func (p *memoryProducer) Publish(_ context.Context, event Event) error {
	p.events = append(p.events, event)
	return nil
}

func cloneTicket(ticket *model.Ticket) *model.Ticket {
	cloned := *ticket
	return &cloned
}
