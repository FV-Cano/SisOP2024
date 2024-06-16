package mmu

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/sisoputnfrba/tp-golang/cpu/globals"
	solicitudesmemoria "github.com/sisoputnfrba/tp-golang/cpu/solicitudesMemoria"
	"github.com/sisoputnfrba/tp-golang/cpu/tlb"
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

	tamTabla, err := io.ReadAll(respuesta.Body)
	if err != nil {
		log.Fatal("Error al leer el cuerpo de la respuesta")
	}

	//fmt.Println("QUE RECIBISTE VERSION TABLA: ", globals.BytesToInt(tamTabla))
	
	tamTablaString := string(tamTabla)
	tamTablaInt := globals.PasarAInt(tamTablaString)
	fmt.Println("QUE RECIBISTE VERSION INT: ", tamTablaInt)

	return tamTablaInt

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

	frameEnString := string(frame)
	frameEnInt := globals.PasarAInt(frameEnString)
	log.Printf("PID: %d - OBTENER MARCO - Página: %d - Marco: %d", pid, pagina, frameEnInt)
	return frameEnInt
}

//------------------------------------------------------------------------------------------

func ObtenerDireccionesFisicas(direccionLogica int, tamanio int, pid int) []globals.DireccionTamanio {
	var direccion_y_tamanio []globals.DireccionTamanio
	tamPagina := SolicitarTamPagina()
	numeroPagina := direccionLogica / tamPagina
	desplazamiento := direccionLogica - numeroPagina * tamPagina 
	cantidadPaginas := tamanio / tamPagina
	var frame int
	var tamanioTotal int

	if (PedirTamTablaPaginas(pid) == 0){
		tamanioTotal = desplazamiento + tamanio 
		fmt.Println("ME METI X PRIM VEZ A LA TABLA, TAMANIO", tamanioTotal)
		} else { 
		if(tlb.BuscarEnTLB(pid, numeroPagina)){
			log.Printf("PID: %d - TLB HIT - Pagina: %d", pid, numeroPagina)
			frame = tlb.FrameEnTLB(pid, numeroPagina)
			} else { 
				log.Printf("PID: %d - TLB MISS - Pagina: %d", pid, numeroPagina)
				frame = Frame_rcv(&globals.CurrentJob, numeroPagina)
				tlb.ActualizarTLB(pid, numeroPagina, frame)
				}		
				tamanioTotal = frame * tamPagina + desplazamiento + tamanio
			}
			
	if (tamanioTotal > PedirTamTablaPaginas(pid) * tamPagina) {	
		fmt.Println("VOY A SOLICITAR RESIZE")
		solicitudesmemoria.Resize(tamanioTotal)
	}
	
	//Primer pagina teniendo en cuenta el desplazamiento
	slice.Push(&direccion_y_tamanio, globals.DireccionTamanio{DireccionFisica: frame*tamPagina + desplazamiento, Tamanio: tamPagina - desplazamiento})
	tamanioRestante := tamanio - (tamPagina - desplazamiento)
	fmt.Println("AGREGUE DIREC FISICA A LA LISTA")

	for i := 1; i < cantidadPaginas; i++ {
		if i == cantidadPaginas-1 {
			//Ultima pagina teniendo en cuenta el tamanio
			numeroPagina++
			if(tlb.BuscarEnTLB(pid, numeroPagina)){ 
				log.Printf("PID: %d - TLB HIT - Pagina: %d", pid, numeroPagina)
				frame = tlb.FrameEnTLB(pid, numeroPagina)
				fmt.Printf("BUSQUE EN TLB PARA EL PID %d LA PAG %d EL FRAME %d ", pid, numeroPagina, frame)

			} else { 
				log.Printf("PID: %d - TLB MISS - Pagina: %d", pid, numeroPagina)
				frame = Frame_rcv(&globals.CurrentJob, numeroPagina)
				tlb.ActualizarTLB(pid, numeroPagina, frame)
				fmt.Printf("Busco FRAME MEMORIA para el PID %d Y EL FRAME ES %d ",pid,frame)
			}
			slice.Push(&direccion_y_tamanio, globals.DireccionTamanio{DireccionFisica: frame * tamPagina, Tamanio: tamanioRestante})
		} else { //Paginas del medio sin tener en cuenta el desplazamiento
			numeroPagina++
			if(tlb.BuscarEnTLB(pid, numeroPagina)){
				log.Printf("PID: %d - TLB HIT - Pagina: %d", pid, numeroPagina)
				frame = tlb.FrameEnTLB(pid, numeroPagina)
				fmt.Printf("BUSQUE EN TLB PARA EL PID %d LA PAG %d EL FRAME %d ", pid, numeroPagina, frame)

			} else { 
				log.Printf("PID: %d - TLB MISS - Pagina: %d", pid, numeroPagina)
				frame = Frame_rcv(&globals.CurrentJob, numeroPagina)
				fmt.Printf("Busco FRAME MEMORIA para el PID %d Y EL FRAME ES %d ",pid,frame)
				tlb.ActualizarTLB(pid, numeroPagina, frame)
				}		
			slice.Push(&direccion_y_tamanio, globals.DireccionTamanio{DireccionFisica: frame * tamPagina, Tamanio: tamPagina})
			tamanioRestante -= tamPagina
		}
	}
	return direccion_y_tamanio
}

