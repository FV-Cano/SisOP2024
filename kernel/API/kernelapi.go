package kernel_api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/sisoputnfrba/tp-golang/kernel/globals"
	"github.com/sisoputnfrba/tp-golang/utils/device"
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
 	[ ] Testeada
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
		Quantum: 		uint32(globals.Configkernel.Quantum),
		CPU_reg: 		map[string]interface{}{
							"AX": uint8(0),
							"BX": uint8(0),
							"CX": uint8(0),
							"DX": uint8(0),
							"EAX": uint32(0),
							"EBX": uint32(0),
							"ECX": uint32(0),
							"EDX": uint32(0),
							"SI": uint32(0),
							"DI": uint32(0),
						},
		State: 			"NEW",
		EvictionReason: "",
		Resources: 		make(map[string]int),	// * El valor por defecto es 0, tener en cuenta por las dudas a la hora de testear
		RequestedResource: "",
	}

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
		log.Printf("POST request failed (No se pueden cargar instrucciones): %v", err)
	}
	
	cliente := &http.Client{}
	requerirInstrucciones.Header.Set("Content-Type", "application/json")
	recibirRespuestaInstrucciones, err := cliente.Do(requerirInstrucciones)
	if (err != nil || recibirRespuestaInstrucciones.StatusCode != http.StatusOK) {
		log.Fatal("Error en CargarInstrucciones (memoria)", err)
	}

	// Si la lista está vacía, la desbloqueo
	if len(globals.LTS) == 0 {
		globals.EmptiedListMutex.Unlock()
	}

	globals.LTSMutex.Lock()
	slice.Push(&globals.LTS, *newPcb)
	defer globals.LTSMutex.Unlock()

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

	[ ] Testeada
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
	[ ] Testeada
