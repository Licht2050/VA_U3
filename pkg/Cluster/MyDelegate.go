package Cluster

import (
	"VAA_Uebung1/pkg/Cluster/Bank"
	"VAA_Uebung1/pkg/Election"
	"VAA_Uebung1/pkg/Graph"
	Lamportclock "VAA_Uebung1/pkg/LamportClock"
	"VAA_Uebung1/pkg/Neighbour"
	RicartAndAgrawala "VAA_Uebung1/pkg/Ricart_And_Agrawala"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/hashicorp/memberlist"
)

type SyncerDelegate struct {
	MasterNode       *memberlist.Node
	LocalNode        *memberlist.Node
	Node             *memberlist.Memberlist
	Neighbours       *Neighbour.NeighboursList
	NodesNeighbour   *Neighbour.NodesAndNeighbours
	NeighbourNum     *int
	NodeList         *Neighbour.NodesList
	Graph            *Graph.Graph
	RumorsList       *RumorsList
	ElectionExplorer *Election.ElectionExplorer
	EchoMessage      *Election.Echo
	RingMessage      *Election.RingMessage
	EchoCounter      *int
	//ElectionProtokol []*Election.ElectionExplorer
	// Broadcasts        *memberlist.TransmitLimitedQueue
	neighbourFilePath    *string
	BelievableRumorsRNum *int
	Local_Appointment    *Appointment
	Local_AP_Protocol    *Appointment_Protocol
	Cluster_AP_Protocol  *Appointment_Protocol
	Double_Counting1     *int
	Double_Counting2     *int
	Chanel               *chan Message
	//Account
	Account *Bank.Account
	//Ricart and Agrawala Algorithm
	LamportTime *Lamportclock.LamportClock
	R_A_Algrth  *RicartAndAgrawala.RicartAndAgrawala
}

//compare the incoming byte message to structs
func CompareJson(msg []byte) int {
	var receivedMsg map[string]interface{}
	err := json.Unmarshal(msg, &receivedMsg)
	_ = err

	// emptyValue := reflect.ValueOf(NeigbourGraph).Type()
	for key := range receivedMsg {

		if key == "neighbours" {
			return NEIGHBOUR_INFO_MESSAGE

		}
		if key == "m" {
			return ELECTION_EXPLORER_MESSAGE
		}
		if key == "coordinator" {
			return ELECTION_ECHO_MESSAGE
		}
		if key == "appointment_time" {
			return APPOINTMENT_MESSAGE
		}
		if key == "req_initiator_time" {
			return REQUEST_ACCOUNT_ACCESS
		}
		if key == "ack_status" {
			return ACCESS_ACCOUNT_ACKNOWLEDGE
		}
		if key == "access_seekers" {
			return ACCOUNT_NEGOTIATION
		}
		if key == "free-lock" {
			return FREE_LOCK
		}
	}
	return MESSAGE
}

// NotifyMsg is called when a user-data message is received.
// Care should be taken that this method does not block, since doing
// so would block the entire UDP packet receive loop. Additionally, the byte
// slice may be modified after the call returns, so it should be copied if needed.
func (sd *SyncerDelegate) NotifyMsg(msg []byte) {

	check := CompareJson(msg)

	switch check {
	case MESSAGE:
		//all incoming message as Message struct will be handeld
		Message_Handling(msg, sd)
	case ELECTION_EXPLORER_MESSAGE:
		//It will handle the explorer message
		election_explorer_message_handling(msg, sd)
	case ELECTION_ECHO_MESSAGE:
		//it handeld the echo message. echo message contains the answer for coordinator election message.
		echo_message_handling(msg, sd)
	case NEIGHBOUR_INFO_MESSAGE:
		//MasterNode recieve's neighbours from every node in the cluster afther any update occurred
		//afther recieved the message it will insert nodes and their neighbour's in to "NodesAndNeighbours" list
		neighbour_info_message_handling(sd, msg)

	case REQUEST_ACCOUNT_ACCESS:
		reqAA := RicartAndAgrawala.RequesAccountAccess{}
		err := json.Unmarshal(msg, &reqAA)
		_ = err

		fmt.Println("Recieved Req")

		sd.LamportTime.Update(reqAA.Req_Sender_Time.GetTime())
		reqAA.Req_Sender_Time = *sd.LamportTime

		//Request Send per flooding alg to neighbours
		//if the request is already recieved, then dont handle it
		reqAA.Sender = sd.LocalNode
		if Req_Send_To_Neighbours(reqAA, sd) {
			//Recieve Access to Critical Section Request
			request_account_access(reqAA, sd)
		}

	case ACCESS_ACCOUNT_ACKNOWLEDGE:

		recieved_access_acknowledge(msg, sd)

	case ACCOUNT_NEGOTIATION:

		account_Negotiation(msg, sd)

	case FREE_LOCK:
		fmt.Println("*********************************** Free_Lock Message is recieved ***************************")
		lock_handling(msg, sd)
	}
}

