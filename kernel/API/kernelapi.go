package kernel_api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sisoputnfrba/tp-golang/kernel/globals"
	"github.com/sisoputnfrba/tp-golang/utils/pcb"
	"github.com/sisoputnfrba/tp-golang/utils/slice"
)

/* Glossary:

- BRQ: Body Request
- BRS: Body Response

*/

type ProcessStart_BRQ struct {
	Path string `json:"path"`
}

type ProcessStart_BRS struct {
	PID uint32 `json:"pid"`
}

/**
 * ProcessInit: Inicia un proceso en base a un archivo dentro del FS de Linux. // ?: Qué contiene el path del archivo? las instrucciones?
	[x] Creación de PCB
	[x] Asignación de PID incrementando en 1 por cada proceso creado
	[ ] Estado de proceso: NEW
*/
func ProcessInit(w http.ResponseWriter, r *http.Request) {
	var request ProcessStart_BRQ
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	// En algún lugar voy a tener que usar el path
	pcb := &pcb.T_PCB{
		PID: 		generatePID(),
		PC: 		0,
		Quantum: 	0,
		CPU_reg: 	[8]int{0, 0, 0, 0, 0, 0, 0, 0},
		State: 		"READY", // TODO: La idea es que el estado sea NEW cuando implementemos el LTS
	}

	globals.PidMutex.Lock()
	slice.Push(&globals.Processes, *pcb)
	slice.Push(&globals.STS, *pcb)	// TODO: Implementar LTS
	globals.PidMutex.Unlock()

	var respBody ProcessStart_BRS = ProcessStart_BRS{PID: pcb.PID}

	response, err := json.Marshal(respBody)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

func generatePID() uint32 {
	globals.PidMutex.Lock()
	defer globals.PidMutex.Unlock()
	globals.NextPID++
	return globals.NextPID
}

/**
 * ProcessDelete: Elimina un proceso en base a un PID. Realiza las operaciones como si el proceso llegase a EXIT
	[ ] Cambio de estado de proceso: EXIT
	[ ] Liberación de recursos
	[ ] Liberación de archivos
	[ ] Liberación de memoria
*/
func ProcessDelete(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Job deleted"))
}

type ProcessStatus_BRS struct {
	State string `json:"state"`
}

/**
 * ProcessState: Devuelve el estado de un proceso en base a un PID
	[ ] Devuelve el estado del proceso

	Por el momento devuelve un dato hardcodeado
*/
func ProcessState(w http.ResponseWriter, r *http.Request) {
	var respBody ProcessStatus_BRS = ProcessStatus_BRS{State: "EXEC"}

	response, err := json.Marshal(respBody)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

/**
 * PlanificationStart: Retoma el STS y LTS en caso de que la planificación se encuentre pausada. Si no, ignora la petición.
*/
func PlanificationStart(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Scheduler started"))
}

/**
 * PlanificationStop: Detiene el STS y LTS en caso de que la planificación se encuentre en ejecución. Si no, ignora la petición.
	El proceso que se encuentra en ejecución NO es desalojado. Una vez que salga de EXEC se pausa el manejo de su motivo de desalojo.
	El resto de procesos bloqueados van a pausar su transición a la cola de Ready
*/
func PlanificationStop(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Scheduler stopped"))
}

// TODO: Reemplazar el response con la futura struct de PCB. Preguntar cómo retornar varias struct
type ProcessList_BRS struct {
	Pid int `json:"pid"`
	State string `json:"state"`
}

/**
 * ProcessList: Devuelve una lista de procesos con su PID y estado
*/
func ProcessList(w http.ResponseWriter, r *http.Request) {
	var respBody ProcessList_BRS = ProcessList_BRS{Pid: 5, State: "BLOCK"}

	response, err := json.Marshal(respBody)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

// -----------------------------------------------------------------

func PCB_Send(pcb pcb.T_PCB) error {
	//Encode data
	jsonData, err := json.Marshal(pcb)
	if err != nil {
		return fmt.Errorf("failed to encode PCB: %v", err)
	}

	// Send data
	url := fmt.Sprintf("http://%s:%d/pcb-recv", globals.Configkernel.IP_cpu, globals.Configkernel.Port_cpu)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("POST request failed. Failed to send PCB: %v", err)
	}

	// Wait for response
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected response status: %s", resp.Status)
	}

	// Decode response and update value
	err = json.NewDecoder(resp.Body).Decode(&pcb)
	if err != nil {
		return fmt.Errorf("failed to decode PCB response: %v", err)
	}
	// Operar desalojo

	return nil
}