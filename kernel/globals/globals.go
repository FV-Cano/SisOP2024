package globals

// Global varibles:

import (
	"sync"

	"github.com/sisoputnfrba/tp-golang/utils/pcb"
)

var (
	NextPID 			uint32 		= 0
	Processes 			[]pcb.T_PCB
	LTS 				[]pcb.T_PCB
	STS 				[]pcb.T_PCB
)

// Global semaphores
var (
	PidMutex 	sync.Mutex
)

type T_ConfigKernel struct {
	Port 				int 		`json:"port"`
	IP_memory 			string 		`json:"ip_memory"`
	Port_memory 		int 		`json:"port_memory"`
	IP_cpu 				string 		`json:"ip_cpu"`
	Port_cpu 			int 		`json:"port_cpu"`
	Planning_algorithm 	string 		`json:"planning_algorithm"`
	Quantum 			int 		`json:"quantum"`
	Resources 			[]string 	`json:"resources"`
	Resource_instances 	[]int 		`json:"resource_instances"`
	Multiprogramming 	int 		`json:"multiprogramming"`
}

var Configkernel *T_ConfigKernel