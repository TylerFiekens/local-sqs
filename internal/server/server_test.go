package server

import (
	"encoding/xml"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

type errReader struct{}

func (errReader) Read([]byte) (int, error) { return 0, errors.New("read error") }

func post(t *testing.T, srv *Server, body url.Values) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, req)
	return rr
}

func parseError(t *testing.T, body string) errorResponse {
	t.Helper()
	var resp errorResponse
	if err := xml.Unmarshal([]byte(body), &resp); err != nil {
		t.Fatalf("failed to parse error XML: %v\nbody: %s", err, body)
	}
	return resp
}

func TestNew_emptyQueues(t *testing.T) {
	s := New("http://localhost:9324")
	if s.queues == nil {
		t.Fatal("queues map must be initialized")
	}
	if len(s.queues) != 0 {
		t.Fatal("new server must have no queues")
	}
}

func TestHandleSQS_methodNotAllowed(t *testing.T) {
	s := New("http://localhost:9324")
	for _, method := range []string{http.MethodGet, http.MethodPut, http.MethodDelete, http.MethodPatch} {
		req := httptest.NewRequest(method, "/", nil)
		rr := httptest.NewRecorder()
		s.ServeHTTP(rr, req)
		if rr.Code != http.StatusMethodNotAllowed {
			t.Errorf("method %s: want 405, got %d", method, rr.Code)
		}
	}
}

func TestHandleSQS_missingAction(t *testing.T) {
	s := New("http://localhost:9324")
	rr := post(t, s, url.Values{})

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", rr.Code)
	}
	resp := parseError(t, rr.Body.String())
	if resp.Error.Code != "MissingAction" {
		t.Fatalf("want MissingAction, got %q", resp.Error.Code)
	}
}

func TestHandleSQS_unknownAction(t *testing.T) {
	s := New("http://localhost:9324")
	rr := post(t, s, url.Values{"Action": {"DoSomethingFake"}})

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", rr.Code)
	}
	resp := parseError(t, rr.Body.String())
	if resp.Error.Code != "InvalidAction" {
		t.Fatalf("want InvalidAction, got %q", resp.Error.Code)
	}
	if !strings.Contains(resp.Error.Message, "DoSomethingFake") {
		t.Fatalf("error message should contain action name, got %q", resp.Error.Message)
	}
}

func TestHandleSQS_parseFormError(t *testing.T) {
	s := New("http://localhost:9324")
	req := httptest.NewRequest(http.MethodPost, "/", errReader{})
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	s.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", rr.Code)
	}
	resp := parseError(t, rr.Body.String())
	if resp.Error.Code != "InvalidParameterValue" {
		t.Fatalf("want InvalidParameterValue, got %q", resp.Error.Code)
	}
}

func TestHandleSQS_contentType(t *testing.T) {
	s := New("http://localhost:9324")
	rr := post(t, s, url.Values{})

	if ct := rr.Header().Get("Content-Type"); ct != "text/xml" {
		t.Fatalf("want Content-Type=text/xml, got %q", ct)
	}
}

func TestWriteError_xmlStructure(t *testing.T) {
	rr := httptest.NewRecorder()
	writeError(rr, http.StatusBadRequest, "SomeCode", "some message")

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", rr.Code)
	}

	resp := parseError(t, rr.Body.String())
	if resp.Error.Code != "SomeCode" {
		t.Fatalf("want Code=SomeCode, got %q", resp.Error.Code)
	}
	if resp.Error.Message != "some message" {
		t.Fatalf("want Message=%q, got %q", "some message", resp.Error.Message)
	}
	if resp.Error.Type != "Sender" {
		t.Fatalf("want Type=Sender, got %q", resp.Error.Type)
	}
	if resp.RequestID == "" {
		t.Fatal("RequestId must not be empty")
	}
}
