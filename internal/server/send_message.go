package server

import (
	"encoding/xml"
	"net/http"
	"time"

	"github.com/tylerfiekens/local-sqs/internal/queue"
)

// sendMessageResponse is the XML response returned on a successful SendMessage call.
type sendMessageResponse struct {
	XMLName           xml.Name          `xml:"SendMessageResponse"`
	Xmlns             string            `xml:"xmlns,attr"`
	SendMessageResult sendMessageResult `xml:"SendMessageResult"`
	ResponseMetadata  responseMetadata  `xml:"ResponseMetadata"`
}

// sendMessageResult holds the identifiers returned by SendMessage.
type sendMessageResult struct {
	MD5OfMessageBody string `xml:"MD5OfMessageBody"`
	MessageID        string `xml:"MessageId"`
}

// handleSendMessage processes a SendMessage request. It enqueues the message
// on the target queue and returns the message ID and body MD5.
func (s *Server) handleSendMessage(w http.ResponseWriter, r *http.Request) {
	queueURL := r.FormValue("QueueUrl")
	if queueURL == "" {
		writeError(w, http.StatusBadRequest, "MissingParameter", "QueueUrl is required")
		return
	}

	body := r.FormValue("MessageBody")
	if body == "" {
		writeError(w, http.StatusBadRequest, "MissingParameter", "MessageBody is required")
		return
	}

	q := s.queueByURL(w, queueURL)
	if q == nil {
		return
	}

	id := newID()
	msg := &queue.Message{
		ID:            id,
		Body:          body,
		ReceiptHandle: newID(),
		SentAt:        time.Now(),
	}
	q.Enqueue(msg)

	resp := sendMessageResponse{
		Xmlns: "http://queue.amazonaws.com/doc/2012-11-05/",
		SendMessageResult: sendMessageResult{
			MD5OfMessageBody: md5Hex(body),
			MessageID:        id,
		},
		ResponseMetadata: responseMetadata{RequestID: "local"},
	}
	w.Header().Set("Content-Type", "text/xml")
	xml.NewEncoder(w).Encode(resp)
}
