package resource

import (
	"log"

	kernel_api "github.com/sisoputnfrba/tp-golang/kernel/API"
	"github.com/sisoputnfrba/tp-golang/kernel/globals"
	"github.com/sisoputnfrba/tp-golang/utils/pcb"
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
	globals.MapMutex.Lock()
	defer globals.MapMutex.Unlock()

	globals.ResourceMap[resource] = append(globals.ResourceMap[resource], pcb)
	globals.Blocked = append(globals.Blocked, pcb)
}

/**
 * Desencola un proceso de la cola de bloqueo de un recurso

 * @param resource: recurso al que se quiere acceder
 * @return pcb: proceso desencolado
*/
func DequeueProcess(resource string) pcb.T_PCB {
	globals.MapMutex.Lock()
	defer globals.MapMutex.Unlock()

	pcb := globals.ResourceMap[resource][0]
	globals.ResourceMap[resource] = globals.ResourceMap[resource][1:]

	kernel_api.RemoveFromBlocked(uint32(pcb.PID))
	return pcb
}

/**
 * Solicita la consumisión una instancia de un recurso

 * @param resource: recurso a consumir
*/
func RequestConsumption(resource string) {
	globals.MapMutex.Lock()
	defer globals.MapMutex.Unlock()

	if IsAvailable(resource) {
		globals.Resource_instances[resource]--
		log.Print("Se consumio una instancia del recurso: ", resource, "\n")
	} else {
		// * No debería ocurrir un problema de sincronización con esto pero por las dudas dejo el comentario
		globals.ChangeState(&globals.CurrentJob, "BLOCKED")
		QueueProcess(resource, globals.CurrentJob)
		log.Print("Entra el proceso PID: ", globals.CurrentJob.PID, " a la cola de bloqueo del recurso ", resource,  "\n")
	}
}

func ReleaseConsumption(resource string) {
	globals.MapMutex.Lock()
	defer globals.MapMutex.Unlock()

	globals.Resource_instances[resource]++
	log.Print("Se libero una instancia del recurso: ", resource, "\n")
	ReleaseJobIfBlocked(resource)
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
	globals.MapMutex.Lock()
	defer globals.MapMutex.Unlock()

	return globals.Resource_instances[resource] > 0
}

/**
 * Libera un proceso bloqueado por un recurso

 * @param resource: recurso del que se quiere liberar un proceso
*/
func ReleaseJobIfBlocked(resource string) {
	globals.MapMutex.Lock()
	defer globals.MapMutex.Unlock()

	if len(globals.ResourceMap[resource]) > 0 {
		pcb := DequeueProcess(resource)
		globals.ChangeState(&pcb, "READY")
		globals.STS = append(globals.STS, pcb)
		log.Print("Se desbloqueo el proceso PID: ", pcb.PID, " del recurso ", resource, "\n")
	}
}