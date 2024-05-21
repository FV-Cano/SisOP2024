package operaciones

import (
	"log"
)

type T_Registers struct {
	PC  uint32 `json:"pc"`
	SI  uint32 `json:"si"`
	DI  uint32 `json:"di"`
}

/* type T_CPU_reg struct {
	registers map[string]interface{}
} */

/* func NewT_CPU_reg() *T_CPU_reg {
	return &T_CPU_reg{
		registers: make(map[string]interface{}){
			"AX": uint8(0), "BX": uint8(0), "CX": uint8(0), "DX": uint8(0),
			"EAX": uint32(0), "EBX": uint32(0), "ECX": uint32(0), "EDX": uint32(0),
		}
	}
} */

type Uint interface {~uint8 | ~uint32}

func IO_GEN_SLEEP(cantidadUnidadesTrabajo int, cantTiempoDeTrabajo int) {
	
}

func JNZ[T Uint](registro T,  parametro T) {
	if 	registro != 0 {
		log.Printf("El PC de la instruccion actual es %d", parametro)
	}
	
}

// para llamarla SET(&registro, valor)
func SET[T Uint](registro *T, valor T) {
	*registro = valor
}

// para llamarla SUM(&registroDestino, registroOrigen)
func SUM[T Uint](registroDestino *T, registroOrigen T) {
	*registroDestino = *registroDestino + registroOrigen
}

// para llamarla SUB(&registroDestino, registroOrigen)
func SUB[T Uint](registroDestino *T, registroOrigen T) {
	if *registroDestino  >= registroOrigen {
		*registroDestino = *registroDestino - registroOrigen
	}
}

