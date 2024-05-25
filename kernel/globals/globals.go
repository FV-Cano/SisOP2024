package globals

// Global variables:

import (
	"log"
	"sync"

	"github.com/sisoputnfrba/tp-golang/utils/pcb"
)

var (
	NextPID 			uint32 		= 0
	Processes 			= []pcb.T_PCB{
		{PID: 90, PC: 0, Quantum: 0, CPU_reg: map[string]interface{}{"AX": uint8(0), "BX": uint8(0), "CX": uint8(0), "DX": uint8(0), "EAX": uint32(0), "EBX": uint32(0), "ECX": uint32(0), "EDX": uint32(0),
						}, State: "READY", EvictionReason: ""},
		{PID: 91, PC: 0, Quantum: 0, CPU_reg: map[string]interface{}{"AX": uint8(0), "BX": uint8(0), "CX": uint8(0), "DX": uint8(0), "EAX": uint32(0), "EBX": uint32(0), "ECX": uint32(0), "EDX": uint32(0),
						}, State: "BLOCKED", EvictionReason: ""},
		{PID: 92, PC: 0, Quantum: 0, CPU_reg: map[string]interface{}{"AX": uint8(0),"BX": uint8(0),
		"CX": uint8(0),"DX": uint8(0),"EAX": uint32(0),"EBX": uint32(0),"ECX": uint32(0),"EDX": uint32(0),
						}, State: "READY", EvictionReason: ""},
		{PID: 93, PC: 0, Quantum: 0, CPU_reg: map[string]interface{}{"AX": uint8(0), "BX": uint8(0), "CX": uint8(0), "DX": uint8(0), "EAX": uint32(0), "EBX": uint32(0), "ECX": uint32(0), "EDX": uint32(0),
						}, State: "EXIT", EvictionReason: ""},
		{PID: 94, PC: 0, Quantum: 0, CPU_reg: map[string]interface{}{"AX": uint8(0),"BX": uint8(0),"CX": uint8(0),"DX": uint8(0),"EAX": uint32(0),"EBX": uint32(0),"ECX": uint32(0),"EDX": uint32(0),
						}, State: "READY", EvictionReason: ""},
	}
	LTS 				[]pcb.T_PCB
	STS 				[]pcb.T_PCB
)

// Global semaphores
var (
	// * Mutex
		PidMutex 				sync.Mutex
		ProcessesMutex 			sync.Mutex
		STSMutex 				sync.Mutex
		LTSMutex 				sync.Mutex
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
		
type IOInterface struct {
	InterfaceType string `json:"interfaceType"`
	InterfacePort int    `json:"interfacePort"`
}

var InterfaceIP = "127.0.0.1"
	
var IO_Interface IOInterface
var BlockedJob_by_IO pcb.T_PCB