package Cluster

import (
	"fmt"

	"github.com/hashicorp/memberlist"
)

type Appointment struct {
	Time    int             `json:"appointment_time"`
	Inviter memberlist.Node `json:"appointment_inviter"`
	Message Message         `json:"appointment_message"`
}

func (a *Appointment) ToString() string {
	out := "\n"
	out += fmt.Sprintf("\t%s Appointment: %d:00 o'clock\n", a.Inviter.Name, a.Time)

	return out

}
func (a *Appointment) Create_Available_Time(time int, node memberlist.Node) {
	a.Time = time
	a.Inviter = node
}

func (a *Appointment) Clear() Appointment {
	return Appointment{}
}

func (a *Appointment) Make_an_Appointment(recieved_time int, node memberlist.Node) {
	if recieved_time <= 0 {
		return
	}
	a.Time = (a.Time + recieved_time) / 2
	a.Inviter = node
}

type Appointment_Protocol struct {
	Self                     memberlist.Node
	A_Max                    int
	Recieved_Counter         int
	Waited_Counter           int
	Appointments             map[string]Appointment
	Available_Appointments   []int
	Rand_Selected_Neighbours map[string]memberlist.Node
}

func (ap *Appointment_Protocol) ToString() string {
	out := ""
	out += fmt.Sprintf("\t%s\n", ap.Self.Name)
	for _, ap := range ap.Appointments {
		out += ap.ToString()
	}

	return out
}

func CreateAppointmentProtocol(self memberlist.Node, aMax int, avialable_ap []int) *Appointment_Protocol {
	return &Appointment_Protocol{
		Self:                     self,
		A_Max:                    aMax,
		Recieved_Counter:         0,
		Waited_Counter:           0,
		Appointments:             map[string]Appointment{},
		Available_Appointments:   avialable_ap,
		Rand_Selected_Neighbours: make(map[string]memberlist.Node),
	}
}

func (ap *Appointment_Protocol) CopyRandSelected_Neighbour(neighbour_map map[string]memberlist.Node) {
	for _, neighbour := range neighbour_map {
		if _, ok := ap.Rand_Selected_Neighbours[neighbour.Name]; !ok {
			ap.Rand_Selected_Neighbours[neighbour.Name] = neighbour
		}
	}
}

func (ap *Appointment_Protocol) Add_Appointment(a Appointment) {
	if _, ok := ap.Appointments[a.Inviter.Name]; !ok {
		ap.Appointments[a.Inviter.Name] = a
	}
}

func (ap *Appointment_Protocol) Start_Value() {

	ap.A_Max = 8
	ap.Recieved_Counter = 0
	ap.Waited_Counter = 0

	ap.Appointments = map[string]Appointment{}
}
