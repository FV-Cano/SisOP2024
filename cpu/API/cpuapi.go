package cpu_api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sisoputnfrba/tp-golang/cpu/cicloInstruccion"
	"github.com/sisoputnfrba/tp-golang/cpu/globals"
	"github.com/sisoputnfrba/tp-golang/utils/pcb"
)

/**
 * PCB_Send: Envía un PCB al Kernel y recibe la respuesta

 * @return error: Error en caso de que falle el envío
 */
func PCB_Send(pcb *pcb.T_PCB) error {
	//Encode data
	jsonData, err := json.Marshal(pcb)
	if err != nil {
		return fmt.Errorf("failed to encode PCB: %v", err)
	}

	client := &http.Client{
		Timeout: 0,
	}

	// Send data
	url := fmt.Sprintf("http://%s:%d/dispatch", globals.Configcpu.IP_kernel, globals.Configcpu.Port_kernel)
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
	err = json.NewDecoder(resp.Body).Decode(&pcb) // ? Semaforo?
	if err != nil {
		return fmt.Errorf("failed to decode PCB response: %v", err)
	}

	return nil
}

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

	fmt.Println("ABER MOSTRAMELON: ", pcb.EvictionFlag)
	pcb.EvictionFlag = false
	fmt.Println("C PUSO FOLS ", pcb.EvictionFlag)
	
	// Encode PCB
	jsonResp, err := json.Marshal(globals.CurrentJob)
	if err != nil {
		http.Error((w), "Failed to encode PCB response", http.StatusInternalServerError)
	}

	PCB_Send(&globals.CurrentJob)

	// Send back PCB
	w.WriteHeader(http.StatusOK)
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

	if request.Pid == globals.CurrentJob.PID && globals.CurrentJob.EvictionReason != "EXIT" {
		switch request.InterruptionReason {
			case "QUANTUM":
				pcb.EvictionFlag = true
				globals.CurrentJob.EvictionReason = "TIMEOUT"
		}
	} else {
		fmt.Println("Se ignora la interrupción")
	}

	w.WriteHeader(http.StatusOK)
}