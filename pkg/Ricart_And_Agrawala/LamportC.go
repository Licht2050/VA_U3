package RicartAndAgrawala

import (
	"sync/atomic"
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
