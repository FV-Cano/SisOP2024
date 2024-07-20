package kernelutils

import (
	"fmt"
	"log"
	"time"

	kernel_api "github.com/sisoputnfrba/tp-golang/kernel/API"
	"github.com/sisoputnfrba/tp-golang/kernel/globals"
	resource "github.com/sisoputnfrba/tp-golang/kernel/resources"
	"github.com/sisoputnfrba/tp-golang/utils/pcb"
	"github.com/sisoputnfrba/tp-golang/utils/slice"
)

func LTS_Plan() {
	for {

		if globals.PlanningState == "STOPPED" {
			// fmt.Print("ESTOY PARADO MI REY\n")
			// Si funciona es la engaña pichanga máxima
			globals.LTSPlanBinary <- true
			<- globals.LTSPlanBinary
			continue
		}

		// Si la lista de jobs está vacía, esperar a que tenga al menos uno
		if len(globals.LTS) == 0 {
			globals.EmptiedList <- true
			continue
		}

		log.Println("Comienza el LTS")
		log.Println("La lista es: ", globals.LTS)
		log.Println("La lista tiene longitud: ", len(globals.LTS))

		globals.LTSMutex.Lock()
		auxJob := slice.Shift(&globals.LTS)
		globals.LTSMutex.Unlock()
		if auxJob.PID != 0 {
			// Los procesos en READY, EXEC y BLOCKED afectan al grado de multiprogramación
			globals.MultiprogrammingCounter <- int(auxJob.PID)
			globals.ChangeState(&auxJob, "READY")
			slice.Push(&globals.STS, auxJob)
			globals.STSCounter <- int(auxJob.PID)
		}
	}
}

func STS_Plan() {
	switch globals.Configkernel.Planning_algorithm {
	case "FIFO":
		log.Println("FIFO algorithm")
		for {
			if globals.PlanningState == "STOPPED" {
				globals.STSPlanBinary <- true
				<- globals.STSPlanBinary
				continue
			}

			<-globals.STSCounter
			//log.Println("FIFO Planificandoooo")
			//globals.JobExecBinary <- true
			FIFO_Plan()
		}

	case "RR":
		log.Println("ROUND ROBIN algorithm")
		quantum := uint32(time.Duration(globals.Configkernel.Quantum) * time.Millisecond)	// TODO change
		for {
			if globals.PlanningState == "STOPPED" {
				globals.STSPlanBinary <- true
				<- globals.STSPlanBinary
				continue
			}
			<-globals.STSCounter
			//globals.JobExecBinary <- true
			RR_Plan(quantum)
		}

	case "VRR":
		log.Println("VIRTUAL ROUND ROBIN algorithm")
		for {
			if globals.PlanningState == "STOPPED" {
				globals.STSPlanBinary <- true
				<- globals.STSPlanBinary
				continue
			}

			<-globals.STSCounter
			//log.Println("VRR Planificandoooo")
			//globals.JobExecBinary <- true
			VRR_Plan()
		}

	default:
		log.Fatalf("Not a planning algorithm")
	}
}

type T_Quantum struct {
	TimeExpired chan bool
}

/*
*
  - FIFO_Plan
*/
func FIFO_Plan() {
	// 1. Tomo el primer proceso de la lista y lo quito de la misma
	globals.CurrentJob = slice.Shift(&globals.STS)

	// 2. Cambio su estado a EXEC
	globals.ChangeState(&globals.CurrentJob, "EXEC")

	// 3. Envío el PCB al CPU
	kernel_api.PCB_Send()

	<-globals.PcbReceived

	// 4. Manejo de desalojo
	EvictionManagement()
}

/**
  - RR_Plan
*/
/*
func RR_Plan(quantum uint32) {
	//globals.EnganiaPichangaMutex.Lock()
	// 1. Tomo el primer proceso de la lista y lo quito de la misma
	globals.CurrentJob = slice.Shift(&globals.STS)

	// 2. Cambio su estado a EXEC
	globals.ChangeState(&globals.CurrentJob, "EXEC")
	//globals.EnganiaPichangaMutex.Unlock()

	// 3. Envío el PCB al CPU
	kernel_api.PCB_Send() // <-- Envía proceso y espera respuesta
	go startTimer(quantum)

	// 4. Esperar a que el proceso termine o sea desalojado por el timer
	<- globals.PcbReceived

	// 5. Manejo de desalojo
	EvictionManagement()
}
*/

