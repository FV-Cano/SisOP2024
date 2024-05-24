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

var quantum int

func Plan() {
	switch globals.Configkernel.Planning_algorithm {
	case "FIFO":
		log.Println("FIFO algorithm")
		for {
			globals.PlanBinary <- true
			<- globals.MultiprogrammingCounter
			globals.JobExecBinary <- true
			log.Println("Planificandoooo")
			FIFO_Plan()
			<- globals.PlanBinary
			
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
  - RR_Plan

  - [x] Tomar proceso de lista de procesos

  - [x] Enviar CE a CPU

  - [x] Ejecutar Quantum -> // [x] Mandar interrupción a CPU por endpoint interrupt si termina el quantum

  - [ ] Esperar respuesta de CPU (Bloqueado)

  - [ ] Recibir respuesta de CPU
*/
func RR_Plan() {
	globals.CurrentJob = slice.Shift(&globals.STS)
	
	globals.ChangeState(&globals.CurrentJob, "EXEC")

	go startTimer()                 // ? Puedo arrancar el timer antes de enviar la pcb?
	kernel_api.PCB_Send() // <-- Envía proceso y espera respuesta (la respuesta teóricamente actualiza la variable enviada como parámetro)s

	// Esperar a que el proceso termine o sea desalojado por el timer
}

func startTimer() {
	quantumTime := time.Duration(quantum)
	time.Sleep(quantumTime)
	quantumInterrupt()
}

func quantumInterrupt() {
	pcb.EvictionFlag = true
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

	log.Printf("PID: %d - Desalojado por fin de quantum\n", globals.CurrentJob.PID)
}

/**
  - FIFO_Plan

  - [x] Tomar proceso de lista de procesos

  - [x] Enviar CE a CPU

  - [ ] Recibir respuesta de CPU

  - [ ] Agregar semáforos
*/
func FIFO_Plan() {

	// 1. Tomo el primer proceso de la lista y lo quito de la misma
	globals.CurrentJob = slice.Shift(&globals.STS)
	
	// 2. Cambio su estado a EXEC
	globals.ChangeState(&globals.CurrentJob, "EXEC")

	// 3. Envío el PCB al CPU
	kernel_api.PCB_Send()
	
	// 4. Manejo de desalojo
	EvictionManagement()
	<- globals.JobExecBinary
	
	// 5. Logueo el estado del proceso
	// log.Printf("Proceso %d: %s\n", globals.CurrentJob.PID, globals.CurrentJob.State)
}

/*
*

  - EvictionManagement

  - [ ] Implementar caso de desalojo por bloqueo

  - [ ] Implementar caso de desalojo por timeout

  - [x] Implementar caso de desalojo por finalización

*
*/
func EvictionManagement() {
	evictionReason := globals.CurrentJob.EvictionReason

	switch evictionReason {
	case "BLOCKED_IO":
		globals.ChangeState(&globals.CurrentJob, "BLOCKED")
		go kernel_api.SolicitarGenSleep(globals.CurrentJob)
		// Cabe la posibilidad de que este envío tenga que ser una goroutine paralela
	case "TIMEOUT":
		globals.ChangeState(&globals.CurrentJob, "READY")
		
		var pids []uint32
		for _, job := range globals.STS {
			pids = append(pids, job.PID)
		}
		log.Printf("Cola ready [%d]\n", pids)

	case "EXIT":
		globals.ChangeState(&globals.CurrentJob, "TERMINATED") // ? Cambiar a EXIT?

		// <- globals.MultiprogrammingCounter

		// * VERIFICAR SI SE DEBE AGREGAR A LA LISTA LTS
		// slice.Push(&globals.LTS, process)

		log.Printf("Finaliza el proceso %d - Motivo: %s\n", globals.CurrentJob.PID, evictionReason)

	case "":
		// ? Es necesario?
	default:
		log.Fatalf("'%s' no es una razón de desalojo válida", evictionReason)
	}
}