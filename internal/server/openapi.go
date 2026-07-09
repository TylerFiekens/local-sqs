package server

import (
	"net/http"

	"github.com/tylerfiekens/local-sqs/api"
)

// handleOpenAPI serves the OpenAPI specification as YAML.
func (s *Server) handleOpenAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/yaml")
	w.Write(api.Spec)
}
