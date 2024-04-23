package globals

// Global varibles:

import (
	"sync"

	"github.com/sisoputnfrba/tp-golang/utils/pcb"
)

var (
	NextPID 	uint32 		= 0
	Processes 	[]pcb.T_PCB
	LTS 		[]pcb.T_PCB
	STS 		[]pcb.T_PCB
	PidMutex 	sync.Mutex
)