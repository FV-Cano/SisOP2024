package operaciones

import (
	"log"
)

type T_Registers struct {
	PC  uint32 `json:"pc"`
	SI  uint32 `json:"si"`
	DI  uint32 `json:"di"`
}

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
func SUM[T Uint, Z Uint](registroDestino *T, registroOrigen Z) {
	registroOrigenConvertido := T(registroOrigen)

	*registroDestino = *registroDestino + registroOrigenConvertido
}

// para llamarla SUB(&registroDestino, registroOrigen)
func SUB[T Uint, Z Uint](registroDestino *T, registroOrigen Z) {
	registroOrigenConvertido := T(registroOrigen)

	if *registroDestino  >= registroOrigenConvertido {
		*registroDestino = *registroDestino - registroOrigenConvertido
	}
}