func lock_handling(msg []byte, sd *SyncerDelegate) {
	ac_operation_ack := Bank.Account_Operation_Ack{}
	err := json.Unmarshal(msg, &ac_operation_ack)
	_ = err
	sd.R_A_Algrth.AccInformation = true

	sd.LamportTime.Update(ac_operation_ack.Sender_Time.GetTime())

	ch := make(chan bool, 1)
	go freeLock(ch, sd)
	done := <-ch
	fmt.Println("*********************************** Before ***************************: ", done)
	if done {
		//Start the same process again.
		//Criticle Section Access Request
		// ch := make(chan int, 1)
		// go Account_Access_Channel(ch, sd)
		Account_Access_Channel(sd)
		fmt.Println("*********************************** Inside ***************************: ", done)
	}
}

func account_Negotiation(msg []byte, sd *SyncerDelegate) {
	ac_Negotiation := Bank.Account_Message{}
	err := json.Unmarshal(msg, &ac_Negotiation)
	_ = err

	sd.LamportTime.Update(ac_Negotiation.Sender_Time.GetTime())

	// acoountChannel := make(chan Bank.Account_Message, 1)
	Change_Account_Amount(ac_Negotiation, sd)
	// acoountChannel <- ac_Negotiation
}

func recieved_access_acknowledge(msg []byte, sd *SyncerDelegate) {
	ackAA := RicartAndAgrawala.AccountAccess_Acknowledge{}
	err := json.Unmarshal(msg, &ackAA)
	_ = err
	fmt.Printf("------------------------------Acknowledge Recieved from : \"%s Sender Time: %d\"------------------------------\n", ackAA.Ack_Sender.Name, ackAA.Ack_Sender_Time.GetTime())
	sd.LamportTime.Update(ackAA.Ack_Sender_Time.GetTime())
	acknowledge_Handling(ackAA, sd)
}

//Acknowledge handling Funcktion
func acknowledge_Handling(ack RicartAndAgrawala.AccountAccess_Acknowledge, sd *SyncerDelegate) {
	if sd.R_A_Algrth.Interested_Resource_isEmpty() {
		return
	}

	//check if the recieved acknowledge match to interested account
	if ack.ReqAA.Account_Holder.Name == sd.R_A_Algrth.Interested_Resource2.Account_Holder.Name {

		//check if the reciever could remove from the Ack_Waited_Queue
		if sd.R_A_Algrth.Remove_From_Ack_Waited_Queue(ack.Ack_Sender) {

			if len(sd.R_A_Algrth.Ack_Waited_Queue) == 0 {
				sd.R_A_Algrth.CurrentlyUsingR = sd.R_A_Algrth.Interested_Resource2

				//
				Clean_Lockal_Var(sd)

				//select randomly number between 0-100 to increase or decrease the balance
				rand_Percentage := rand.Intn(100)
				//Send Own account info to node intrested node and wait for his account info
				temp_Account_Message := Bank.Account_Message{Account_Holder: *sd.LocalNode, Sender: *sd.LocalNode,
					Balance: sd.Account.Balance, Percentage: rand_Percentage, Sender_Time: *sd.LamportTime}
				sd.SendMesgToMember(sd.R_A_Algrth.CurrentlyUsingR.Account_Holder, temp_Account_Message)

				fmt.Printf("The Acknowledge is succesfully recieved from all members now the account Info with randomly selected percentage \"%d\" is send to: \"%s\"\n",
					rand_Percentage, sd.R_A_Algrth.CurrentlyUsingR.Account_Holder.Name,
				)
			}
		}
	}

	time.Sleep(1 * time.Microsecond)

	//print the Waited_for_Ack_Queue
	for _, m := range sd.R_A_Algrth.Ack_Waited_Queue {
		fmt.Printf("The Acknowledge Reciever in Que: %s\n", m.Name)
	}

}

