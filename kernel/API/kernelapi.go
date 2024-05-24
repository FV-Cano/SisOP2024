package kernel_api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

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

type GetInstructions_BRQ struct {
	Path string `json:"path"`
	Pid  uint32 `json:"pid"`
	Pc 	uint32  `json:"pc"`
}

/**
 * ProcessInit: Inicia un proceso en base a un archivo dentro del FS de Linux.
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

	pathInst, err := json.Marshal(fmt.Sprintf(request.Path))
    if err != nil {
        http.Error(w, "Error al codificar los datos como JSON", http.StatusInternalServerError)
        return
    }
	pathInstString := string(pathInst)
	

	newPcb := &pcb.T_PCB{
		PID: 			generatePID(),
		PC: 			0,
		Quantum: 		0,
		CPU_reg: 		map[string]interface{}{
							"AX": uint8(0),
							"BX": uint8(0),
							"CX": uint8(0),
							"DX": uint8(0),
							"EAX": uint32(0),
							"EBX": uint32(0),
							"ECX": uint32(0),
							"EDX": uint32(0),
						},
		State: 			"READY", // TODO: La idea es que el estado sea NEW cuando implementemos el LTS
		EvictionReason: "",
	}

	globals.ProcessesMutex.Lock()
	slice.Push(&globals.Processes, *newPcb)
	defer globals.ProcessesMutex.Unlock()

	globals.STSMutex.Lock()
	slice.Push(&globals.STS, *newPcb)	// TODO: Implementar LTS
	defer globals.STSMutex.Unlock()

	globals.MultiprogrammingCounter <- int(newPcb.PID)

	var respBody ProcessStart_BRS = ProcessStart_BRS{PID: newPcb.PID}

	response, err := json.Marshal(respBody)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Obtengo las instrucciones del proceso
	url := fmt.Sprintf("http://%s:%d/instrucciones", globals.Configkernel.IP_memory, globals.Configkernel.Port_memory)

	bodyInst, err := json.Marshal(GetInstructions_BRQ{
		Path: pathInstString,
		Pid: newPcb.PID,
		Pc: newPcb.PC,
	})
	if err != nil {
		return
	}
	
	requerirInstrucciones, err := http.NewRequest("POST", url, bytes.NewBuffer(bodyInst))
	if err != nil {
		log.Fatalf("POST request failed (No se pueden cargar instrucciones): %v", err)
	}
	
	cliente := &http.Client{}
	requerirInstrucciones.Header.Set("Content-Type", "application/json")
	recibirRespuestaInstrucciones, err := cliente.Do(requerirInstrucciones)
	if (err != nil || recibirRespuestaInstrucciones.StatusCode != http.StatusOK) {
		log.Fatal("Error en CargarInstrucciones (memoria)", err)
	}

	log.Printf("Se crea el proceso %d en %s\n", newPcb.PID, newPcb.State)

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
	pidString := r.PathValue("pid")
	pid, err := GetPIDFromString(pidString)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Elimino el proceso de la lista de procesos
	RemoveByID(pid)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Job deleted")) // ! No tiene que devolver nada
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
	pidString := r.PathValue("pid")
	pid, err := GetPIDFromString(pidString)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	process, _ := SearchByID(pid, globals.Processes)

	result := ProcessStatus_BRS{State: process.State}

	response, err := json.Marshal(result)
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
	<- globals.PlanBinary
}

/**
 * PlanificationStop: Detiene el STS y LTS en caso de que la planificación se encuentre en ejecución. Si no, ignora la petición.
	El proceso que se encuentra en ejecución NO es desalojado. Una vez que salga de EXEC se pausa el manejo de su motivo de desalojo.
	El resto de procesos bloqueados van a pausar su transición a la cola de Ready
*/
func PlanificationStop(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	globals.PlanBinary <- false
}

type ProcessList_BRS struct {
	Pid int `json:"pid"`
	State string `json:"state"`
}

