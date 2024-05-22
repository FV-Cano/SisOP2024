package semaphores

import "sync"

// Global semaphores
var (
	// * Mutex
		PCBMutex 			sync.Mutex
	// * Binario
		//Binary  			= make (chan bool, 1)
	// * Contadores
		//Contador = make (chan int, 10)
)