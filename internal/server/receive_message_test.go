package server

import (
	"crypto/md5"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"testing"
)

func parseReceiveMessageResponse(t *testing.T, body string) receiveMessageResponse {
	t.Helper()
	var resp receiveMessageResponse
	if err := xml.Unmarshal([]byte(body), &resp); err != nil {
		t.Fatalf("failed to parse ReceiveMessageResponse XML: %v\nbody: %s", err, body)
	}
	return resp
}

func sendMsg(t *testing.T, s *Server, queueURL, body string) {
	t.Helper()
	post(t, s, url.Values{
		"Action":      {"SendMessage"},
		"QueueUrl":    {queueURL},
		"MessageBody": {body},
	})
}

func TestReceiveMessage_success(t *testing.T) {
	s := New("http://localhost:9324")
	post(t, s, url.Values{"Action": {"CreateQueue"}, "QueueName": {"q"}})
	sendMsg(t, s, "http://localhost:9324/queue/q", "hello")

	rr := post(t, s, url.Values{
		"Action":   {"ReceiveMessage"},
		"QueueUrl": {"http://localhost:9324/queue/q"},
	})

	if rr.Code != http.StatusOK {
		t.Fatalf("want 200, got %d\nbody: %s", rr.Code, rr.Body.String())
	}
	resp := parseReceiveMessageResponse(t, rr.Body.String())
	if len(resp.ReceiveMessageResult.Messages) != 1 {
		t.Fatalf("want 1 message, got %d", len(resp.ReceiveMessageResult.Messages))
	}
	msg := resp.ReceiveMessageResult.Messages[0]
	if msg.Body != "hello" {
		t.Fatalf("want Body=hello, got %q", msg.Body)
	}
	if msg.MessageID == "" {
		t.Fatal("MessageId must not be empty")
	}
	if msg.ReceiptHandle == "" {
		t.Fatal("ReceiptHandle must not be empty")
	}
	wantMD5 := fmt.Sprintf("%x", md5.Sum([]byte("hello")))
	if msg.MD5OfBody != wantMD5 {
		t.Fatalf("want MD5=%q, got %q", wantMD5, msg.MD5OfBody)
	}
}

func TestReceiveMessage_empty(t *testing.T) {
	s := New("http://localhost:9324")
	post(t, s, url.Values{"Action": {"CreateQueue"}, "QueueName": {"q"}})

	rr := post(t, s, url.Values{
		"Action":   {"ReceiveMessage"},
		"QueueUrl": {"http://localhost:9324/queue/q"},
	})

	if rr.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rr.Code)
	}
	resp := parseReceiveMessageResponse(t, rr.Body.String())
	if len(resp.ReceiveMessageResult.Messages) != 0 {
		t.Fatalf("want 0 messages, got %d", len(resp.ReceiveMessageResult.Messages))
	}
}

func TestReceiveMessage_defaultMaxOne(t *testing.T) {
	s := New("http://localhost:9324")
	post(t, s, url.Values{"Action": {"CreateQueue"}, "QueueName": {"q"}})
	sendMsg(t, s, "http://localhost:9324/queue/q", "a")
	sendMsg(t, s, "http://localhost:9324/queue/q", "b")

	rr := post(t, s, url.Values{
		"Action":   {"ReceiveMessage"},
		"QueueUrl": {"http://localhost:9324/queue/q"},
	})

	resp := parseReceiveMessageResponse(t, rr.Body.String())
	if len(resp.ReceiveMessageResult.Messages) != 1 {
		t.Fatalf("default MaxNumberOfMessages must be 1, got %d", len(resp.ReceiveMessageResult.Messages))
	}
}

func TestReceiveMessage_maxMessages(t *testing.T) {
	s := New("http://localhost:9324")
	post(t, s, url.Values{"Action": {"CreateQueue"}, "QueueName": {"q"}})
	for i := range 5 {
		sendMsg(t, s, "http://localhost:9324/queue/q", fmt.Sprintf("msg%d", i))
	}

	rr := post(t, s, url.Values{
		"Action":              {"ReceiveMessage"},
		"QueueUrl":            {"http://localhost:9324/queue/q"},
		"MaxNumberOfMessages": {"3"},
	})

	resp := parseReceiveMessageResponse(t, rr.Body.String())
	if len(resp.ReceiveMessageResult.Messages) != 3 {
		t.Fatalf("want 3 messages, got %d", len(resp.ReceiveMessageResult.Messages))
	}
	if s.queues["q"].Len() != 2 {
		t.Fatalf("want 2 remaining in queue, got %d", s.queues["q"].Len())
	}
}

