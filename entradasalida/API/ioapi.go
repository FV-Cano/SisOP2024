package IO_api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/sisoputnfrba/tp-golang/entradasalida/globals"
)

type CantUnidadesTrabajo struct {
	Unidades int `json:"cantUnidades"`
}

type IOInterface struct {
	InterfaceType string `json:"interfaceType"`
	InterfacePort int    `json:"interfacePort"`
}

func HandshakeKernel() error {
	genInterface := IOInterface{
		InterfaceType: globals.ConfigIO.Type,
		InterfacePort: globals.ConfigIO.Port,
	}
	// Tal vez usa la misma IP que el kernel

	jsonData, err := json.Marshal(genInterface)
	if err != nil {
		return fmt.Errorf("failed to encode interface: %v", err)
	}

	url := fmt.Sprintf("http://%s:%d/io-handshake", globals.ConfigIO.Ip_kernel, globals.ConfigIO.Port_kernel)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("POST request failed. Failed to send interface: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected response status: %s", resp.Status)
	}

	log.Println("Handshake con Kernel exitoso")

	return nil
}

func IO_GEN_SLEEP() {
	sleepTime := globals.TiempoEspera * globals.ConfigIO.Unit_work_time
	log.Printf("PID: %d - Operacion: <IO_GEN_SLEEP>", globals.SleepPCB.PID)
	log.Printf("Bloqueado por %d segundos\n", sleepTime)

	time.Sleep(time.Duration(sleepTime) * time.Second)
}

func Resp_TiempoEsperaIO (w http.ResponseWriter, r *http.Request) {
	var aux int
	err := json.NewDecoder(r.Body).Decode(&aux)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	// * Posible implementación de semáforo
	globals.TiempoEspera = aux

	w.WriteHeader(http.StatusOK)
}

func IOGenSleep(w http.ResponseWriter, r *http.Request) {
	err := json.NewDecoder(r.Body).Decode(&globals.SleepPCB)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	IO_GEN_SLEEP()

	log.Println("Fin de bloqueo")

	jsonData, err := json.Marshal(globals.SleepPCB)
	if err != nil {
		http.Error(w, "Failed to encode PCB response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)

	w.WriteHeader(http.StatusOK)
}