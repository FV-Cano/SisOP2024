package tlb

import (
	"fmt"

	"github.com/sisoputnfrba/tp-golang/cpu/globals"
)

type Pagina_marco struct {
	Pagina int
	Marco  int
}

type TLB []map[int]Pagina_marco

var CurrentTLB TLB
var OrderedKeys []int //mantiene el orden de las claves en la TLB

func BuscarEnTLB(pid, pagina int) bool {
	for _, entradaTLB := range CurrentTLB {
		if entry, exists := entradaTLB[pid]; exists && entry.Pagina == pagina {
			return true
		}
	}
	return false
}

func FrameEnTLB(pid int, pagina int) int {
	for _, entradaTLB := range CurrentTLB {
		if entry, exists := entradaTLB[pid]; exists && entry.Pagina == pagina {
			ActualizarTLB(pid, pagina, entry.Marco)
			return entry.Marco
		}
	}
	return -1
}

func ObtenerPagina(direccionLogica int, nroPag int, tamanio int) int {
	pagina := (direccionLogica + nroPag*tamanio) / tamanio

	return pagina
}

func ObtenerOffset(direccionLogica int, nroPag int, tamanio int) int {

	offset := (direccionLogica + nroPag*tamanio) % tamanio

	return offset
}

func CalcularDireccionFisica(frame int, offset int, tamanio int) int {

	direccionBase := frame * tamanio

	return direccionBase + offset

}

/*func ActualizarTLB(pid, pagina, marco int) {
	if globals.Configcpu.Algorithm_tlb == "FIFO" {
	if len(tlb.CurrentTLB) >= globals.Configcpu.Number_felling_tlb {
		// Si la TLB está llena, eliminar la entrada más antigua (FIFO)
		for key := range tlb.CurrentTLB {
			delete(tlb.CurrentTLB, key)
			break
		}
	}
	tlb.CurrentTLB[pid] = tlb.Pagina_marco{Pagina: pagina, Marco: marco}

	}
}*/

