package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/D3rille/kk-fast-food-system/internal/handlers"
)

func TestHealthz(t *testing.T) {
	handler := handlers.NewHealthHandler(nil)

	req, err := http.NewRequest("GET", "/healthz", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	http.HandlerFunc(handler.Healthz).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status OK, got %v", rr.Code)
	}

	expected := `{"status":"OK"}`
	if rr.Body.String() != expected {
		t.Errorf("expected body %q, got %q", expected, rr.Body.String())
	}
}

func TestReadyz_Disconnected(t *testing.T) {
	handler := handlers.NewHealthHandler(nil)

	req, err := http.NewRequest("GET", "/readyz", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	http.HandlerFunc(handler.Readyz).ServeHTTP(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Errorf("expected status ServiceUnavailable, got %v", rr.Code)
	}

	var res map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &res); err != nil {
		t.Fatal(err)
	}

	if res["status"] != "Unavailable" {
		t.Errorf("expected status 'Unavailable', got %q", res["status"])
	}
}
