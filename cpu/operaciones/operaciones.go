package operaciones

import (
	"log"
	//"time"
	"github.com/sisoputnfrba/tp-golang/utils/pcb"
)

type T_Registers struct {
	PC  uint32 `json:"pc"`
	AX  uint8  `json:"ax"`
	BX  uint8  `json:"bx"`
	CX  uint8  `json:"cx"`
	DX  uint8  `json:"dx"`
	EAX uint32 `json:"eax"`
	EBX uint32 `json:"ebx"`
	ECX uint32 `json:"ecx"`
	EDX uint32 `json:"edx"`
	SI  uint32 `json:"si"`
	DI  uint32 `json:"di"`
}

func IO_GEN_SLEEP(cantidadUnidadesTrabajo int, cantTiempoDeTrabajo int) {
	
}

// para llamarla SET(&registro, valor)
func SET(registro *uint32, valor uint32) {
	*registro = valor
}

// para llamarla SUM(&registroDestino, registroOrigen)
func SUM(registroDestino *uint32, registroOrigen uint32) {
	*registroDestino = *registroDestino + registroOrigen
}

// para llamarla SUB(&registroDestino, registroOrigen)
func SUB(registroDestino *uint32, registroOrigen uint32) {
	if *registroDestino  >= registroOrigen {
		*registroDestino = *registroDestino - registroOrigen
	}
}

var instruccionActual pcb.T_PCB //Temporal, debería ser la instruccion actual

func JNZ(registro *uint32,  parametro uint32) {//instruccion pcb.T_PCB) { //A QUE SE REFIERE CON INSTRUCCIOOOOOON
	if *registro != 0 {
		instruccionActual.PC = parametro //instruccion.PC
		log.Printf("El PC de la instruccion actual es %d", parametro)
	}
}