/**
 * ProcessList: Devuelve una lista de procesos con su PID y estado
*/
func ProcessList(w http.ResponseWriter, r *http.Request) {
	// Me traigo los procesos de la lista de procesos
	globals.ProcessesMutex.Lock()
	allProcesses := globals.Processes
	defer globals.ProcessesMutex.Unlock()

	// Formateo los procesos para devolverlos
	respBody := make([]ProcessList_BRS, len(allProcesses))
	for i, process := range allProcesses {
		respBody[i] = ProcessList_BRS{Pid: int(process.PID), State: process.State}
	}

	response, err := json.Marshal(respBody)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

/**
 * PCB_Send: Envía un PCB al CPU y recibe la respuesta

 * @param pcb: PCB a enviar
 * @return error: Error en caso de que falle el envío
*/
func PCB_Send() error {
	//Encode data
	jsonData, err := json.Marshal(globals.CurrentJob) // ? Semaforo?
	if err != nil {
		return fmt.Errorf("failed to encode PCB: %v", err)
	}

	client := &http.Client{
		Timeout: 0,
	}

	// Send data
	url := fmt.Sprintf("http://%s:%d/dispatch", globals.Configkernel.IP_cpu, globals.Configkernel.Port_cpu)
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("POST request failed. Failed to send PCB: %v", err)
	}

	// Wait for response
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected response status: %s", resp.Status)
	}

	// Decode response and update value
	err = json.NewDecoder(resp.Body).Decode(&globals.CurrentJob) // ? Semaforo?
	if err != nil {
		return fmt.Errorf("failed to decode PCB response: %v", err)
	}

	return nil
}

/**
 * SearchByID: Busca un proceso en la lista de procesos en base a su PID

 * @param pid: PID del proceso a buscar
 * @param processList: Lista de procesos
 * @return *pcb.T_PCB: Proceso encontrado
*/
func SearchByID(pid uint32, processList []pcb.T_PCB) (*pcb.T_PCB, int) {
	for i, process := range processList {
		if process.PID == pid {
			return &process, i
		}
	}
	return nil, -1
}

/**
 * RemoveByID: Remueve un proceso de la lista de procesos en base a su PID

 * @param pid: PID del proceso a remover
*/
func RemoveByID(pid uint32) error {
	_, generalIndex := SearchByID(pid, globals.Processes)
	
	if (generalIndex == -1) {
		return fmt.Errorf("process with PID %d not found", pid)
	} else {
		globals.ProcessesMutex.Lock()
		defer globals.ProcessesMutex.Unlock()
		slice.RemoveAtIndex(&globals.Processes, generalIndex)
	}
	
	_, ltsIndex := SearchByID(pid, globals.LTS)
	
	_, stsIndex := SearchByID(pid, globals.STS)
	
	if ltsIndex != -1 {
		globals.LTSMutex.Lock()
		defer globals.LTSMutex.Unlock()
		slice.RemoveAtIndex(&globals.LTS, ltsIndex)	
	}
	
	if stsIndex != -1 {
		globals.STSMutex.Lock()
		defer globals.STSMutex.Unlock()
		slice.RemoveAtIndex(&globals.STS, stsIndex)
	}
	
	return nil
}

/**
 * GetPIDFromQueryPath: Obtiene el PID de un path de query

 * @param path: Path de query
 * @return uint32: PID extraído
*/
func GetPIDFromString(pidString string) (uint32, error) {
	pid64, error := strconv.ParseUint(pidString, 10, 32)
	return uint32(pid64), error
}

// ----------------- IO -----------------

func GetIOInterface(w http.ResponseWriter, r *http.Request) {
	// Decode body
	
	err := json.NewDecoder(r.Body).Decode(&globals.IO_Interface)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// log received interface
	log.Printf("Interface received, type: %s, port: %d\n", globals.IO_Interface.InterfaceType, globals.IO_Interface.InterfacePort)

	w.WriteHeader(http.StatusOK)
}

// TODO: Borrar
// Comunicarle a CPU dispositivos de IO disponibles
	/*
		Nota: La comunicación IO --handshake--> Kernel --> CPU es la única forma que se me ocurrió para que CPU sepa qué dispositivos de IO están disponibles.
		- CPU no puede comunicarse con IO directamente para preguntar
		- CPU no puede comunicarse directamente con kernel, solo lo hace en respuesta a peticiones
		- Si bien podría hacer IO --handshake--> Kernel (variable global común a todos los módulos), no me parece correcto porque rompe la idea de que cada módulo es independiente y conoce únicamente la información justa (De igual manera, idea sujeta a cambios) 
	*/

	/* // Encode data
	jsonData, err := json.Marshal(globals.IO_Interface)
	if err != nil {
		http.Error(w, "Failed to encode interface", http.StatusInternalServerError)
		return
	}

	// Send data
	url := fmt.Sprintf("http://%s:%d/io-data", globals.Configkernel.IP_cpu, globals.Configkernel.Port_cpu)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		http.Error(w, "Failed to send interface", http.StatusInternalServerError)
		return
	}

	if resp.StatusCode != http.StatusOK {
		http.Error(w, "Unexpected response status", http.StatusInternalServerError)
		return
	} */

func Resp_ExisteInterfazGen (w http.ResponseWriter, r *http.Request) {

	var received_data string

	err := json.NewDecoder(r.Body).Decode(&received_data)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	var response bool
	if received_data == globals.IO_Interface.InterfaceType {
		response = true
	} else {
		response = false
	}

	jsonResp, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResp)
}