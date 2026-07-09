package server

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"regexp"

	"github.com/tylerfiekens/local-sqs/internal/queue"
)

// validQueueName matches queue names allowed by the SQS API:
// 1–80 characters, alphanumeric, hyphens, and underscores.
var validQueueName = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,80}$`)

// createQueueResponse is the XML response returned on a successful CreateQueue call.
type createQueueResponse struct {
	XMLName           xml.Name          `xml:"CreateQueueResponse"`
	Xmlns             string            `xml:"xmlns,attr"`
	CreateQueueResult createQueueResult `xml:"CreateQueueResult"`
	ResponseMetadata  responseMetadata  `xml:"ResponseMetadata"`
}

// createQueueResult holds the queue URL returned by CreateQueue.
type createQueueResult struct {
	QueueURL string `xml:"QueueUrl"`
}

// handleCreateQueue processes a CreateQueue request. It is idempotent:
// calling it with an existing queue name returns the existing queue URL.
func (s *Server) handleCreateQueue(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("QueueName")
	if name == "" {
		writeError(w, http.StatusBadRequest, "MissingParameter", "QueueName is required")
		return
	}
	if !validQueueName.MatchString(name) {
		writeError(w, http.StatusBadRequest, "InvalidParameterValue",
			fmt.Sprintf("QueueName %q must be 1-80 alphanumeric characters, hyphens, or underscores", name))
		return
	}

	s.mu.Lock()
	q, exists := s.queues[name]
	if !exists {
		url := fmt.Sprintf("%s/queue/%s", s.baseURL, name)
		q = queue.New(name, url)
		s.queues[name] = q
	}
	s.mu.Unlock()

	resp := createQueueResponse{
		Xmlns:             "http://queue.amazonaws.com/doc/2012-11-05/",
		CreateQueueResult: createQueueResult{QueueURL: q.URL},
		ResponseMetadata:  responseMetadata{RequestID: "local"},
	}
	w.Header().Set("Content-Type", "text/xml")
	xml.NewEncoder(w).Encode(resp)
}