func request_account_access(reqAA RicartAndAgrawala.RequesAccountAccess, sd *SyncerDelegate) {

	fmt.Printf("------------------------------Request Recieved from : \"%s  Sender Time: %d\"------------------------------\n", reqAA.Initator, reqAA.ReqInitiator_Time.GetTime())
	account_access_handling(reqAA, sd)
}

//Send Acknowledge to Request for Account Access
func send_AcA_Acknowledge(req RicartAndAgrawala.RequesAccountAccess, sd *SyncerDelegate) {
	status := "Ok"
	ack := RicartAndAgrawala.New_AccountAccess_Acknowledge(*req.Account, *sd.LamportTime, *sd.LocalNode, status)
	fmt.Printf("\t\t\t\t\t\tAcknowledge Send To: \"%s\", for Access to \"%s's and %s's\" Accounts\n\n",
		req.Initator.Name, req.Account.Account_Holder.Name, req.Initator.Name)

	sd.SendMesgToMember(*req.Initator, ack)
}

//Request handling
func account_access_handling(req RicartAndAgrawala.RequesAccountAccess, sd *SyncerDelegate) {

	//if not interested and currently not using the critical resources, then send acknowledgement.
	sender_resource := Bank.Account{Account_Holder: *req.Initator}
	if !(sd.R_A_Algrth.Is_CurrentlyUsing(*req.Account, sender_resource) ||
		sd.R_A_Algrth.Is_Interested(*req.Account, sender_resource)) {
		//Send Ack
		fmt.Printf("\n\t\t\t\t\t\tNot Interested And Not Currently Using\n")
		send_AcA_Acknowledge(req, sd)

		return

	} else if sd.R_A_Algrth.Is_CurrentlyUsing(*req.Account, sender_resource) {
		//Add to Queue
		fmt.Printf("\n\t\t\t\t\t\tRequest Add To Queue, Because CurrentlyUsing\n")
		AddToQueue(sd, req)
		return

	} else if sd.R_A_Algrth.Is_Interested(*req.Account, sender_resource) {
		fmt.Printf("########################## requested Account is interesting to %s  and local interting account: %s\n",
			req.Account.Account_Holder.Name, sd.R_A_Algrth.Interested_Resource2.Account_Holder.Name,
		)

		if req.ReqInitiator_Time.GetTime() == sd.R_A_Algrth.Sending_Req_Time.GetTime() {
			//if both have the same time, then the one with smalle id will gain access to the critical section
			ifBoth_Have_Same_time(sd, req)
			return
			//if requested sender time is less than local time, then should queue the request
		} else if req.ReqInitiator_Time.GetTime() < sd.R_A_Algrth.Sending_Req_Time.GetTime() {

			fmt.Printf("\n\t\t\t\t\t\tAcknowledge Send, Because The Sender Time \"%d < %d\" Local Time\n",
				req.ReqInitiator_Time.GetTime(), sd.LamportTime.GetTime(),
			)
			send_AcA_Acknowledge(req, sd)

			return

		} else {
			fmt.Printf("\n\t\t\t\t\t\tRequest Add To Queue, Because The Sender Time \"%d > %d\" Local Time\n",
				req.ReqInitiator_Time.GetTime(), sd.LamportTime.GetTime(),
			)
			AddToQueue(sd, req)
		}
	}

}

func ifBoth_Have_Same_time(sd *SyncerDelegate, req RicartAndAgrawala.RequesAccountAccess) {
	localId, _ := strconv.Atoi(ParseNodeId(sd.LocalNode.Name))
	senderId, _ := strconv.Atoi(ParseNodeId(req.Initator.Name))

	if senderId < localId {
		fmt.Printf("\n\t\t\t\t\t\tAcknowledge Send, Because The Sender Time \"%d == %d\" Local Time But \"SenderId < LocalId\"\n",
			req.ReqInitiator_Time.GetTime(), sd.R_A_Algrth.Sending_Req_Time.GetTime(),
		)
		send_AcA_Acknowledge(req, sd)
	} else {
		fmt.Printf("\n\t\t\t\t\t\tRequest Add To Queue, Because The Sender Time \"%d == %d\" Local Time But \" LocalId < SenderId\"\n",
			req.ReqInitiator_Time.GetTime(), sd.R_A_Algrth.Sending_Req_Time.GetTime(),
		)
		AddToQueue(sd, req)
	}
}

