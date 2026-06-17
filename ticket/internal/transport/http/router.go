package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"ticket/internal/model"
	"ticket/internal/service"
	ticketswagger "ticket/internal/swagger"
	mw "ticket/internal/transport/http/middleware"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
)

type TicketService interface {
	CreateTicket(ctx context.Context, req model.CreateTicketRequest) (*model.Ticket, error)
	GetTicket(ctx context.Context, id string) (*model.Ticket, error)
	GetTickets(ctx context.Context) ([]model.Ticket, error)
	UpdateTicket(ctx context.Context, id string, req model.UpdateTicketRequest) (*model.Ticket, error)
	DeleteTicket(ctx context.Context, id string) error
	GetCurrentDecision(ctx context.Context, ticketID string) (*model.DecisionResponse, error)
	ChooseProblem(ctx context.Context, ticketID string, req model.ChooseProblemRequest) (*model.DecisionResponse, error)
	GetDecisionNode(id string) (model.Node, error)
	Ping(ctx context.Context) error
}

type Metrics interface {
	Handler() http.Handler
	Middleware(handlerName string, next http.Handler) http.Handler
}

type Handler struct {
	service TicketService
	metrics Metrics
	logger  zerolog.Logger
}

func NewRouter(service TicketService, metrics Metrics, logger zerolog.Logger, authServiceURL string) http.Handler {
	h := &Handler{service: service, metrics: metrics, logger: logger}

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(h.logRequest)

	r.Get("/health/live", h.liveness)
	r.Get("/health/ready", h.readiness)
	r.Handle("/metrics", metrics.Handler())
	r.Get("/swagger", ticketswagger.UIHandler)
	r.Get("/swagger/", ticketswagger.UIHandler)
	r.Get("/swagger/openapi.yaml", ticketswagger.SpecHandler)

	r.Route("/tickets", func(r chi.Router) {

		r.Use(mw.Auth(authServiceURL))

		r.Method(http.MethodPost, "/", metrics.Middleware("create_ticket", http.HandlerFunc(h.createTicket)))
		r.Method(http.MethodGet, "/", metrics.Middleware("get_tickets", http.HandlerFunc(h.getTickets)))
		r.Method(http.MethodGet, "/{ticketID}", metrics.Middleware("get_ticket", http.HandlerFunc(h.getTicket)))
		r.Method(http.MethodPatch, "/{ticketID}", metrics.Middleware("update_ticket", http.HandlerFunc(h.updateTicket)))
		r.Method(http.MethodDelete, "/{ticketID}", metrics.Middleware("delete_ticket", http.HandlerFunc(h.deleteTicket)))
		r.Method(http.MethodGet, "/{ticketID}/decision", metrics.Middleware("get_current_decision", http.HandlerFunc(h.getCurrentDecision)))
		r.Method(http.MethodPost, "/{ticketID}/decision", metrics.Middleware("choose_problem", http.HandlerFunc(h.chooseProblem)))
	})

	r.Method(http.MethodGet, "/decision-tree/{nodeID}", metrics.Middleware("get_decision_node", http.HandlerFunc(h.getDecisionNode)))

	return r
}

func (h *Handler) createTicket(w http.ResponseWriter, r *http.Request) {
	var req model.CreateTicketRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	ticket, err := h.service.CreateTicket(r.Context(), req)
	if err != nil {
		h.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, ticket)
}

func (h *Handler) getTicket(w http.ResponseWriter, r *http.Request) {
	ticket, err := h.service.GetTicket(r.Context(), chi.URLParam(r, "ticketID"))
	if err != nil {
		h.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, ticket)
}

func (h *Handler) getTickets(w http.ResponseWriter, r *http.Request) {
	tickets, err := h.service.GetTickets(r.Context())
	if err != nil {
		h.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, tickets)
}

func (h *Handler) updateTicket(w http.ResponseWriter, r *http.Request) {
	var req model.UpdateTicketRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	ticket, err := h.service.UpdateTicket(r.Context(), chi.URLParam(r, "ticketID"), req)
	if err != nil {
		h.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, ticket)
}

func (h *Handler) deleteTicket(w http.ResponseWriter, r *http.Request) {
	if err := h.service.DeleteTicket(r.Context(), chi.URLParam(r, "ticketID")); err != nil {
		h.writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) getCurrentDecision(w http.ResponseWriter, r *http.Request) {
	resp, err := h.service.GetCurrentDecision(r.Context(), chi.URLParam(r, "ticketID"))
	if err != nil {
		h.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) chooseProblem(w http.ResponseWriter, r *http.Request) {
	var req model.ChooseProblemRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	resp, err := h.service.ChooseProblem(r.Context(), chi.URLParam(r, "ticketID"), req)
	if err != nil {
		h.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) getDecisionNode(w http.ResponseWriter, r *http.Request) {
	node, err := h.service.GetDecisionNode(chi.URLParam(r, "nodeID"))
	if err != nil {
		h.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, node)
}

func (h *Handler) liveness(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) readiness(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	if err := h.service.Ping(ctx); err != nil {
		h.logger.Error().Err(err).Msg("readiness failed")
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "not_ready"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

func (h *Handler) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(rec, r)
		h.logger.Info().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Int("status", rec.statusCode).
			Dur("duration", time.Since(start)).
			Str("request_id", middleware.GetReqID(r.Context())).
			Msg("http request")
	})
}

func (h *Handler) writeError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	message := "internal error"

	switch {
	case errors.Is(err, service.ErrInvalidInput), errors.Is(err, service.ErrInvalidStep):
		status = http.StatusBadRequest
		message = err.Error()
	case errors.Is(err, service.ErrNotFound):
		status = http.StatusNotFound
		message = "not found"
	default:
		h.logger.Error().Err(err).Msg("request failed")
	}

	writeJSON(w, status, map[string]string{"error": message})
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return false
	}
	return true
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.statusCode = code
	r.ResponseWriter.WriteHeader(code)
}
