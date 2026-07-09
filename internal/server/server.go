package server

import (
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/tylerfiekens/local-sqs/internal/queue"
)

type Server struct {
	mux    *http.ServeMux
	queues map[string]*queue.Queue
	mu     sync.RWMutex
}

func New() *Server {
	s := &Server{
		mux:    http.NewServeMux(),
		queues: make(map[string]*queue.Queue),
	}
	s.mux.HandleFunc("/", s.handleSQS)
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

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

type errorResponse struct {
	XMLName   xml.Name `xml:"ErrorResponse"`
	Xmlns     string   `xml:"xmlns,attr"`
	Error     sqsError `xml:"Error"`
	RequestID string   `xml:"RequestId"`
}

type sqsError struct {
	Type    string `xml:"Type"`
	Code    string `xml:"Code"`
	Message string `xml:"Message"`
}

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
