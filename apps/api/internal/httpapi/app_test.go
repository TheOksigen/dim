package httpapi

import (
	"bytes"
	"log/slog"
	"net/http/httptest"
	"testing"
	"time"

	"exam-results-platform/api/internal/config"
)

func TestLookupRejectsInvalidFIN(t *testing.T) {
	app := New(config.Config{
		DBTimeout:        100 * time.Millisecond,
		RequestTimeout:   time.Second,
		MaxBodyBytes:     1024,
		RateLimitMax:     10,
		RateLimitWindow:  time.Minute,
		CacheTTL:         time.Minute,
		NegativeCacheTTL: time.Second,
		CORSOrigins:      "http://localhost:3000",
	}, nil, nil, slog.New(slog.NewTextHandler(bytes.NewBuffer(nil), nil)))

	request := httptest.NewRequest("POST", "/api/v1/results/lookup", bytes.NewBufferString("{\"finCode\":\"BAD\"}"))
	request.Header.Set("Content-Type", "application/json")
	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != 400 {
		t.Fatalf("status = %d, want 400", response.StatusCode)
	}
	if response.Header.Get("Cache-Control") != "no-store, private" {
		t.Fatalf("unexpected cache header %q", response.Header.Get("Cache-Control"))
	}
}
