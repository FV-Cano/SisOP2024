package semaphores

import "sync"

// Global semaphores
var (
	// * Mutex
		PCBMutex 			sync.Mutex
		CurrentJobMutex 	sync.Mutex
	// * Binario
		EndProcessBinary  	= make (chan bool, 1)
		ExitChan 			= make (chan bool, 1)
		TimerChan 			= make (chan bool, 1)
	// * Contadores
		//Contador 			= make (chan int, 10)
)