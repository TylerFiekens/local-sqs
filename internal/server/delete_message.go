package server

import (
	"encoding/xml"
	"net/http"
)

// deleteMessageResponse is the XML response returned on a successful DeleteMessage call.
type deleteMessageResponse struct {
	XMLName          xml.Name         `xml:"DeleteMessageResponse"`
	Xmlns            string           `xml:"xmlns,attr"`
	ResponseMetadata responseMetadata `xml:"ResponseMetadata"`
}

// handleDeleteMessage processes a DeleteMessage request. It removes the message
// identified by ReceiptHandle from the target queue. The operation is idempotent:
// an unknown ReceiptHandle returns 200 with no error, matching AWS behaviour.
func (s *Server) handleDeleteMessage(w http.ResponseWriter, r *http.Request) {
	queueURL := r.FormValue("QueueUrl")
	if queueURL == "" {
		writeError(w, http.StatusBadRequest, "MissingParameter", "QueueUrl is required")
		return
	}

	receiptHandle := r.FormValue("ReceiptHandle")
	if receiptHandle == "" {
		writeError(w, http.StatusBadRequest, "MissingParameter", "ReceiptHandle is required")
		return
	}

	q := s.queueByURL(w, queueURL)
	if q == nil {
		return
	}

	q.Delete(receiptHandle)

	resp := deleteMessageResponse{
		Xmlns:            "http://queue.amazonaws.com/doc/2012-11-05/",
		ResponseMetadata: responseMetadata{RequestID: "local"},
	}
	w.Header().Set("Content-Type", "text/xml")
	xml.NewEncoder(w).Encode(resp)
}
