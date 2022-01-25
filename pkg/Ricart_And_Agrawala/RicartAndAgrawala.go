package RicartAndAgrawala

import (
	"VAA_Uebung1/pkg/Cluster/Bank"
	"container/list"

	"github.com/hashicorp/memberlist"
)

type RequesAccountAccess struct {
	Time    *LamportClock
	Sender  *memberlist.Node
	Account *Bank.Account
}

type RicartAndAgrawala struct {
	RequestQueue      *list.List
	CurrentlyUsingR   *Bank.Account
	RequestedResource *Bank.Account
	AskedResource     *Bank.Account
}

func New() *RicartAndAgrawala {
	return &RicartAndAgrawala{RequestQueue: list.New()}
}

func compaire(req1 RequesAccountAccess, req2 RequesAccountAccess) bool {
	if req1.Sender.Name == req2.Sender.Name {
		if req1.Account.Account_Holder.Name == req2.Account.Account_Holder.Name {
			return true
		}
	}
	return false
}

func (ra *RicartAndAgrawala) Contains(requestAA RequesAccountAccess) (bool, RequesAccountAccess) {
	for element := ra.RequestQueue.Front(); element != nil; element = element.Next() {
		temp := element.Value.(RequesAccountAccess)
		if compaire(temp, requestAA) {
			return true, element.Value.(RequesAccountAccess)
		}
	}

	return false, RequesAccountAccess{}
}

func (ra *RicartAndAgrawala) AddRequestToQueue(requestAA RequesAccountAccess) {
	contain, _ := ra.Contains(requestAA)
	if !contain {
		ra.RequestQueue.PushBack(requestAA)
	}
}

//this will return and remove the first value in the queue
func (ra *RicartAndAgrawala) GetFirstQueueRequest() RequesAccountAccess {
	value := ra.RequestQueue.Front()
	ra.RequestQueue.Remove(value)
	return value.Value.(RequesAccountAccess)
}

// func (ra *RequesAccountAccess)
