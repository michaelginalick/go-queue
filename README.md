# goqueue


[![test](https://github.com/michaelginalick/go-queue/actions/workflows/test.yml/badge.svg)](https://github.com/michaelginalick/go-queue/actions/workflows/test.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/michaelginalick/go-queue)](https://goreportcard.com/report/github.com/michaelginalick/go-queue)
[![GoDoc](https://godoc.org/github.com/michaelginalick/go-queue?status.svg)](https://godoc.org/github.com/michaelginalick/go-queue)

`goqueue` provides a small concurrency-limited FIFO work queue for Go.

This implementation is derived from the `par` package in the Go standard library (see `cmd/go/internal/par`). The design is adapted into a reusable package.

The queue ensures that no more than a fixed number of functions execute concurrently. Additional functions are queued and executed in first-in-first-out (FIFO) order.

## Installation

```bash
go get github.com/michaelginalick/go-queue
```

## Overview

# goqueue:

- Limits the number of concurrently running functions.
- Executes queued functions in FIFO order.
- Is safe for concurrent use by multiple goroutines.
- Provides a way to wait until all work has completed.
- The zero value of Queue is not usable. Use NewQueue to construct one.

# Usage

```golang
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/michaelginalick/go-queue"
)

func main() {
	q, err := goqueue.NewQueue(2)
	if err != nil {
		panic(err)
	}

	ctx := context.Background()

	for i := 1; i <= 5; i++ {
		i := i
		q.Add(ctx, func(ctx context.Context) {
			fmt.Printf("start %d\n", i)
			time.Sleep(100 * time.Millisecond)
			fmt.Printf("done %d\n", i)
		})
	}

	// Wait until all queued and active work completes.
	<-q.Idle()
	fmt.Println("all done")
}
```

See ```/examples``` for additional usage.

# API
### ```NewQueue(maxActive int) (*Queue, error)```
- Creates a new queue that allows at most maxActive functions to run concurrently.
Returns an error if maxActive < 1.

### ```(*Queue) Add(ctx context.Context, f func(context.Context))```
- Submits a function for execution.
- If fewer than maxActive functions are currently running, f begins immediately in a new goroutine.
- Otherwise, f is added to a FIFO backlog.
- Add does not block.
- The provided context.Context is passed to f when it executes.

### ```(*Queue) Idle() <-chan struct{}```
- Returns a channel that is closed when the queue becomes idle (no active functions and no backlog).
- If the queue is already idle, the returned channel is already closed.

### ```(*Queue) Len() int64```
- Returns the number of functions currently waiting in the backlog.
- Active functions are not included.

### Behavior Notes
- Concurrency is limited to maxActive.
- Execution order of queued tasks is FIFO.
- Each task runs in its own goroutine.
- If a task panics, the queue may enter an inconsistent state.

### Relationship to the Go Standard Library
This implementation is adapted from the ```par``` package in the Go toolchain (cmd/go/internal/par) in the Go standard library. That package is internal to the Go command and cannot be imported directly, so this repository provides a reusable version of the same core idea.