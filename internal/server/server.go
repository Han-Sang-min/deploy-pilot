// Package server contains the HTTP surface of the deploy-pilot sample service.
package server

import (
	"encoding/json"
	"log/slog"
	"math"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/Han-Sang-min/deploy-pilot/internal/metrics"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Config configures a Server. Logger is optional.
type Config struct {
	Version    string
	Commit     string
	Logger     *slog.Logger
	ReadyDelay time.Duration
}

// Server is the application. It is ready only after ReadyDelay has elapsed,
// which lets readiness-probe behaviour be demonstrated during rollouts.
type Server struct {
	cfg     Config
	metrics *metrics.Registry
	ready   atomic.Bool
}

// New constructs a Server and starts its warm-up timer.
func New(cfg Config) *Server {
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}
	s := &Server{
		cfg:     cfg,
		metrics: metrics.New(cfg.Version, cfg.Commit),
	}
	time.AfterFunc(cfg.ReadyDelay, func() {
		s.ready.Store(true)
		cfg.Logger.Info("service marked ready")
	})
	return s
}

// Handler returns the fully instrumented HTTP handler.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleRoot)
	mux.HandleFunc("/healthz", s.handleHealthz)
	mux.HandleFunc("/readyz", s.handleReadyz)
	mux.HandleFunc("/work", s.handleWork)
	mux.HandleFunc("/boom", s.handleBoom)
	mux.HandleFunc("/spin", s.handleSpin)
	mux.Handle("/metrics", promhttp.HandlerFor(s.metrics.Gatherer(), promhttp.HandlerOpts{}))
	return s.instrument(mux)
}

func (s *Server) handleRoot(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"service": "deploy-pilot",
		"version": s.cfg.Version,
		"commit":  s.cfg.Commit,
	})
}

// handleHealthz is the liveness probe: the process is up and serving.
func (s *Server) handleHealthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// handleReadyz is the readiness probe: ready only after warm-up.
func (s *Server) handleReadyz(w http.ResponseWriter, _ *http.Request) {
	if !s.ready.Load() {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "warming-up"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

// handleWork simulates business work with a configurable ?ms= latency.
func (s *Server) handleWork(w http.ResponseWriter, r *http.Request) {
	ms := queryInt(r, "ms", 50)
	time.Sleep(time.Duration(ms) * time.Millisecond)
	writeJSON(w, http.StatusOK, map[string]any{"slept_ms": ms})
}

// handleBoom always fails. Used to drive the high-error-rate runbook/alert.
func (s *Server) handleBoom(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "synthetic failure"})
}

// handleSpin burns CPU for ?ms= milliseconds. Drives the cpu-spike runbook.
func (s *Server) handleSpin(w http.ResponseWriter, r *http.Request) {
	ms := queryInt(r, "ms", 200)
	deadline := time.Now().Add(time.Duration(ms) * time.Millisecond)
	x := 0.0
	for time.Now().Before(deadline) {
		x += math.Sqrt(float64(time.Now().UnixNano() % 7919))
	}
	writeJSON(w, http.StatusOK, map[string]any{"spun_ms": ms, "sink": x > -1})
}

// instrument records request count and latency for every route.
func (s *Server) instrument(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		route := normalizeRoute(r.URL.Path)

		next.ServeHTTP(rec, r)

		elapsed := time.Since(start).Seconds()
		s.metrics.RequestsTotal.WithLabelValues(r.Method, route, strconv.Itoa(rec.status)).Inc()
		s.metrics.RequestDuration.WithLabelValues(r.Method, route).Observe(elapsed)
	})
}

// normalizeRoute keeps metric label cardinality bounded to known routes.
func normalizeRoute(path string) string {
	switch path {
	case "/", "/healthz", "/readyz", "/work", "/boom", "/spin", "/metrics":
		return path
	default:
		return "other"
	}
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func writeJSON(w http.ResponseWriter, code int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(body)
}

func queryInt(r *http.Request, key string, def int) int {
	if v := r.URL.Query().Get(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 && n <= 60000 {
			return n
		}
	}
	return def
}
