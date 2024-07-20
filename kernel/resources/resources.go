package resource

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/sisoputnfrba/tp-golang/kernel/globals"
	"github.com/sisoputnfrba/tp-golang/utils/pcb"
	"github.com/sisoputnfrba/tp-golang/utils/slice"
)

/**
 * Inicializa el mapa de recursos y la cantidad de instancias de cada recurso
 */
func InitResourceMap() {
	globals.ResourceMap = make(map[string][]pcb.T_PCB)
	globals.Resource_instances = make(map[string]int)

	for i, resource := range globals.Configkernel.Resources {
		globals.ResourceMap[resource] = []pcb.T_PCB{}
		globals.Resource_instances[resource] = globals.Configkernel.Resource_instances[i]
	}
}

/**
 * Encola un proceso en la cola de bloqueo de un recurso

 * @param resource: recurso al que se quiere acceder
 * @param pcb: proceso a encolar
*/
func QueueProcess(resource string, pcb pcb.T_PCB) {
	globals.ResourceMap[resource] = append(globals.ResourceMap[resource], pcb)
	slice.Push(&globals.Blocked, pcb)
}

/**
 * Desencola un proceso de la cola de bloqueo de un recurso

 * @param resource: recurso al que se quiere acceder
 * @return pcb: proceso desencolado
*/
func DequeueProcess(resource string) pcb.T_PCB {
	pcb := globals.ResourceMap[resource][0]
	globals.ResourceMap[resource] = globals.ResourceMap[resource][1:]

	RemoveFromBlocked(uint32(pcb.PID))
	return pcb
}

// No es la implementación más linda pero es la solución inmediata para evitar dependencias circulares
func RemoveFromBlocked(pid uint32) {
	for i, pcb := range globals.Blocked {
		if pcb.PID == pid {
			slice.RemoveAtIndex(&globals.Blocked, i)
		}
	}
}

/**
 * Solicita la consumisión una instancia de un recurso

 * @param resource: recurso a consumir
*/
func RequestConsumption(resource string) {
	globals.MapMutex.Lock()
	defer globals.MapMutex.Unlock()
	if IsAvailable(resource) {
		globals.ChangeState(&globals.CurrentJob, "READY")
		globals.Resource_instances[resource]--
		globals.CurrentJob.Resources[resource]++
		log.Print("Se consumio una instancia del recurso: ", resource, "\n")
		globals.CurrentJob.RequestedResource = ""
		slice.Push(&globals.STS, globals.CurrentJob)
		globals.STSCounter <- 1
	} else {
		log.Print("No hay instancias del recurso solicitado\n")
		// * No debería ocurrir un problema de sincronización con esto pero por las dudas dejo el comentario
		globals.ChangeState(&globals.CurrentJob, "BLOCKED")
		globals.CurrentJob.PC--	// Se decrementa el PC para que no avance en la próxima ejecución
		log.Print("PID: ", globals.CurrentJob.PID, " - Bloqueado por: ", resource, "\n")
		log.Print("Entra el proceso PID: ", globals.CurrentJob.PID, " a la cola de bloqueo del recurso ", resource,  "\n")
		QueueProcess(resource, globals.CurrentJob)
	}
}

/**
 * Solicita la liberación de una instancia de un recurso

 * @param resource: recurso a liberar
*/
func ReleaseConsumption(resource string) {
	globals.MapMutex.Lock()
	defer globals.MapMutex.Unlock()

	if globals.CurrentJob.Resources[resource] == 0 {
		log.Print("El proceso PID: ", globals.CurrentJob.PID, " no tiene instancias del recurso ", resource, " para liberar\n")
		return
	}

	globals.CurrentJob.Resources[resource]--
	globals.Resource_instances[resource]++
	log.Print("Se libero una instancia del recurso: ", resource, "\n")
	slice.InsertAtIndex(&globals.STS, 0, globals.CurrentJob)
	ReleaseJobIfBlocked(resource)
	globals.STSCounter <- 1	// TODO: dudas
}

/**
 * Consulta si existe un recurso

 * @param resource: recurso a consultar
 * @return bool: true si existe, false en caso contrario
*/
func Exists(resource string) bool {
	globals.MapMutex.Lock()
	defer globals.MapMutex.Unlock()

	_, ok := globals.Resource_instances[resource]
	return ok
}

/**
 * Consulta si hay instancias disponibles de un recurso

 * @param resource: recurso a consultar
 * @return bool: true si hay instancias disponibles, false en caso contrario
*/
func IsAvailable(resource string) bool {
	return globals.Resource_instances[resource] > 0
}

/**
 * Libera un proceso bloqueado por un recurso

 * @param resource: recurso del que se quiere liberar un proceso
*/
func ReleaseJobIfBlocked(resource string) {
	if len(globals.ResourceMap[resource]) > 0 {
		pcb := DequeueProcess(resource)
		globals.ChangeState(&pcb, "READY")
		globals.STS = append(globals.STS, pcb)
		log.Print("Se desbloqueo el proceso PID: ", pcb.PID, " del recurso ", resource, "\n")
		globals.STSCounter <- 1
	}
}

/**
 * Libera todos los recursos de un proceso

 * @param pcb: proceso al que se le quieren liberar los recursos
 * @return pcb: proceso con los recursos liberados
*/
func ReleaseAllResources(pcb pcb.T_PCB) pcb.T_PCB {
	for resource, instances := range globals.CurrentJob.Resources {
		for i := 0; i < instances; i++ {
			ReleaseConsumption(resource)
		}
	}
	
	return pcb
}

/**
 * Consulta si un proceso tiene recursos

 * @param pcb: proceso a consultar
 * @return bool: true si tiene recursos, false en caso contrario
*/
func HasResources(pcb pcb.T_PCB) bool {
	for _, instances := range pcb.Resources {
		if instances > 0 {
			return true
		}
	}
	return false
}


// --------------------- API ------------------------

func GETResourcesInstances(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(globals.Resource_instances)
}

func GETResourceBlockedJobs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(globals.ResourceMap)
}