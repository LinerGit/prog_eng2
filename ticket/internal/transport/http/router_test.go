package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"ticket/internal/metrics"
	"ticket/internal/model"
	"ticket/internal/service"

	"github.com/rs/zerolog"
)

func TestRouterCreateAndChooseProblem(t *testing.T) {
	repo := newRouterMemoryRepo()
	svc := service.NewTicketService(repo, serviceProducer{}, nil, zerolog.Nop())
	router := NewRouter(svc, metrics.New("ticket_service_test"), zerolog.Nop(), "http://users-service:8080")

	body := bytes.NewBufferString(`{"title":"problem","description":"details"}`)
	req := httptest.NewRequest(http.MethodPost, "/tickets/", body)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}

	var ticket model.Ticket
	if err := json.NewDecoder(rec.Body).Decode(&ticket); err != nil {
		t.Fatalf("decode ticket: %v", err)
	}

	body = bytes.NewBufferString(`{"next_node_id":"product_problem"}`)
	req = httptest.NewRequest(http.MethodPost, "/tickets/"+ticket.ID+"/decision", body)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
}

func TestRouterServesSwagger(t *testing.T) {
	repo := newRouterMemoryRepo()
	svc := service.NewTicketService(repo, serviceProducer{}, nil, zerolog.Nop())
	router := NewRouter(svc, metrics.New("ticket_service_test"), zerolog.Nop(), "http://users-service:8080")

	req := httptest.NewRequest(http.MethodGet, "/swagger/openapi.yaml", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("openapi status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("openapi: 3.0.3")) {
		t.Fatalf("openapi document missing version: %s", rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/swagger/", nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("swagger ui status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("SwaggerUIBundle")) {
		t.Fatalf("swagger ui body missing bundle bootstrap")
	}
}

type routerMemoryRepo struct {
	tickets map[string]*model.Ticket
}

func newRouterMemoryRepo() *routerMemoryRepo {
	return &routerMemoryRepo{tickets: make(map[string]*model.Ticket)}
}

func (r *routerMemoryRepo) CreateTicket(_ context.Context, ticket *model.Ticket) error {
	cloned := *ticket
	r.tickets[ticket.ID] = &cloned
	return nil
}

func (r *routerMemoryRepo) GetTicket(_ context.Context, id string) (*model.Ticket, error) {
	ticket, ok := r.tickets[id]
	if !ok {
		return nil, service.ErrNotFound
	}
	cloned := *ticket
	return &cloned, nil
}

func (r *routerMemoryRepo) GetTickets(context.Context) ([]model.Ticket, error) {
	result := make([]model.Ticket, 0, len(r.tickets))
	for _, ticket := range r.tickets {
		result = append(result, *ticket)
	}
	return result, nil
}

func (r *routerMemoryRepo) UpdateTicket(_ context.Context, ticket *model.Ticket) error {
	cloned := *ticket
	r.tickets[ticket.ID] = &cloned
	return nil
}

func (r *routerMemoryRepo) DeleteTicket(_ context.Context, id string) error {
	delete(r.tickets, id)
	return nil
}

func (r *routerMemoryRepo) Ping(context.Context) error {
	return nil
}

type serviceProducer struct{}

func (serviceProducer) Publish(context.Context, service.Event) error {
	return nil
}
