package server

import (
	"encoding/xml"
	"net/http"
	"net/url"
	"testing"
)

func parseDeleteQueueResponse(t *testing.T, body string) deleteQueueResponse {
	t.Helper()
	var resp deleteQueueResponse
	if err := xml.Unmarshal([]byte(body), &resp); err != nil {
		t.Fatalf("failed to parse DeleteQueueResponse XML: %v\nbody: %s", err, body)
	}
	return resp
}

func TestDeleteQueue_success(t *testing.T) {
	s := New("http://localhost:9324")
	post(t, s, url.Values{"Action": {"CreateQueue"}, "QueueName": {"my-queue"}})

	rr := post(t, s, url.Values{
		"Action":   {"DeleteQueue"},
		"QueueUrl": {"http://localhost:9324/queue/my-queue"},
	})

	if rr.Code != http.StatusOK {
		t.Fatalf("want 200, got %d\nbody: %s", rr.Code, rr.Body.String())
	}
	resp := parseDeleteQueueResponse(t, rr.Body.String())
	if resp.ResponseMetadata.RequestID == "" {
		t.Fatal("ResponseMetadata.RequestId must not be empty")
	}
	if len(s.queues) != 0 {
		t.Fatalf("queue must be removed, got %d queues", len(s.queues))
	}
}

func TestDeleteQueue_idempotent(t *testing.T) {
	s := New("http://localhost:9324")

	vals := url.Values{
		"Action":   {"DeleteQueue"},
		"QueueUrl": {"http://localhost:9324/queue/non-existent"},
	}

	rr1 := post(t, s, vals)
	rr2 := post(t, s, vals)

	if rr1.Code != http.StatusOK || rr2.Code != http.StatusOK {
		t.Fatalf("both calls must return 200, got %d and %d", rr1.Code, rr2.Code)
	}
}

func TestDeleteQueue_missingQueueUrl(t *testing.T) {
	s := New("http://localhost:9324")
	rr := post(t, s, url.Values{"Action": {"DeleteQueue"}})

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", rr.Code)
	}
	resp := parseError(t, rr.Body.String())
	if resp.Error.Code != "MissingParameter" {
		t.Fatalf("want MissingParameter, got %q", resp.Error.Code)
	}
}

func TestDeleteQueue_invalidQueueUrl(t *testing.T) {
	s := New("http://localhost:9324")
	rr := post(t, s, url.Values{
		"Action":   {"DeleteQueue"},
		"QueueUrl": {"://not-a-url"},
	})

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", rr.Code)
	}
	resp := parseError(t, rr.Body.String())
	if resp.Error.Code != "InvalidParameterValue" {
		t.Fatalf("want InvalidParameterValue, got %q", resp.Error.Code)
	}
}

func TestDeleteQueue_contentType(t *testing.T) {
	s := New("http://localhost:9324")
	rr := post(t, s, url.Values{
		"Action":   {"DeleteQueue"},
		"QueueUrl": {"http://localhost:9324/queue/q"},
	})

	if ct := rr.Header().Get("Content-Type"); ct != "text/xml" {
		t.Fatalf("want Content-Type=text/xml, got %q", ct)
	}
}
