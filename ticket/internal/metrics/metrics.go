package metrics

import (
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Metrics struct {
	registry *prometheus.Registry

	httpInFlight *prometheus.GaugeVec
	httpRequests *prometheus.CounterVec
	httpDuration *prometheus.HistogramVec
	dbDuration   *prometheus.HistogramVec
	dbErrors     *prometheus.CounterVec
	ticketEvents *prometheus.CounterVec
	rabbitEvents *prometheus.CounterVec
}

func New(serviceName string) *Metrics {
	namespace := sanitizeNamespace(serviceName)
	m := &Metrics{
		registry: prometheus.NewRegistry(),
		httpInFlight: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "http_requests_in_flight",
			Help:      "HTTP requests currently being served.",
		}, []string{"handler"}),
		httpRequests: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "http_requests_total",
			Help:      "Total HTTP requests by handler, method, and code.",
		}, []string{"handler", "method", "code"}),
		httpDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "http_request_duration_seconds",
			Help:      "HTTP request duration.",
			Buckets:   prometheus.DefBuckets,
		}, []string{"handler", "method"}),
		dbDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "db_query_duration_seconds",
			Help:      "Database query duration by operation.",
			Buckets:   prometheus.DefBuckets,
		}, []string{"operation"}),
		dbErrors: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "db_errors_total",
			Help:      "Database errors by operation.",
		}, []string{"operation"}),
		ticketEvents: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "ticket_events_total",
			Help:      "Ticket business events by type.",
		}, []string{"event"}),
		rabbitEvents: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "rabbitmq_events_total",
			Help:      "RabbitMQ publish attempts by result and exchange.",
		}, []string{"exchange", "result"}),
	}

	m.registry.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		m.httpInFlight,
		m.httpRequests,
		m.httpDuration,
		m.dbDuration,
		m.dbErrors,
		m.ticketEvents,
		m.rabbitEvents,
	)

	return m
}

func sanitizeNamespace(serviceName string) string {
	serviceName = strings.TrimSpace(serviceName)
	if serviceName == "" {
		return "ticket_service"
	}
	re := regexp.MustCompile(`[^a-zA-Z0-9_]`)
	return re.ReplaceAllString(serviceName, "_")
}

func (m *Metrics) Handler() http.Handler {
	return promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{
		Registry:          m.registry,
		EnableOpenMetrics: true,
	})
}

func (m *Metrics) Middleware(handlerName string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, statusCode: http.StatusOK}

		m.httpInFlight.WithLabelValues(handlerName).Inc()
		defer m.httpInFlight.WithLabelValues(handlerName).Dec()
		defer func() {
			code := strconv.Itoa(rec.statusCode)
			m.httpRequests.WithLabelValues(handlerName, r.Method, code).Inc()
			m.httpDuration.WithLabelValues(handlerName, r.Method).Observe(time.Since(start).Seconds())
		}()

		next.ServeHTTP(rec, r)
	})
}

func (m *Metrics) ObserveDB(operation string, startedAt time.Time, err error) {
	m.dbDuration.WithLabelValues(operation).Observe(time.Since(startedAt).Seconds())
	if err != nil {
		m.dbErrors.WithLabelValues(operation).Inc()
	}
}

func (m *Metrics) IncTicketEvent(event string) {
	m.ticketEvents.WithLabelValues(event).Inc()
}

func (m *Metrics) IncRabbitMQEvent(exchange, result string) {
	m.rabbitEvents.WithLabelValues(exchange, result).Inc()
}

type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.statusCode = code
	r.ResponseWriter.WriteHeader(code)
}
