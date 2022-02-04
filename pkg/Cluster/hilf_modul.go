package Cluster

import (
	"VAA_Uebung1/pkg/Cluster/Bank"
	"VAA_Uebung1/pkg/Exception"
	"VAA_Uebung1/pkg/Neighbour"
	RicartAndAgrawala "VAA_Uebung1/pkg/Ricart_And_Agrawala"
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/hashicorp/memberlist"
)

type Message struct {
	Msg      string    `json:"msg"`
	FilePath string    `json:"filepath"`
	Snder    string    `json:"sender"`
	Receiver string    `json:"receiver"`
	SendTime time.Time `json:"send_time"`
}

type MenuEnum int

const (
	Kill_Node = iota
	Kill_AllNodes
	Print_Cluster_Nodes
	Send_Rumor
	ParseNeighbourG_To_PNG
	Print_Cluster_Nodes_Neighbours
	Print_Cluster_Nodes_Neighbours_As_DiG
	Pars_DiG_To_Dot
	Read_DiG_From_DotFile
	Read_Neighbours_From_DotFile
	Pars_Random_DiGraph_PNG
	Pars_Random_UnDiGraph_PNG
	Ricart_And_Agrawala
	Default
)

const (
	MESSAGE = iota
	ELECTION_EXPLORER_MESSAGE
	ELECTION_ECHO_MESSAGE
	NEIGHBOUR_INFO_MESSAGE
	APPOINTMENT_MESSAGE
	REQUEST_ACCOUNT_ACCESS
	ACCESS_ACCOUNT_ACKNOWLEDGE
	ACCOUNT_NEGOTIATION
	FREE_LOCK
)

const (
	RECIEVED_ExplorerID_EQUAL_THEN_Local_ID = iota
	RECIEVED_ExplorerID_GREATER_THEN_Local_ID
	IS_LEAF_NODE
	IS_NOT_LEAF_NODE
)

func Menu() {

	fmt.Println("Menu")
	fmt.Println("0. Kill node")
	fmt.Println("1. Kill all nodes")
	fmt.Println("2. print cluster nodes")
	fmt.Println("3. start election process")
	fmt.Println("4. parse neigbour graph to PNG file")
	fmt.Println("5. print cluster Node's neigbours")
	fmt.Println("6. print Node's neigbours as directed graph")
	fmt.Println("7. Parse directed graph to .dot file")
	fmt.Println("8. Read graph from file")
	fmt.Println("9. Read neighbour from gv.dot file")
	fmt.Println("10. Parse Random genereted DirectGraph to PNG file")
	fmt.Println("11. Parse Random genereted UnDirectGraph to PNG and .dot file")
	fmt.Println("12. Ricart and Agrawala Algorithm")
}

type Error_And_Msg struct {
	Text string
	Err  error
}

func Check(err Error_And_Msg) {
	if err.Err != nil && len(err.Text) == 0 {
		Exception.ErrorHandler(err.Err)
	}
	if err.Err != nil && len(err.Text) > 0 {
		panic(err.Text + string(err.Err.Error()))
	}
}

func write_key_to_file(clusterKey string, bindPort int, fileName string) {
	f, err := os.Create(fileName)
	err_st := Error_And_Msg{Err: err}
	Check(err_st)
	defer f.Close()

	w := bufio.NewWriter(f)
	_, err = fmt.Fprintf(w, "%s\n", clusterKey)
	err_st.Err = err
	Check(err_st)

	_, err = fmt.Fprintf(w, "%d\n", bindPort)
	err_st.Err = err
	Check(err_st)

	w.Flush()
}

func AddClusterMemberToNodeList(ml *memberlist.Memberlist, nodeList *Neighbour.NodesList) {
	for _, node := range ml.Members() {
		if node.Name != ml.LocalNode().Name {
			nodeList.AddNode(node)
		}
	}
}

func UserInputInt(result *int) {
	var input string

	for {
		_, err := fmt.Scan(&input)
		_ = err
		*result, err = strconv.Atoi(input)
		if err != nil {
			fmt.Println("Enter a valid number")
		} else {
			break
		}
	}
}

func SearchMemberbyName(nodeName string, ml *memberlist.Memberlist) *memberlist.Node {
	foundNode := memberlist.Node{}
	for _, member := range ml.Members() {
		if member.Name == nodeName {
			foundNode = *member
			break
		}
	}
	return &foundNode
}

