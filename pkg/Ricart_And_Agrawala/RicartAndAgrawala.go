package RicartAndAgrawala

import (
	"VAA_Uebung1/pkg/Cluster/Bank"
	"container/list"
	"fmt"

	"github.com/hashicorp/memberlist"
)

type RequesAccountAccess struct {
	Req_Sender_Time LamportClock     `json:"lamport_time"`
	Sender          *memberlist.Node `json:"req_ac_access_sender"`
	Account         *Bank.Account    `json:"account"`
}

type AccountAccess_Acknowledge struct {
	Ack_Sender_Time *LamportClock   `json:"ac_sender_lamport_time"`
	Ack_Sender      memberlist.Node `json:"ack_to_ac_access_sender"`
	ReqAA           Bank.Account    `json:"reqted_ac"`
	Status          string          `json:"ack_status"`
}

type RicartAndAgrawala struct {
	RequestQueue        *list.List
	CurrentlyUsingR     *Bank.Account
	RequestedResource   *Bank.Account
	Interested_Resource *Bank.Account
	Ack_Waited_Queue    map[string]memberlist.Node
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
	return (ra.Interested_Resource == nil)
}

func (ra *RicartAndAgrawala) CurrentlyUnsingR_isEmpty() bool {
	return (ra.CurrentlyUsingR == nil)
}

func (ra *RicartAndAgrawala) Is_CurrentlyUsing(reqAccount Bank.Account) bool {
	if ra.CurrentlyUnsingR_isEmpty() {
		return false
	}

	return ra.CurrentlyUsingR.Account_Holder.Name == reqAccount.Account_Holder.Name
}

func (ra *RicartAndAgrawala) Is_Interested(reqAccount Bank.Account) bool {
	if ra.Interested_Resource_isEmpty() {
		return false
	}

	return ra.Interested_Resource.Account_Holder.Name == reqAccount.Account_Holder.Name
}

func New_AccountAccess_Acknowledge(reqAC Bank.Account, ack_sender_time LamportClock,
	ack_sender memberlist.Node, status string) AccountAccess_Acknowledge {

	return AccountAccess_Acknowledge{
		Ack_Sender_Time: &ack_sender_time,
		Ack_Sender:      ack_sender,
		ReqAA:           reqAC,
		Status:          status,
	}
}

func New_RequesAccountAccess(t LamportClock,
	sender memberlist.Node, account Bank.Account) *RequesAccountAccess {

	return &RequesAccountAccess{
		Req_Sender_Time: t,
		Sender:          &sender,
		Account:         &account}
}

func New_R_and_A_Algrthm() *RicartAndAgrawala {
	return &RicartAndAgrawala{RequestQueue: list.New(), Ack_Waited_Queue: make(map[string]memberlist.Node)}
}

func compaire(req1 RequesAccountAccess, req2 RequesAccountAccess) bool {
	if req1.Sender.Name == req2.Sender.Name {
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
		out += fmt.Sprintf("\t\t\t\t\t\t\tRequest Sender: \t%s\n\t\t\t\t\t\t\tSender Time: \t\t%d\n\t\t\t\t\t\t\tInterested Account: \t%s\n",
			temp.Sender.Name, temp.Req_Sender_Time.GetTime(), temp.Account.Account_Holder.Name,
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
