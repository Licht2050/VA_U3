package Snapshot

import (
	Lamportclock "VAA_Uebung1/pkg/LamportClock"
	"container/list"
)

type Ch struct {
	C      chan interface{}
	Closed bool
}

type Chandy_Lamport struct {
	Incommint_Channel map[string]Ch
	Outgoing_Channel  map[string]Ch
	Snapshot_Time     *Lamportclock.LamportClock
	Incommint_Message *list.List
	Snapshot_Start    bool
	Message_recivied  bool
}

func (ch *Ch) Close_Channel() {
	ch.Closed = true

}

func NewChandy_Lamport() *Chandy_Lamport {
	return &Chandy_Lamport{
		Incommint_Channel: make(map[string]Ch),
		Outgoing_Channel:  make(map[string]Ch),
		Snapshot_Start:    false,
	}

}
