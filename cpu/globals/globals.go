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

type T_Registers struct {
	PC 	uint32 	`json:"pc"`
	AX 	uint8 	`json:"ax"`
	BX 	uint8 	`json:"bx"`
	CX 	uint8 	`json:"cx"`
	DX 	uint8 	`json:"dx"`
	EAX uint32 	`json:"eax"`
	EBX uint32 	`json:"ebx"`
	ECX uint32 	`json:"ecx"`
	EDX uint32 	`json:"edx"`
	SI 	uint32 	`json:"si"`
	DI 	uint32 	`json:"di"`
}

var Configcpu *T_CPU

// TODO: Mover a instrucciones
func IO_GEN_SLEEP(cantidadUnidadesTrabajo int, cantTiempoDeTrabajo int) {
	time.Sleep(time.Duration(cantTiempoDeTrabajo * cantidadUnidadesTrabajo))
	log.Println("Se cumplio el tiempo de espera")
}