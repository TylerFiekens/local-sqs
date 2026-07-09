// Package server implements an HTTP server that speaks the AWS SQS Query API.
// Clients send POST requests with Content-Type application/x-www-form-urlencoded
// and an Action field. Responses are SQS-compatible XML.
package server

import (
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/tylerfiekens/local-sqs/internal/queue"
)

// Server is an HTTP server that implements the SQS Query API.
// Queues and messages are held in memory and lost on shutdown.
// Server is safe for concurrent use.
type Server struct {
	mux    *http.ServeMux
	queues map[string]*queue.Queue
	mu     sync.RWMutex
}

// New creates a Server with no queues and registers all HTTP routes.
func New() *Server {
	s := &Server{
		mux:    http.NewServeMux(),
		queues: make(map[string]*queue.Queue),
	}
	s.mux.HandleFunc("/", s.handleSQS)
	return s
}

// ServeHTTP implements http.Handler, allowing Server to be passed directly
// to http.ListenAndServe or used in tests via httptest.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

// handleSQS is the single entry point for all SQS API requests.
// It parses the Action field from the form body and dispatches to the
// appropriate handler. Unknown or unimplemented actions return an XML error.
func (s *Server) handleSQS(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		writeError(w, http.StatusBadRequest, "InvalidParameterValue", "failed to parse request body")
		return
	}

	action := r.FormValue("Action")
	if action == "" {
		writeError(w, http.StatusBadRequest, "MissingAction", "Action is required")
		return
	}

	log.Printf("action=%s", action)

	switch action {
	default:
		writeError(w, http.StatusBadRequest, "InvalidAction", fmt.Sprintf("action %q not implemented", action))
	}
}

// errorResponse is the top-level XML envelope returned on any SQS API error.
type errorResponse struct {
	XMLName   xml.Name `xml:"ErrorResponse"`
	Xmlns     string   `xml:"xmlns,attr"`
	Error     sqsError `xml:"Error"`
	RequestID string   `xml:"RequestId"`
}

// sqsError holds the error detail nested inside an errorResponse.
type sqsError struct {
	Type    string `xml:"Type"`
	Code    string `xml:"Code"`
	Message string `xml:"Message"`
}

// writeError writes an SQS-compatible XML error response with the given HTTP
// status code, error code, and human-readable message.
func writeError(w http.ResponseWriter, status int, code, message string) {
	resp := errorResponse{
		Xmlns: "http://queue.amazonaws.com/doc/2012-11-05/",
		Error: sqsError{
			Type:    "Sender",
			Code:    code,
			Message: message,
		},
		RequestID: "local",
	}
	w.Header().Set("Content-Type", "text/xml")
	w.WriteHeader(status)
	xml.NewEncoder(w).Encode(resp)
}
