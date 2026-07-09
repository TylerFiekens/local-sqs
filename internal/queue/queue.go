// Package queue provides an in-memory SQS queue implementation.
package queue

import (
	"sync"
	"time"
)

// Message represents a single SQS message held in a queue.
type Message struct {
	// ID is the unique identifier assigned to the message on send.
	ID string
	// Body is the raw message payload.
	Body string
	// ReceiptHandle is the opaque token used to delete the message after processing.
	ReceiptHandle string
	// SentAt is the time the message was enqueued.
	SentAt time.Time
}

// Queue is a thread-safe, in-memory FIFO message queue.
// All state is lost when the process exits.
type Queue struct {
	// Name is the queue name as provided at creation.
	Name string
	// URL is the full endpoint URL used by SQS clients to address this queue.
	URL      string
	mu       sync.Mutex
	messages []*Message
}

// New creates an empty Queue with the given name and URL.
func New(name, url string) *Queue {
	return &Queue{
		Name: name,
		URL:  url,
	}
}

// Enqueue appends msg to the tail of the queue. Safe for concurrent use.
func (q *Queue) Enqueue(msg *Message) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.messages = append(q.messages, msg)
}

// Dequeue removes and returns up to maxMessages messages from the head of the
// queue in FIFO order. Returns nil if the queue is empty. Safe for concurrent use.
func (q *Queue) Dequeue(maxMessages int) []*Message {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.messages) == 0 {
		return nil
	}

	n := min(maxMessages, len(q.messages))
	out := make([]*Message, n)
	copy(out, q.messages[:n])
	q.messages = q.messages[n:]
	return out
}

// Delete removes the message identified by receiptHandle from the queue.
// Returns true if the message was found and removed, false otherwise.
// Safe for concurrent use.
func (q *Queue) Delete(receiptHandle string) bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	for i, m := range q.messages {
		if m.ReceiptHandle == receiptHandle {
			q.messages = append(q.messages[:i], q.messages[i+1:]...)
			return true
		}
	}
	return false
}

// Len returns the number of messages currently in the queue. Safe for concurrent use.
func (q *Queue) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.messages)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