*/
func ProcessState(w http.ResponseWriter, r *http.Request) {
	fmt.Println("HOLAAA ESTOY EN PROCESS STATE")
	pidString := r.PathValue("pid")
	pid, err := GetPIDFromString(pidString)
	if err != nil {
		fmt.Println("Error al convertir PID a string: ", pidString, err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Println("Vamos a buscar el proceso con PID: ", pid)

	process, _ := SearchByID(pid, getProcessList())
	if process == nil {
		http.Error(w, "Process not found", http.StatusNotFound)
		return
	}

	fmt.Println("Encontré esto: ", process)

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
	<- globals.LTSPlanBinary
	<- globals.STSPlanBinary
}

/**
 * PlanificationStop: Detiene el STS y LTS en caso de que la planificación se encuentre en ejecución. Si no, ignora la petición.
	El proceso que se encuentra en ejecución NO es desalojado. Una vez que salga de EXEC se pausa el manejo de su motivo de desalojo.
	El resto de procesos bloqueados van a pausar su transición a la cola de Ready
*/
func PlanificationStop(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	globals.LTSPlanBinary <- false
	globals.STSPlanBinary <- false
}

type ProcessList_BRS struct {
	Pid int `json:"pid"`
	State string `json:"state"`
}

/**
 * ProcessList: Devuelve una lista de procesos con su PID y estado
*/
func ProcessList(w http.ResponseWriter, r *http.Request) {
	allProcesses := getProcessList()

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
 * getProcessList: Devuelve una lista de todos los procesos en el sistema (LTS, STS, Blocked, STS_Priority, CurrentJob)

 * @return []pcb.T_PCB: Lista de procesos
*/
func getProcessList() []pcb.T_PCB {
	var allProcesses []pcb.T_PCB
	allProcesses = append(allProcesses, globals.LTS...)
	allProcesses = append(allProcesses, globals.STS...)
	allProcesses = append(allProcesses, globals.STS_Priority...)
	allProcesses = append(allProcesses, globals.Blocked...)
	if globals.CurrentJob.PID != 0 {
		allProcesses = append(allProcesses, globals.CurrentJob)
	}
	return allProcesses
}

/**
 * PCB_Send: Envía un PCB al CPU y recibe la respuesta

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

	globals.PcbReceived <- true

	return nil
}

func PCB_recv(w http.ResponseWriter, r *http.Request) {
	var received_pcb pcb.T_PCB

	// Decode PCB
	err := json.NewDecoder(r.Body).Decode(&received_pcb)
	if err != nil {
		http.Error(w, "Failed to decode PCB", http.StatusBadRequest)
		return
	}
		
	globals.CurrentJob = received_pcb
	globals.PcbReceived <- true

	// Encode PCB
	jsonResp, err := json.Marshal(received_pcb)
	if err != nil {
		http.Error((w), "Failed to encode PCB response", http.StatusInternalServerError)
	}

	// Send back PCB
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResp)	
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
	_, ltsIndex := SearchByID(pid, globals.LTS)
	_, stsIndex := SearchByID(pid, globals.STS)
	
	if ltsIndex != -1 {
		globals.LTSMutex.Lock()
		defer globals.LTSMutex.Unlock()
		slice.RemoveAtIndex(&globals.LTS, ltsIndex)	
	} else if stsIndex != -1 {
		globals.STSMutex.Lock()
		defer globals.STSMutex.Unlock()
		slice.RemoveAtIndex(&globals.STS, stsIndex)
	}
	
	return nil
}

/**
 * GetPIDFromQueryPath: Convierte un PID en formato string a uint32

 * @param pidString: PID en formato string
 * @return uint32: PID extraído
*/
func GetPIDFromString(pidString string) (uint32, error) {
	pid64, error := strconv.ParseUint(pidString, 10, 32)
	return uint32(pid64), error
}

func RemoveFromBlocked(pid uint32) {
	for i, pcb := range globals.Blocked {
		if pcb.PID == pid {
			slice.RemoveAtIndex(&globals.Blocked, i)
		}
	}

}

// ----------------- IO -----------------
func GetIOInterface(w http.ResponseWriter, r *http.Request) {
	var interf device.T_IOInterface
	
	err := json.NewDecoder(r.Body).Decode(&interf)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	slice.Push(&globals.Interfaces, interf)

	log.Printf("Interface received, type: %s, port: %d\n", interf.InterfaceType, interf.InterfacePort)

	w.WriteHeader(http.StatusOK)
}
type SearchInterface struct {
	Name string `json:"name"`
	Type string `json:"type"`
}
func ExisteInterfaz(w http.ResponseWriter, r *http.Request) {
	var received_data SearchInterface
	err := json.NewDecoder(r.Body).Decode(&received_data)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
	}
	
	log.Printf("Received data: %s, %s\n", received_data.Name, received_data.Type)

	aux, err := SearchDeviceByName(received_data.Name)
	if err != nil {
		http.Error(w, "Device not found", http.StatusNotFound)
	}
	
	var response bool
	if aux.InterfaceType == received_data.Type {
		response = true
	} else {
		response = false
	}

	jsonResp, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResp)
}

func SearchDeviceByName(deviceName string) (device.T_IOInterface, error) {
	for _, interf := range globals.Interfaces {
		if interf.InterfaceName == deviceName  {
			fmt.Println("Interfaz encontrada: ", interf)
			return interf, nil
		}
	}
	return device.T_IOInterface{}, fmt.Errorf("device not found")
}


// * Types para realizar solicitudes a IO
type GenSleep struct {
	Pcb	 				pcb.T_PCB
	Inter 				device.T_IOInterface
	TimeToSleep 		int
}

type StdinRead struct {
	Pcb 				pcb.T_PCB
	Inter 				device.T_IOInterface
	DireccionesFisicas 	[]globals.DireccionTamanio
}

type StdoutWrite struct {
	Pcb 				pcb.T_PCB
	Inter 				device.T_IOInterface
	DireccionesFisicas 	[]globals.DireccionTamanio
}


func SolicitarGenSleep(pcb pcb.T_PCB) {
	genSleepDataDecoded := genericInterfaceBody.(struct {
		InterfaceName string
		SleepTime     int
	})

	newInter, err := SearchDeviceByName(genSleepDataDecoded.InterfaceName)
	if err != nil {
		log.Printf("Device not found: %v", err)
	}
	
	genSleep := GenSleep{
		Pcb: 			pcb,
		Inter: 			newInter, 
		TimeToSleep: 	genSleepDataDecoded.SleepTime,
	}

	globals.EnganiaPichangaMutex.Unlock()
	
	jsonData, err := json.Marshal(genSleep)
	if err != nil {
		log.Printf("Failed to encode GenSleep request: %v", err)
	}

	url := fmt.Sprintf("http://%s:%d/io-operate", newInter.InterfaceIP, newInter.InterfacePort)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Failed to send PCB: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("Unexpected response status: %s", resp.Status)
	}
}

