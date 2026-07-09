package server

import (
	"encoding/xml"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func parseCreateQueueResponse(t *testing.T, body string) createQueueResponse {
	t.Helper()
	var resp createQueueResponse
	if err := xml.Unmarshal([]byte(body), &resp); err != nil {
		t.Fatalf("failed to parse CreateQueueResponse XML: %v\nbody: %s", err, body)
	}
	return resp
}

func TestCreateQueue_success(t *testing.T) {
	s := New("http://localhost:9324")
	rr := post(t, s, url.Values{"Action": {"CreateQueue"}, "QueueName": {"my-queue"}})

	if rr.Code != http.StatusOK {
		t.Fatalf("want 200, got %d\nbody: %s", rr.Code, rr.Body.String())
	}
	resp := parseCreateQueueResponse(t, rr.Body.String())
	want := "http://localhost:9324/queue/my-queue"
	if resp.CreateQueueResult.QueueURL != want {
		t.Fatalf("want QueueUrl=%q, got %q", want, resp.CreateQueueResult.QueueURL)
	}
	if resp.ResponseMetadata.RequestID == "" {
		t.Fatal("ResponseMetadata.RequestId must not be empty")
	}
}

func TestCreateQueue_idempotent(t *testing.T) {
	s := New("http://localhost:9324")
	vals := url.Values{"Action": {"CreateQueue"}, "QueueName": {"my-queue"}}

	rr1 := post(t, s, vals)
	rr2 := post(t, s, vals)

	if rr1.Code != http.StatusOK || rr2.Code != http.StatusOK {
		t.Fatalf("both calls must return 200, got %d and %d", rr1.Code, rr2.Code)
	}

	resp1 := parseCreateQueueResponse(t, rr1.Body.String())
	resp2 := parseCreateQueueResponse(t, rr2.Body.String())
	if resp1.CreateQueueResult.QueueURL != resp2.CreateQueueResult.QueueURL {
		t.Fatalf("idempotent calls must return same URL: %q != %q",
			resp1.CreateQueueResult.QueueURL, resp2.CreateQueueResult.QueueURL)
	}

	if len(s.queues) != 1 {
		t.Fatalf("idempotent create must not duplicate queue, got %d queues", len(s.queues))
	}
}

func TestCreateQueue_missingName(t *testing.T) {
	s := New("http://localhost:9324")
	rr := post(t, s, url.Values{"Action": {"CreateQueue"}})

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", rr.Code)
	}
	resp := parseError(t, rr.Body.String())
	if resp.Error.Code != "MissingParameter" {
		t.Fatalf("want MissingParameter, got %q", resp.Error.Code)
	}
}

func TestCreateQueue_invalidName(t *testing.T) {
	s := New("http://localhost:9324")

	cases := []string{
		"",
		"has spaces",
		"has@symbol",
		strings.Repeat("a", 81),
	}

	for _, name := range cases {
		vals := url.Values{"Action": {"CreateQueue"}, "QueueName": {name}}
		rr := post(t, s, vals)
		if rr.Code != http.StatusBadRequest {
			t.Errorf("name %q: want 400, got %d", name, rr.Code)
		}
	}
}

func TestCreateQueue_validNameBoundaries(t *testing.T) {
	s := New("http://localhost:9324")

	cases := []string{
		"a",
		strings.Repeat("a", 80),
		"with-hyphens",
		"with_underscores",
		"MixedCase123",
	}

	for _, name := range cases {
		vals := url.Values{"Action": {"CreateQueue"}, "QueueName": {name}}
		rr := post(t, s, vals)
		if rr.Code != http.StatusOK {
			t.Errorf("name %q: want 200, got %d\nbody: %s", name, rr.Code, rr.Body.String())
		}
	}
}

func TestCreateQueue_contentType(t *testing.T) {
	s := New("http://localhost:9324")
	rr := post(t, s, url.Values{"Action": {"CreateQueue"}, "QueueName": {"q"}})

	if ct := rr.Header().Get("Content-Type"); ct != "text/xml" {
		t.Fatalf("want Content-Type=text/xml, got %q", ct)
	}
}
