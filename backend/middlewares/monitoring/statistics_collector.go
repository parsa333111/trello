package monitoring

import (
	"log"

	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	Requests *prometheus.CounterVec
	Queries  *prometheus.CounterVec
	Delay    prometheus.Gauge
}

const (
	Successful   = "successful"
	Unsuccessful = "unsuccessful"
)

var Statistics *Metrics
var Registry *prometheus.Registry

func newMetrics() *Metrics {
	metrics := &Metrics{
		Requests: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "total_number_of_requests",
				Help: "How many HTTP requests processed, partitioned by successful and unsuccessful requests.",
			},
			[]string{"status"},
		),
		Queries: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "total_number_of_queries",
				Help: "How many database queries processed, partitioned by successful and unsuccessful queris.",
			},
			[]string{"status"},
		),
		Delay: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "total_delay",
				Help: "How long it has taken to respond to the requests in total.",
			},
		),
	}

	metrics.Delay.Set(0)

	return metrics
}

func InitalizeStatistics() {
	Registry = prometheus.NewRegistry()
	Statistics = newMetrics()

	if err := Registry.Register(Statistics.Requests); err != nil {
		log.Fatal(err)
	}
	if err := Registry.Register(Statistics.Queries); err != nil {
		log.Fatal(err)
	}
	if err := Registry.Register(Statistics.Delay); err != nil {
		log.Fatal(err)
	}
}
