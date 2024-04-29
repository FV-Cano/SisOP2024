package main

import (

	"encoding/json"
	"log"
	"net/http"
	"time"

	client "github.com/sisoputnfrba/tp-golang/utils/client-Functions"
	logger "github.com/sisoputnfrba/tp-golang/utils/log"

	cfg "github.com/sisoputnfrba/tp-golang/utils/config"
)

type T_ConfigIO struct {
	Port               int    `json:"port"`
	Type               string `json:"type"`
	Unit_work_time     int    `json:"unit_work_time"`
	Ip_kernel          string `json:"ip_kernel"`
	Port_kernel        int    `json:"port_kernel"`
	Ip_memory          string `json:"ip_memory"`
	Port_memory        int    `json:"port_memory"`
	Dialfs_path        string `json:"dialfs_path"`
	Dialfs_block_size  int    `json:"dialfs_block_size"`
	Dialfs_block_count int    `json:"dialfs_block_count"`
}

var configio T_ConfigIO

// TODO, este struct va en el globals, en Kernel hay que desarrollar
// una funcion que codifique las unidades de trabajo en un json
type CantUnidadesTrabajo struct {
	Unidades int `json:"cantUnidades"`
}


func main() {
	// Iniciar loggers
	logger.ConfigurarLogger("io.log")
	logger.LogfileCreate("io_debug.log")

	// Inicializar config
	err := cfg.ConfigInit("config-io.json", &configio)
	if err != nil {
		log.Fatalf("Error al cargar la configuracion %v", err)
	}
	log.Printf("Configuración IO cargada")

	client.EnviarMensaje(configio.Ip_kernel, configio.Port_kernel, "Saludo kernel desde IO")
	client.EnviarMensaje(configio.Ip_memory, configio.Port_memory, "Saludo memoria desde IO")


}

func IO_GEN_SLEEP(cantidadUnidadesTrabajo int) {
	time.Sleep(time.Duration(configio.Unit_work_time * cantidadUnidadesTrabajo))
	log.Println("Se cumplio el tiempo de espera")
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

	IO_GEN_SLEEP(cantUnidadesTrabajo.Unidades)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Espera finalizada"))
}