func RR_Plan(quantum uint32) {
	globals.EnganiaPichangaMutex.Lock() // Paso 1: Bloquear el mutex para asegurar la exclusión mutua
	
	globals.CurrentJob = slice.Shift(&globals.STS)   // Paso 2: Tomar el primer proceso
	globals.ChangeState(&globals.CurrentJob, "EXEC") // Cambiar estado a EXEC
	globals.EnganiaPichangaMutex.Unlock()            // Desbloquear el mutex después de modificar las variables compartidas

	go startTimer(quantum)
	kernel_api.PCB_Send()                                             // Paso 3: Enviar el PCB al CPU
	//timer := time.NewTimer(time.Duration(quantum)) // Paso 4: Iniciar el temporizador

	select {
	case <-globals.PcbReceived:

	}
	/* select {
	case <-globals.PcbReceived: // Paso 5: Esperar a que el proceso termine
		// El proceso ha terminado, manejar la finalización
	case <-timer.C: // El temporizador ha expirado antes de que el proceso termine
		globals.EnganiaPichangaMutex.Lock()
		globals.STS = append(globals.STS, globals.CurrentJob) // Paso 6: Agregar el proceso al final de la lista TODO:
		globals.ChangeState(&globals.CurrentJob, "READY")     // Cambiar el estado a READY
		globals.EnganiaPichangaMutex.Unlock()
	} */

	EvictionManagement() // Paso 7: Manejar el desalojo
}

func VRR_Plan() {
    globals.EnganiaPichangaMutex.Lock()
    // Determinar el trabajo actual basado en la prioridad o la cola estándar
    if len(globals.STS_Priority) > 0 {
        globals.CurrentJob = slice.Shift(&globals.STS_Priority)
    } else {
        globals.CurrentJob = slice.Shift(&globals.STS)
    }

    // Cambiar el estado del trabajo actual a EXEC
    globals.ChangeState(&globals.CurrentJob, "EXEC")
    globals.EnganiaPichangaMutex.Unlock()

    // Iniciar el temporizador para el quantum del trabajo actual
    timeBefore := time.Now()
    go startTimer(globals.CurrentJob.Quantum)

    // Enviar el PCB al kernel
    kernel_api.PCB_Send()

    // Esperar a recibir la señal de que el PCB ha sido procesado
    <-globals.PcbReceived

    // Calcular el tiempo que tomó la ejecución
    timeAfter := time.Now()
    diffTime := uint32(timeAfter.Sub(timeBefore))

    if diffTime < globals.CurrentJob.Quantum {
        // Si el trabajo terminó antes de consumir su quantum, ajustar el quantum restante
        globals.CurrentJob.Quantum -= uint32(time.Duration(diffTime) * time.Millisecond)
    } else {
        // Si el trabajo consumió todo su quantum, restablecer el quantum según la configuración del kernel
        globals.CurrentJob.Quantum = uint32(time.Duration(globals.Configkernel.Quantum) * time.Millisecond)
    }

    // Manejar la gestión de expulsión después de la ejecución del trabajo
    EvictionManagement()
}

/*
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

	<-globals.PcbReceived
	timeAfter := time.Now()

	diffTime := uint32(timeAfter.Sub(timeBefore))
	if diffTime < globals.CurrentJob.Quantum {
		globals.CurrentJob.Quantum -= diffTime * uint32(time.Millisecond)
	} else {
		globals.CurrentJob.Quantum = uint32(globals.Configkernel.Quantum * int(time.Millisecond))
	}

	EvictionManagement()
}
*/
func startTimer(quantum uint32) {
	quantumTime := time.Duration(quantum)
	auxPcb := globals.CurrentJob
	time.Sleep(quantumTime)

	quantumInterrupt(auxPcb)
}

func quantumInterrupt(pcb pcb.T_PCB) {
	// Interrumpir proceso actual, response = OK message
	kernel_api.SendInterrupt("QUANTUM", pcb.PID)
}

