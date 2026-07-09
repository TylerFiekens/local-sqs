package queue

import (
	"sync"
	"time"
)

type Message struct {
	ID            string
	Body          string
	ReceiptHandle string
	SentAt        time.Time
}

type Queue struct {
	Name     string
	URL      string
	mu       sync.Mutex
	messages []*Message
}

func New(name, url string) *Queue {
	return &Queue{
		Name: name,
		URL:  url,
	}
}

func (q *Queue) Enqueue(msg *Message) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.messages = append(q.messages, msg)
}

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
