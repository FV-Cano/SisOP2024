package main

import (
	"log"

	client "github.com/sisoputnfrba/tp-golang/utils/client-Functions"
	cfg "github.com/sisoputnfrba/tp-golang/utils/config"
	logger "github.com/sisoputnfrba/tp-golang/utils/log"
	server "github.com/sisoputnfrba/tp-golang/utils/server-Functions"
)

type T_CPU struct {
	Port               int    `json:"port"`
	IP_memory          string `json:"ip_memory"`
	Port_memory        int    `json:"port_memory"`
	Number_felling_tlb int    `json:"number_felling_tlb"`
	Algorithm_tlb      string `json:"algorithm_tlb"`
}

var configcpu T_CPU

func main() {
	// Iniciar loggers
	logger.ConfigurarLogger("cpu.log")
	logger.LogfileCreate("cpu_debug.log")

	// *** CONFIGURACION ***
	err := cfg.ConfigInit("config-cpu.json", &configcpu)
	if err != nil {
		log.Fatalf("Error al cargar la configuracion %v", err)
	}
	log.Println("Configuracion CPU cargada")

	// *** SERVIDOR ***
	go server.ServerStart(configcpu.Port)

	// *** CLIENTE ***
	
	log.Println("Enviando mensaje al servidor")

	client.EnviarMensaje(configcpu.IP_memory, configcpu.Port_memory, "Saludo memoria desde CPU")

	select {}
}