func AddToQueue(sd *SyncerDelegate, req RicartAndAgrawala.RequesAccountAccess) {
	sd.R_A_Algrth.AddRequestToQueue(req)
	fmt.Println(sd.R_A_Algrth.Queue_toString())
}

func ChoosRand_Time_and_Neighbour(sd *SyncerDelegate) (int, map[string]memberlist.Node) {

	preferred_Time := ChoosePreferredTime(sd)
	temp_neighbour := make(map[string]memberlist.Node)

	//Anzahl P randomly ausgewaehlte neighbours wird auch durch zufall gemacht.
	rIndex := rand.Intn(len(sd.Neighbours.Neighbours))
	if rIndex == 0 {
		rIndex++
	}
	sd.Neighbours.ChooseRandomNeighbour(rIndex, temp_neighbour)

	return preferred_Time, temp_neighbour
}

func ChoosePreferredTime(sd *SyncerDelegate) int {
	rIndex := rand.Intn(len(sd.Local_AP_Protocol.Available_Appointments))
	return sd.Local_AP_Protocol.Available_Appointments[rIndex]
}

func neighbour_info_message_handling(sd *SyncerDelegate, msg []byte) {
	if sd.Node.LocalNode().Name == "Master" {

		var receivedMsg Neighbour.NeighboursList
		err := json.Unmarshal(msg, &receivedMsg)
		errorAnd_Msg := Error_And_Msg{Err: err, Text: "Could not encode the NeighboursInfo message"}
		Check(errorAnd_Msg)

		sd.NodesNeighbour.AddNodesAndNeighbours(receivedMsg)
	}
}

func echo_message_handling(msg []byte, sd *SyncerDelegate) {
	echo_message := new(Election.Echo)
	err := json.Unmarshal(msg, echo_message)
	_ = err

	if echo_message.Coordinator == sd.ElectionExplorer.M {
		//update the local echo sender list
		UpdateLocalEchoMessage(sd, echo_message)
		if sd.EchoMessage.EchoWaitedNum == sd.EchoMessage.EchoRecievedNum {
			if sd.LocalNode.Name != sd.ElectionExplorer.Initiator.Name {
				//When er von all seine Nachbarn bekommen hat, traegt er die sender auf die EchoSenderList ein.
				echo_message.EchoSenderList = sd.EchoMessage.EchoSenderList
				sd.SendMesgToMember(*sd.ElectionExplorer.Initiator, sd.EchoMessage)

			} else {
				fmt.Println("I am the Coordinator and i Recieved From : ", sd.EchoMessage.EchoSenderList)
				msg := Message{Msg: "Iam_the_Coordinator", Snder: sd.Node.LocalNode().Name}
				Inform_Cluster_Node(*sd.EchoMessage, *sd, msg)
				time.Sleep(2 * time.Second)

				// Inform_Appointment_Process_Starter(sd)
				// go TestChannel(*sd.Chanel, sd)

			}
		}
	}
}

func ChoseRondomFromMap(mp map[string]*memberlist.Node) string {
	temp_slice := []string{}
	for _, element := range mp {
		temp_slice = append(temp_slice, element.Name)
	}

	randIndex := rand.Intn(len(temp_slice))
	return temp_slice[randIndex]
}

func Inform_Cluster_Node(echo Election.Echo, sd SyncerDelegate, msg Message) {

	// apnt := Appointment{}
	// apnt.Create_Available_Time(11, *sd.LocalNode)
	for _, sender := range echo.EchoSenderList {
		fmt.Println("Echo Send To: ", sender.Name)
		sd.SendMesgToMember(*sender, msg)
		// sd.SendMesgToMember(*sender, apnt)
	}

}

