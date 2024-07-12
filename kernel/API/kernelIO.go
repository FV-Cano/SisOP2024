package kernel_api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/sisoputnfrba/tp-golang/kernel/globals"
	"github.com/sisoputnfrba/tp-golang/utils/device"
	"github.com/sisoputnfrba/tp-golang/utils/pcb"
	"github.com/sisoputnfrba/tp-golang/utils/slice"
)

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
		if interf.InterfaceName == deviceName {
			log.Println("Interfaz encontrada: ", interf)
			return interf, nil
		}
	}
	return device.T_IOInterface{}, fmt.Errorf("device not found")
}

// * Types para realizar solicitudes a IO
type GenSleep struct {
	Pcb         pcb.T_PCB
	Inter       device.T_IOInterface
	TimeToSleep int
}

type StdinRead struct {
	Pcb                pcb.T_PCB
	Inter              device.T_IOInterface
	DireccionesFisicas []globals.DireccionTamanio
}

type StdoutWrite struct {
	Pcb                pcb.T_PCB
	Inter              device.T_IOInterface
	DireccionesFisicas []globals.DireccionTamanio
}

type DialFSRequest struct {
	Pcb           pcb.T_PCB
	Inter         device.T_IOInterface
	NombreArchivo string
	Tamanio       int
	Puntero       int
	Direccion     int
	Operacion     string
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
		Pcb:         pcb,
		Inter:       newInter,
		TimeToSleep: genSleepDataDecoded.SleepTime,
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
		DireccionesFisicas []globals.DireccionTamanio
		InterfaceName      string
		Tamanio            int
	})

	log.Println("RECIBE STDIN READ: ", stdinDataDecoded)

	newInter, err := SearchDeviceByName(stdinDataDecoded.InterfaceName)
	if err != nil {
		log.Printf("Device not found: %v", err)
	}

	stdinRead := StdinRead{
		Pcb:                pcb,
		Inter:              newInter,
		DireccionesFisicas: stdinDataDecoded.DireccionesFisicas,
	}

	log.Println("LE QUIERE MANDAR A IO: ", stdinRead)

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
		DireccionesFisicas []globals.DireccionTamanio
		InterfaceName      string
	})

	newInter, err := SearchDeviceByName(stdoutDataDecoded.InterfaceName)
	if err != nil {
		log.Printf("Device not found: %v", err)
	}

	stdoutWrite := StdoutWrite{
		Pcb:                pcb,
		Inter:              newInter,
		DireccionesFisicas: stdoutDataDecoded.DireccionesFisicas,
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

func SolicitarDialFS(pcb pcb.T_PCB) {
	dialFsDataDecoded := genericInterfaceBody.(struct {
		InterfaceName string
		FileName      string
		Size          int
		Pointer       int
		Address       int
		Operation     string
	})

	newInter, err := SearchDeviceByName(dialFsDataDecoded.InterfaceName)
	if err != nil {
		log.Printf("Device not found: %v", err)
	}

	dialFS := DialFSRequest{
		Pcb:           pcb,
		Inter:         newInter,
		NombreArchivo: dialFsDataDecoded.FileName,
		Tamanio:       dialFsDataDecoded.Size,
		Puntero:       dialFsDataDecoded.Pointer,
		Direccion:     dialFsDataDecoded.Address,
		Operacion:     dialFsDataDecoded.Operation,
	}

	globals.EnganiaPichangaMutex.Unlock()

	jsonData, err := json.Marshal(dialFS)
	if err != nil {
		log.Printf("Failed to encode DialFS request: %v", err)
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

var genericInterfaceBody interface{}

/*
	 RecvData_gensleep: Recibe desde CPU la información necesaria para solicitar un GEN_SLEEP.

	 Opera con estructura:
		- Nombre de la interfaz
		- Tiempo de espera

*
*/
func RecvData_gensleep(w http.ResponseWriter, r *http.Request) {
	var received_data struct {
		InterfaceName string
		SleepTime     int
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
		DireccionesFisicas []globals.DireccionTamanio
		InterfaceName      string
		Tamanio            int
	}

	err := json.NewDecoder(r.Body).Decode(&received_data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Println("Received data: ", received_data)
	genericInterfaceBody = received_data

	w.WriteHeader(http.StatusOK)
}

/*
	 RecvData_stdout: Recibe desde CPU la información necesaria para solicitar un STDOUT_WRITE.

	 Opera con estructura:
		- Direcciones físicas
		- Nombre de Interfaz

*
*/
func RecvData_stdout(w http.ResponseWriter, r *http.Request) {
	var received_data struct {
		DireccionesFisicas []globals.DireccionTamanio
		InterfaceName      string
	}

	err := json.NewDecoder(r.Body).Decode(&received_data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	genericInterfaceBody = received_data

	w.WriteHeader(http.StatusOK)
}

//DIALFS

func RecvData_dialfs(w http.ResponseWriter, r *http.Request) {
	var received_data struct {
		//	Pcb 					pcb.T_PCB
		InterfaceName string
		FileName      string
		Size          int
		Pointer       int
		Address       []globals.DireccionTamanio
		Operation     string
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

*
*/
func RecvPCB_IO(w http.ResponseWriter, r *http.Request) {
	var received_pcb pcb.T_PCB

	err := json.NewDecoder(r.Body).Decode(&received_pcb)
	if err != nil {
		http.Error(w, "Failed to decode PCB", http.StatusBadRequest)
		return
	}

	log.Println("PCB que nos manda IO (Kernel): PC: ", received_pcb.PC, "PID: ", received_pcb.PID)

	log.Println("Blocked: ", globals.Blocked)

	RemoveByID(received_pcb.PID)
	globals.ChangeState(&received_pcb, "READY")
	slice.Push(&globals.STS, received_pcb)
	globals.STSCounter <- int(received_pcb.PID)

	log.Println("LTS: ", globals.LTS)
	log.Println("STS: ", globals.STS)

	w.WriteHeader(http.StatusOK)
}
