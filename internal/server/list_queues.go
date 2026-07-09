package server

import (
	"encoding/xml"
	"net/http"
	"sort"
	"strings"
)

// listQueuesResponse is the XML response returned on a successful ListQueues call.
type listQueuesResponse struct {
	XMLName          xml.Name         `xml:"ListQueuesResponse"`
	Xmlns            string           `xml:"xmlns,attr"`
	ListQueuesResult listQueuesResult `xml:"ListQueuesResult"`
	ResponseMetadata responseMetadata `xml:"ResponseMetadata"`
}

// listQueuesResult holds the queue URLs returned by ListQueues.
type listQueuesResult struct {
	QueueURLs []string `xml:"QueueUrl"`
}

// handleListQueues processes a ListQueues request. An optional QueueNamePrefix
// parameter filters results to queues whose names begin with the given prefix.
// Results are returned in alphabetical order by queue name.
func (s *Server) handleListQueues(w http.ResponseWriter, r *http.Request) {
	prefix := r.FormValue("QueueNamePrefix")

	s.mu.RLock()
	urls := make([]string, 0, len(s.queues))
	for name, q := range s.queues {
		if prefix == "" || strings.HasPrefix(name, prefix) {
			urls = append(urls, q.URL)
		}
	}
	s.mu.RUnlock()

	sort.Strings(urls)

	resp := listQueuesResponse{
		Xmlns:            "http://queue.amazonaws.com/doc/2012-11-05/",
		ListQueuesResult: listQueuesResult{QueueURLs: urls},
		ResponseMetadata: responseMetadata{RequestID: "local"},
	}
	w.Header().Set("Content-Type", "text/xml")
	xml.NewEncoder(w).Encode(resp)
}