func election_explorer_message_handling(msg []byte, sd *SyncerDelegate) {

	explorer := new(Election.ElectionExplorer)
	err := json.Unmarshal(msg, explorer)
	_ = err

	switch sd.CompaireElection(*explorer) {

	case RECIEVED_ExplorerID_EQUAL_THEN_Local_ID:
		If_The_Same_Explorer_Recieved(sd, explorer)
		//if recieved explorer is greater than local explorer id and the neigbour list is greater than 1
	case RECIEVED_ExplorerID_GREATER_THEN_Local_ID:

		switch sd.Check_If_Nod_is_Leaf() {
		case IS_NOT_LEAF_NODE:
			//th clean previous election process
			sd.EchoMessage.Clear()
			sd.ElectionExplorer.Clear()

			var tempExplorer Election.ElectionExplorer
			tempExplorer.Clear()
			tempExplorer = *explorer
			//init the local explorer struct with recieved to save all the info.
			sd.ElectionExplorer = &tempExplorer
			sd.ElectionExplorer.Add_RecievedFrom(*explorer.Initiator)
			explorer.Initiator = sd.LocalNode

			//hier wird bestimmt, dass der Node von alle seine Nachbarn echo Nachrichten erwartet
			//ausser der Sender
			sd.EchoMessage.EchoWaitedNum = len(sd.Neighbours.Neighbours) - 1

			sendExplorer(explorer, sd)
		case IS_LEAF_NODE:

			//send echo
			sd.ElectionExplorer.M = explorer.M
			sd.ElectionExplorer.Initiator = explorer.Initiator

			sd.EchoMessage.Coordinator = sd.ElectionExplorer.M
			sd.EchoMessage.EchoSender = *sd.LocalNode
			sd.EchoMessage.AddSender(*sd.LocalNode)

			sd.SendMesgToMember(*sd.ElectionExplorer.Initiator, sd.EchoMessage)
		}

	}
}

//If the recieves the same explorer.
//if the sender is not containt in the local explorer sender list, then it will save to the local sender list and
//waited for echo var will decrement.
//If the recieved counter and waited counter are equal && the node is not the elction initiator,
//then node start to send echo to node, which recieved for the first time.
func If_The_Same_Explorer_Recieved(sd *SyncerDelegate, explorer *Election.ElectionExplorer) {
	if !sd.ElectionExplorer.ContainsNodeInRecievedFrom(explorer.Initiator) {
		//hier wird bestimmt, dass der Node von diesem Node auch keine Echo erwartet,
		//weil er der selbe Nachricht erhalten hat.
		sd.ElectionExplorer.Add_RecievedFrom(*explorer.Initiator)
		sd.EchoMessage.EchoWaitedNum--

	}
	if sd.EchoMessage.EchoRecievedNum == sd.EchoMessage.EchoWaitedNum {
		if sd.ElectionExplorer.Initiator != sd.LocalNode {

			sd.EchoMessage.Coordinator = sd.ElectionExplorer.M
			sd.EchoMessage.EchoSender = *sd.LocalNode
			sd.EchoMessage.AddSender(*sd.LocalNode)

			sd.SendMesgToMember(*sd.ElectionExplorer.Initiator, sd.EchoMessage)
		}
	}
}

func (sd *SyncerDelegate) CompaireElection(eN Election.ElectionExplorer) int {

	if sd.ElectionExplorer.M < eN.M {
		return RECIEVED_ExplorerID_GREATER_THEN_Local_ID
	} else if sd.ElectionExplorer.M == eN.M {
		return RECIEVED_ExplorerID_EQUAL_THEN_Local_ID
	}
	return Default
}

func (sd *SyncerDelegate) Check_If_Nod_is_Leaf() int {
	if len(sd.Neighbours.Neighbours) == 1 {
		return IS_LEAF_NODE
	}
	if len(sd.Neighbours.Neighbours) > 1 {
		return IS_NOT_LEAF_NODE
	}
	return Default
}

