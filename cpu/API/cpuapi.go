package cpu_api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sisoputnfrba/tp-golang/cpu/cicloInstruccion"
	"github.com/sisoputnfrba/tp-golang/cpu/globals"
	"github.com/sisoputnfrba/tp-golang/utils/pcb"
)

type interruptionRequest struct {
	InterruptionCode int `json:"interruptCode"`
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
	
	fmt.Printf("Received PCB: %v\n", received_pcb) //Log?

	// Sección donde trabajo el pcb recibido (me interesa usar un hilo?)
	globals.EvictionMutex.Lock()
	defer globals.EvictionMutex.Unlock()
	for !pcb.EvictionFlag {
		globals.EvictionMutex.Unlock()

		cicloInstruccion.DecodeAndExecute(received_pcb)
		
		// Check interrupt (Al ser asincrónico no puedo hacer el check, espero a que el handler ejecute y luego cambio el valor de la flag de interrupción)
		globals.EvictionMutex.Lock()
	}

	pcb.EvictionFlag = false
	
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
 * HandleInterruption: Maneja las interrupciones de CPU	
*/
func HandleInterruption(w http.ResponseWriter, r *http.Request) {
	var request interruptionRequest
	
	// Decode json payload
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	switch request.InterruptionCode {
		case pcb.QUANTUM:
			// Cambiar motivo de desalojo a "Quantum"
	}
}