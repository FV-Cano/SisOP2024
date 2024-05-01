package kernelutils

import (
	"log"
	"net/http"

	kernel_api "github.com/sisoputnfrba/tp-golang/kernel/API"
	"github.com/sisoputnfrba/tp-golang/kernel/globals"
	"github.com/sisoputnfrba/tp-golang/utils/pcb"
	"github.com/sisoputnfrba/tp-golang/utils/slice"
)

func Plan() {
	switch globals.Configkernel.Planning_algorithm {
	case "FIFO":
		log.Println("FIFO algorithm")
		// FIFO
	case "RR":
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
	-  [ ] Enviar CE a CPU
	-  [ ] Ejecutar Quantum -> // [ ] Mandar interrupción a CPU por endpoint interrupt si termina el quantum
	-  [ ] Esperar respuesta de CPU (Bloqueado)
	-  [ ] Recibir respuesta de CPU 
*/
func RR_Plan() {
	quantum := globals.Configkernel.Quantum
	var CurrentJob pcb.T_PCB

	for {
		CurrentJob = slice.Shift(&globals.STS)
		// TODO envío de CE a CPU

	}
}

/**
 * FIFO_Plan

	-  [x] Tomar proceso de lista de procesos
	-  [x] Enviar CE a CPU
	-  [ ] Esperar respuesta de CPU
	-  [ ] Recibir respuesta de CPU
*/
func FIFO_Plan(w http.ResponseWriter, r *http.Request) {
	// Proceso actual
	var CurrentJob pcb.T_PCB

	// Mientras haya procesos en la lista de procesos
	for globals.STS != nil {
		// 1. Tomo el primer proceso de la lista y lo quito de la misma
		CurrentJob = slice.Shift(&globals.STS)
		
		// 2. Cambio su estado a EXEC
		CurrentJob.State = "EXEC"

		// 3. Envío el PCB al CPU
		kernel_api.PCB_Send(CurrentJob)

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
		slice.Push(&globals.LTS, process)

	default:
		log.Fatalf("'%s' no es una razón de desalojo válida", process.EvictionReason)
	}
}