package server

import (
	"encoding/xml"
	"net/http"
	"strconv"
)

// receiveMessageResponse is the XML response returned on a successful ReceiveMessage call.
type receiveMessageResponse struct {
	XMLName              xml.Name             `xml:"ReceiveMessageResponse"`
	Xmlns                string               `xml:"xmlns,attr"`
	ReceiveMessageResult receiveMessageResult `xml:"ReceiveMessageResult"`
	ResponseMetadata     responseMetadata     `xml:"ResponseMetadata"`
}

// receiveMessageResult holds the messages returned by ReceiveMessage.
type receiveMessageResult struct {
	Messages []receiveMessageItem `xml:"Message"`
}

// receiveMessageItem is a single message in a ReceiveMessage response.
type receiveMessageItem struct {
	MessageID     string `xml:"MessageId"`
	ReceiptHandle string `xml:"ReceiptHandle"`
	MD5OfBody     string `xml:"MD5OfBody"`
	Body          string `xml:"Body"`
}

// handleReceiveMessage processes a ReceiveMessage request. It dequeues up to
// MaxNumberOfMessages messages (1–10, default 1) from the target queue and
// returns them. Returns an empty result if the queue is empty.
func (s *Server) handleReceiveMessage(w http.ResponseWriter, r *http.Request) {
	queueURL := r.FormValue("QueueUrl")
	if queueURL == "" {
		writeError(w, http.StatusBadRequest, "MissingParameter", "QueueUrl is required")
		return
	}

	maxMessages := 1
	if raw := r.FormValue("MaxNumberOfMessages"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n < 1 || n > 10 {
			writeError(w, http.StatusBadRequest, "InvalidParameterValue",
				"MaxNumberOfMessages must be an integer between 1 and 10")
			return
		}
		maxMessages = n
	}

	q := s.queueByURL(w, queueURL)
	if q == nil {
		return
	}

	msgs := q.Dequeue(maxMessages)

	items := make([]receiveMessageItem, len(msgs))
	for i, m := range msgs {
		items[i] = receiveMessageItem{
			MessageID:     m.ID,
			ReceiptHandle: m.ReceiptHandle,
			MD5OfBody:     md5Hex(m.Body),
			Body:          m.Body,
		}
	}

	resp := receiveMessageResponse{
		Xmlns:                "http://queue.amazonaws.com/doc/2012-11-05/",
		ReceiveMessageResult: receiveMessageResult{Messages: items},
		ResponseMetadata:     responseMetadata{RequestID: "local"},
	}
	w.Header().Set("Content-Type", "text/xml")
	xml.NewEncoder(w).Encode(resp)
}
