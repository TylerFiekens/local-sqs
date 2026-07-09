package server

import (
	"crypto/md5"
	"crypto/rand"
	"fmt"
	"net/http"
	"net/url"
	"path"

	"github.com/tylerfiekens/local-sqs/internal/queue"
)

// queueByURL parses queueURL, looks up the queue by name, and returns it.
// Writes an XML error and returns nil if the URL is invalid or the queue does not exist.
func (s *Server) queueByURL(w http.ResponseWriter, queueURL string) *queue.Queue {
	u, err := url.Parse(queueURL)
	if err != nil {
		writeError(w, http.StatusBadRequest, "InvalidParameterValue", "QueueUrl is not a valid URL")
		return nil
	}

	name := path.Base(u.Path)

	s.mu.RLock()
	q := s.queues[name]
	s.mu.RUnlock()

	if q == nil {
		writeError(w, http.StatusBadRequest, "AWS.SimpleQueueService.NonExistentQueue",
			fmt.Sprintf("queue %q does not exist", name))
	}
	return q
}

// newID returns a random UUID v4.
func newID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%12x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

// md5Hex returns the lowercase hex MD5 digest of s.
func md5Hex(s string) string {
	sum := md5.Sum([]byte(s))
	return fmt.Sprintf("%x", sum)
}
