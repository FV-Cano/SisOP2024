package IO_api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/sisoputnfrba/tp-golang/entradasalida/globals"
	"github.com/sisoputnfrba/tp-golang/utils/device"
	"github.com/sisoputnfrba/tp-golang/utils/pcb"
)

type CantUnidadesTrabajo struct {
	Unidades int `json:"cantUnidades"`
}

func HandshakeKernel(nombre string) error {
	genInterface := device.T_IOInterface{
		InterfaceName: nombre,
		InterfaceType: globals.ConfigIO.Type,
		InterfaceIP: globals.ConfigIO.Ip,
		InterfacePort: globals.ConfigIO.Port,
	}

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

func IO_GEN_SLEEP(sleepTime int, pcb pcb.T_PCB) {
	sleepTimeTotal := sleepTime * globals.ConfigIO.Unit_work_time
	log.Printf("PID: %d - Operacion: <IO_GEN_SLEEP>", pcb.PID)
	log.Printf("Bloqueado por %d segundos\n", sleepTimeTotal)

	time.Sleep(time.Duration(sleepTimeTotal) * time.Second)
}

/* func Resp_TiempoEsperaIO (w http.ResponseWriter, r *http.Request) {
	var aux int
	err := json.NewDecoder(r.Body).Decode(&aux)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	// * Posible implementación de semáforo
	globals.TiempoEspera = aux

	w.WriteHeader(http.StatusOK)
} */

type GenSleep struct {
	Pcb	 		pcb.T_PCB
	Inter 		device.T_IOInterface
	TimeToSleep int
}

func IOGenSleep(w http.ResponseWriter, r *http.Request) {
	var genSleep GenSleep

	err := json.NewDecoder(r.Body).Decode(&genSleep)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	fmt.Println("A ver qué me mandarooon: ", genSleep.Pcb)

	IO_GEN_SLEEP(genSleep.TimeToSleep, genSleep.Pcb)
	log.Println("Fin de bloqueo")

	jsonData, err := json.Marshal(genSleep.Pcb)
	if err != nil {
		http.Error(w, "Failed to encode PCB response", http.StatusInternalServerError)
		return
	}

	fmt.Println("Te mando esto master: ", genSleep.Pcb)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}