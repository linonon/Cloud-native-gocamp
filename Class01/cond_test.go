package c01

import (
	"fmt"
	"strconv"
	"sync"
	"testing"
	"time"
)

type Queue struct {
	queue []string
	cond  *sync.Cond
}

func (q *Queue) Enqueue(item string) {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()

	q.queue = append(q.queue, item)
	fmt.Printf("putting %v to queue, notify all\n", item)
	q.cond.Broadcast()
}

func (q *Queue) Dequeue() string {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()

	if len(q.queue) == 0 {
		fmt.Println("no data available, waiting...")
		q.cond.Wait()
	}
	result := q.queue[0]
	q.queue = q.queue[1:]
	fmt.Printf("getting item: %v, has %d items left\n", result, len(q.queue))
	return result
}

func TestQueue(t *testing.T) {
	q := Queue{
		queue: []string{},
		cond:  sync.NewCond(&sync.Mutex{}),
	}

	go func() {
		var i = 0
		for {
			i++
			q.Enqueue(strconv.Itoa(i))
			i++
			q.Enqueue(strconv.Itoa(i))
			time.Sleep(time.Second * 2)
		}
	}()

	for {
		q.Dequeue()
		time.Sleep(time.Second)
	}
}
