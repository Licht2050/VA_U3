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
	fmt.Println("11. Parse Random genereted UnDirectGraph to PNG file")
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
	sd.R_And_Agra_Algrth.Interested_Resource2 = &interestedAccount
	sd.R_And_Agra_Algrth.Interested_Resource1 = sd.Account

	fmt.Printf("\n\n===================================Request to Access: \"%s's Account\"=========================================\n\n",
		sd.R_And_Agra_Algrth.Interested_Resource2.Account_Holder.Name)

	sd.LamportTime.Increment()
	req_accountAccess := RicartAndAgrawala.New_RequesAccountAccess(
		*sd.LamportTime, *sd.LocalNode, *sd.R_And_Agra_Algrth.Interested_Resource2,
	)

	for _, m := range sd.Node.Members() {
		if m.Name != sd.LocalNode.Name {
			sd.SendMesgToMember(*m, req_accountAccess)
			sd.R_And_Agra_Algrth.Add_Ack_Waited_Queue(*m)
			fmt.Printf("++++++++++++++++++++++++++++++Ricard Agrawala send to: %s and lamport time: %d++++++++++++++++++++\n",
				m.Name, *sd.LamportTime,
			)
		}
	}
}

func Account_Access_Channel(sd *SyncerDelegate) {

	if !sd.R_And_Agra_Algrth.AccInformation && !sd.R_And_Agra_Algrth.Operation_Ack {

		sleepTime := rand.Intn(3)

		time.Sleep(time.Duration(sleepTime) * time.Second)

		fmt.Printf("--------------------------Start to send Account Access Request after \"%d Seconds\"--------------------------\n", sleepTime)
		Star_Account_Access_Process(sd)
	}
}

func Change_Account_Amount(ch chan Bank.Account_Message, sd *SyncerDelegate) {

	recieved_ac_msg := <-ch
	fmt.Printf("--------------------------Account Information recieved to start the increment or decrement Operation--------------------------\n")

	fmt.Printf("$$$$$$$$$$$$$$$$$$$$$$$$$$$$Recieved Account Info$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$4\n%s", recieved_ac_msg.String())

	fmt.Printf("$$$$$$$$$$$$$$$$$$$$$$$$$$$$Own Account Info$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$4\n%s", sd.Account.String())

	if sd.R_And_Agra_Algrth.CurrentlyUsingR == nil {
		fmt.Printf("--------------------------Own Account Information send to \"%s\" --------------------------\n", recieved_ac_msg.Sender.Name)

		own_account_msg := Bank.Account_Message{Account_Holder: *sd.LocalNode,
			Sender: *sd.LocalNode, Balance: sd.Account.Balance, Percentage: recieved_ac_msg.Percentage}

		sd.SendMesgToMember(recieved_ac_msg.Sender, own_account_msg)

		operation(recieved_ac_msg, sd)
		fmt.Printf("$$$$$$$$$$$$$$$$$$$$$$$$$$$$Own Account Info afther Operation$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$4\n%s", sd.Account.String())
		//afther operation send acknowledge to free the block
		ack := Bank.Account_Operation_Ack{Ack: true, Sender: *sd.LocalNode}
		sd.SendMesgToMember(recieved_ac_msg.Sender, ack)
		fmt.Printf("--------------------------After Operation Acknowledgement send to \"%s\" to free the lock\n", recieved_ac_msg.Sender.Name)

	} else {

		operation(recieved_ac_msg, sd)
		fmt.Printf("$$$$$$$$$$$$$$$$$$$$$$$$$$$$Own Account Info afther Operation$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$4\n%s", sd.Account.String())
		sd.R_And_Agra_Algrth.AccInformation = true
		//Freigabe soll passieren
		// ch := make(chan Bank.Account_Operation_Ack, 1)
		// go freeLock(ch, sd)
	}
}

func operation(receivedAc Bank.Account_Message, sd *SyncerDelegate) {
	if receivedAc.Balance >= sd.Account.Balance {
		sd.Account.Increase_Balance(receivedAc.Balance, receivedAc.Percentage)

		return
	}

	sd.Account.Decrease_Balance(sd.Account.Balance, receivedAc.Percentage)
}

func freeLock(sd *SyncerDelegate) bool {
	if sd.R_And_Agra_Algrth.AccInformation && sd.R_And_Agra_Algrth.Operation_Ack {

		fmt.Printf("\n\t\t\t\t\t\tFree The Lock and Send Acknowledge to all nodes in the queue\n")
		for sd.R_And_Agra_Algrth.RequestQueue.Len() > 0 {
			temp := sd.R_And_Agra_Algrth.GetFirstQueueRequest()
			send_AcA_Acknowledge(temp, sd)
		}
		//clean all
		sd.R_And_Agra_Algrth.AccInformation = false
		sd.R_And_Agra_Algrth.Operation_Ack = false
		sd.R_And_Agra_Algrth.Interested_Resource2 = nil
		sd.R_And_Agra_Algrth.CurrentlyUsingR = nil

		return true
	}
	//Resource wieder freigeben
	return false
}
