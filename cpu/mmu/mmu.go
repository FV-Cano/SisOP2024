package mmu

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	cpu_api "github.com/sisoputnfrba/tp-golang/cpu/API"
	"github.com/sisoputnfrba/tp-golang/cpu/globals"
	solicitudesmemoria "github.com/sisoputnfrba/tp-golang/cpu/solicitudesMemoria"
	"github.com/sisoputnfrba/tp-golang/utils/pcb"
	"github.com/sisoputnfrba/tp-golang/utils/slice"
)

func SolicitarTamPagina() int {
	cliente := &http.Client{}
	url := fmt.Sprintf("http://%s:%d/tamPagina", globals.Configcpu.IP_memory, globals.Configcpu.Port_memory)
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
	url := fmt.Sprintf("http://%s:%d/tamTabla", globals.Configcpu.IP_memory, globals.Configcpu.Port_memory)

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
	url := fmt.Sprintf("http://%s:%d/enviarMarco", globals.Configcpu.IP_memory, globals.Configcpu.Port_memory)

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
	//TODO reemplazar estos calculos por las func creadas
	tamPagina := SolicitarTamPagina()
	numeroPagina := direccionLogica / tamPagina
	desplazamiento := direccionLogica - numeroPagina*tamPagina
	cantidadPaginas := tamanio / tamPagina
	frame := Frame_rcv(&globals.CurrentJob, numeroPagina)
	tamanioTotal := frame*tamPagina + desplazamiento + tamanio
	if tamanioTotal > PedirTamTablaPaginas(pid)*tamPagina {
		solicitudesmemoria.Resize(tamanioTotal)
	}
	//Primer pagina teniendo en cuenta el desplazamiento
	slice.Push(&direccion_y_tamanio, globals.DireccionTamanio{DireccionFisica: frame*tamPagina + desplazamiento, Tamanio: tamPagina - desplazamiento})
	tamanioRestante := tamanio - (tamPagina - desplazamiento)
	for i := 1; i < cantidadPaginas; i++ {
		if i == cantidadPaginas-1 {
			//Ultima pagina teniendo en cuenta el tamanio
			slice.Push(&direccion_y_tamanio, globals.DireccionTamanio{DireccionFisica: frame * tamPagina, Tamanio: tamanioRestante})
		} else { //Paginas del medio sin tener en cuenta el desplazamiento
			numeroPagina++
			frame = Frame_rcv(&globals.CurrentJob, direccionLogica)
			slice.Push(&direccion_y_tamanio, globals.DireccionTamanio{DireccionFisica: frame * tamPagina, Tamanio: tamPagina})
			tamanioRestante -= tamPagina
		}
	}
	return direccion_y_tamanio
}

func ObtenerDireccionesFisicasConTLB(direccionLogica int, tamanio int, pid int) []globals.DireccionTamanio {
	var direccion_y_tamanio []globals.DireccionTamanio
	tamPagina := SolicitarTamPagina()
	cantidadPaginas := (tamanio + tamPagina - 1) / tamPagina // Redondear hacia arriba
	tamanioRestante := tamanio

	for i := 0; i < cantidadPaginas; i++ {
		numeroPagina := cpu_api.ObtenerPagina(direccionLogica, i, tamPagina)
		offset := cpu_api.ObtenerOffset(direccionLogica, i, tamPagina)

		var frame int
		var direccionFisica int

		// Verificar en la TLB
		if cpu_api.BuscarEnTLB(pid, numeroPagina) {
			frame = cpu_api.FrameEnTLB(pid, numeroPagina)
			// Calcular dirección física usando la TLB
			direccionFisica = cpu_api.DireccionFisicaTLB(frame, offset, tamPagina)
		} else {
			// TLB miss, calcular dirección física como antes
			frame = Frame_rcv(&globals.CurrentJob, numeroPagina)
			direccionFisica = frame * tamPagina + offset
		}

		// Calcular el tamaño de la página actual
		var tamanioPagina int
		//Para la primera página, se ajusta según el offset.
		if i == 0 {
			tamanioPagina = tamPagina - offset
		//Para la última página, se ajusta según el tamanioRestante.
		} else if i == cantidadPaginas-1 {
			tamanioPagina = tamanioRestante
		//Para las páginas intermedias, se utiliza el tamaño completo de la página.
		} else {
			tamanioPagina = tamPagina
		}

		if tamanioPagina > tamanioRestante {
			tamanioPagina = tamanioRestante
		}

		// Agregar la dirección física y el tamaño a la lista
		slice.Push(&direccion_y_tamanio, globals.DireccionTamanio{
			DireccionFisica: direccionFisica,
			Tamanio:         tamanioPagina,
		})

		//Se reduce en cada iteración para reflejar el tamaño ya procesado.
		tamanioRestante -= tamanioPagina

		// Actualizar TLB (opcionalmente)
		// cpu_api.ActualizarTLB(pid, numeroPagina, frame)
	}

	return direccion_y_tamanio
}