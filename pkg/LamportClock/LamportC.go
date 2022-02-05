package Lamportclock

type LamportClock struct {
	Lamport_time int
}

func NewLamportClock() *LamportClock {
	return &LamportClock{Lamport_time: 0}
}

//Return the current value of the lamport clock
func (lc *LamportClock) GetTime() int {
	return lc.Lamport_time
}

// Increment is used to increment and return the value of the lamport clock
func (l *LamportClock) Increment() int {
	l.Lamport_time++
	return l.Lamport_time
}

//This updates the local clock if necessary after
//a clock value received from another process
func (lc *LamportClock) Update(v int) bool {
	// If the param value is less then local, we do not need to do anything
	if v < lc.Lamport_time {
		return false
	}

	lc.Lamport_time = v + 1

	return true
}
