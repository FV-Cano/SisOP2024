package main

import (
	"log"

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
