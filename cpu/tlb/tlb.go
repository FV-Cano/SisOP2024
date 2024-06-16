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
		/**Lista “jenga” con números de págs -> con cada referencia se coloca (o se mueve, si ya existe) la pág al final de la lista.
		 Se elige como víctima la primera de la lista.*/
	
		if !BuscarEnTLB(pid, pagina) { // La página no está en la TLB
			if len(CurrentTLB) < globals.Configcpu.Number_felling_tlb { // Hay lugar en la TLB
				CurrentTLB = append(CurrentTLB, map[int]Pagina_marco{pid: {Pagina: pagina, Marco: marco}})
			} else { // No hay lugar en la TLB, se reemplaza la página menos recientemente utilizada
				CurrentTLB = append(CurrentTLB[1:], map[int]Pagina_marco{pid: {Pagina: pagina, Marco: marco}})
			}
		} else { // La página está en la TLB, se mueve al final de la lista
			var indice int
			for i := range CurrentTLB {
				if BuscarEnTLB(pid, pagina) {
					indice = i				//indica el valor de la lista de mápas en donde se encuentra la pagina
					break
				}
			}
			CurrentTLB = append(CurrentTLB[:indice], CurrentTLB[indice+1:]...)
			CurrentTLB = append(CurrentTLB, map[int]Pagina_marco{pid: {Pagina: pagina, Marco: marco}})
		}

		// Imprimir la TLB
		fmt.Println("LA TLB QUEDO ASI: ")
		for i := range CurrentTLB {
			fmt.Println(CurrentTLB[i])
		}
		/* if !BuscarEnTLB(pid, pagina) { // Si la página no está en la TLB
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
		*/

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
