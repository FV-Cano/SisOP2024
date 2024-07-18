package cpu_api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/sisoputnfrba/tp-golang/cpu/cicloInstruccion"
	"github.com/sisoputnfrba/tp-golang/cpu/globals"
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

	globals.CurrentJob = &received_pcb

	for !pcb.EvictionFlag {
		cicloInstruccion.DecodeAndExecute(globals.CurrentJob)

		log.Println("Los registros de la cpu son", globals.CurrentJob.CPU_reg)
	}

	//log.Println("ABER MOSTRAMELON: ", pcb.EvictionFlag) // * Se recordará su contribución a la ciencia
	pcb.EvictionFlag = false
	//log.Println("C PUSO FOLS ", pcb.EvictionFlag)

	jsonResp, err := json.Marshal(globals.CurrentJob)
	if err != nil {
		http.Error((w), "Failed to encode PCB response", http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResp)
}

type InterruptionRequest struct {
	InterruptionReason string `json:"InterruptionReason"`
	Pid                uint32 `json:"pid"`
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
		"EXIT":          		{},
		"BLOCKED_IO_GEN": 		{},
		"BLOCKED_IO_STDIN":		{},
		"BLOCKED_IO_STDOUT":	{},
		"BLOCKED_IO_DIALFS":    {},
		"OUT_OF_MEMORY": 		{},
		"WAIT":		 			{},
		"SIGNAL":		 		{},
	}

	if _, ok := evictionReasons[globals.CurrentJob.EvictionReason]; !ok && request.Pid == globals.CurrentJob.PID {
		pcb.EvictionFlag = true

		switch request.InterruptionReason {
		case "QUANTUM":
			globals.CurrentJob.EvictionReason = "TIMEOUT"

		case "DELETE":
			globals.CurrentJob.EvictionReason = "INTERRUPTED_BY_USER"
		}
	}

	w.WriteHeader(http.StatusOK)
}
