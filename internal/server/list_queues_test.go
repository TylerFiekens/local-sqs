package server

import (
	"encoding/xml"
	"net/http"
	"net/url"
	"testing"
)

func parseListQueuesResponse(t *testing.T, body string) listQueuesResponse {
	t.Helper()
	var resp listQueuesResponse
	if err := xml.Unmarshal([]byte(body), &resp); err != nil {
		t.Fatalf("failed to parse ListQueuesResponse XML: %v\nbody: %s", err, body)
	}
	return resp
}

func TestListQueues_empty(t *testing.T) {
	s := New("http://localhost:9324")
	rr := post(t, s, url.Values{"Action": {"ListQueues"}})

	if rr.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rr.Code)
	}
	resp := parseListQueuesResponse(t, rr.Body.String())
	if len(resp.ListQueuesResult.QueueURLs) != 0 {
		t.Fatalf("want 0 queues, got %d", len(resp.ListQueuesResult.QueueURLs))
	}
}

func TestListQueues_all(t *testing.T) {
	s := New("http://localhost:9324")
	post(t, s, url.Values{"Action": {"CreateQueue"}, "QueueName": {"alpha"}})
	post(t, s, url.Values{"Action": {"CreateQueue"}, "QueueName": {"beta"}})
	post(t, s, url.Values{"Action": {"CreateQueue"}, "QueueName": {"gamma"}})

	rr := post(t, s, url.Values{"Action": {"ListQueues"}})

	if rr.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rr.Code)
	}
	resp := parseListQueuesResponse(t, rr.Body.String())
	if len(resp.ListQueuesResult.QueueURLs) != 3 {
		t.Fatalf("want 3 queues, got %d", len(resp.ListQueuesResult.QueueURLs))
	}
}

func TestListQueues_sorted(t *testing.T) {
	s := New("http://localhost:9324")
	post(t, s, url.Values{"Action": {"CreateQueue"}, "QueueName": {"gamma"}})
	post(t, s, url.Values{"Action": {"CreateQueue"}, "QueueName": {"alpha"}})
	post(t, s, url.Values{"Action": {"CreateQueue"}, "QueueName": {"beta"}})

	rr := post(t, s, url.Values{"Action": {"ListQueues"}})
	resp := parseListQueuesResponse(t, rr.Body.String())

	urls := resp.ListQueuesResult.QueueURLs
	for i := 1; i < len(urls); i++ {
		if urls[i] < urls[i-1] {
			t.Fatalf("results not sorted: %q before %q", urls[i-1], urls[i])
		}
	}
}

func TestListQueues_withPrefix(t *testing.T) {
	s := New("http://localhost:9324")
	post(t, s, url.Values{"Action": {"CreateQueue"}, "QueueName": {"order-placed"}})
	post(t, s, url.Values{"Action": {"CreateQueue"}, "QueueName": {"order-shipped"}})
	post(t, s, url.Values{"Action": {"CreateQueue"}, "QueueName": {"payment-received"}})

	rr := post(t, s, url.Values{"Action": {"ListQueues"}, "QueueNamePrefix": {"order-"}})

	if rr.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rr.Code)
	}
	resp := parseListQueuesResponse(t, rr.Body.String())
	if len(resp.ListQueuesResult.QueueURLs) != 2 {
		t.Fatalf("want 2 queues matching prefix, got %d", len(resp.ListQueuesResult.QueueURLs))
	}
	for _, u := range resp.ListQueuesResult.QueueURLs {
		if !containsStr(u, "order-") {
			t.Errorf("unexpected queue URL in filtered results: %q", u)
		}
	}
}

func TestListQueues_prefixNoMatch(t *testing.T) {
	s := New("http://localhost:9324")
	post(t, s, url.Values{"Action": {"CreateQueue"}, "QueueName": {"my-queue"}})

	rr := post(t, s, url.Values{"Action": {"ListQueues"}, "QueueNamePrefix": {"no-match"}})

	resp := parseListQueuesResponse(t, rr.Body.String())
	if len(resp.ListQueuesResult.QueueURLs) != 0 {
		t.Fatalf("want 0 queues, got %d", len(resp.ListQueuesResult.QueueURLs))
	}
}

func TestListQueues_contentType(t *testing.T) {
	s := New("http://localhost:9324")
	rr := post(t, s, url.Values{"Action": {"ListQueues"}})

	if ct := rr.Header().Get("Content-Type"); ct != "text/xml" {
		t.Fatalf("want Content-Type=text/xml, got %q", ct)
	}
}

func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsRune(s, substr))
}

func containsRune(s, substr string) bool {
	for i := range s {
		if i+len(substr) <= len(s) && s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
