package IO_api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/sisoputnfrba/tp-golang/cpu/operaciones"
	"github.com/sisoputnfrba/tp-golang/entradasalida/globals"
)

type CantUnidadesTrabajo struct {
	Unidades int `json:"cantUnidades"`
}

func RecibirPeticionKernel(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var cantUnidadesTrabajo CantUnidadesTrabajo
	err := decoder.Decode(&cantUnidadesTrabajo)
	if err != nil {
		log.Printf("Error al decodificar mensaje: %s\n", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Error al decodificar mensaje"))
		return
	}

	log.Println("Me llego una petición de Kernel")
	log.Printf("%+v\n", cantUnidadesTrabajo)

	operaciones.IO_GEN_SLEEP(cantUnidadesTrabajo.Unidades, globals.ConfigIO.Unit_work_time)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Espera finalizada"))
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
