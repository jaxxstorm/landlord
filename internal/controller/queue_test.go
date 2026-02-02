package controller

import (
	"testing"
	"time"
)

func TestNewRateLimitingQueue(t *testing.T) {
	q := NewRateLimitingQueue()
	if q == nil {
		t.Fatal("NewRateLimitingQueue() returned nil")
	}
	if q.Len() != 0 {
		t.Errorf("NewRateLimitingQueue() len = %d, want 0", q.Len())
	}
}

func TestQueueAdd(t *testing.T) {
	q := NewRateLimitingQueue()
	q.Add("key1")
	if q.Len() != 1 {
		t.Errorf("Queue.Add() len = %d, want 1", q.Len())
	}

	item, shutdown := q.Get()
	if shutdown {
		t.Fatal("Queue.Get() returned shutdown=true, want false")
	}
	if item != "key1" {
		t.Errorf("Queue.Get() = %v, want key1", item)
	}
}

func TestQueueAddDuplicate(t *testing.T) {
	q := NewRateLimitingQueue()
	q.Add("key1")
	q.Add("key1")
	q.Add("key1")

	if q.Len() != 1 {
		t.Errorf("Queue with duplicates len = %d, want 1 (deduped)", q.Len())
	}

	item, _ := q.Get()
	if item != "key1" {
		t.Errorf("Queue.Get() = %v, want key1", item)
	}
}

func TestQueueAddRateLimited(t *testing.T) {
	q := NewRateLimitingQueue()
	q.Add("key1")
	item, _ := q.Get()
	q.Done(item)

	q.AddRateLimited("key1")
	// AddRateLimited uses exponential backoff, so the item won't be immediately available
	// Just verify it doesn't panic and the queue is in a valid state
	if q.ShuttingDown() {
		t.Fatal("Queue should not be shutting down")
	}

	q.ShutDown()
}

func TestQueueForget(t *testing.T) {
	q := NewRateLimitingQueue()
	q.Add("key1")
	item, _ := q.Get()

	q.Forget("key1")
	q.Done(item)

	if q.Len() != 0 {
		t.Errorf("Queue.Forget() len = %d, want 0", q.Len())
	}
}

func TestQueueShutdown(t *testing.T) {
	q := NewRateLimitingQueue()
	q.Add("key1")
	item, _ := q.Get()

	q.ShutDown()
	if !q.ShuttingDown() {
		t.Fatal("Queue.ShuttingDown() = false, want true after ShutDown()")
	}

	q.Done(item)

	_, shutdown := q.Get()
	if !shutdown {
		t.Fatal("Queue.Get() returned shutdown=false after ShutDown(), want true")
	}
}

func TestQueueMultipleOperations(t *testing.T) {
	q := NewRateLimitingQueue()

	q.Add("key1")
	q.Add("key2")
	q.Add("key3")

	if q.Len() != 3 {
		t.Errorf("Queue with 3 items len = %d, want 3", q.Len())
	}

	item1, _ := q.Get()
	item2, _ := q.Get()
	item3, _ := q.Get()

	expectedSet := map[string]bool{"key1": true, "key2": true, "key3": true}
	if !expectedSet[item1.(string)] || !expectedSet[item2.(string)] || !expectedSet[item3.(string)] {
		t.Errorf("Queue.Get() returned unexpected items")
	}

	q.Done(item1)
	q.Done(item2)
	q.Done(item3)

	if q.Len() != 0 {
		t.Errorf("Queue after processing all items len = %d, want 0", q.Len())
	}
}

func TestQueueGetTimeout(t *testing.T) {
	q := NewRateLimitingQueue()
	done := make(chan bool, 1)

	go func() {
		_, shutdown := q.Get()
		done <- shutdown
	}()

	time.Sleep(100 * time.Millisecond)
	q.ShutDown()

	select {
	case shutdown := <-done:
		if !shutdown {
			t.Fatal("Queue.Get() should return shutdown=true when queue is shut down")
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Queue.Get() did not return in time")
	}
}
