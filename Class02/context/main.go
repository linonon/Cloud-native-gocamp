package main

import (
	"context"
	"fmt"
	"time"
)

type T struct {
	A, B string
}

func main() {
	// 最頂層上下文
	baseCtx := context.Background()
	a := T{A: "a"}
	b := T{B: "b"}
	ctx := context.WithValue(baseCtx, a, b)
	go func(ctx context.Context) {
		fmt.Println(ctx.Value(a)) // { b}
	}(ctx)
	timeoutCtx, cancel := context.WithTimeout(baseCtx, 5*time.Second)
	defer cancel() // releases resources if slowOperation completes before timeout elapses
	go func(ctx context.Context) {
		ticker := time.NewTicker(1 * time.Second)
		for range ticker.C {
			select {
			case <-ctx.Done():
				fmt.Println("child process interrupt...")
				return
			default:
				fmt.Println("enter default")
			}
		}
	}(timeoutCtx)
	select {
	case <-timeoutCtx.Done():
		fmt.Println("main process exit")
		time.Sleep(10 * time.Second)
	}
}
