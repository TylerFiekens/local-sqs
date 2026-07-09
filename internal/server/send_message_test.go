package server

import (
	"crypto/md5"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"testing"
)

func parseSendMessageResponse(t *testing.T, body string) sendMessageResponse {
	t.Helper()
	var resp sendMessageResponse
	if err := xml.Unmarshal([]byte(body), &resp); err != nil {
		t.Fatalf("failed to parse SendMessageResponse XML: %v\nbody: %s", err, body)
	}
	return resp
}

func TestSendMessage_success(t *testing.T) {
	s := New("http://localhost:9324")
	post(t, s, url.Values{"Action": {"CreateQueue"}, "QueueName": {"my-queue"}})

	rr := post(t, s, url.Values{
		"Action":      {"SendMessage"},
		"QueueUrl":    {"http://localhost:9324/queue/my-queue"},
		"MessageBody": {"hello world"},
	})

	if rr.Code != http.StatusOK {
		t.Fatalf("want 200, got %d\nbody: %s", rr.Code, rr.Body.String())
	}
	resp := parseSendMessageResponse(t, rr.Body.String())
	if resp.SendMessageResult.MessageID == "" {
		t.Fatal("MessageId must not be empty")
	}

	wantMD5 := fmt.Sprintf("%x", md5.Sum([]byte("hello world")))
	if resp.SendMessageResult.MD5OfMessageBody != wantMD5 {
		t.Fatalf("want MD5=%q, got %q", wantMD5, resp.SendMessageResult.MD5OfMessageBody)
	}
}

func TestSendMessage_enqueuesMessage(t *testing.T) {
	s := New("http://localhost:9324")
	post(t, s, url.Values{"Action": {"CreateQueue"}, "QueueName": {"q"}})

	post(t, s, url.Values{
		"Action":      {"SendMessage"},
		"QueueUrl":    {"http://localhost:9324/queue/q"},
		"MessageBody": {"msg1"},
	})
	post(t, s, url.Values{
		"Action":      {"SendMessage"},
		"QueueUrl":    {"http://localhost:9324/queue/q"},
		"MessageBody": {"msg2"},
	})

	if got := s.queues["q"].Len(); got != 2 {
		t.Fatalf("want 2 messages in queue, got %d", got)
	}
}

func TestSendMessage_uniqueMessageIDs(t *testing.T) {
	s := New("http://localhost:9324")
	post(t, s, url.Values{"Action": {"CreateQueue"}, "QueueName": {"q"}})

	vals := url.Values{
		"Action":      {"SendMessage"},
		"QueueUrl":    {"http://localhost:9324/queue/q"},
		"MessageBody": {"body"},
	}
	rr1 := post(t, s, vals)
	rr2 := post(t, s, vals)

	resp1 := parseSendMessageResponse(t, rr1.Body.String())
	resp2 := parseSendMessageResponse(t, rr2.Body.String())
	if resp1.SendMessageResult.MessageID == resp2.SendMessageResult.MessageID {
		t.Fatal("each SendMessage must produce a unique MessageId")
	}
}

func TestSendMessage_missingQueueUrl(t *testing.T) {
	s := New("http://localhost:9324")
	rr := post(t, s, url.Values{
		"Action":      {"SendMessage"},
		"MessageBody": {"hello"},
	})

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", rr.Code)
	}
	resp := parseError(t, rr.Body.String())
	if resp.Error.Code != "MissingParameter" {
		t.Fatalf("want MissingParameter, got %q", resp.Error.Code)
	}
}

func TestSendMessage_missingMessageBody(t *testing.T) {
	s := New("http://localhost:9324")
	rr := post(t, s, url.Values{
		"Action":   {"SendMessage"},
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

func TestSendMessage_nonExistentQueue(t *testing.T) {
	s := New("http://localhost:9324")
	rr := post(t, s, url.Values{
		"Action":      {"SendMessage"},
		"QueueUrl":    {"http://localhost:9324/queue/no-such-queue"},
		"MessageBody": {"hello"},
	})

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", rr.Code)
	}
	resp := parseError(t, rr.Body.String())
	if resp.Error.Code != "AWS.SimpleQueueService.NonExistentQueue" {
		t.Fatalf("want NonExistentQueue, got %q", resp.Error.Code)
	}
}

func TestSendMessage_invalidQueueUrl(t *testing.T) {
	s := New("http://localhost:9324")
	rr := post(t, s, url.Values{
		"Action":      {"SendMessage"},
		"QueueUrl":    {"://not-a-url"},
		"MessageBody": {"hello"},
	})

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", rr.Code)
	}
	resp := parseError(t, rr.Body.String())
	if resp.Error.Code != "InvalidParameterValue" {
		t.Fatalf("want InvalidParameterValue, got %q", resp.Error.Code)
	}
}

func TestSendMessage_contentType(t *testing.T) {
	s := New("http://localhost:9324")
	post(t, s, url.Values{"Action": {"CreateQueue"}, "QueueName": {"q"}})

	rr := post(t, s, url.Values{
		"Action":      {"SendMessage"},
		"QueueUrl":    {"http://localhost:9324/queue/q"},
		"MessageBody": {"hello"},
	})

	if ct := rr.Header().Get("Content-Type"); ct != "text/xml" {
		t.Fatalf("want Content-Type=text/xml, got %q", ct)
	}
}
