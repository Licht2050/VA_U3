package Cluster

import (
	"VAA_Uebung1/pkg/Cluster/Bank"
	"VAA_Uebung1/pkg/Election"
	Lamportclock "VAA_Uebung1/pkg/LamportClock"
	"VAA_Uebung1/pkg/Neighbour"
	RicartAndAgrawala "VAA_Uebung1/pkg/Ricart_And_Agrawala"
	"VAA_Uebung1/pkg/Snapshot"
	"encoding/base64"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/hashicorp/memberlist"
)

func JoinCluster(nodeName, bindIP, bindPort, httpPort, clusterKey, knownIP string) {

	//config
	config := memberlist.DefaultLocalConfig()
	bIP, _ := strconv.Atoi(bindPort)
	config.BindPort = bIP
	// config.AdvertisePort = 8000
	config.BindAddr = bindIP
	config.Name = nodeName
	config.SecretKey, _ = base64.StdEncoding.DecodeString(clusterKey)
	//dadurch wird alle logging vo Memberlist ausgeschaltet
	config.LogOutput = ioutil.Discard

	//create a memberlist
	ml, err := memberlist.Create(config)
	err_st := Error_And_Msg{Err: err}
	Check(err_st)

	//join the Cluster with the help of passt param: clusterIP, clusterKey, clusterPort
	_, err = ml.Join([]string{knownIP})
	//if err not nil print the following Text
	err_st.Err = err
	err_st.Text = "Failed to join cluster: "
	Check(err_st)

	log.Printf("Joined the cluster")

	// var broadcast *memberlist.TransmitLimitedQueue

	masterNode := SearchMemberbyName("Master", ml)

	//neibour variable
	neigbours := Neighbour.NewNeighbourList()
	neigbours.Node = *ml.LocalNode()
	//It defines the number of neighbours that each node can have
	neigbourNum := 3
	nodeList := new(Neighbour.NodesList)

	//Nodelist could be accessable from eventdelegate
	AddClusterMemberToNodeList(ml, nodeList)
	// neigbours.UpdateNeighbourList(neigbourNum, *nodeList)
	// body, _ := json.Marshal(neigbours)
	// ml.SendBestEffort(masterNode, body)

	//rumors var
	rumors_list := NewRumorsList()
	blievableRRNum := 2

	//ElectionExplorer
	// nodeId, _ := strconv.Atoi(ParseNodeId(ml.LocalNode().Name))
	electionExplorer := Election.NewElection(-1, *ml.LocalNode())
	echoMessage := new(Election.Echo)
	ringMessage := new(Election.RingMessage)
	echoCounter := new(int)
	echoMessage.Clear()

	//Appointment
	appointment := Appointment{}

	//Ricart And Agrawala Algorithm
	lamportClock := Lamportclock.NewLamportClock()
	r_and_agra_algrthm := RicartAndAgrawala.New_R_and_A_Algrthm()

	appointment_Protocol := CreateAppointmentProtocol(*ml.LocalNode(), A_MAX, AVAILABLE_Appointment)
	cluster_appointment_Protocol := CreateAppointmentProtocol(*ml.LocalNode(), 0, AVAILABLE_Appointment)

	doubleCounting1 := 0
	doubleCounting2 := 0

	test := make(chan Message, 1)

	bank_account := create_Account(ml.LocalNode())

	snapshot := Snapshot.NewChandy_Lamport()

	sm_ac := Bank.NewSumAccount()

	sd := &SyncerDelegate{
		Node: ml, Neighbours: neigbours, NeighbourNum: &neigbourNum,
		NodeList: nodeList, MasterNode: masterNode,
		LocalNode:            ml.LocalNode(),
		RumorsList:           rumors_list,
		BelievableRumorsRNum: &blievableRRNum,
		ElectionExplorer:     electionExplorer,
		EchoMessage:          echoMessage,
		RingMessage:          ringMessage,
		EchoCounter:          echoCounter,
		Local_Appointment:    &appointment,
		Local_AP_Protocol:    appointment_Protocol,
		Double_Counting1:     &doubleCounting1,
		Double_Counting2:     &doubleCounting2,
		Chanel:               &test,
		Cluster_AP_Protocol:  cluster_appointment_Protocol,
		Account:              bank_account,
		LamportTime:          lamportClock,
		R_A_Algrth:           r_and_agra_algrthm,
		Snapshot:             snapshot,
		Sum_Account:          sm_ac,
	}

	sd.R_A_Algrth.Own_Rsource = &Bank.Account{Account_Holder: *sd.LocalNode}

	config.Delegate = sd
	config.Events = sd

	node := Node{
		Memberlist: ml,
		Neigbour:   neigbours,
	}

	http.HandleFunc("/", node.handler)

	go func() {
		http.ListenAndServe(":"+httpPort, nil)
	}()

	log.Printf("webserver is up. URL: http://%s:%s/ \n", bindIP, httpPort)

	msg := Message{Msg: "I am a new Member", Snder: ml.LocalNode().Name, SendTime: *sd.LamportTime}
	time.Sleep(time.Second * 2)
	BroadcastClusterMessage(ml, &msg)

	incomingSigs := make(chan os.Signal, 1)
	signal.Notify(incomingSigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, os.Interrupt)

	select {
	case <-incomingSigs:
		log.Println("test")

		if err := ml.Leave(time.Second * 5); err != nil {
			err_st.Err = err
			Check(err_st)
		}
	case <-incomingSigs:

	}

}
