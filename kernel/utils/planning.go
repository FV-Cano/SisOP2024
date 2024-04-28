package kernelutils

import (
	"log"

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