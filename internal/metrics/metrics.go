// Package metrics owns a private Prometheus registry so that the exposition
// format (the actual scrape contract Prometheus/Grafana depend on) is explicit
// and testable, rather than relying on the global default registry.
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

// Registry bundles the application metrics and the underlying Prometheus
// registry used to serve /metrics.
type Registry struct {
	reg *prometheus.Registry

	RequestsTotal   *prometheus.CounterVec
	RequestDuration *prometheus.HistogramVec
}

// New builds the registry, wires the standard Go/process collectors, and
// records build metadata as app_build_info{version,commit} = 1.
func New(version, commit string) *Registry {
	reg := prometheus.NewRegistry()
	reg.MustRegister(collectors.NewGoCollector())
	reg.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))

	buildInfo := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "app_build_info",
		Help: "Build metadata exported as a constant gauge with value 1.",
	}, []string{"version", "commit"})
	buildInfo.WithLabelValues(version, commit).Set(1)
	reg.MustRegister(buildInfo)

	requestsTotal := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total number of HTTP requests by method, route and status code.",
	}, []string{"method", "route", "status"})
	reg.MustRegister(requestsTotal)

	requestDuration := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "HTTP request latency in seconds by method and route.",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "route"})
	reg.MustRegister(requestDuration)

	return &Registry{
		reg:             reg,
		RequestsTotal:   requestsTotal,
		RequestDuration: requestDuration,
	}
}

// Gatherer exposes the registry for promhttp without leaking write access.
func (r *Registry) Gatherer() prometheus.Gatherer { return r.reg }
