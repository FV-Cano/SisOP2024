package mmu

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/sisoputnfrba/tp-golang/cpu/globals"
	solicitudesmemoria "github.com/sisoputnfrba/tp-golang/cpu/solicitudesMemoria"
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

func PedirTamTablaPaginas(pid int) int {
	cliente := &http.Client{}
	url := fmt.Sprintf("http://%s:%d/tamTabla", globals.Configcpu.IP_memory,globals.Configcpu.Port_memory)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal("Error al hacer el request")
	}
	q := req.URL.Query()
	q.Add("pid", strconv.Itoa(int(pid)))
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
	tamTabla, err := io.ReadAll(respuesta.Body)
	if err != nil {
		log.Fatal("Error al leer el cuerpo de la respuesta")
	}

	return int(globals.BytesToInt(tamTabla))

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

	return int(globals.BytesToInt(frame))
}


func ObtenerDireccionesFisicas(direccionLogica int, tamanio int, pid int) []globals.DireccionTamanio { 
	var direccion_y_tamanio []globals.DireccionTamanio
	tamPagina := SolicitarTamPagina()
	numeroPagina := direccionLogica/tamPagina
	desplazamiento := direccionLogica - numeroPagina * tamPagina
	cantidadPaginas := tamanio/tamPagina
	frame := Frame_rcv(&globals.CurrentJob, numeroPagina) 
	tamanioTotal := frame * tamPagina + desplazamiento + tamanio
	if (tamanioTotal > PedirTamTablaPaginas(pid) * tamPagina) {
		solicitudesmemoria.Resize(tamanioTotal)
	}
	//Primer pagina teniendo en cuenta el desplazamiento
	slice.Push[globals.DireccionTamanio](&direccion_y_tamanio, globals.DireccionTamanio{frame * tamPagina + desplazamiento, tamPagina - desplazamiento})
	tamanioRestante := tamanio - (tamPagina - desplazamiento)
	for  i:= 1; i < cantidadPaginas; i++ {
		if (i == cantidadPaginas-1) {
			//Ultima pagina teniendo en cuenta el tamanio
			slice.Push[globals.DireccionTamanio](&direccion_y_tamanio, globals.DireccionTamanio{frame * tamPagina, tamanioRestante})
		} else { //Paginas del medio sin tener en cuenta el desplazamiento
			numeroPagina++
			frame = Frame_rcv(&globals.CurrentJob, direccionLogica)
			slice.Push[globals.DireccionTamanio](&direccion_y_tamanio, globals.DireccionTamanio{frame * tamPagina, tamPagina})
			tamanioRestante -= tamPagina
		}
	}
	return direccion_y_tamanio
}