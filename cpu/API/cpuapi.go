package cpu_api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/sisoputnfrba/tp-golang/cpu/cicloInstruccion"
	"github.com/sisoputnfrba/tp-golang/cpu/globals"
	"github.com/sisoputnfrba/tp-golang/cpu/mmu"
	"github.com/sisoputnfrba/tp-golang/utils/pcb"
)

/**
 * PCB_recv: Recibe un PCB, lo "procesa" y lo devuelve
 * Cumple con la funcionalidad principal de CPU.
	* Procesar = Fetch -> Decode -> Execute
*/
func PCB_recv(w http.ResponseWriter, r *http.Request) {
	var received_pcb pcb.T_PCB

	// Decode PCB
	err := json.NewDecoder(r.Body).Decode(&received_pcb)
	if err != nil {
		http.Error(w, "Failed to decode PCB", http.StatusBadRequest)
		return
	}

	globals.CurrentJob = received_pcb
		
	for !pcb.EvictionFlag {
		cicloInstruccion.DecodeAndExecute(&globals.CurrentJob)

		fmt.Println("Los registros de la cpu son", globals.CurrentJob.CPU_reg)
	}

	//fmt.Println("ABER MOSTRAMELON: ", pcb.EvictionFlag) // * Se recordará su contribución a la ciencia
	pcb.EvictionFlag = false
	//fmt.Println("C PUSO FOLS ", pcb.EvictionFlag)
	
	jsonResp, err := json.Marshal(globals.CurrentJob)
	if err != nil {
		http.Error((w), "Failed to encode PCB response", http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResp)
}

type InterruptionRequest struct {
	InterruptionReason string `json:"InterruptionReason"`
	Pid uint32 `json:"pid"`
}
/**
 * HandleInterruption: Maneja las interrupciones de CPU	
*/
func HandleInterruption(w http.ResponseWriter, r *http.Request) {
	var request InterruptionRequest
	
	// Decode json payload
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	evictionReasons := map[string]struct{}{
		"EXIT":       {},
		"BLOCKED_IO": {},
	}
	
	if _, ok := evictionReasons[globals.CurrentJob.EvictionReason]; !ok && request.Pid == globals.CurrentJob.PID {
		switch request.InterruptionReason {
			case "QUANTUM":
				pcb.EvictionFlag = true
				globals.CurrentJob.EvictionReason = "TIMEOUT"
		}
	}

	w.WriteHeader(http.StatusOK)
}
//peticion para RESIZE de memoria (DESDE CPU A MEMORIA)

func Resize(tamanio int) {
	cliente := &http.Client{}
	url := fmt.Sprintf("http://%s:%d/resize", globals.Configcpu.IP_memory,globals.Configcpu.Port_memory)
	req, err := http.NewRequest("PATCH", url, nil)
	if err != nil {
		return 
	}

	q := req.URL.Query()
	q.Add("tamanio", "tamanio")
	req.URL.RawQuery = q.Encode()

	req.Header.Set("Content-Type", "application/json")
	respuesta, err := cliente.Do(req)
	if err != nil {
		return 
	}

	// Verificar el código de estado de la respuesta
	if respuesta.StatusCode != http.StatusOK {
		return 
	}

	bodyBytes, err := io.ReadAll(respuesta.Body)
	if err != nil {
		return 	
	}
	//En caso de que la respuesta de la memoria sea Out of Memory, se deberá devolver el contexto de ejecución al Kernel informando de esta situación
	// Y Avisar que el error es por out of memory
	var respuestaResize = string(bodyBytes)
	if respuestaResize != "OK" {
		//TODO: pcb.EvictionFlag = true
		//hacer esto
	} 
}
/* type DireccionTamanio struct {
	DireccionFisica int 
	Tamanio         int 
} */

type BodyRequestLeer struct {
	DireccionesTamanios []mmu.DireccionTamanio `json:"direcciones_tamanios"`
}


type BodyRequestEscribir struct {
	DireccionesTamanios []mmu.DireccionTamanio `json:"direcciones_tamanios"`
	Valor_a_escribir    string             `json:"valor_a_escribir"`
	Pid                 int                `json:"pid"`
}

func SolicitarEscritura(direccionesTamanios []mmu.DireccionTamanio  , valorAEscribir string, pid int) {

	body, err := json.Marshal(BodyRequestEscribir{
		DireccionesTamanios : direccionesTamanios,
		Valor_a_escribir    : valorAEscribir,
		Pid                 : pid,
	})          
	if err != nil {
		return
	}

	cliente := &http.Client{}
	url := fmt.Sprintf("http://%s:%d/write", globals.Configcpu.IP_memory, globals.Configcpu.Port_memory)
	escribirEnMemoria, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return
	}

	escribirEnMemoria.Header.Set("Content-Type", "application/json")
	respuesta, err := cliente.Do(escribirEnMemoria)
	if err != nil {
		fmt.Println("error")
	}

	if respuesta.StatusCode != http.StatusOK {
		fmt.Println("Error al realizar la escritura")
	}

	bodyBytes, err := io.ReadAll(respuesta.Body)
	if err != nil {
		fmt.Println("error")
	}

	// La respuesta puede ser un "Ok" o u "Error: dirección o tamanio fuera de rango"

	respuestaEnString := string(bodyBytes)
	if respuestaEnString != "OK" {
		fmt.Println("Se produjo un error al escribir", respuestaEnString)
	} else {
		fmt.Println("Se realizó la escritura correctamente", respuestaEnString)
	}

}

// -------------------------------------------------------------------------------
// LE SOLICITO A MEMORIA LEER Y DEVOLVER LO QUE ESTÉ EN LA DIREC FISICA INDICADA
func SolicitarLectura(direccionesFisicas []mmu.DireccionTamanio) string {

	jsonDirecYTamanio, err := json.Marshal(BodyRequestLeer{
		DireccionesTamanios: direccionesFisicas,
	})
	if err != nil {
		return "error"
	}

	cliente := &http.Client{}
	url := fmt.Sprintf("http://%s:%d/read", globals.Configcpu.IP_memory, globals.Configcpu.Port_memory)
	leerMemoria, err := http.NewRequest("GET", url, bytes.NewBuffer(jsonDirecYTamanio))
	if err != nil {
		return "error"
	}

	leerMemoria.Header.Set("Content-Type", "application/json")
	respuesta, err := cliente.Do(leerMemoria)
	if err != nil {
		return "error"
	}

	if respuesta.StatusCode != http.StatusOK {
		return "Error al realizar la lectura"	
	}

	bodyBytes, err := io.ReadAll(respuesta.Body)
	if err != nil {
		return "error"
	}

	return string(bodyBytes)

}

func PasarAInt(cadena string) int {
	num, err := strconv.Atoi(cadena)
	if err != nil {
		fmt.Println("Error: ", err)
	}
	return num
}