func TestReceiveMessage_maxMessagesInvalidValues(t *testing.T) {
	s := New("http://localhost:9324")
	post(t, s, url.Values{"Action": {"CreateQueue"}, "QueueName": {"q"}})

	cases := []string{"0", "11", "-1", "abc"}
	for _, v := range cases {
		rr := post(t, s, url.Values{
			"Action":              {"ReceiveMessage"},
			"QueueUrl":            {"http://localhost:9324/queue/q"},
			"MaxNumberOfMessages": {v},
		})
		if rr.Code != http.StatusBadRequest {
			t.Errorf("MaxNumberOfMessages=%q: want 400, got %d", v, rr.Code)
		}
		resp := parseError(t, rr.Body.String())
		if resp.Error.Code != "InvalidParameterValue" {
			t.Errorf("MaxNumberOfMessages=%q: want InvalidParameterValue, got %q", v, resp.Error.Code)
		}
	}
}

func TestReceiveMessage_fifoOrder(t *testing.T) {
	s := New("http://localhost:9324")
	post(t, s, url.Values{"Action": {"CreateQueue"}, "QueueName": {"q"}})
	sendMsg(t, s, "http://localhost:9324/queue/q", "first")
	sendMsg(t, s, "http://localhost:9324/queue/q", "second")

	rr := post(t, s, url.Values{
		"Action":              {"ReceiveMessage"},
		"QueueUrl":            {"http://localhost:9324/queue/q"},
		"MaxNumberOfMessages": {"2"},
	})

	resp := parseReceiveMessageResponse(t, rr.Body.String())
	if len(resp.ReceiveMessageResult.Messages) != 2 {
		t.Fatalf("want 2 messages, got %d", len(resp.ReceiveMessageResult.Messages))
	}
	if resp.ReceiveMessageResult.Messages[0].Body != "first" {
		t.Fatalf("want FIFO order, first message body=first, got %q", resp.ReceiveMessageResult.Messages[0].Body)
	}
}

func TestReceiveMessage_missingQueueUrl(t *testing.T) {
	s := New("http://localhost:9324")
	rr := post(t, s, url.Values{"Action": {"ReceiveMessage"}})

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", rr.Code)
	}
	resp := parseError(t, rr.Body.String())
	if resp.Error.Code != "MissingParameter" {
		t.Fatalf("want MissingParameter, got %q", resp.Error.Code)
	}
}

func TestReceiveMessage_invalidQueueUrl(t *testing.T) {
	s := New("http://localhost:9324")
	rr := post(t, s, url.Values{
		"Action":   {"ReceiveMessage"},
		"QueueUrl": {"://bad"},
	})

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", rr.Code)
	}
	resp := parseError(t, rr.Body.String())
	if resp.Error.Code != "InvalidParameterValue" {
		t.Fatalf("want InvalidParameterValue, got %q", resp.Error.Code)
	}
}

func TestReceiveMessage_nonExistentQueue(t *testing.T) {
	s := New("http://localhost:9324")
	rr := post(t, s, url.Values{
		"Action":   {"ReceiveMessage"},
		"QueueUrl": {"http://localhost:9324/queue/no-such-queue"},
	})

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", rr.Code)
	}
	resp := parseError(t, rr.Body.String())
	if resp.Error.Code != "AWS.SimpleQueueService.NonExistentQueue" {
		t.Fatalf("want NonExistentQueue, got %q", resp.Error.Code)
	}
}

func TestReceiveMessage_contentType(t *testing.T) {
	s := New("http://localhost:9324")
	post(t, s, url.Values{"Action": {"CreateQueue"}, "QueueName": {"q"}})

	rr := post(t, s, url.Values{
		"Action":   {"ReceiveMessage"},
		"QueueUrl": {"http://localhost:9324/queue/q"},
	})

	if ct := rr.Header().Get("Content-Type"); ct != "text/xml" {
		t.Fatalf("want Content-Type=text/xml, got %q", ct)
	}
}
