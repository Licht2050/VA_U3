package main

import (
	"container/list"
	"fmt"
	"sync/atomic"
	"time"
)

type LamportClock struct {
	lamport_time uint64
}

func NewLamportClock() *LamportClock {
	return &LamportClock{lamport_time: 0}
}

//Return the current value of the lamport clock
func (lc *LamportClock) GetTime() uint64 {
	return lc.lamport_time
}

// Increment is used to increment and return the value of the lamport clock
func (l *LamportClock) Increment() uint64 {
	return atomic.AddUint64(&l.lamport_time, 1)
}

//This updates the local clock if necessary after
//a clock value received from another process
func (lc *LamportClock) Update(v uint64) {
	// If the other value is old, we do not need to do anything
	cur := atomic.LoadUint64(&lc.lamport_time)
	other := uint64(v)
	if other < cur {
		return
	}

	atomic.SwapUint64(&lc.lamport_time, other+1)
}

func contain(queue *list.List, lc LamportClock) (bool, *LamportClock) {
	for e := queue.Front(); e != nil; e = e.Next() {
		temp := e.Value.(*LamportClock)
		if *temp == lc {
			return true, e.Value.(*LamportClock)
		}
		// fmt.Println(e.Value)
	}

	return false, &LamportClock{}
}

func test(ch chan int) {
	i := 0
	for i < 10 {
		time.Sleep(1 * time.Second)

		fmt.Println("Test time: ", i)

		if i == 5 {
			ch <- 10
			break
		}
		i++
	}

}

func main() {
	// lc := NewLamportClock()
	// lc1 := NewLamportClock()

	// lc1.Increment()
	// lc1.Increment()

	// fmt.Println(lc.GetTime())

	// lc.Update(lc1.GetTime())
	// fmt.Println(lc.GetTime())

	// queue := list.New()

	// queue.PushBack(*lc)
	// queue.PushBack(*lc1)

	// // fmt.Println(contain(queue, *lc1))
	// // queue.PushBack(lc1)

	// for queue.Len() > 0 {
	// 	test := queue.Front().Value.(LamportClock)
	// 	fmt.Println(test.GetTime())

	// 	queue.Remove(queue.Front())
	// }
	ch := make(chan int, 1)

	go test(ch)

	v := <-ch

	if v == 10 {
		fmt.Println("Gooooooooooooooooooooooooooooooooooooooooooooo")
	}

	i := 0
	for i < 10 {
		fmt.Println("Main: ", i)
		time.Sleep(1 * time.Second)
		i++
	}

}
