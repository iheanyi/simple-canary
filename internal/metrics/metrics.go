package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Node is a wrapper around Prometheus's registerer interface
type Node struct {
	registry prometheus.Registerer
}

// Prometheus returns an instance of the metrics node and the prometheus
// handler.
func Prometheus() (*Node, http.Handler) {
	registry := prometheus.NewRegistry()
	handler := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	registry.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
	registry.MustRegister(prometheus.NewGoCollector())

	return &Node{
		registry: registry,
	}, handler
}

// Labels returns a new instance of the metrics node with the added labels to
// the registry.
func (n *Node) Labels(labels map[string]string) *Node {
	promLabels := prometheus.Labels(labels)
	newNode := prometheus.WrapRegistererWith(promLabels, n.registry)

	return &Node{
		registry: newNode,
	}
}

// Counter returns a new CounterVec on the metrics node.
func (n *Node) Counter(name, description string, labels ...string) *prometheus.CounterVec {
	counter := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: name,
		Help: description,
	}, labels)

	// Register the counter for usage.
	n.registry.MustRegister(counter)

	return counter
}

// Gauge returns a new Gauge on the metrics node.
func (n *Node) Gauge(name, description string) prometheus.Gauge {
	gauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: name,
		Help: description,
	})

	n.registry.MustRegister(gauge)
	return gauge
}

// Summary returns a new SummaryVector on the metrics node.
func (n *Node) Summary(name, description string, buckets []float64, labels ...string) *prometheus.SummaryVec {
	calculatedBuckets := make(map[float64]float64, len(buckets))

	for _, value := range buckets {
		calculatedBuckets[value] = (1.0 - value) / 10
	}
	summary := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name:       name,
		Help:       description,
		Objectives: calculatedBuckets,
	}, labels)

	n.registry.MustRegister(summary)

	return summary
}
