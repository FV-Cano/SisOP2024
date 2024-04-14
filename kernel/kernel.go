package main

import (
	"log"

	client "github.com/sisoputnfrba/tp-golang/utils/client-Functions"
	cfg "github.com/sisoputnfrba/tp-golang/utils/config"
	logger "github.com/sisoputnfrba/tp-golang/utils/log"
	server "github.com/sisoputnfrba/tp-golang/utils/server-Functions"
)

type T_ConfigKernel struct {
	Port 				int 		`json:"port"`
	IP_memory 			string 		`json:"ip_memory"`
	Port_memory 		int 		`json:"port_memory"`
	IP_cpu 				string 		`json:"ip_cpu"`
	Port_cpu 			int 		`json:"port_cpu"`
	Planning_algorithm 	string 		`json:"planning_algorithm"`
	Quantum 			int 		`json:"quantum"`
	Resources 			[]string 	`json:"resources"`
	Resource_instances 	[]int 		`json:"resource_instances"`
	Multiprogramming 	int 		`json:"multiprogramming"`
}

var configkernel T_ConfigKernel

func main() {
	// Iniciar loggers
	logger.ConfigurarLogger()
	logger.LogfileCreate("kernel.log")

	// Inicializamos la config y tomamos valores
	err := cfg.ConfigInit("config_kernel.json", &configkernel)
	if err != nil {
		log.Fatalf("Error al cargar la configuracion %v", err)
	}

	ipMemory := configkernel.IP_memory
	portMemory := configkernel.Port_memory
	ipCpu := configkernel.IP_cpu
	portCpu := configkernel.Port_cpu

	// TODO check si funciona
	go server.ServerStart(configkernel.Port)

	client.EnviarMensaje(ipMemory, portMemory, "Saludo memoria desde Kernel")
	client.EnviarMensaje(ipCpu, portCpu, "Saludo cpu desde Kernel")
}