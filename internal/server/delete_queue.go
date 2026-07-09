package server

import (
	"encoding/xml"
	"net/http"
	"net/url"
	"path"
)

// deleteQueueResponse is the XML response returned on a successful DeleteQueue call.
type deleteQueueResponse struct {
	XMLName          xml.Name         `xml:"DeleteQueueResponse"`
	Xmlns            string           `xml:"xmlns,attr"`
	ResponseMetadata responseMetadata `xml:"ResponseMetadata"`
}

// handleDeleteQueue processes a DeleteQueue request. It is idempotent:
// deleting a queue that does not exist returns 200 with no error, matching AWS behaviour.
func (s *Server) handleDeleteQueue(w http.ResponseWriter, r *http.Request) {
	queueURL := r.FormValue("QueueUrl")
	if queueURL == "" {
		writeError(w, http.StatusBadRequest, "MissingParameter", "QueueUrl is required")
		return
	}

	u, err := url.Parse(queueURL)
	if err != nil {
		writeError(w, http.StatusBadRequest, "InvalidParameterValue", "QueueUrl is not a valid URL")
		return
	}

	name := path.Base(u.Path)

	s.mu.Lock()
	delete(s.queues, name)
	s.mu.Unlock()

	resp := deleteQueueResponse{
		Xmlns:            "http://queue.amazonaws.com/doc/2012-11-05/",
		ResponseMetadata: responseMetadata{RequestID: "local"},
	}
	w.Header().Set("Content-Type", "text/xml")
	xml.NewEncoder(w).Encode(resp)
}
