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
	resource "github.com/sisoputnfrba/tp-golang/kernel/resources"
	"github.com/sisoputnfrba/tp-golang/utils/pcb"
	"github.com/sisoputnfrba/tp-golang/utils/slice"
)

func LTS_Plan() {
	for {
		// Si la lista de jobs está vacía, esperar a que tenga al menos uno
		fmt.Println("La lista es: ", globals.LTS)
		fmt.Println("La lista tiene longitud: ", len(globals.LTS))

		if len(globals.LTS) == 0 {
			globals.EmptiedListMutex.Lock()
			//continue
		}
		auxJob := slice.Shift(&globals.LTS)
		//globals.MultiprogrammingCounter <- int(auxJob.PID)
		globals.ChangeState(&auxJob, "READY")
		slice.Push(&globals.STS, auxJob)
		//globals.ControlMutex.Unlock()
		globals.STSCounter <- int(auxJob.PID)

		// Los procesos en READY, EXEC y BLOCKED afectan al grado de multiprogramación
		globals.MultiprogrammingCounter <- int(auxJob.PID) // ! Lo cambiamos de linea porque tecnicamente debería ser después de ser agregado a la cola de listos
	}
}

func STS_Plan() {
	switch globals.Configkernel.Planning_algorithm {
	case "FIFO":
		
		log.Println("FIFO algorithm")
		for {
			//if len(globals.STS) == 0 {
				//globals.ControlMutex.Lock()
				//continue
			//}

			globals.PlanBinary <- true
			<- globals.STSCounter
			//log.Println("FIFO Planificandoooo")
			FIFO_Plan()
			<- globals.JobExecBinary
			<- globals.PlanBinary
		}
		
	case "RR":
		log.Println("ROUND ROBIN algorithm")
		quantum := uint32(globals.Configkernel.Quantum * int(time.Millisecond))
		for {
			globals.PlanBinary <- true
			<- globals.STSCounter
			//log.Println("RR Planificandoooo")
			RR_Plan(quantum)
			<- globals.JobExecBinary
			<- globals.PlanBinary
		}
		
	case "VRR":
		log.Println("VIRTUAL ROUND ROBIN algorithm")
		for {
			globals.PlanBinary <- true
			<- globals.STSCounter
			VRR_Plan()
			<- globals.JobExecBinary
			<- globals.PlanBinary
		}

	default:
		log.Fatalf("Not a planning algorithm")
	}
}

type T_Quantum struct {
	TimeExpired chan bool
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
  - RR_Plan
*/
func RR_Plan(quantum uint32) {
	globals.EnganiaPichangaMutex.Lock()
	// 1. Tomo el primer proceso de la lista y lo quito de la misma
	globals.CurrentJob = slice.Shift(&globals.STS)

	// 2. Cambio su estado a EXEC
	globals.ChangeState(&globals.CurrentJob, "EXEC")
	globals.EnganiaPichangaMutex.Unlock()

	// 3. Envío el PCB al CPU
	go startTimer(quantum)
	kernel_api.PCB_Send() // <-- Envía proceso y espera respuesta

	// 4. Esperar a que el proceso termine o sea desalojado por el timer
	<- globals.PcbReceived

	// fmt.Println("REGISTROS: ", globals.CurrentJob.CPU_reg)
	// fmt.Println("EVICTION REASON: ", globals.CurrentJob.EvictionReason)

	// 5. Manejo de desalojo
	EvictionManagement()
}

func VRR_Plan() {
	globals.EnganiaPichangaMutex.Lock()
	if len(globals.STS_Priority) == 0 {
		globals.CurrentJob = slice.Shift(&globals.STS)
	} else {
		globals.CurrentJob = slice.Shift(&globals.STS_Priority)
	}
	
	globals.ChangeState(&globals.CurrentJob, "EXEC")
	globals.EnganiaPichangaMutex.Unlock()

	timeBefore := time.Now()
	go startTimer(globals.CurrentJob.Quantum)
	kernel_api.PCB_Send()
	//timeAfter := time.Now()

	<- globals.PcbReceived
	timeAfter := time.Now() // ! Se cambió de lugar para que se tome el tiempo después de recibir el PCB

	diffTime := uint32(timeAfter.Sub(timeBefore))
	if diffTime < globals.CurrentJob.Quantum {
		globals.CurrentJob.Quantum -= diffTime
	} else {
		globals.CurrentJob.Quantum = uint32(globals.Configkernel.Quantum)
	}

	EvictionManagement()
}

func startTimer(quantum uint32) {
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
  - EvictionManagement

  - [ ] Implementar caso de desalojo por bloqueo
  - [x] Implementar caso de desalojo por timeout
  - [x] Implementar caso de desalojo por finalización
*
*/
func EvictionManagement() {
	evictionReason := globals.CurrentJob.EvictionReason

	switch evictionReason {
	case "BLOCKED_IO_GEN":
		globals.EnganiaPichangaMutex.Lock()
		globals.ChangeState(&globals.CurrentJob, "BLOCKED")
		slice.Push(&globals.Blocked, globals.CurrentJob)
		go func(){
			kernel_api.SolicitarGenSleep(globals.CurrentJob)
		}()
		globals.JobExecBinary <- true
		
	case "BLOCKED_IO_STD":
		globals.ChangeState(&globals.CurrentJob, "BLOCKED")
		
		<- globals.AvailablePcb

		globals.ChangeState(&globals.CurrentJob, "READY")
		globals.STS = append(globals.STS, globals.CurrentJob) // diferente en el caso de VRR
		globals.JobExecBinary <- true

	case "TIMEOUT":
		globals.ChangeState(&globals.CurrentJob, "READY")
		globals.STS = append(globals.STS, globals.CurrentJob)
		globals.JobExecBinary <- true

	case "EXIT":
		globals.ChangeState(&globals.CurrentJob, "TERMINATED")
		globals.JobExecBinary <- true
		<- globals.MultiprogrammingCounter
		log.Printf("Finaliza el proceso %d - Motivo: %s\n", globals.CurrentJob.PID, evictionReason)

	case "WAIT":
		if resource.Exists(globals.CurrentJob.RequestedResource) {
			resource.RequestConsumption(globals.CurrentJob.RequestedResource)
		}

	case "SIGNAL":
		if resource.Exists(globals.CurrentJob.RequestedResource) {
			resource.ReleaseConsumption(globals.CurrentJob.RequestedResource)
		}
	case "OUT_OF_MEMORY":
		globals.ChangeState(&globals.CurrentJob, "TERMINATED")
		globals.JobExecBinary <- true
		<- globals.MultiprogrammingCounter
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