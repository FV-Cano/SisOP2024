package mmu

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/sisoputnfrba/tp-golang/cpu/cicloInstruccion"
	"github.com/sisoputnfrba/tp-golang/cpu/globals"
	"github.com/sisoputnfrba/tp-golang/utils/pcb"
)

type BodyRequestEscribir struct {
	Direccion_fisica int
	Valor_a_escribir string
	Desplazamiento   int
}

type BodyRequestLeer struct {
	Direccion_fisica int
	Tamanio          int
}

//hago este archivo para que no se rompa nada si hago pull de cpu, pero va para cpu.api!!!
//peticion para RESIZE de memoria (DESDE CPU A MEMORIA)

func Resize(tamaño int, w http.ResponseWriter, r *http.Request) string {
	cliente := &http.Client{}
	url := fmt.Sprintf("http://%s:%d/resize", globals.Configcpu.IP_memory, globals.Configcpu.Port_memory)
	req, err := http.NewRequest("PATCH", url, nil)
	if err != nil {
		return "Error al hacer el request"
	}
	tamañoEnString := strconv.Itoa(tamaño)
	pid := strconv.Itoa(int(globals.CurrentJob.PID))

	q := req.URL.Query()
	q.Add("tamaño", tamañoEnString)
	q.Add("pid", pid)
	req.URL.RawQuery = q.Encode()

	req.Header.Set("Content-Type", "application/json")
	respuesta, err := cliente.Do(req)
	if err != nil {
		return "Error al hacer el request"
	}
	// Verificar el código de estado de la respuesta
	if respuesta.StatusCode != http.StatusOK {
		return "Error en el estado de la respuesta"
	}
	bodyBytes, err := io.ReadAll(respuesta.Body)
	if err != nil {
		return "Error al leer el cuerpo de la respuesta"
	}

	//En caso de que la respuesta de la memoria sea Out of Memory, se deberá devolver el contexto de ejecución al Kernel informando de esta situación
	// Y Avisar que el error es por out of memory
	var respuestaResize = string(bodyBytes)
	if respuestaResize != "OK" {
		Return_EC(w, r)
		return "Error: Out of memory"
	} else {
		return respuestaResize
	}
}

// Se ocupa de enviar el pcb actual
// TODO: solicitud desde kernel, o puedo mandarlo como cliente?

