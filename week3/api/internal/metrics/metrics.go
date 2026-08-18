package metrics

import (
	"net/http"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// create counter, format: GET, 200
var HTTPRequestsTotal = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "cloud3_http_requests_total",
		Help: "Total number of HTTP requests handled by the Cloud3 API.",
	},
	[]string{"method", "status"},
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
		// Prometheus scraping /metrics doesn't count as an API request.
		// if r.URL.Path == "/metrics" {
		// 	next.ServeHTTP(w, r)
		// 	return
		// }

		// wrapper for request method and the return code
		recorder := &statusRecorder{
			ResponseWriter: w,
			statusCode:     http.StatusOK, // by default
		}

		// let response go through recorder, so we can have status code
		next.ServeHTTP(recorder, r)

		// request total +1
		HTTPRequestsTotal.
			WithLabelValues(
				r.Method,
				strconv.Itoa(recorder.statusCode),
			).Inc()
	})
}

