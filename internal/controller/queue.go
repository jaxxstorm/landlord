package controller

import (
	"time"

	"k8s.io/client-go/util/workqueue"
)

// Queue wraps a rate-limiting workqueue for tenant reconciliation
type Queue struct {
	queue workqueue.RateLimitingInterface
}

// NewRateLimitingQueue creates a new workqueue with exponential backoff
// Base delay: 1 second, max delay: 5 minutes
func NewRateLimitingQueue() *Queue {
	rateLimiter := workqueue.NewItemExponentialFailureRateLimiter(
		1*time.Second, // base delay
		5*time.Minute, // max delay
	)

	return &Queue{
		queue: workqueue.NewRateLimitingQueue(rateLimiter),
	}
}

// Add adds an item to the queue
func (q *Queue) Add(item interface{}) {
	q.queue.Add(item)
}

// Get retrieves an item from the queue (blocks if empty)
func (q *Queue) Get() (item interface{}, shutdown bool) {
	return q.queue.Get()
}

// Done marks an item as finished processing
func (q *Queue) Done(item interface{}) {
	q.queue.Done(item)
}

// AddRateLimited adds an item with rate limiting (for retries)
func (q *Queue) AddRateLimited(item interface{}) {
	q.queue.AddRateLimited(item)
}

// Forget indicates successful processing (resets backoff for this item)
func (q *Queue) Forget(item interface{}) {
	q.queue.Forget(item)
}

// ShutDown signals the queue to shut down
func (q *Queue) ShutDown() {
	q.queue.ShutDown()
}

// ShuttingDown returns whether the queue is shutting down
func (q *Queue) ShuttingDown() bool {
	return q.queue.ShuttingDown()
}

// Len returns the number of items in the queue
func (q *Queue) Len() int {
	return q.queue.Len()
}
