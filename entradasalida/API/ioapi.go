package IO_api

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
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

	IO_GEN_SLEEP(genSleep.TimeToSleep, genSleep.Pcb)
	log.Println("Fin de bloqueo")

	jsonData, err := json.Marshal(genSleep.Pcb)
	if err != nil {
		http.Error(w, "Failed to encode PCB response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}

func IOStdinRead(w http.ResponseWriter, r *http.Request) {
	var infoRecibida struct {
		direccionesFisicas []globals.DireccionTamanio
		tamanio int
	}
	
	err := json.NewDecoder(r.Body).Decode(&infoRecibida)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	// Lee datos de la entrada
	reader := bufio.NewReader(os.Stdin)
	data, _ := reader.ReadString('\n')

	// Le pido a memoria que me guarde los datos
	url := fmt.Sprintf("http://%s:%d/write", globals.ConfigIO.Ip_memory, globals.ConfigIO.Port_memory)

	bodyWrite, err := json.Marshal(struct {
		data string
		direccionesFisicas []globals.DireccionTamanio
		tamanio int
	} {data, infoRecibida.direccionesFisicas, infoRecibida.tamanio})
	if err != nil {
		log.Printf("Failed to encode data: %v", err)
	}

	response, err := http.Post(url, "application/json", bytes.NewBuffer(bodyWrite))
	if err != nil {
		log.Printf("Failed to send data: %v", err)
	}

	if response.StatusCode != http.StatusOK {
		log.Printf("Unexpected response status: %s", response.Status)
	}

	// Le aviso a kernel que terminé
	w.WriteHeader(http.StatusOK)
}

func IOStdoutWrite(w http.ResponseWriter, r *http.Request) {
	var direccionesRecibidas []globals.DireccionTamanio

	err := json.NewDecoder(r.Body).Decode(&direccionesRecibidas)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Le pido a memoria que me lea los datos
	url := fmt.Sprintf("http://%s:%d/read", globals.ConfigIO.Ip_memory, globals.ConfigIO.Port_memory)

	bodyRead, err := json.Marshal(direccionesRecibidas)
	if err != nil {
		log.Printf("Failed to encode data: %v", err)
	}

	datosLeidos, err := http.Post(url, "application/json", bytes.NewBuffer(bodyRead))
	if err != nil {
		log.Printf("Failed to receive data: %v", err)
	}

	if datosLeidos.StatusCode != http.StatusOK {
		log.Printf("Unexpected response status: %s", datosLeidos.Status)
	}

	// Lee los datos de la respuesta
	response, err := io.ReadAll(datosLeidos.Body)
	if err != nil {
		log.Printf("Failed to read response body: %v", err)
	}

	// Convierto los datos a string
	responseString := string(response)

	// Consumo una unidad de trabajo
	time.Sleep(time.Duration(globals.ConfigIO.Unit_work_time) * time.Millisecond)

	// Escribo los datos en la salida (los muestro por pantalla)
	writer := bufio.NewWriter(os.Stdout)
	writer.WriteString(responseString)
	writer.Flush()

	// Le aviso a kernel que terminé
	w.WriteHeader(http.StatusOK)
}