package mmu

import (
	//"encoding/json"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/sisoputnfrba/tp-golang/cpu/globals"
	"github.com/sisoputnfrba/tp-golang/utils/pcb"
	"github.com/sisoputnfrba/tp-golang/utils/slice"
)


func SolicitarTamPagina() int {
	cliente := &http.Client{}
	url := fmt.Sprintf("http://%s:%d/tamPagina", globals.Configcpu.IP_memory,globals.Configcpu.Port_memory)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal("Error al hacer el request")
	}
	req.Header.Set("Content-Type", "application/json")
	respuesta, err := cliente.Do(req)
	if err != nil {
		log.Fatal("Error al hacer el request")
	}
	if respuesta.StatusCode != http.StatusOK {
		log.Fatal("Error en el estado de la respuesta")
	}
	tamPagina, err := io.ReadAll(respuesta.Body)
	if err != nil {
		log.Fatal("Error al leer el cuerpo de la respuesta")
	}
	tamPaginaEnInt, err := strconv.Atoi(string(tamPagina))
	if err != nil {
		log.Fatal("Error al hacer el request")
	}
	return tamPaginaEnInt

}

func Frame_rcv(currentPCB *pcb.T_PCB, pagina int) int {
	//Enviamos el PID y la PAGINA a memoria
	pid := currentPCB.PID
	cliente := &http.Client{}
	url := fmt.Sprintf("http://%s:%d/enviarMarco", globals.Configcpu.IP_memory,globals.Configcpu.Port_memory)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal("Error al hacer el request")
	}
	q := req.URL.Query()
	q.Add("pid", strconv.Itoa(int(pid)))
	q.Add("pagina", strconv.Itoa(pagina)) //paso la direccionLogica completa y no la página porque quien tiene el tamanio de la página es memoria
	req.URL.RawQuery = q.Encode()

	req.Header.Set("Content-Type", "application/json")
	respuesta, err := cliente.Do(req)
	if err != nil {
		log.Fatal("Error al hacer el request")
	}

	if respuesta.StatusCode != http.StatusOK {
		log.Fatal("Error en el estado de la respuesta")
	}

	//Memoria nos devuelve un frame a partir de la data enviada
	frame, err := io.ReadAll(respuesta.Body)
	if err != nil {
		log.Fatal("Error al leer el cuerpo de la respuesta")
	}

	return int(bytesToInt(frame))
}

type Direccion_y_tamanio struct {
	direccion_fisica int
	tamanio int
}

func ObtenerDireccionesFisicas(direccionLogica int, tamanio int, pid int) []Direccion_y_tamanio { 
	var direccion_y_tamanio []Direccion_y_tamanio
	tamPagina := SolicitarTamPagina()

	numeroPagina := direccionLogica/tamPagina
	frame := Frame_rcv(&globals.CurrentJob, numeroPagina)
	
	desplazamiento := direccionLogica - numeroPagina*tamPagina
	cantidadPaginas := tamanio/tamPagina

	if (desplazamiento != 0){
		for i := 0; i < cantidadPaginas; i++ {
			slice.Push[Direccion_y_tamanio](&direccion_y_tamanio, Direccion_y_tamanio{frame * tamPagina, tamPagina - desplazamiento})
			numeroPagina++
			frame = Frame_rcv(&globals.CurrentJob, numeroPagina)
		}
	} else {
		for  i:= 1; i < cantidadPaginas; i++ {
			slice.Push[Direccion_y_tamanio](&direccion_y_tamanio, Direccion_y_tamanio{frame * tamPagina, tamPagina})
			numeroPagina++
			frame = Frame_rcv(&globals.CurrentJob, direccionLogica + tamPagina)
		}
	}
	return direccion_y_tamanio
}


func bytesToInt(b []byte) uint32 {
    return binary.BigEndian.Uint32(b)
}