func Return_EC(w http.ResponseWriter, r *http.Request) {

	contextoEjecucion := globals.CurrentJob

	respuesta, err := json.Marshal(&contextoEjecucion)
	if err != nil {
		http.Error(w, "Error al codificar los datos como JSON", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(respuesta)

}

//-----------------------------------------------------------------------
//MOV_OUT (Registro Dirección, Registro Datos):
// Lee el valor del Registro Datos y lo escribe en la
// dirección física de memoria obtenida a partir de la Dirección
//	Lógica almacenada en el Registro Dirección.

func DecodeAndExecute(w http.ResponseWriter, r *http.Request, currentPCB *pcb.T_PCB) {
	instActual := cicloInstruccion.Fetch(currentPCB)
	instruccionDecodificada := cicloInstruccion.Delimitador(instActual)

	if instruccionDecodificada[0] == "EXIT" {
		currentPCB.EvictionReason = "EXIT"
		pcb.EvictionFlag = true

		log.Printf("PID: %d - Ejecutando: %s", currentPCB.PID, instruccionDecodificada[0])
	} else {
		log.Printf("PID: %d - Ejecutando: %s - %s", currentPCB.PID, instruccionDecodificada[0], instruccionDecodificada[1:])
	}

	switch instruccionDecodificada[0] {
	case "MOV_OUT":
		//TODO,DE DONDE SALE EL TAM?*/
		direcsFisicas := ObtenerDireccionesFisicas(PasarAInt(instruccionDecodificada[1]), 5, int(currentPCB.PID))
		cantDirecciones := len(direcsFisicas)

		for i := 0; i < cantDirecciones; i++ {
			tamanio := currentPCB.CPU_reg["DI"] - currentPCB.CPU_reg["SI"]
			// a chequear lo de tamanio y el error que tira
			requestBody := BodyRequestEscribir{
				Direccion_fisica: direcsFisicas[i].direccion_fisica,
				Valor_a_escribir: instruccionDecodificada[2],
				Desplazamiento:   tamanio,
			}
			SolicitarEscritura(w, r, requestBody)

		}
		currentPCB.PC++
		//----------------------------------------------------------------------------

		// MOV_IN (Registro Datos, Registro Dirección): Lee el valor
		// de memoria correspondiente a la Dirección Lógica que se encuentra
		// en el Registro Dirección y lo almacena en el Registro Datos.

	case "MOV_IN":

		direcsFisicas := ObtenerDireccionesFisicas(PasarAInt(instruccionDecodificada[2]), 5, int(currentPCB.PID))
		cantDirecciones := len(direcsFisicas)

		for i := 0; i < cantDirecciones; i++ {
			tamanio := currentPCB.CPU_reg["DI"] - currentPCB.CPU_reg["SI"]
			// a chequear lo de tamanio
			requestBody := BodyRequestLeer{
				Direccion_fisica: direcsFisicas[i].direccion_fisica,
				Tamanio:          tamanio,
			}
			datos := SolicitarLectura(w, r, requestBody)
			currentPCB.CPU_reg[instruccionDecodificada[1]] = datos
		}

		currentPCB.PC++
	}

}

// -------------------------------------------------------------------------------
// LE SOLICITO A MEMORIA ESCRIBIR EN LA DIRECCION FISICA INDICADA
func SolicitarEscritura(w http.ResponseWriter, r *http.Request, requestBody BodyRequestEscribir) {

	bodyEscritura, err := json.Marshal(requestBody)
	if err != nil {
		return
	}

	cliente := &http.Client{}
	url := fmt.Sprintf("http://%s:%d/write", globals.Configcpu.IP_memory, globals.Configcpu.Port_memory)
	escribirEnMemoria, err := http.NewRequest("POST", url, bytes.NewBuffer(bodyEscritura))
	if err != nil {
		return
	}

	escribirEnMemoria.Header.Set("Content-Type", "application/json")
	respuesta, err := cliente.Do(escribirEnMemoria)
	if err != nil {
		fmt.Println("error")
	}

	if respuesta.StatusCode != http.StatusOK {
		fmt.Println("Error al realizar la escritura")
	}

	bodyBytes, err := io.ReadAll(respuesta.Body)
	if err != nil {
		fmt.Println("error")
	}

	// La respuesta puede ser un "Ok" o u "Error: dirección o tamanio fuera de rango"

	respuestaEnString := string(bodyBytes)
	if respuestaEnString != "OK" {
		fmt.Println("Se produjo un error al escribir", respuestaEnString)
	} else {
		fmt.Println("Se realizó la escritura correctamente", respuestaEnString)
	}

}

// -------------------------------------------------------------------------------
// LE SOLICITO A MEMORIA LEER Y DEVOLVER LO QUE ESTÉ EN LA DIREC FISICA INDICADA
func SolicitarLectura(w http.ResponseWriter, r *http.Request, requestBody BodyRequestLeer) string {

	jsonDirecYTamanio, err := json.Marshal(requestBody)
	if err != nil {
		return "error"
	}

	cliente := &http.Client{}
	url := fmt.Sprintf("http://%s:%d/read", globals.Configcpu.IP_memory, globals.Configcpu.Port_memory)
	leerMemoria, err := http.NewRequest("GET", url, bytes.NewBuffer(jsonDirecYTamanio))
	if err != nil {
		return "error"
	}

	leerMemoria.Header.Set("Content-Type", "application/json")
	respuesta, err := cliente.Do(leerMemoria)
	if err != nil {
		return "error"
	}

	if respuesta.StatusCode != http.StatusOK {
		return fmt.Sprintf("Error al realizar la lectura")
	}

	bodyBytes, err := io.ReadAll(respuesta.Body)
	if err != nil {
		return "error"
	}

	return string(bodyBytes)

}

func PasarAInt(cadena string) int {
	num, err := strconv.Atoi(cadena)
	if err != nil {
		fmt.Println("Error: ", err)
	}
	return num
}
