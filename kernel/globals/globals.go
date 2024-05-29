package globals

// Global variables:

import (
	"log"
	"sync"

	"github.com/sisoputnfrba/tp-golang/utils/device"
	"github.com/sisoputnfrba/tp-golang/utils/pcb"
)

var (
	NextPID 			uint32 = 0
	Processes 			[]pcb.T_PCB
	LTS 				[]pcb.T_PCB
	STS 				[]pcb.T_PCB
	Interfaces 			[]device.T_IOInterface
)

// Global semaphores
var (
	// * Mutex
		PidMutex 				sync.Mutex
		ProcessesMutex 			sync.Mutex
		STSMutex 				sync.Mutex
		LTSMutex 				sync.Mutex
		EnganiaPichangaMutex 	sync.Mutex
	// * Binarios
		PlanBinary  			= make (chan bool, 1)
		JobExecBinary			= make (chan bool, 1)
		PcbReceived				= make (chan bool, 1)
	// * Contadores
		// Chequea si hay procesos en la cola de listos, lo usamos en EvictionManagement y en ProcessInit
		MultiprogrammingCounter = make (chan int, 10)
)

// CurrentJob (kernel_api funcion PCB_Send) se lee

var CurrentJob pcb.T_PCB

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

func ChangeState(pcb *pcb.T_PCB, newState string) {
	ProcessesMutex.Lock()
	defer ProcessesMutex.Unlock()
	
	prevState := pcb.State
	pcb.State = newState
	log.Printf("PID: %d - Estado anterior: %s - Estado actual: %s \n", pcb.PID, prevState, pcb.State)
}
		
var BlockedJob_by_IO pcb.T_PCB