//url=https://stackoverflow.com/questions/51275036/split-a-string-of-one-side-letters-one-side-numbers
func ParseNodeId(s string) (numbers string) {
	// var l, n []rune
	var n []rune
	for _, r := range s {
		switch {
		// case r >= 'A' && r <= 'Z':
		// 	l = append(l, r)
		// case r >= 'a' && r <= 'z':
		// 	l = append(l, r)
		case r >= '0' && r <= '9':

			n = append(n, r)
		}
	}
	// return string(l), string(n)
	return string(n)
}

func PassSlicetoMap(ml memberlist.Memberlist, memberMap map[string]memberlist.Node) {
	for _, node := range ml.Members() {
		memberMap[node.Name] = *node
	}
}

func One_Random_Member(sd *SyncerDelegate) memberlist.Node {
	randMember := memberlist.Node{}

	for {
		rIndex := rand.Intn(len(sd.Node.Members()))
		if sd.Node.Members()[rIndex].Name != sd.LocalNode.Name {
			randMember = *sd.Node.Members()[rIndex]
			break
		}
	}

	return randMember
}

func Star_Account_Access_Process(sd *SyncerDelegate) {

	//Choose randomly a member to access his account
	randMember := One_Random_Member(sd)
	interestedAccount := Bank.Account{Account_Holder: randMember}
	sd.R_A_Algrth.Interested_Resource2 = &interestedAccount
	sd.R_A_Algrth.Own_Rsource = sd.Account

	fmt.Printf("\n\n===================================I Want to Access: \"%s's Account\"=========================================\n\n",
		sd.R_A_Algrth.Interested_Resource2.Account_Holder.Name)

	sd.LamportTime.Increment()
	req_accountAccess := RicartAndAgrawala.New_RequesAccountAccess(
		*sd.LamportTime, *sd.LocalNode, *sd.R_A_Algrth.Interested_Resource2,
	)

	for _, m := range sd.Node.Members() {
		if m.Name != sd.LocalNode.Name {
			sd.R_A_Algrth.Add_Ack_Waited_Queue(*m)
		}
	}

	//Request Send per flooding Alg
	Req_Send_Per_Flooding(*req_accountAccess, sd)
}

func Req_Send_Per_Flooding(req RicartAndAgrawala.RequesAccountAccess, sd *SyncerDelegate) {
	//
	for _, ne := range sd.Neighbours.Neighbours {
		fmt.Printf("++++++++++++++++++++++++++++++ Ricard Agrawala send to: %s and lamport time: %d++++++++++++++++++++\n",
			ne.Name, *sd.LamportTime,
		)

		sd.SendMesgToMember(ne, req)
	}
}

func Req_Send_To_Neighbours(req RicartAndAgrawala.RequesAccountAccess, sd *SyncerDelegate) {
	contain, _ := sd.R_A_Algrth.Is_Flooding_Queue_Contains(req)
	// fmt.Printf("\n\n---- Initiator: %s ----Sender : %s interested account: %s\n\n", req.Initator.Name, req.Sender.Name, req.Account.Account_Holder.Name)
	if !contain {

		sd.R_A_Algrth.AddTo_Flooding_Queue(req)
		for _, ne := range sd.Neighbours.Neighbours {
			if ne.Name != req.Sender.Name && ne.Name != req.Initator.Name {
				fmt.Printf("++++++++++++++++++++++++++++++ Req using flooding Alg send to: %s and local lamport time: %d++++++++++++++++++++\n",
					ne.Name, *sd.LamportTime,
				)
				sd.SendMesgToMember(ne, req)
			}
		}
	}
	// fmt.Println(sd.R_A_Algrth.Flooding_Queue_toString())
	//if the queue contain more then 10 request, then the remove the oldest
	if sd.R_A_Algrth.Flooding_Queue.Len() > 10 {

		fmt.Printf("\n\n\n----------------------------------------------------------  \"%s\" req as initiator is removed from flooding queue because is too old\n\n\n",
			sd.R_A_Algrth.Flooding_Queue.Front().Value.(RicartAndAgrawala.RequesAccountAccess).Initator.Name)
		// sd.R_A_Algrth.Remove_Old_Flooding_element()
	}
}