/*
*
  - EvictionManagement
  - Maneja los desalojos de los procesos

*
*/
func EvictionManagement() {
	evictionReason := globals.CurrentJob.EvictionReason

	switch evictionReason {
	case "BLOCKED_IO_GEN":
		globals.EnganiaPichangaMutex.Lock()
		globals.ChangeState(&globals.CurrentJob, "BLOCKED")
		slice.Push(&globals.Blocked, globals.CurrentJob)

		pcbAux := globals.CurrentJob
		log.Printf("PID: %d - Bloqueado por I/O genérico\n", globals.CurrentJob.PID)
		go func() {
			kernel_api.SolicitarGenSleep(pcbAux)
		}()
		//<-globals.JobExecBinary

	case "BLOCKED_IO_STDIN":
		globals.EnganiaPichangaMutex.Lock()
		globals.ChangeState(&globals.CurrentJob, "BLOCKED")
		slice.Push(&globals.Blocked, globals.CurrentJob)

		pcbAux := globals.CurrentJob
		log.Printf("PID: %d - Bloqueado por I/O de entrada\n", globals.CurrentJob.PID)
		go func() {
			kernel_api.SolicitarStdinRead(pcbAux)
		}()
		//<-globals.JobExecBinary

	case "BLOCKED_IO_STDOUT":
		globals.EnganiaPichangaMutex.Lock()
		globals.ChangeState(&globals.CurrentJob, "BLOCKED")
		slice.Push(&globals.Blocked, globals.CurrentJob)

		pcbAux := globals.CurrentJob
		go func() {
			kernel_api.SolicitarStdoutWrite(pcbAux)
		}()
		//<-globals.JobExecBinary

	case "BLOCKED_IO_DIALFS":
		globals.EnganiaPichangaMutex.Lock()
		globals.ChangeState(&globals.CurrentJob, "BLOCKED")

		pcbAux := globals.CurrentJob
		slice.Push(&globals.Blocked, globals.CurrentJob)
		go func() {
			kernel_api.SolicitarDialFS(pcbAux)
		}()
		//<-globals.JobExecBinary

	case "TIMEOUT":
		// TODO: Doble inserción en STS
		globals.ChangeState(&globals.CurrentJob, "READY")
		globals.STS = append(globals.STS, globals.CurrentJob)
		log.Printf("PID: %d - Desalojado por fin de quantum\n", globals.CurrentJob.PID)
		//<-globals.JobExecBinary
		globals.STSCounter <- int(globals.CurrentJob.PID)

	case "EXIT":
		kernel_api.KillJob(globals.CurrentJob)
		//<-globals.JobExecBinary
		<-globals.MultiprogrammingCounter
		log.Printf("Finaliza el proceso %d - Motivo: %s\n", globals.CurrentJob.PID, evictionReason)

	case "WAIT":
		if resource.Exists(globals.CurrentJob.RequestedResource) {
			resource.RequestConsumption(globals.CurrentJob.RequestedResource)
			//<-globals.JobExecBinary
		} else {
			fmt.Print("El recurso no existe\n")
			globals.CurrentJob.EvictionReason = "EXIT"
			EvictionManagement()
		}

	case "SIGNAL":
		if resource.Exists(globals.CurrentJob.RequestedResource) {
			resource.ReleaseConsumption(globals.CurrentJob.RequestedResource)
			//<-globals.JobExecBinary
		} else {
			fmt.Print("El recurso no existe\n")
			globals.CurrentJob.EvictionReason = "EXIT"
			EvictionManagement()
		}

	case "OUT_OF_MEMORY":	// ? En qué caso llega acá?
		globals.ChangeState(&globals.CurrentJob, "TERMINATED")
		kernel_api.KillJob(globals.CurrentJob)
		//<-globals.JobExecBinary
		<-globals.MultiprogrammingCounter
		log.Printf("Finaliza el proceso %d - Motivo: %s\n", globals.CurrentJob.PID, evictionReason)

	case "INTERRUPTED_BY_USER":
		globals.ChangeState(&globals.CurrentJob, "TERMINATED")
		kernel_api.KillJob(globals.CurrentJob)
		//<-globals.JobExecBinary
		<-globals.MultiprogrammingCounter
		log.Printf("Finaliza el proceso %d - Motivo: %s\n", globals.CurrentJob.PID, evictionReason)

	default:
		log.Fatalf("'%s' no es una razón de desalojo válida", evictionReason)
	}
}
