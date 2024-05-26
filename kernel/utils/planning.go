package kernelutils

import (
	"bytes"
	"encoding/json"
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
			<- globals.MultiprogrammingCounter
			globals.PlanBinary <- true
			//log.Println("FIFO Planificandoooo")
			FIFO_Plan()
			<- globals.JobExecBinary
			<- globals.PlanBinary
			
		}
		// FIFO
	case "RR":
		quantum = globals.Configkernel.Quantum * int(time.Millisecond)
		log.Println("ROUND ROBIN algorithm")
		for {
			<- globals.MultiprogrammingCounter
			globals.PlanBinary <- true
			//log.Println("RR Planificandoooo")
			RR_Plan()
			<- globals.JobExecBinary
			<- globals.PlanBinary
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
*/
func RR_Plan() {
	// 1. Tomo el primer proceso de la lista y lo quito de la misma
	globals.CurrentJob = slice.Shift(&globals.STS)

	// 2. Cambio su estado a EXEC
	globals.ChangeState(&globals.CurrentJob, "EXEC")

	// 3. Envío el PCB al CPU
	go startTimer()
	kernel_api.PCB_Send() // <-- Envía proceso y espera respuesta

	// 4. Esperar a que el proceso termine o sea desalojado por el timer
	<- globals.PcbReceived

	// fmt.Println("REGISTROS: ", globals.CurrentJob.CPU_reg)
	// fmt.Println("EVICTION REASON: ", globals.CurrentJob.EvictionReason)

	// 5. Manejo de desalojo
	EvictionManagement()
}

func startTimer() {
	quantumTime := time.Duration(quantum)
	auxPcb := globals.CurrentJob
	time.Sleep(quantumTime)
	
	quantumInterrupt(auxPcb)
}

func quantumInterrupt(pcb pcb.T_PCB) {
	// Interrumpir proceso actual, response = OK message
	SendInterrupt("QUANTUM", pcb.PID)
	
	if globals.CurrentJob.EvictionReason == "TIMEOUT" {
		log.Printf("PID: %d - Desalojado por fin de quantum\n", globals.CurrentJob.PID)
	}
}

/**
  - FIFO_Plan
*/
func FIFO_Plan() {
	// 1. Tomo el primer proceso de la lista y lo quito de la misma

	globals.CurrentJob = slice.Shift(&globals.STS)
	
	// 2. Cambio su estado a EXEC
	globals.ChangeState(&globals.CurrentJob, "EXEC")

	// 3. Envío el PCB al CPU
	kernel_api.PCB_Send()
	
	<- globals.PcbReceived

	// 4. Manejo de desalojo
	EvictionManagement()
}

/**
  - EvictionManagement

  - [ ] Implementar caso de desalojo por bloqueo
  - [x] Implementar caso de desalojo por timeout
  - [x] Implementar caso de desalojo por finalización
*
*/
func EvictionManagement() {
	evictionReason := globals.CurrentJob.EvictionReason

	switch evictionReason {
	case "BLOCKED_IO":
		globals.ChangeState(&globals.CurrentJob, "BLOCKED")
		enganiaPichanga := globals.CurrentJob
		go func(){
			kernel_api.SolicitarGenSleep(enganiaPichanga)
			globals.MultiprogrammingCounter <- int(globals.CurrentJob.PID)
		}()
		globals.JobExecBinary <- true
		
	case "TIMEOUT":
		globals.ChangeState(&globals.CurrentJob, "READY")
		globals.STS = append(globals.STS, globals.CurrentJob)
		globals.MultiprogrammingCounter <- int(globals.CurrentJob.PID)
		globals.JobExecBinary <- true
		
		/* var pids []uint32
		for _, job := range globals.STS {
			pids = append(pids, job.PID)
		}
		log.Printf("Cola ready %d\n", pids) */

	case "NOT_FOUND_IO":
		globals.ChangeState(&globals.CurrentJob, "READY")
		globals.STS = append(globals.STS, globals.CurrentJob)
		globals.MultiprogrammingCounter <- int(globals.CurrentJob.PID)
		globals.JobExecBinary <- true

	case "EXIT":
		globals.ChangeState(&globals.CurrentJob, "TERMINATED") // ? Cambiar a EXIT?
		globals.JobExecBinary <- true

		// * VERIFICAR SI SE DEBE AGREGAR A LA LISTA LTS
		// slice.Push(&globals.LTS, process)

		log.Printf("Finaliza el proceso %d - Motivo: %s\n", globals.CurrentJob.PID, evictionReason)

	default:
		log.Fatalf("'%s' no es una razón de desalojo válida", evictionReason)
	}	
}

type InterruptionRequest struct {
	InterruptionReason string `json:"InterruptionReason"`
	Pid uint32 `json:"pid"`
}

func SendInterrupt(reason string, pid uint32) {
	url := fmt.Sprintf("http://%s:%d/interrupt", globals.Configkernel.IP_cpu, globals.Configkernel.Port_cpu)

	bodyInt, err := json.Marshal(InterruptionRequest{
		InterruptionReason: reason,
		Pid: pid,
	})
	if err != nil {
		return
	}
	
	enviarInterrupcion, err := http.NewRequest("POST", url, bytes.NewBuffer(bodyInt))
	if err != nil {
		log.Fatalf("POST request failed (No se puede enviar interrupción): %v", err)
	}
	
	cliente := &http.Client{}
	enviarInterrupcion.Header.Set("Content-Type", "application/json")
	recibirRta, err := cliente.Do(enviarInterrupcion)
	if (err != nil || recibirRta.StatusCode != http.StatusOK) {
		log.Fatal("Error al interrupir proceso", err)
	}
}