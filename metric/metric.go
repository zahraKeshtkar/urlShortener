package metric

import (
	"log"
	"net/http"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"url-shortner/config"
)

type metric struct {
	RequestCounter   *prometheus.CounterVec
	RequestDurations prometheus.Histogram
}

var MuxMetric metric

func Monitor() {
	p := &http.ServeMux{}
	p.Handle("/metrics", promhttp.Handler())

	log.Fatal(http.ListenAndServe(strconv.Itoa(config.DefaultConfig.Metric.Port), p))
}

func NewMuxMetric() {
	MuxMetric = metric{
		RequestCounter: prometheus.NewCounterVec(prometheus.CounterOpts{Name: "number_of_requests"}, []string{"url"}),
		RequestDurations: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "A histogram of the HTTP request durations in seconds.",
			Buckets: prometheus.ExponentialBuckets(0.1, 1.5, 5),
		})}
}
