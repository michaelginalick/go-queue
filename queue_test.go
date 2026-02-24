package goqueue

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"
)

func BenchmarkGoQueue(t *testing.B) {
	q, _ := NewQueue(10)

	for i := 0; i < 1_000_000; i++ {
		q.Add(context.Background(), func(ctx context.Context) {
			s := "new string"
			strings.Join([]string{s}, "new")
		})
	}

	<-q.Idle()
}

func TestQueueIdle(t *testing.T) {
	q, _ := NewQueue(1)
	select {
	case <-q.Idle():
	default:
		t.Errorf("NewQueue(1) is not initially idle.")
	}
	started := make(chan struct{})
	unblock := make(chan struct{})

	q.Add(
		context.TODO(),
		func(context.Context) {
			close(started)
			<-unblock
		},
	)
	<-started
	idle := q.Idle()
	select {
	case <-idle:
		t.Errorf("NewQueue(1) is marked idle while processing work.")
	default:
	}
	close(unblock)
	<-idle // Should be closed as soon as the Add callback returns.
}

func TestQueueBacklog(t *testing.T) {
	const (
		maxActive = 3
		totalWork = 3 * maxActive
	)

	q, _ := NewQueue(maxActive)
	t.Logf("q = NewQueue(%d)", maxActive)
	var wg sync.WaitGroup
	wg.Add(totalWork)

	started := make([]chan struct{}, totalWork)
	unblock := make(chan struct{})
	for i := range started {
		started[i] = make(chan struct{})
		i := i
		q.Add(context.Background(),
			func(context.Context) {
				close(started[i])
				<-unblock
				wg.Done()
			},
		)
	}

	for i, c := range started {
		if i < maxActive {
			<-c
		} else {
			select {
			case <-c:
				t.Errorf("Work item %d started before previous items finished.", i)
			default:

			}
		}
	}
	close(unblock)
	for _, c := range started[maxActive:] {
		<-c
	}
	wg.Wait()
}

func TestQueueLen(t *testing.T) {
	q, _ := NewQueue(1)

	ctx := context.Background()

	q.Add(ctx, func(context.Context) {
		time.Sleep(1 * time.Second)
	})
	q.Add(ctx, func(context.Context) {
		time.Sleep(1 * time.Second)
	})

	l := q.BacklogLen()
	if l != 1 {
		t.Errorf("queue len expected 1 got %d", l)
	}
}

func TestNewQueue(t *testing.T) {
	_, err := NewQueue(0)

	if err == nil {
		t.Errorf("expected error for non-positive queue length")
	}
}