//as message the following action will be handeld:
//body=leave -> the process will exist
//body=readNeighbour -> the node will read its neighbour from file
//body=Start_Election -> the node will start elcetion process
func Message_Handling(msg []byte, sd *SyncerDelegate) {
	var receivedMsg Message
	err := json.Unmarshal(msg, &receivedMsg)
	_ = err
	switch receivedMsg.Msg {
	case "Iam_the_Coordinator":
		fmt.Println("Message : ", receivedMsg.Msg, "Coordinator: ", receivedMsg.Snder)
		coordinator := SearchMemberbyName(receivedMsg.Snder, sd.Node)
		sd.ElectionExplorer.Initiator = coordinator
		sd.Local_AP_Protocol.Start_Value()
		sd.Local_Appointment.Clear()
	case "leave":
		//Leave will kill the process
		//and the node will remove from Cluster-Memberlist
		sd.Leave()

	case "readNeighbour":
		fmt.Println("Readed .dot file -----------------------------------------------")
		read_neighbours_from_dot_file(sd, receivedMsg)

	case "Start_Election":
		fmt.Println("I Have to Start Election Process Message to become coordinator +++++++++++++++++++++++++")
		start_election(sd)

	case "Start_Req_Critical_Section":
		//Criticle Section Access Request
		// ch := make(chan int, 1)
		// go Account_Access_Channel(ch, sd)
		Account_Access_Channel(sd)
	}

}

func start_election(sd *SyncerDelegate) {
	//th clean previous election process
	sd.EchoMessage.Clear()
	sd.ElectionExplorer.Clear()

	nodeId, _ := strconv.Atoi(ParseNodeId(sd.LocalNode.Name))
	tempExplorer := Election.NewElection(nodeId, *sd.LocalNode)

	sd.ElectionExplorer = tempExplorer

	sd.EchoMessage.EchoWaitedNum = len(sd.Neighbours.Neighbours)
	sendExplorer(tempExplorer, sd)
}

//clear the available neighbour list
//Read Graph from file
//add Nodes to Neighbourlist if there is a releastionship for this nod found
//Send the new neighbour list to MasterNode
func read_neighbours_from_dot_file(sd *SyncerDelegate, receivedMsg Message) {
	sd.neighbourFilePath = &receivedMsg.FilePath
	ReadNeighbourFromDot(sd)

	for _, ne := range sd.Neighbours.Neighbours {
		fmt.Println("Neighbours: ", ne.Name)
	}
}

//update the local echo sender list
func UpdateLocalEchoMessage(sd *SyncerDelegate, echo_message *Election.Echo) {
	sd.EchoMessage.EchoRecievedNum++
	if sd.ElectionExplorer.Initiator.Name != sd.LocalNode.Name {
		sd.EchoMessage.AddSender(*sd.LocalNode)
	}
	sd.EchoMessage.AddSender(echo_message.EchoSender)
	sd.EchoMessage.Coordinator = sd.ElectionExplorer.M
	sd.EchoMessage.EchoSender = *sd.LocalNode

	for _, sender := range echo_message.EchoSenderList {
		sd.EchoMessage.EchoSenderList[sender.Name] = sender
	}

}

//Echo wird nur an sender der ExplorerNachricht gesendet
func sendExplorer(explorer *Election.ElectionExplorer, sd *SyncerDelegate) {

	for _, neighbour := range sd.Neighbours.Neighbours {
		if neighbour.Name != sd.ElectionExplorer.Initiator.Name {
			sd.SendMesgToMember(neighbour, explorer)
		}
	}
}

func ReadNeighbourFromDot(sd *SyncerDelegate) {
	if sd.neighbourFilePath != nil && *sd.neighbourFilePath != "" {

		g := Graph.NewDiGraph()
		g.ParseFileToGraph(*sd.neighbourFilePath)

		// if AddNodesToNeighbourList(g, sd) {
		AddNodesToNeighbourList(g, sd)
		body, _ := json.Marshal(sd.Neighbours)
		sd.Node.SendBestEffort(sd.MasterNode, body)
		// }
	}
}

func AddNodesToNeighbourList(g *Graph.Graph, sd *SyncerDelegate) bool {
	sd.Neighbours.ClearNeighbours()
	for _, node := range g.Nodes {
		if node.Name == sd.LocalNode.Name {

			neighbours := g.GetEdges(node.Name)
			if len(neighbours.Nodes) > 0 {

				for _, neighbour := range neighbours.Nodes {
					//it add to the neighbour list if the node is a cluster memeber
					found_Node := SearchMemberbyName(neighbour.Name, sd.Node)
					if found_Node.Name == neighbour.Name {

						sd.Neighbours.AddNeighbour(*found_Node)
					}
				}
				return true
			}
			return false
		}
	}
	return false
}

