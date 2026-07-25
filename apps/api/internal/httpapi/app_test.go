package httpapi

import (
	"bytes"
	"io"
	"log/slog"
	"net/http/httptest"
	"strings"
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

func TestOpenAPIAndSwaggerDocsAreAvailable(t *testing.T) {
	app := New(config.Config{
		DBTimeout:        100 * time.Millisecond,
		RequestTimeout:   time.Second,
		MaxBodyBytes:     1024,
		RateLimitMax:     10,
		RateLimitWindow:  time.Minute,
		CacheTTL:         time.Minute,
		NegativeCacheTTL: time.Second,
	}, nil, nil, slog.New(slog.NewTextHandler(bytes.NewBuffer(nil), nil)))

	openAPIResponse, err := app.Test(httptest.NewRequest("GET", "/openapi.json", nil))
	if err != nil {
		t.Fatalf("openapi app.Test() error = %v", err)
	}
	defer openAPIResponse.Body.Close()
	openAPIBody, err := io.ReadAll(openAPIResponse.Body)
	if err != nil {
		t.Fatalf("read OpenAPI response: %v", err)
	}
	if openAPIResponse.StatusCode != 200 || !strings.Contains(string(openAPIBody), "\"openapi\": \"3.0.3\"") {
		t.Fatalf("unexpected OpenAPI response: status=%d body=%q", openAPIResponse.StatusCode, openAPIBody)
	}

	docsResponse, err := app.Test(httptest.NewRequest("GET", "/docs/index.html", nil))
	if err != nil {
		t.Fatalf("docs app.Test() error = %v", err)
	}
	defer docsResponse.Body.Close()
	docsBody, err := io.ReadAll(docsResponse.Body)
	if err != nil {
		t.Fatalf("read docs response: %v", err)
	}
	if docsResponse.StatusCode != 200 || !strings.Contains(string(docsBody), "swagger-ui") || !strings.Contains(string(docsBody), "/openapi.json") {
		t.Fatalf("unexpected docs response: status=%d", docsResponse.StatusCode)
	}
}
