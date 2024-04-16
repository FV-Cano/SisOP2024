package main

import (
	"log"

	cfg "github.com/sisoputnfrba/tp-golang/utils/config"
	logger "github.com/sisoputnfrba/tp-golang/utils/log"
	"github.com/sisoputnfrba/tp-golang/utils/server-Functions"
)

type T_ConfigMemory struct {
	Port 				int 	`json:"port"`
	Memory_size 		int 	`json:"memory_size"`
	Page_size		 	int 	`json:"page_size"`
	Instructions_path 	string 	`json:"instructions_path"`
	Delay_response 		int 	`json:"delay_response"`
}

var configmemory T_ConfigMemory

func main() {
	// Iniciar loggers
	logger.ConfigurarLogger("memory.log")
	logger.LogfileCreate("memory_debug.log")

	// Inicializamos la config
	err := cfg.ConfigInit("config-memory.json", &configmemory)
	if err != nil {
		log.Fatalf("Error al cargar la configuracion %v", err)
	}
	log.Println("Configuracion MEMORIA cargada")
	// Handlers

	// Iniciar servidor

	go server.ServerStart(configmemory.Port)

	select {}
}