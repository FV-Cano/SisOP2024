package globals

import "sync"

var Configcpu *T_CPU
type T_CPU struct {
	Port               	int    `json:"port"`
	IP_memory          	string `json:"ip_memory"`
	Port_memory        	int    `json:"port_memory"`
	Number_felling_tlb 	int    `json:"number_felling_tlb"`
	Algorithm_tlb      	string `json:"algorithm_tlb"`
}

// Global semaphores
var (
	// * Mutex
		EvictionMutex 					sync.Mutex
		OperationMutex 					sync.Mutex
	// * Binario
		PlanBinary  					= make (chan bool, 1)
	// * Contadores
		MultiprogrammingCounter 		= make (chan int, 10)
)