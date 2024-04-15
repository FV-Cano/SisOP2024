package main

import (
	"log"

	
	cfg "github.com/sisoputnfrba/tp-golang/utils/config"
	datasend "github.com/sisoputnfrba/tp-golang/utils/data-send"
)

type T_CPU struct {
	Port 				int 	`json:"port"`
	IP_memory 			string 	`json:"ip_memory"`
	Port_memory 		int 	`json:"port_memory"`
	Number_felling_tlb 	int 	`json:"number_felling_tlb"`
	Algorithm_tlb 		string 	`json:"algorithm_tlb"`
	Message				string	`json:"Message"`
}

var configcpu T_CPU

func main() {
	
	err := cfg.ConfigInit("config_cpu.json", &configcpu)

	if err != nil {
		log.Fatalf("Error al cargar la configuracion %v", err)
	}

	log.Printf(("Algoritmo de reemplazo de TLB: %s"), configcpu.Algorithm_tlb)

	// loggeamos el valor de la config
	log.Println(configcpu.Message)

	// ADVERTENCIA: Antes de continuar, tenemos que asegurarnos que el servidor esté corriendo para poder conectarnos a él

	// enviar un mensaje al servidor con el valor de la config
	//TODO PONER globals.ClientConfig.Port_memory
	datasend.EnviarMensaje(configcpu.IP_memory, configcpu.Port, configcpu.Message)
	
	// generamos un paquete y lo enviamos al servidor
	datasend.GenerarYEnviarPaquete(configcpu.IP_memory, configcpu.Port)

}
