package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

type mockDBChecker struct {
	err error
}

func (m *mockDBChecker) Ping(ctx context.Context) error {
	return m.err
}

func TestHealthCheck_Success(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := NewHandler(&mockDBChecker{err: nil}, "0.1.0")

	if err := handler.HealthCheck(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var resp HealthResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse json response: %v", err)
	}

	if resp.Status != "ok" {
		t.Errorf("expected status ok, got %s", resp.Status)
	}
	if resp.Database.Status != "connected" {
		t.Errorf("expected db status connected, got %s", resp.Database.Status)
	}
}

func TestHealthCheck_DBDisconnected(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := NewHandler(&mockDBChecker{err: errors.New("connection refused")}, "0.1.0")

	if err := handler.HealthCheck(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var resp HealthResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse json response: %v", err)
	}

	if resp.Status != "degraded" {
		t.Errorf("expected status degraded, got %s", resp.Status)
	}
	if resp.Database.Status != "disconnected" {
		t.Errorf("expected db status disconnected, got %s", resp.Database.Status)
	}
}
