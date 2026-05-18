package server

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func newTestServer(t *testing.T) http.Handler {
	t.Helper()
	// ReadyDelay 0 => ready almost immediately for deterministic tests.
	s := New(Config{Version: "test", Commit: "abc123", ReadyDelay: time.Millisecond})
	time.Sleep(5 * time.Millisecond)
	return s.Handler()
}

func TestHealthzAlwaysOK(t *testing.T) {
	rr := httptest.NewRecorder()
	newTestServer(t).ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("healthz: got %d, want 200", rr.Code)
	}
}

func TestReadyzReadyAfterWarmup(t *testing.T) {
	rr := httptest.NewRecorder()
	newTestServer(t).ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/readyz", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("readyz: got %d, want 200", rr.Code)
	}
}

func TestBoomReturns500(t *testing.T) {
	rr := httptest.NewRecorder()
	newTestServer(t).ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/boom", nil))
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("boom: got %d, want 500", rr.Code)
	}
}

func TestMetricsExposesRequestCounter(t *testing.T) {
	h := newTestServer(t)

	// Generate one request so the counter has a sample.
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/healthz", nil))

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	body, _ := io.ReadAll(rr.Body)
	if !strings.Contains(string(body), "http_requests_total") {
		t.Fatalf("/metrics missing http_requests_total; body:\n%s", body)
	}
	if !strings.Contains(string(body), `app_build_info{commit="abc123",version="test"}`) {
		t.Fatalf("/metrics missing app_build_info")
	}
}
