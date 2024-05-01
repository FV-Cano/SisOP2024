package cpu

import (
	"log"
	"time"

)


func IO_GEN_SLEEP(cantidadUnidadesTrabajo int, cantTiempoDeTrabajo int) {
	time.Sleep(time.Duration(cantTiempoDeTrabajo * cantidadUnidadesTrabajo))
	log.Println("Se cumplio el tiempo de espera")
}