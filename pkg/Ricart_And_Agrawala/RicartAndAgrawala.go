package RicartAndAgrawala

import (
	"VAA_Uebung1/pkg/Cluster/Bank"
	Lamportclock "VAA_Uebung1/pkg/LamportClock"
	"container/list"
	"fmt"

	"github.com/hashicorp/memberlist"
)

type RequesAccountAccess struct {
	Req_Sender_Time   Lamportclock.LamportClock `json:"req_sender_time"`
	ReqInitiator_Time Lamportclock.LamportClock `json:"req_initiator_time"`
	Initator          *memberlist.Node          `json:"req_ac_access_sender"`
	Account           *Bank.Account             `json:"account"`
	Sender            *memberlist.Node          `json:"flooding_sender"`
}

type AccountAccess_Acknowledge struct {
	Ack_Sender_Time Lamportclock.LamportClock `json:"ac_sender_lamport_time"`
	Ack_Sender      memberlist.Node           `json:"ack_to_ac_access_sender"`
	ReqAA           Bank.Account              `json:"reqted_ac"`
	Status          string                    `json:"ack_status"`
}

type RicartAndAgrawala struct {
	Sending_Req_Time     *Lamportclock.LamportClock
	RequestQueue         *list.List
	CurrentlyUsingR      *Bank.Account
	Own_Rsource          *Bank.Account
	Interested_Resource2 *Bank.Account
	Ack_Waited_Queue     map[string]memberlist.Node
	AccInformation       bool
	Operation_Ack        bool
	Flooding_Queue       *list.List
}

func (ra *RicartAndAgrawala) Add_Ack_Waited_Queue(req_reciever memberlist.Node) {
	if _, ok := ra.Ack_Waited_Queue[req_reciever.Name]; !ok {
		ra.Ack_Waited_Queue[req_reciever.Name] = req_reciever
	}
}

func (ra *RicartAndAgrawala) Remove_From_Ack_Waited_Queue(ack_sender memberlist.Node) bool {
	if ra.Ack_Waited_Queue == nil || (len(ra.Ack_Waited_Queue) == 0) {
		return false
	}
	if _, ok := ra.Ack_Waited_Queue[ack_sender.Name]; !ok {
		return false
	}

	delete(ra.Ack_Waited_Queue, ack_sender.Name)
	return true
}

func (ra *RicartAndAgrawala) Interested_Resource_isEmpty() bool {
	return (ra.Interested_Resource2 == nil)
}

func (ra *RicartAndAgrawala) CurrentlyUnsingR_isEmpty() bool {
	return (ra.CurrentlyUsingR == nil)
}

func (ra *RicartAndAgrawala) Is_CurrentlyUsing(reqAccount1 Bank.Account, reqAccount2 Bank.Account) bool {
	if ra.CurrentlyUnsingR_isEmpty() {
		return false
	}

	//check for resource2
	if ra.CurrentlyUsingR.Account_Holder.Name == reqAccount2.Account_Holder.Name ||
		ra.Own_Rsource.Account_Holder.Name == reqAccount2.Account_Holder.Name {
		return true
	}
	//check for resource1
	return ra.CurrentlyUsingR.Account_Holder.Name == reqAccount1.Account_Holder.Name ||
		ra.Own_Rsource.Account_Holder.Name == reqAccount1.Account_Holder.Name

}

func (ra *RicartAndAgrawala) Is_Interested(reqAccount1 Bank.Account, reqAccount2 Bank.Account) bool {

	if ra.Interested_Resource_isEmpty() {
		return false
	}

	//check for resource2
	if ra.Interested_Resource2.Account_Holder.Name == reqAccount2.Account_Holder.Name ||
		ra.Own_Rsource.Account_Holder.Name == reqAccount2.Account_Holder.Name {
		return true
	}

	//check for resource1
	return ra.Interested_Resource2.Account_Holder.Name == reqAccount1.Account_Holder.Name ||
		ra.Own_Rsource.Account_Holder.Name == reqAccount1.Account_Holder.Name
}

func New_AccountAccess_Acknowledge(reqAC Bank.Account, ack_sender_time Lamportclock.LamportClock,
	ack_sender memberlist.Node, status string) AccountAccess_Acknowledge {

	return AccountAccess_Acknowledge{
		Ack_Sender_Time: ack_sender_time,
		Ack_Sender:      ack_sender,
		ReqAA:           reqAC,
		Status:          status,
	}
}

