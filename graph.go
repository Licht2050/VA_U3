package main

import (
	"fmt"
	"sync"
)

type M struct {
	Ch chan interface{}
}

type Test struct {
	Ch map[string]M
}

func Sen(ch chan interface{}, str string, list map[string]int, wg *sync.WaitGroup) {

	test1 := <-ch
	list[str] = test1.(int)
	fmt.Println("Inside function: ", test1.(int))

	wg.Done()

}

func main() {

	ch := new(Test)
	ch.Ch = make(map[string]M)
	ch.Ch["new"] = M{Ch: make(chan interface{})}

	var wg sync.WaitGroup
	list := make(map[string]int)
	wg.Add(1)
	go Sen(ch.Ch["new"].Ch, "Hello", list, &wg)
	ch.Ch["new"].Ch <- 88
	wg.Wait()
	fmt.Println(list["Hello"])
}



case REQUEST_ACCOUNT_ACCESS:
	reqAA := RicartAndAgrawala.RequesAccountAccess{}
	err := json.Unmarshal(msg, &reqAA)
	_ = err

	fmt.Println("Recieved Req")

	var wg sync.WaitGroup
	sender := reqAA.Sender.Name
	if !sd.Snapshot.Incommint_Channel[sender].Closed {
		wg.Add(1)
		go Recieve_UsingChannel_Snapshot(&wg, sd.Snapshot.Incommint_Channel[sender].C, sd)
		sd.Snapshot.Incommint_Channel[sender].C <- reqAA
		wg.Wait()
	}
case ACCESS_ACCOUNT_ACKNOWLEDGE:
	ackAA := RicartAndAgrawala.AccountAccess_Acknowledge{}
	err := json.Unmarshal(msg, &ackAA)
	_ = err

	var wg sync.WaitGroup
	sender := ackAA.Ack_Sender.Name
	fmt.Println("Acknowledge ========================== received from :", sender)
	if !sd.Snapshot.Incommint_Channel[sender].Closed {
		wg.Add(1)

		go recieved_access_acknowledge(&wg, sd.Snapshot.Incommint_Channel[sender].C, sd)
		sd.Snapshot.Incommint_Channel[sender].C <- ackAA

		wg.Wait()
	}

case ACCOUNT_NEGOTIATION:

	ac_Negotiation := Bank.Account_Message{}
	err := json.Unmarshal(msg, &ac_Negotiation)
	_ = err

	var wg sync.WaitGroup
	sender := ac_Negotiation.Sender.Name
	if !sd.Snapshot.Incommint_Channel[sender].Closed {
		wg.Add(1)

		go account_Negotiation(&wg, sd.Snapshot.Incommint_Channel[sender].C, sd)
		sd.Snapshot.Incommint_Channel[sender].C <- ac_Negotiation

		wg.Wait()
	}

case FREE_LOCK:
	ac_operation_ack := Bank.Account_Operation_Ack{}
	err := json.Unmarshal(msg, &ac_operation_ack)
	_ = err

	fmt.Println("*********************************** Free_Lock Message is recieved ***************************")
	var wg sync.WaitGroup
	sender := ac_operation_ack.Sender.Name
	if !sd.Snapshot.Incommint_Channel[sender].Closed {
		wg.Add(1)

		go lock_handling(&wg, sd.Snapshot.Incommint_Channel[sender].C, sd)
		sd.Snapshot.Incommint_Channel[sender].C <- ac_operation_ack
		wg.Wait()
	}
