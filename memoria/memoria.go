package main

import (
	"log"

	cfg "github.com/sisoputnfrba/tp-golang/utils/config"
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
	err := cfg.ConfigInit("config_memory.json", &configmemory)
	if err != nil {
		log.Fatalf("Error al cargar la configuracion %v", err)
	}

	log.Printf(("Tamanio de la memoria: %d"), configmemory.Memory_size)
}