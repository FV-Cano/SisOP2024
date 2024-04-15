package main

import (
	"log"

	cfg "github.com/sisoputnfrba/tp-golang/utils/config"
)

type T_ConfigKernel struct {
	Port 				int 	`json:"port"`
	IP_memory 			string 	`json:"ip_memory"`
	Port_memory 		int 	`json:"port_memory"`
	IP_cpu 				string 	`json:"ip_cpu"`
	Port_cpu 			int 	`json:"port_cpu"`
	Planning_algorithm 	string 	`json:"planning_algorithm"`
	Quantum 			int 	`json:"quantum"`
	Multiprogramming 	int 	`json:"multiprogramming"`
}

// Resources 			list `json:"resources"`
// Resource_instances 	list `json:"resource_instances"`
// TODO check list type

var configkernel T_ConfigKernel

func main() {
    err := cfg.ConfigInit("config_kernel.json", &configkernel)
	if err != nil {
		log.Fatalf("Error al cargar la configuracion %v", err)
	}

	log.Printf(("Algoritmo de planificacion: %s"), configkernel.Planning_algorithm)
}