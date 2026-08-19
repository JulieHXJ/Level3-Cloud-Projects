package metrics

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// ------------API-----------
// create counter, format: GET, /instance， 200
var HTTPRequestsTotal = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "cloud3_http_requests_total",
		Help: "Total number of HTTP requests handled by API.",
	},
	[]string{"method", "route", "status"},
)

var HTTPRequestsDuration = promauto.NewHistogramVec(
	prometheus.HistogramOpts{
		Name: "cloud3_http_request_duration_seconds",
		Help: "Duration of HTTP requests handled by API",
	},
	[]string{"method", "route"},
	// add request_id, instance_id later
)

// ---------Paas business-------
// count only create, patch and delete
var InstanceOperationsTotal = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "cloud3_instance_operations_total",
		Help: "Total number of Cloud3 instance operations.",
	},
	[]string{"operation", "result"},
)

// get ListInstance + Inc()/Dec()
var InstancesCurrent = promauto.NewGauge(
	prometheus.GaugeOpts{
		Name: "cloud3_instances_current",
		Help: "Current number of instances managed by Cloud3.",
	},
)

// -------Protection---------
var RateLimitExceededTotal = promauto.NewCounter(
	prometheus.CounterOpts{
		Name: "cloud3_rate_limit_exceeded_total",
		Help: "Total number of requests rejected by rate limiting.",
	},
)

type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (w *statusRecorder) WriteHeader(status int) {
	w.statusCode = status
	w.ResponseWriter.WriteHeader(status)
}

// record http request and response status, count +1
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		start := time.Now()

		// wrapper for request method and the return code
		recorder := &statusRecorder{
			ResponseWriter: w,
			statusCode:     http.StatusOK, // by default
		}

		// let response go through recorder, so we can have status code
		next.ServeHTTP(recorder, r)

		route := routePaterrn(r)
		duration := time.Since(start).Seconds()

		// request total +1
		HTTPRequestsTotal.WithLabelValues(
			r.Method,
			route,
			strconv.Itoa(recorder.statusCode),
		).Inc()

		HTTPRequestsDuration.WithLabelValues(
			r.Method,
			route,
		).Observe(duration)
	})
}

// helper: retrive the path
func routePaterrn(r *http.Request) string {
	paterrn := r.Pattern

	if paterrn != "" && paterrn != "/" {
		if _, route, found := strings.Cut(paterrn, " "); found {
			return route
		}
		return paterrn
	}

	// when request are rejected during JWT validation
	path := r.URL.Path

	switch {
	case path == "/api/v1/health":
		return "/api/v1/health"

	case path == "/api/v1/auth/login":
		return "/api/v1/auth/login"

	case path == "/api/v1/auth/register":
		return "/api/v1/auth/register"

	case path == "/api/v1/instances":
		return "/api/v1/instances"

	case strings.HasSuffix(path, "/connection") &&
		strings.HasPrefix(path, "/api/v1/instances/"):
		return "/api/v1/instances/{id}/connection"

	case strings.HasSuffix(path, "/logs") &&
		strings.HasPrefix(path, "/api/v1/instances/"):
		return "/api/v1/instances/{id}/logs"

	case strings.HasPrefix(path, "/api/v1/instances/"):
		return "/api/v1/instances/{id}"

	default:
		return "unmatched"
	}
}