// func Account_Access_Channel(ch chan int, sd *SyncerDelegate)
func Account_Access_Channel(sd *SyncerDelegate) {
	sleepTime := rand.Intn(3)
	time.Sleep(time.Duration(sleepTime) * time.Second)

	fmt.Printf("--------------------------Start to send Account Access Request after \"%d Seconds\"--------------------------\n", sleepTime)
	Star_Account_Access_Process(sd)

}

//func Change_Account_Amount(ch chan Bank.Account_Message, sd *SyncerDelegate)
func Change_Account_Amount(received_ac_msg Bank.Account_Message, sd *SyncerDelegate) {

	// received_ac_msg := <-ch
	fmt.Printf("--------------------------Account Information recieved to start the increment or decrement Operation--------------------------\n\n")

	fmt.Printf("$$$$$$$$$$$$$$$$$$$$$$$$$$$$ Recieved Account Info $$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$4\n%s", received_ac_msg.String())

	fmt.Printf("$$$$$$$$$$$$$$$$$$$$$$$$$$$$ Own Account Info $$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$4\n%s", sd.Account.String())

	if sd.R_A_Algrth.CurrentlyUsingR == nil {

		fmt.Printf("--------------------------Own Account Information send to \"%s\" --------------------------\n", received_ac_msg.Sender.Name)
		own_account_msg := Bank.Account_Message{Account_Holder: *sd.LocalNode,
			Sender: *sd.LocalNode, Balance: sd.Account.Balance, Percentage: received_ac_msg.Percentage}

		sd.SendMesgToMember(received_ac_msg.Sender, own_account_msg)
		time.Sleep(1 * time.Second)

		operation(received_ac_msg, sd)
		fmt.Printf("$$$$$$$$$$$$$$$$$$$$$$$$$$$$ Target Account Info after Operation $$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$4\n%s", sd.Account.String())
		//afther operation send acknowledge to free the block
		ack := Bank.Account_Operation_Ack{Ack: true, Sender: *sd.LocalNode}
		sd.SendMesgToMember(received_ac_msg.Sender, ack)
		fmt.Printf("\n\n\n--------------------------After Operation Acknowledgement send to \"%s\" to free the lock\n\n\n", received_ac_msg.Sender.Name)

	} else if received_ac_msg.Account_Holder.Name == sd.R_A_Algrth.CurrentlyUsingR.Account_Holder.Name {
		operation(received_ac_msg, sd)
		fmt.Printf("\n\n$$$$$$$$$$$$$$$$$$$$$$$$$$$$ Access Seekers Account Info after Operation $$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$4\n%s", sd.Account.String())

		sd.R_A_Algrth.Operation_Ack = true

	}
}

func operation(receivedAc Bank.Account_Message, sd *SyncerDelegate) {
	if receivedAc.Balance >= sd.Account.Balance {
		sd.Account.Increase_Balance(receivedAc.Balance, receivedAc.Percentage)

		return
	}
	sd.Account.Decrease_Balance(sd.Account.Balance, receivedAc.Percentage)
}

func freeLock(ch chan bool, sd *SyncerDelegate) {

	i := 0
	for i < 10 {
		fmt.Println("***********************************  ***************************: ", sd.R_A_Algrth.AccInformation, sd.R_A_Algrth.Operation_Ack)
		if sd.R_A_Algrth.AccInformation && sd.R_A_Algrth.Operation_Ack {
			sd.R_A_Algrth.CurrentlyUsingR = nil
			Clean_Lockal_Var(sd)
			fmt.Printf("\n\t\t\t\t\t\tFree The Lock and Send Acknowledge to all nodes in the queue\n")
			for sd.R_A_Algrth.RequestQueue.Len() > 0 {
				temp := sd.R_A_Algrth.GetFirstQueueRequest()
				send_AcA_Acknowledge(temp, sd)
			}
			ch <- true
			break
		}
		time.Sleep(1 * time.Second)
		i++
	}

}

func Clean_Lockal_Var(sd *SyncerDelegate) {
	//clean all
	sd.R_A_Algrth.Interested_Resource2 = nil
	sd.R_A_Algrth.AccInformation = false
	sd.R_A_Algrth.Operation_Ack = false
}
