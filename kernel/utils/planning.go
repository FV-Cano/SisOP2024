package kernelutils

import (
	"log"
	"time"

	kernel_api "github.com/sisoputnfrba/tp-golang/kernel/API"
	"github.com/sisoputnfrba/tp-golang/kernel/globals"
	"github.com/sisoputnfrba/tp-golang/utils/pcb"
	"github.com/sisoputnfrba/tp-golang/utils/slice"
)

var CurrentJob pcb.T_PCB
var quantum int

func Plan() {
	switch globals.Configkernel.Planning_algorithm {
	case "FIFO":
		log.Println("FIFO algorithm")
		for {
			FIFO_Plan()
		}
		// FIFO
	case "RR":
		quantum = globals.Configkernel.Quantum * int(time.Millisecond)
		log.Println("ROUND ROBIN algorithm")
		// RR
	case "VRR":
		log.Println("VIRTUAL ROUND ROBIN algorithm")
		// VRR
	default:
		log.Fatalf("Not a planning algorithm")
	}
}

/**
 * RR_Plan

	-  [x] Tomar proceso de lista de procesos
	-  [x] Enviar CE a CPU
	-  [x] Ejecutar Quantum -> // [ ] Mandar interrupción a CPU por endpoint interrupt si termina el quantum
	-  [ ] Esperar respuesta de CPU (Bloqueado)
	-  [ ] Recibir respuesta de CPU 
*/
type T_Quantum struct {
	TimeExpired chan bool
}

func RR_Plan() {
	CurrentJob = slice.Shift(&globals.STS)
	CurrentJob.State = "EXEC"
	kernel_api.PCB_Send(CurrentJob) // <-- Envía proceso y espera respuesta (la respuesta teóricamente actualiza la variable enviada como parámetro) // ? Bloquea?

	// Timer
	timer := &T_Quantum{TimeExpired: make(chan bool)}
	go startTimer(timer)

	// Esperar a que el proceso termine o sea desalojado por el timer
	select {
	case <-timer.TimeExpired:
		// Procesar desalojo por fin de quantum
	case <-pcb.Finished: // TODO: Actualizar canal con true cuando el proceso termine
		// Desalojo normal
	}
}

func startTimer(timer *T_Quantum) {
	quantumTime := time.Duration(quantum)
	time.Sleep(quantumTime)
	timer.TimeExpired <- true
}

func quantumInterrupt() {
	// TODO: Cómo funciona la interrupción? Si me comunico por el endpoint de dispatch que debería hacer? Puedo parar la ejecución de un proceso que está siendo llevada a cabo por un HTTP request?
}

/**
 * FIFO_Plan

	-  [x] Tomar proceso de lista de procesos
	-  [x] Enviar CE a CPU
	-  [ ] Recibir respuesta de CPU
	-  [ ] Agregar semáforos
*/
func FIFO_Plan() {
	// Proceso actual

	// Mientras haya procesos en la lista de procesos
	for globals.STS != nil {
		// 1. Tomo el primer proceso de la lista y lo quito de la misma
		CurrentJob = slice.Shift(&globals.STS)
		
		// 2. Cambio su estado a EXEC
		CurrentJob.State = "EXEC"

		// 3. Envío el PCB al CPU
		kernel_api.PCB_Send(CurrentJob)

		// ! Simulación de finalización de proceso - BORRAR
		CurrentJob.State = "EXIT"

		// 4. Manejo de desalojo
		EvictionManagement(CurrentJob)

		// 5. Logueo el estado del proceso
		log.Printf("Proceso %d: %s\n", CurrentJob.PID, CurrentJob.State)
	}
}

/**
 * EvictionManagement
	
	-  [ ] Implementar caso de desalojo por bloqueo
	-  [ ] Implementar caso de desalojo por timeout
	-  [x] Implementar caso de desalojo por finalización
**/
func EvictionManagement(process pcb.T_PCB) {
	switch process.EvictionReason {
	case "BLOCKED_IO":

	case "TIMEOUT":
		
	case "EXIT": 
		process.State = "FINISHED"
		// * VERIFICAR SI SE DEBE AGREGAR A LA LISTA LTS
		slice.Push(&globals.LTS, process)

	case "": 

	default:
		log.Fatalf("'%s' no es una razón de desalojo válida", process.EvictionReason)
	}
}