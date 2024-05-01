package globals

import (
	"log"
	"time"
)

type T_CPU struct {
	Port               	int    `json:"port"`
	IP_memory          	string `json:"ip_memory"`
	Port_memory        	int    `json:"port_memory"`
	Number_felling_tlb 	int    `json:"number_felling_tlb"`
	Algorithm_tlb      	string `json:"algorithm_tlb"`
}

var Configcpu *T_CPU

// TODO: Mover a instrucciones
func IO_GEN_SLEEP(cantidadUnidadesTrabajo int, cantTiempoDeTrabajo int) {
	time.Sleep(time.Duration(cantTiempoDeTrabajo * cantidadUnidadesTrabajo))
	log.Println("Se cumplio el tiempo de espera")
}