package cpu_api

import (
	"encoding/json"
	"fmt"
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
		"EXIT":       {},
		"BLOCKED_IO": {},
		"OUT_OF_MEMORY": {},
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
/*
func BuscarEnTLB(pid int, pagina int) bool {

	if entry, exists := tlb.CurrentTLB[pid]; exists && entry.Pagina == pagina {
		return true
	}
	return false
}

func FrameEnTLB(pid int, pagina int) int {

	if entry, exists := tlb.CurrentTLB[pid]; exists && entry.Pagina == pagina {
		ActualizarTLB(pid, pagina, tlb.CurrentTLB[pid].Marco)
		return tlb.CurrentTLB[pid].Marco
	}
	return -1

}

func ObtenerPagina(direccionLogica int, nroPag int, tamanio int) int {
	pagina := (direccionLogica + nroPag*tamanio) / tamanio

	return pagina
}

func ObtenerOffset(direccionLogica int, nroPag int, tamanio int) int {

	offset := (direccionLogica + nroPag*tamanio) % tamanio

	return offset
}

func CalcularDireccionFisica(frame int, offset int, tamanio int) int {

	direccionBase := frame * tamanio

	return direccionBase + offset

}

/*func ActualizarTLB(pid, pagina, marco int) { 
	if globals.Configcpu.Algorithm_tlb == "FIFO" {
	if len(tlb.CurrentTLB) >= globals.Configcpu.Number_felling_tlb {
		// Si la TLB está llena, eliminar la entrada más antigua (FIFO)
		for key := range tlb.CurrentTLB {
			delete(tlb.CurrentTLB, key)
			break
		}
	}
	tlb.CurrentTLB[pid] = tlb.Pagina_marco{Pagina: pagina, Marco: marco}

	}
}

func ActualizarTLB(pid, pagina, marco int) { 
	switch globals.Configcpu.Algorithm_tlb {
		case "FIFO":
        if !BuscarEnTLB(pid, pagina) { //Si la página no está en la tlb
            if len(tlb.CurrentTLB) < globals.Configcpu.Number_felling_tlb {
                // Si la TLB no está llena, agregar la entrada
                tlb.CurrentTLB[pid] = tlb.Pagina_marco{Pagina: pagina, Marco: marco}
                tlb.OrderedKeys = append(tlb.OrderedKeys, pid) // Agregar la clave al final de la lista
            } else {
                // Si la TLB está llena, eliminar la entrada más antigua (FIFO)
                oldestKey := tlb.OrderedKeys[0] // Obtener la clave más antigua
                delete(tlb.CurrentTLB, oldestKey) // Eliminar la entrada más antigua
                tlb.OrderedKeys = tlb.OrderedKeys[1:] // Eliminar la clave más antigua de la lista
                tlb.CurrentTLB[pid] = tlb.Pagina_marco{Pagina: pagina, Marco: marco} // Agregar la nueva entrada
                tlb.OrderedKeys = append(tlb.OrderedKeys, pid) // Agregar la nueva clave al final de la lista
            } 
        } 
			
		case "LRU":
			 if !BuscarEnTLB(pid, pagina) { //Si la página no está en la tlb
				if len(tlb.CurrentTLB) < globals.Configcpu.Number_felling_tlb {
					// Si la TLB no está llena, agregar la entrada
					tlb.CurrentTLB[pid] = tlb.Pagina_marco{Pagina: pagina, Marco: marco}
					tlb.OrderedKeys = append(tlb.OrderedKeys, pid) // Agregar la clave al final de la lista
				} else {
					// Si la TLB está llena, eliminar la entrada más antigua (FIFO)
					oldestKey := tlb.OrderedKeys[0] // Obtener la clave más antigua
					delete(tlb.CurrentTLB, oldestKey) // Eliminar la entrada más antigua
					tlb.OrderedKeys = tlb.OrderedKeys[1:] // Eliminar la clave más antigua de la lista
					tlb.CurrentTLB[pid] = tlb.Pagina_marco{Pagina: pagina, Marco: marco} // Agregar la nueva entrada
					tlb.OrderedKeys = append(tlb.OrderedKeys, pid) // Agregar la nueva clave al final de la lista
				}
			} else { //SI LA PAGINA YA EXISTE EN LA TLB, LLEVARLA AL FINAL DE LA LISTA
				// Eliminar la entrada existente y agregarla nuevamente
				for i, key := range tlb.OrderedKeys {
					if key == pid {
						// Eliminar la clave de la lista
						tlb.OrderedKeys = append(tlb.OrderedKeys[:i], tlb.OrderedKeys[i+1:]...)
						break
					}
				}
				tlb.CurrentTLB[pid] = tlb.Pagina_marco{Pagina: pagina, Marco: marco} // Agregar la nueva entrada
				tlb.OrderedKeys = append(tlb.OrderedKeys, pid) // Agregar la nueva clave al final de la lista
			}
	}
}*/