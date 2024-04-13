package main

import (
	"log"

	cfg "github.com/sisoputnfrba/tp-golang/utils/config"
)

type T_CPU struct {
	Port 				int 	`json:"port"`
	IP_memory 			string 	`json:"ip_memory"`
	Port_memory 		int 	`json:"port_memory"`
	Number_felling_tlb 	int 	`json:"number_felling_tlb"`
	Algorithm_tlb 		string 	`json:"algorithm_tlb"`
}

var configcpu T_CPU

func main() {
	err := cfg.ConfigInit("config_cpu.json", &configcpu)
	if err != nil {
		log.Fatalf("Error al cargar la configuracion %v", err)
	}

	log.Printf(("Algoritmo de reemplazo de TLB: %s"), configcpu.Algorithm_tlb)
}