func New_RequesAccountAccess(t Lamportclock.LamportClock,
	sender memberlist.Node, account Bank.Account) *RequesAccountAccess {

	return &RequesAccountAccess{
		ReqInitiator_Time: t,
		Req_Sender_Time:   t,
		Initator:          &sender,
		Account:           &account,
		Sender:            &sender,
	}
}

func New_R_and_A_Algrthm() *RicartAndAgrawala {
	return &RicartAndAgrawala{RequestQueue: list.New(),
		Ack_Waited_Queue: make(map[string]memberlist.Node),
		Flooding_Queue:   list.New(),
	}
}

func compaire(req1 RequesAccountAccess, req2 RequesAccountAccess) bool {
	if req1.Initator.Name == req2.Initator.Name {
		if req1.Account.Account_Holder.Name == req2.Account.Account_Holder.Name {
			return true
		}
	}
	return false
}

func (ra *RicartAndAgrawala) Queue_toString() string {
	out := "\n"
	out += "----------------------------------------------------------Print Queue Element----------------------------------------------------------\n\n"
	for element := ra.RequestQueue.Front(); element != nil; element = element.Next() {
		temp := element.Value.(RequesAccountAccess)
		out += fmt.Sprintf("\t\t\t\t\t\t\tRequest Initiator: \t%s\n\t\t\t\t\t\t\tSender Time: \t\t%d\n\t\t\t\t\t\t\tInterested Account: \t%s\n",
			temp.Initator.Name, temp.ReqInitiator_Time.GetTime(), temp.Account.Account_Holder.Name,
		)
	}
	out += "\n\n---------------------------------------------------------------End Queue---------------------------------------------------------------\n\n"

	return out
}

func (ra *RicartAndAgrawala) Is_Queue_Contains(requestAA RequesAccountAccess) (bool, RequesAccountAccess) {
	for element := ra.RequestQueue.Front(); element != nil; element = element.Next() {
		temp := element.Value.(RequesAccountAccess)
		if compaire(temp, requestAA) {
			return true, element.Value.(RequesAccountAccess)
		}
	}

	return false, RequesAccountAccess{}
}

func (ra *RicartAndAgrawala) AddRequestToQueue(requestAA RequesAccountAccess) {
	contain, _ := ra.Is_Queue_Contains(requestAA)
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

//---------------------------------------------------------------- Flooding Alg

//saves the flooding request
func (ra *RicartAndAgrawala) Is_Flooding_Queue_Contains(requestAA RequesAccountAccess) (bool, RequesAccountAccess) {
	for element := ra.Flooding_Queue.Front(); element != nil; element = element.Next() {
		temp := element.Value.(RequesAccountAccess)
		if compaireFlooding_req(temp, requestAA) {
			return true, element.Value.(RequesAccountAccess)
		}
	}

	return false, RequesAccountAccess{}
}

func (ra *RicartAndAgrawala) AddTo_Flooding_Queue(requestAA RequesAccountAccess) {
	contain, _ := ra.Is_Flooding_Queue_Contains(requestAA)
	// fmt.Printf("\n\n\nInside Add To flooding queue initiator: %s and contain %t", requestAA.Initator.Name, contain)
	if !contain {
		ra.Flooding_Queue.PushBack(requestAA)
	}
}

func (ra *RicartAndAgrawala) Remove_Old_Flooding_element() {

	value := ra.RequestQueue.Front()
	ra.Flooding_Queue.Remove(value)

}

func compaireFlooding_req(req1 RequesAccountAccess, req2 RequesAccountAccess) bool {
	if req1.Initator.Name == req2.Initator.Name {
		if req1.Account.Account_Holder.Name == req2.Account.Account_Holder.Name {
			if req1.ReqInitiator_Time.GetTime() == req2.ReqInitiator_Time.GetTime() {
				return true
			}
		}
	}
	return false
}

func (ra *RicartAndAgrawala) Flooding_Queue_toString() string {
	out := "\n"
	out += "----------------------------------------------------------Print Queue Element----------------------------------------------------------\n\n"
	for element := ra.Flooding_Queue.Front(); element != nil; element = element.Next() {
		temp := element.Value.(RequesAccountAccess)
		out += fmt.Sprintf("\t\t\t\t\t\t\tRequest Initiator: \t%s Request sender: %s\n\t\t\t\t\t\t\tSender Time: \t\t%d\n\t\t\t\t\t\t\tInterested Account: \t%s\n",
			temp.Initator.Name, temp.Sender.Name, temp.ReqInitiator_Time.GetTime(), temp.Account.Account_Holder.Name,
		)
	}
	out += "\n\n---------------------------------------------------------------End Queue---------------------------------------------------------------\n\n"

	return out
}