func SolicitarStdinRead(pcb pcb.T_PCB) {
	stdinDataDecoded := genericInterfaceBody.(struct {
		DireccionesFisicas 	[]globals.DireccionTamanio
		InterfaceName 		string
		Tamanio 			int
	})

	fmt.Println("RECIBE STDIN READ: ", stdinDataDecoded)

	newInter, err := SearchDeviceByName(stdinDataDecoded.InterfaceName)
	if err != nil {
		log.Printf("Device not found: %v", err)
	}

	stdinRead := StdinRead {
		Pcb: 					pcb,
		Inter:	 				newInter,
		DireccionesFisicas:		stdinDataDecoded.DireccionesFisicas,
	}

	fmt.Println("LE QUIERE MANDAR A IO: ", stdinRead)

	globals.EnganiaPichangaMutex.Unlock()

	jsonData, err := json.Marshal(stdinRead)
	if err != nil {
		log.Printf("Failed to encode StdinRead request: %v", err)
	}

	url := fmt.Sprintf("http://%s:%d/io-operate", newInter.InterfaceIP, newInter.InterfacePort)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Failed to send PCB: %v", err)
	}

	log.Println("IO STDIN FUE AVISADO POR KERNEL")

	if resp.StatusCode != http.StatusOK {
		log.Printf("Unexpected response status: %s", resp.Status)
	}
}

func SolicitarStdoutWrite(pcb pcb.T_PCB) {
	stdoutDataDecoded := genericInterfaceBody.(struct {
		DireccionesFisicas 	[]globals.DireccionTamanio
		InterfaceName 		string
	})

	newInter, err := SearchDeviceByName(stdoutDataDecoded.InterfaceName)
	if err != nil {
		log.Printf("Device not found: %v", err)
	}

	stdoutWrite := StdoutWrite {
		Pcb: 					pcb,
		Inter: 					newInter,
		DireccionesFisicas: 	stdoutDataDecoded.DireccionesFisicas,
	}

	globals.EnganiaPichangaMutex.Unlock()

	jsonData, err := json.Marshal(stdoutWrite)
	if err != nil {
		log.Printf("Failed to encode StdoutWrite request: %v", err)
	}

	url := fmt.Sprintf("http://%s:%d/io-operate", newInter.InterfaceIP, newInter.InterfacePort)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Failed to send PCB: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("Unexpected response status: %s", resp.Status)
	}
}

// -------------------------- Caos de Interfaces --------------------------

var genericInterfaceBody 	interface{}

/*
 RecvData_gensleep: Recibe desde CPU la información necesaria para solicitar un GEN_SLEEP.

 Opera con estructura:
	- Nombre de la interfaz
	- Tiempo de espera
**/
func RecvData_gensleep(w http.ResponseWriter, r *http.Request) {
	var received_data struct {
		InterfaceName 	string
		SleepTime 		int
	}

	err := json.NewDecoder(r.Body).Decode(&received_data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	genericInterfaceBody = received_data
	
	w.WriteHeader(http.StatusOK)
}

/*
 RecvData_stdin: Recibe desde CPU la información necesaria para solicitar un STDIN_READ.

 Opera con estructura:
	- Direcciones físicas
	- Nombre de Interfaz
	- Tamaño
**/
func RecvData_stdin(w http.ResponseWriter, r *http.Request) {
	var received_data struct {
		DireccionesFisicas 	[]globals.DireccionTamanio
		InterfaceName		string
		Tamanio 			int
	}

	err := json.NewDecoder(r.Body).Decode(&received_data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Println("Received data: ", received_data)
	genericInterfaceBody = received_data

	w.WriteHeader(http.StatusOK)
}

/*
 RecvData_stdout: Recibe desde CPU la información necesaria para solicitar un STDOUT_WRITE.

 Opera con estructura:
	- Direcciones físicas
	- Nombre de Interfaz
**/
func RecvData_stdout(w http.ResponseWriter, r *http.Request) {
	var received_data struct {
		DireccionesFisicas 	[]globals.DireccionTamanio
		InterfaceName		string
	}

	err := json.NewDecoder(r.Body).Decode(&received_data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	genericInterfaceBody = received_data

	w.WriteHeader(http.StatusOK)
}

/*
 RecvPCB_IO: Recibe el PCB bloqueado por IO, lo desbloquea y lo agrega a la cola de STS.
**/
func RecvPCB_IO(w http.ResponseWriter, r *http.Request) {
	var received_pcb pcb.T_PCB

	err := json.NewDecoder(r.Body).Decode(&received_pcb)
	if err != nil {
		http.Error(w, "Failed to decode PCB", http.StatusBadRequest)
		return
	}
	
	globals.ChangeState(&received_pcb, "READY")
	slice.Push(&globals.STS, received_pcb)
	globals.STSCounter <- int(received_pcb.PID)

	w.WriteHeader(http.StatusOK)
}