func (d *SyncerDelegate) NotifyJoin(node *memberlist.Node) {

	d.NodeList.AddNode(node)

	log.Printf("notify join %s(%s:%d)", node.Name, node.Addr.To4().String(), node.Port)

}

func (d *SyncerDelegate) NotifyLeave(node *memberlist.Node) {

	log.Printf("notify leave %s(%s:%d)", node.Name, node.Addr.To4().String(), node.Port)

	if d.LocalNode.Name == "Master" {
		d.NodesNeighbour.RemoveNodesNeighbours(*node)

	}

	d.NodeList.RemoveNode(node)
	if d.Neighbours.Contains(node) {
		// d.Neighbours.UpdateNeighbourList(*d.NeighbourNum, *d.NodeList)
		d.Neighbours.RemoveNeighbour(*node)
		body, _ := json.Marshal(d.Neighbours)
		d.Node.SendBestEffort(d.MasterNode, body)
	}

}
func (d *SyncerDelegate) NotifyUpdate(node *memberlist.Node) {

	log.Printf("notify update %s(%s:%d)", node.Name, node.Addr.To4().String(), node.Port)
}

func BroadcastClusterMessage(ml *memberlist.Memberlist, msg *Message) {
	if msg == nil {
		errMessage := "Could not broadcast an empty message"
		log.Println(errMessage)
	}

	body, err := json.Marshal(msg)
	error_and_Message := Error_And_Msg{Err: err, Text: "Could not encode and broadcast the message"}
	Check(error_and_Message)

	for _, mem := range ml.Members() {
		// if mem.Name == "Node02" {
		ml.SendBestEffort(mem, body)
		// }
	}
}

// GetBroadcasts is called when user data messages can be broadcast.
// It can return a list of buffers to send. Each buffer should assume an
// overhead as provided with a limit on the total byte size allowed.
// The total byte size of the resulting data to send must not exceed
// the limit.
func (sd *SyncerDelegate) GetBroadcasts(overhead, limit int) [][]byte {
	return [][]byte{}
}

// func (sd *SyncerDelegate) QueueBroadcast(msg []byte) {
// 	sd.Broadcasts.QueueBroadcast(&MemberlistBroadcast{"test", msg})
// }

// LocalState is used for a TCP Push/Pull. This is sent to
// the remote side in addition to the membership information. Any
// data can be sent here. See MergeRemoteState as well. The `join`
// boolean indicates this is for a join instead of a push/pull.
func (sd *SyncerDelegate) LocalState(join bool) []byte {
	return nil
}

// MergeRemoteState is invoked after a TCP Push/Pull. This is the
// state received from the remote side and is the result of the
// remote side's LocalState call. The 'join'
// boolean indicates this is for a join instead of a push/pull.
func (sd *SyncerDelegate) MergeRemoteState(buf []byte, join bool) {

}

// NodeMeta is used to retrieve meta-data about the current node
// when broadcasting an alive message. It's length is limited to
// the given byte size. This metadata is available in the Node structure.
func (sd *SyncerDelegate) NodeMeta(limit int) []byte {
	return []byte{}
}

func (sd *SyncerDelegate) SendMesgToList(list map[string]memberlist.Node, value interface{}) {
	if len(list) <= 0 {
		return
	}

	body, err := json.Marshal(value)
	error_and_msg := Error_And_Msg{Err: err, Text: "Encode the Struct faild!"}
	Check(error_and_msg)

	for _, member := range list {
		sd.Node.SendBestEffort(&member, body)
	}
}

func (sd *SyncerDelegate) SendMesgToMember(node memberlist.Node, value interface{}) {
	body, err := json.Marshal(value)
	error_and_msg := Error_And_Msg{Err: err, Text: "Encode the Struct faild!"}
	Check(error_and_msg)

	sd.Node.SendBestEffort(&node, body)
	// fmt.Println("Msg Send To: ", node.Name)

}

func (sd *SyncerDelegate) Leave() {
	log.Println("Leave Node: ", sd.Node.LocalNode().Addr)
	sd.Node.Leave(1 * time.Second)
	sd.Node.Shutdown()
	os.Exit(1)
}
