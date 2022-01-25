package Cluster

import (
	"VAA_Uebung1/pkg/Exception"
	"VAA_Uebung1/pkg/Neighbour"
	"bufio"
	"fmt"
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
