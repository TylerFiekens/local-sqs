package queue

import (
	"sync"
	"testing"
	"time"
)

func msg(id, body, receipt string) *Message {
	return &Message{ID: id, Body: body, ReceiptHandle: receipt, SentAt: time.Now()}
}

func TestNew(t *testing.T) {
	q := New("test-queue", "http://localhost:9324/queue/test-queue")
	if q.Name != "test-queue" {
		t.Fatalf("want Name=test-queue, got %q", q.Name)
	}
	if q.URL != "http://localhost:9324/queue/test-queue" {
		t.Fatalf("unexpected URL %q", q.URL)
	}
	if q.Len() != 0 {
		t.Fatal("new queue must be empty")
	}
}

func TestEnqueue(t *testing.T) {
	q := New("q", "u")
	q.Enqueue(msg("1", "hello", "r1"))
	if q.Len() != 1 {
		t.Fatalf("want len=1, got %d", q.Len())
	}
	q.Enqueue(msg("2", "world", "r2"))
	if q.Len() != 2 {
		t.Fatalf("want len=2, got %d", q.Len())
	}
}

func TestDequeue_empty(t *testing.T) {
	q := New("q", "u")
	got := q.Dequeue(10)
	if got != nil {
		t.Fatalf("want nil on empty queue, got %v", got)
	}
}

func TestDequeue_fifo(t *testing.T) {
	q := New("q", "u")
	q.Enqueue(msg("1", "first", "r1"))
	q.Enqueue(msg("2", "second", "r2"))
	q.Enqueue(msg("3", "third", "r3"))

	got := q.Dequeue(2)
	if len(got) != 2 {
		t.Fatalf("want 2 messages, got %d", len(got))
	}
	if got[0].ID != "1" || got[1].ID != "2" {
		t.Fatalf("want FIFO order [1,2], got [%s,%s]", got[0].ID, got[1].ID)
	}
	if q.Len() != 1 {
		t.Fatalf("want 1 remaining, got %d", q.Len())
	}
}

func TestDequeue_capped_at_queue_size(t *testing.T) {
	q := New("q", "u")
	q.Enqueue(msg("1", "a", "r1"))
	q.Enqueue(msg("2", "b", "r2"))

	got := q.Dequeue(100)
	if len(got) != 2 {
		t.Fatalf("want 2 (capped at queue size), got %d", len(got))
	}
	if q.Len() != 0 {
		t.Fatal("queue must be empty after full dequeue")
	}
}

func TestDelete_found(t *testing.T) {
	q := New("q", "u")
	q.Enqueue(msg("1", "a", "r1"))
	q.Enqueue(msg("2", "b", "r2"))
	q.Enqueue(msg("3", "c", "r3"))

	if !q.Delete("r2") {
		t.Fatal("want true for existing receipt handle")
	}
	if q.Len() != 2 {
		t.Fatalf("want len=2 after delete, got %d", q.Len())
	}
	// verify r2 gone, order preserved
	remaining := q.Dequeue(10)
	if remaining[0].ID != "1" || remaining[1].ID != "3" {
		t.Fatalf("want [1,3] after deleting 2, got [%s,%s]", remaining[0].ID, remaining[1].ID)
	}
}

func TestDelete_not_found(t *testing.T) {
	q := New("q", "u")
	q.Enqueue(msg("1", "a", "r1"))

	if q.Delete("no-such-handle") {
		t.Fatal("want false for missing receipt handle")
	}
	if q.Len() != 1 {
		t.Fatal("queue must be unchanged after failed delete")
	}
}

func TestDelete_empty_queue(t *testing.T) {
	q := New("q", "u")
	if q.Delete("anything") {
		t.Fatal("want false on empty queue")
	}
}

func TestConcurrency(t *testing.T) {
	q := New("q", "u")
	const goroutines = 50

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := range goroutines {
		go func(i int) {
			defer wg.Done()
			id := string(rune('A' + i%26))
			q.Enqueue(msg(id, "body", id+"-handle"))
		}(i)
	}
	wg.Wait()

	if q.Len() != goroutines {
		t.Fatalf("want %d messages, got %d", goroutines, q.Len())
	}
}
