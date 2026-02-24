package main

import (
	"context"
	"fmt"
	"os"
	"time"

	goqueue "github.com/michaelginalick/go-queue"
)

func main() {
	deadline := time.Now().Add(4 * time.Second)
	ctx, ctxFunc := context.WithDeadline(context.Background(), deadline)
	defer ctxFunc()
	qu, err := goqueue.NewQueue(2)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	qu.Add(
		ctx, func(ctx context.Context) {
			time.Sleep(5 * time.Second)
			fmt.Println("5 sleep")
		},
	)
	qu.Add(
		ctx, func(ctx context.Context) {
			time.Sleep(3 * time.Second)
			fmt.Println("3 sleep")
		},
	)

	qu.Add(
		ctx, func(ctx context.Context) {
			time.Sleep(6 * time.Second)
			fmt.Println("6 sleep")
		},
	)

	select {
	case <-ctx.Done():
		fmt.Println("Done")
	case <-qu.Idle():
		fmt.Println("Working")
	}
}
