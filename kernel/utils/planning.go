package kernelutils

import (
	"log"
	"sync"

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
	-  [ ] Enviar CE a CPU
	-  [ ] Esperar respuesta de CPU
	-  [ ] Recibir respuesta de CPU
*/
func FIFO_Plan(wg *sync.WaitGroup) {
	defer wg.Done()
	// Proceso actual
	var CurrentJob pcb.T_PCB

	// Mientras haya procesos en la lista de procesos
	for globals.STS != nil {
		// 1. Tomo el primer proceso de la lista y lo quito de la misma
		CurrentJob = slice.Shift(&globals.STS)
		
		// 2. Envío el PCB al CPU

		// 3. Espero y recibo la respuesta del CPU

		// 4. Actualizo el proceso

		// TODO Operar desalojo: función con switch para cada estado del proceso
		// 5. Agrego el proceso a la lista de procesos terminados
		if CurrentJob.State == "BLOCKED" {
			// Se lo mando a IO
		}

		// IO me devuelve el PCB con el estado actualizado
		if CurrentJob.State == "READY" {
			slice.Push(&globals.STS, CurrentJob)
		}

		// 6. Logueo el estado del proceso
		log.Printf("Proceso %d: %s\n", CurrentJob.PID, CurrentJob.State)

		wg.Add(1)
	}

	wg.Wait()
}