package server

import (
	"encoding/xml"
	"net/http"
	"net/url"
	"testing"
)

func parseDeleteMessageResponse(t *testing.T, body string) deleteMessageResponse {
	t.Helper()
	var resp deleteMessageResponse
	if err := xml.Unmarshal([]byte(body), &resp); err != nil {
		t.Fatalf("failed to parse DeleteMessageResponse XML: %v\nbody: %s", err, body)
	}
	return resp
}

func TestDeleteMessage_success(t *testing.T) {
	s := New("http://localhost:9324")
	post(t, s, url.Values{"Action": {"CreateQueue"}, "QueueName": {"q"}})
	sendMsg(t, s, "http://localhost:9324/queue/q", "hello")

	rr := post(t, s, url.Values{
		"Action":   {"ReceiveMessage"},
		"QueueUrl": {"http://localhost:9324/queue/q"},
	})
	recv := parseReceiveMessageResponse(t, rr.Body.String())
	receiptHandle := recv.ReceiveMessageResult.Messages[0].ReceiptHandle

	// message was already dequeued by ReceiveMessage; enqueue another to verify delete works on in-flight messages
	sendMsg(t, s, "http://localhost:9324/queue/q", "second")
	rr2 := post(t, s, url.Values{
		"Action":        {"DeleteMessage"},
		"QueueUrl":      {"http://localhost:9324/queue/q"},
		"ReceiptHandle": {receiptHandle},
	})

	if rr2.Code != http.StatusOK {
		t.Fatalf("want 200, got %d\nbody: %s", rr2.Code, rr2.Body.String())
	}
	resp := parseDeleteMessageResponse(t, rr2.Body.String())
	if resp.ResponseMetadata.RequestID == "" {
		t.Fatal("ResponseMetadata.RequestId must not be empty")
	}
}

func TestDeleteMessage_removesFromQueue(t *testing.T) {
	s := New("http://localhost:9324")
	post(t, s, url.Values{"Action": {"CreateQueue"}, "QueueName": {"q"}})
	sendMsg(t, s, "http://localhost:9324/queue/q", "msg1")
	sendMsg(t, s, "http://localhost:9324/queue/q", "msg2")

	// peek at receipt handle without dequeuing by using receive then re-enqueue isn't possible;
	// directly access queue to get receipt handle
	msgs := s.queues["q"].Dequeue(1)
	receiptHandle := msgs[0].ReceiptHandle

	// re-enqueue so Delete has something to find
	s.queues["q"].Enqueue(msgs[0])

	post(t, s, url.Values{
		"Action":        {"DeleteMessage"},
		"QueueUrl":      {"http://localhost:9324/queue/q"},
		"ReceiptHandle": {receiptHandle},
	})

	if s.queues["q"].Len() != 1 {
		t.Fatalf("want 1 message remaining after delete, got %d", s.queues["q"].Len())
	}
}

func TestDeleteMessage_idempotent(t *testing.T) {
	s := New("http://localhost:9324")
	post(t, s, url.Values{"Action": {"CreateQueue"}, "QueueName": {"q"}})

	vals := url.Values{
		"Action":        {"DeleteMessage"},
		"QueueUrl":      {"http://localhost:9324/queue/q"},
		"ReceiptHandle": {"non-existent-handle"},
	}

	rr1 := post(t, s, vals)
	rr2 := post(t, s, vals)

	if rr1.Code != http.StatusOK || rr2.Code != http.StatusOK {
		t.Fatalf("both calls must return 200, got %d and %d", rr1.Code, rr2.Code)
	}
}

func TestDeleteMessage_missingQueueUrl(t *testing.T) {
	s := New("http://localhost:9324")
	rr := post(t, s, url.Values{
		"Action":        {"DeleteMessage"},
		"ReceiptHandle": {"handle"},
	})

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", rr.Code)
	}
	resp := parseError(t, rr.Body.String())
	if resp.Error.Code != "MissingParameter" {
		t.Fatalf("want MissingParameter, got %q", resp.Error.Code)
	}
}

func TestDeleteMessage_missingReceiptHandle(t *testing.T) {
	s := New("http://localhost:9324")
	rr := post(t, s, url.Values{
		"Action":   {"DeleteMessage"},
		"QueueUrl": {"http://localhost:9324/queue/q"},
	})

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", rr.Code)
	}
	resp := parseError(t, rr.Body.String())
	if resp.Error.Code != "MissingParameter" {
		t.Fatalf("want MissingParameter, got %q", resp.Error.Code)
	}
}

func TestDeleteMessage_invalidQueueUrl(t *testing.T) {
	s := New("http://localhost:9324")
	rr := post(t, s, url.Values{
		"Action":        {"DeleteMessage"},
		"QueueUrl":      {"://bad"},
		"ReceiptHandle": {"handle"},
	})

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", rr.Code)
	}
	resp := parseError(t, rr.Body.String())
	if resp.Error.Code != "InvalidParameterValue" {
		t.Fatalf("want InvalidParameterValue, got %q", resp.Error.Code)
	}
}

func TestDeleteMessage_nonExistentQueue(t *testing.T) {
	s := New("http://localhost:9324")
	rr := post(t, s, url.Values{
		"Action":        {"DeleteMessage"},
		"QueueUrl":      {"http://localhost:9324/queue/no-such-queue"},
		"ReceiptHandle": {"handle"},
	})

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", rr.Code)
	}
	resp := parseError(t, rr.Body.String())
	if resp.Error.Code != "AWS.SimpleQueueService.NonExistentQueue" {
		t.Fatalf("want NonExistentQueue, got %q", resp.Error.Code)
	}
}

func TestDeleteMessage_contentType(t *testing.T) {
	s := New("http://localhost:9324")
	post(t, s, url.Values{"Action": {"CreateQueue"}, "QueueName": {"q"}})

	rr := post(t, s, url.Values{
		"Action":        {"DeleteMessage"},
		"QueueUrl":      {"http://localhost:9324/queue/q"},
		"ReceiptHandle": {"any"},
	})

	if ct := rr.Header().Get("Content-Type"); ct != "text/xml" {
		t.Fatalf("want Content-Type=text/xml, got %q", ct)
	}
}