func ActualizarTLB(pid, pagina, marco int) {
	switch globals.Configcpu.Algorithm_tlb {
	case "FIFO":
		if !BuscarEnTLB(pid, pagina) { //Si la página no está en la tlb
			if len(CurrentTLB) < globals.Configcpu.Number_felling_tlb {
				nuevoElemento := map[int]Pagina_marco{
					pid: {Pagina: pagina, Marco: marco},
				}
				CurrentTLB = append(CurrentTLB, nuevoElemento)
				fmt.Printf("Se agregó la entrada %d a la TLB", CurrentTLB)
				fmt.Println("LA TLB QUEDO ASI: ")
				for i := range CurrentTLB {
					fmt.Println(CurrentTLB[i])
				}
			} else {
				// Remover el primer elemento (FIFO) y agregar el nuevo
				CurrentTLB = append(CurrentTLB[1:], map[int]Pagina_marco{
					pid: {Pagina: pagina, Marco: marco},
				})
				fmt.Printf("Se agregó la entrada %d a la TLB", CurrentTLB)
				fmt.Println("LA TLB QUEDO ASI: ")
				for i := range CurrentTLB {
					fmt.Println(CurrentTLB[i])
				}
			}
		}

	case "LRU":
		if !BuscarEnTLB(pid, pagina) { // Si la página no está en la TLB
			if len(CurrentTLB) < globals.Configcpu.Number_felling_tlb {
				nuevoElemento := map[int]Pagina_marco{
					pid: {Pagina: pagina, Marco: marco},
				}
				CurrentTLB = append(CurrentTLB, nuevoElemento)
				fmt.Printf("Se agregó la entrada pid: %d, pagina: %d, marco: %d a la TLB\n", pid, pagina, marco)
				fmt.Println("LA TLB QUEDO ASI: ")
				for i := range CurrentTLB {
					fmt.Println(CurrentTLB[i])
				}
			} else {
				// Remover el elemento menos recientemente utilizado
				var lruIndex int
				for i, entry := range CurrentTLB {
					for key, value := range entry {
						if key == OrderedKeys[0] && value.Pagina == pagina && value.Marco == marco {
							lruIndex = i
							break
						}
					}
				}
				// Eliminar el elemento LRU de CurrentTLB
				CurrentTLB = append(CurrentTLB[:lruIndex], CurrentTLB[lruIndex+1:]...)
				fmt.Printf("Se eliminó el elemento pid: %d, pagina: %d, marco: %d de la TLB\n", OrderedKeys[0], pagina, marco)

				// Añadir el nuevo elemento al final de CurrentTLB
				nuevoElemento := map[int]Pagina_marco{
					pid: {Pagina: pagina, Marco: marco},
				}
				CurrentTLB = append(CurrentTLB, nuevoElemento)
				fmt.Printf("Se agregó la entrada pid: %d, pagina: %d, marco: %d a la TLB\n", pid, pagina, marco)
				fmt.Println("LA TLB QUEDO ASI: ")
				for i := range CurrentTLB {
					fmt.Println(CurrentTLB[i])
				}
			}
			ActualizarOrdenDeAcceso(pid, pagina, marco) // Actualizar el orden de acceso
		}

		/*case "LRU":
		 if !BuscarEnTLB(pid, pagina) { //Si la página no está en la tlb
			if len(CurrentTLB) < globals.Configcpu.Number_felling_tlb {
				// Si la TLB no está llena, agregar la entrada
				CurrentTLB[pid] = Pagina_marco{Pagina: pagina, Marco: marco}
				OrderedKeys = append(OrderedKeys, pid) // Agregar la clave al final de la lista
			} else {
				// Si la TLB está llena, eliminar la entrada más antigua (FIFO)
				oldestKey := OrderedKeys[0] // Obtener la clave más antigua
				delete(CurrentTLB, oldestKey) // Eliminar la entrada más antigua
				OrderedKeys = OrderedKeys[1:] // Eliminar la clave más antigua de la lista
				CurrentTLB[pid] = Pagina_marco{Pagina: pagina, Marco: marco} // Agregar la nueva entrada
				OrderedKeys = append(OrderedKeys, pid) // Agregar la nueva clave al final de la lista
			}
		} else { //SI LA PAGINA YA EXISTE EN LA TLB, LLEVARLA AL FINAL DE LA LISTA
			// Eliminar la entrada existente y agregarla nuevamente
			for i, key := range OrderedKeys {
				if key == pid {
					// Eliminar la clave de la lista
					OrderedKeys = append(OrderedKeys[:i], OrderedKeys[i+1:]...)
					break
				}
			}
			CurrentTLB[pid] = Pagina_marco{Pagina: pagina, Marco: marco} // Agregar la nueva entrada
			OrderedKeys = append(OrderedKeys, pid) // Agregar la nueva clave al final de la lista*/
	}
}

func ActualizarOrdenDeAcceso(pid, pagina, marco int) {
	// Elimina la clave si ya existe
	for i, key := range OrderedKeys {
		if key == pid {
			OrderedKeys = append(OrderedKeys[:i], OrderedKeys[i+1:]...)
			break
		}
	}
	// Añade la clave al final (más recientemente utilizada)
	OrderedKeys = append(OrderedKeys, pid)

	// Actualizar o agregar la entrada en CurrentTLB
	encontrado := false
	for _, entrada := range CurrentTLB {
		if entrada[pid].Pagina == pagina && entrada[pid].Marco == marco {
			encontrado = true
			break
		}
	}
	if !encontrado {
		nuevoElemento := map[int]Pagina_marco{
			pid: {Pagina: pagina, Marco: marco},
		}
		CurrentTLB = append(CurrentTLB, nuevoElemento)
		fmt.Printf("Se agregó la entrada %d a la TLB\n", pid)
		fmt.Println("LA TLB QUEDO ASI: ")
		for i := range CurrentTLB {
			fmt.Println(CurrentTLB[i])
		}
	}
}
