package kernelutils

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
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
		for {
			RR_Plan()
		}
		// RR
	case "VRR":
		log.Println("VIRTUAL ROUND ROBIN algorithm")
		// VRR
	default:
		log.Fatalf("Not a planning algorithm")
	}
}

type T_Quantum struct {
	TimeExpired chan bool
}

/**
 * RR_Plan

	-  [x] Tomar proceso de lista de procesos
	-  [x] Enviar CE a CPU
	-  [x] Ejecutar Quantum -> // [x] Mandar interrupción a CPU por endpoint interrupt si termina el quantum
	-  [ ] Esperar respuesta de CPU (Bloqueado)
	-  [ ] Recibir respuesta de CPU 
*/
func RR_Plan() {
	CurrentJob = slice.Shift(&globals.STS)
	CurrentJob.State = "EXEC"
	go startTimer() // ? Puedo arrancar el timer antes de enviar la pcb?
	kernel_api.PCB_Send(CurrentJob) // <-- Envía proceso y espera respuesta (la respuesta teóricamente actualiza la variable enviada como parámetro)s

	// Esperar a que el proceso termine o sea desalojado por el timer
}

func startTimer() {
	quantumTime := time.Duration(quantum)
	time.Sleep(quantumTime)
	quantumInterrupt()
}

func quantumInterrupt() {
	pcb.InterruptFlag = true
	interruptionCode := pcb.QUANTUM

	// Interrumpir proceso actual, response = OK message
	url := fmt.Sprintf("http://%s:%d/interrupt", globals.Configkernel.IP_cpu, globals.Configkernel.Port_cpu)
	
	// Json payload
	jsonStr := fmt.Sprintf("interruption code: %d", interruptionCode)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(jsonStr)))
	if err != nil {
		log.Fatalf("Error al crear request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending HTTP request: ", err)
	}
	defer resp.Body.Close()

	fmt.Println("HTTP response status: ", resp.Status)
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
		// ? Es necesario?
	default:
		log.Fatalf("'%s' no es una razón de desalojo válida", process.EvictionReason)
	}
}