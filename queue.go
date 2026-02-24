// Package goqueue provides a concurrency-limited FIFO work queue.
//
// A Queue executes submitted functions while ensuring that no more than a
// configured number run concurrently. Additional functions are queued and
// executed in first-in-first-out (FIFO) order as running functions complete.
//
// The zero value of Queue is not usable. Use NewQueue to construct a Queue.
package goqueue

import (
	"container/list"
	"context"
	"fmt"
)

// Queue represents a concurrency-limited FIFO work queue.
//
// A Queue guarantees that at most maxActive functions are running at any
// given time. If the limit has been reached, additional functions submitted
// with Add are placed in a backlog and executed in submission order.
//
// Queue is safe for concurrent use by multiple goroutines.
type Queue struct {
	maxActive int
	st        chan queueState
}

type queueState struct {
	active  int
	backlog *list.List
	idle    chan struct{}
}

// NewQueue creates a new Queue that allows at most maxActive functions
// to run concurrently.
//
// maxActive must be greater than zero. If maxActive is less than 1,
// NewQueue returns an error.
func NewQueue(maxActive int) (*Queue, error) {
	if maxActive < 1 {
		return nil, fmt.Errorf("goQueue called with nonpositive limit (%d)", maxActive)
	}

	q := &Queue{maxActive: maxActive, st: make(chan queueState, 1)}
	q.st <- queueState{backlog: list.New()}
	return q, nil
}
// Add submits a function to the Queue for execution.
//
// If fewer than the maximum number of functions are currently running,
// f is executed immediately in its own goroutine. Otherwise, f is added
// to the backlog and will be executed in FIFO order when capacity becomes
// available.
//
// The provided context is passed to the function when it executes.
// Add does not block waiting for execution to begin.
//
// The function f must not panic. If f panics, the behavior of the Queue
// is undefined.
func (q *Queue) Add(ctx context.Context, f func(context.Context)) {
	st := <-q.st
	if st.active == q.maxActive {
		st.backlog.PushBack(f)
		q.st <- st
		return
	}

	if st.active == 0 {
		// Mark q as non-idle
		st.idle = nil
	}

	st.active++
	q.st <- st

	go func() {
		for {
			f(ctx)

			st := <-q.st
			if st.backlog.Len() == 0 {
				if st.active--; st.active == 0 && st.idle != nil {
					close(st.idle)
				}
				q.st <- st
				return
			}
			f = st.backlog.Remove(st.backlog.Front()).(func(context.Context))
			q.st <- st
		}
	}()

}

// Idle returns a channel that is closed when the Queue becomes idle.
//
// The returned channel is closed when there are no active functions running
// and the backlog is empty. If the Queue is already idle at the time of the
// call, the returned channel is already closed.
//
// Multiple calls to Idle may return the same channel while the Queue
// remains non-idle.
func (q *Queue) Idle() <-chan struct{} {
	st := <-q.st
	defer func() { q.st <- st }()
	if st.idle == nil {
		st.idle = make(chan struct{})
		if st.active == 0 {
			close(st.idle)
		}
	}
	return st.idle
}

// BacklogLen returns the number of functions currently waiting in the backlog.
//
// This does not include functions that are actively running.
func (q *Queue) BacklogLen() int64 {
	st := <-q.st
	defer func() { q.st <- st }()
	return int64(st.backlog.Len())
}
