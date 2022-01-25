package Bank

import (
	"fmt"
	"log"
	"math/rand"

	"github.com/hashicorp/memberlist"
)

var NEGATIV_MAX = -500

type Account struct {
	Account_Holder memberlist.Node `json:"account_holder"`
	Balance        int             `json:"balance"`
}

type NodesList struct {
	Nodes []memberlist.Node
}

type Customers_Accounts struct {
	Customers_Accounts []*Account
}

func NewAccount(accontHoler memberlist.Node, balance int) *Account {
	return &Account{
		Account_Holder: accontHoler,
		Balance:        balance,
	}
}

func (ac *Account) Get_Balance() int {
	return ac.Balance
}

//This method add rondmly balance ammount between 0 to 10000
func (ac *Account) Add_Rand_Ammount(ammount_max int) {
	rIndex := rand.Intn(ammount_max)
	ac.Balance = rIndex
}

//this increases the account balance by a randomly generated percentage
//multiply with the given ammount
func (ac *Account) Increase_Balance(ammount int) {
	//chose randomly percentage
	rIndex := rand.Intn(100)
	percentage := rIndex / 100
	ac.Balance += ammount * percentage
}

//this decrease the account balance by a randomly generated percentage
//multiply with the given ammount
func (ac *Account) Decrease_Balance(ammount int) {
	if ac.Balance <= NEGATIV_MAX {
		log.Printf("The minimum negaativ boarder is reached!\n")
		return
	}

	//chose randomly percentage
	rIndex := rand.Intn(100)
	percentage := rIndex / 100
	ac.Balance -= ammount * percentage
}

//this will add accounts in a list
func (cAC *Customers_Accounts) AddAccount(ac Account) {
	if ac.Account_Holder.Name == "" {
		log.Println("Add Node and there Neighbours faild: param is empty")
		return
	}

	//if the account already exist, than just update the new balance
	for _, node := range cAC.Customers_Accounts {
		if node.Account_Holder.Name == ac.Account_Holder.Name {
			node.Balance = ac.Balance
			return
		}
	}
	cAC.Customers_Accounts = append(cAC.Customers_Accounts, &ac)
}

func (ac *Account) String() string {
	out := fmt.Sprintf("\tAccount Holder: %s\t account balance: %d Euro\n",
		ac.Account_Holder.Name, ac.Balance)
	return out
}

func (cAC *Customers_Accounts) String() string {
	out := " Customers accounts {\n"
	for _, account := range cAC.Customers_Accounts {
		out += account.String()
	}
	out += "}\n"

	return out
}
