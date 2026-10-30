package web_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"flowledger/web"
)

func TestWebHandler(t *testing.T) {
	handler := web.Handler()

	// 1. Root route serves index.html
	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for root, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `<div id="app"></div>`) {
		t.Errorf("expected HTML body to contain #app div")
	}

	// 2. SPA client-side route fallback to index.html
	req = httptest.NewRequest("GET", "/dashboard/transactions", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for SPA fallback route, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `<div id="app"></div>`) {
		t.Errorf("expected fallback route to return index.html")
	}
}

