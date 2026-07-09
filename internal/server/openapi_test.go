package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandleOpenAPI_success(t *testing.T) {
	s := New("http://localhost:9324")
	req := httptest.NewRequest(http.MethodGet, "/openapi.yaml", nil)
	rr := httptest.NewRecorder()
	s.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); ct != "application/yaml" {
		t.Fatalf("want Content-Type=application/yaml, got %q", ct)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "openapi:") {
		t.Fatal("response must contain OpenAPI spec")
	}
	if !strings.Contains(body, "CreateQueue") {
		t.Fatal("spec must document CreateQueue operation")
	}
}
