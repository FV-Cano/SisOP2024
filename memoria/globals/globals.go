package globals

import "sync"

// Global variables:
var InstruccionesProceso = make(map[int][]string)

// Global semaphores
var (
	// * Mutex
		InstructionsMutex 			sync.Mutex
	// * Binarios
		// Binary  					= make (chan bool, 1)
	// * Contadores
		// Contador 				= make (chan int, 10)
)

type T_ConfigMemory struct {
	Port              int    `json:"port"`
	Memory_size       int    `json:"memory_size"`
	Page_size         int    `json:"page_size"`
	Instructions_path string `json:"instructions_path"`
	Delay_response    int    `json:"delay_response"`
}

var Configmemory *T_ConfigMemory