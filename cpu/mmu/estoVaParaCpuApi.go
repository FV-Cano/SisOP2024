package mmu

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"unsafe"

	"github.com/sisoputnfrba/tp-golang/cpu/cicloInstruccion"
	"github.com/sisoputnfrba/tp-golang/cpu/globals"
	"github.com/sisoputnfrba/tp-golang/utils/pcb"
)

// ENTIENDO QUE UNA VEZ QUE SOLUCIONEMOS LO DE LAS DIRECS FISICAS TODOS
// NUESTROS PROBLEMAS SERAN RESUELTOS
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

func Resize(tamanio int) string {
	cliente := &http.Client{}
	url := fmt.Sprintf("http://%s:%d/resize", globals.Configcpu.IP_memory,globals.Configcpu.Port_memory)
	req, err := http.NewRequest("PATCH", url, nil)
	if err != nil {
		return "Error: " + err.Error()
	}

	q := req.URL.Query()
	q.Add("tamanio", "tamanio")
	req.URL.RawQuery = q.Encode()

	req.Header.Set("Content-Type", "application/json")
	respuesta, err := cliente.Do(req)
	if err != nil {
		return "Error: " + err.Error()
	}

	// Verificar el código de estado de la respuesta
	if respuesta.StatusCode != http.StatusOK {
		return "Error: " + respuesta.Status
	}

	bodyBytes, err := io.ReadAll(respuesta.Body)
	if err != nil {
		return 	"Error: " + err.Error()	
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

		tamanio := int(unsafe.Sizeof(instruccionDecodificada[2]))

		direcsFisicas := ObtenerDireccionesFisicas(PasarAInt(instruccionDecodificada[1]), tamanio, int(currentPCB.PID))
		tipoTamanio := reflect.TypeOf(instruccionDecodificada[2]).String()

		cantDirecciones := len(direcsFisicas)

		if tipoTamanio == "uint8" {
			for i := 0; i < cantDirecciones; i++ {

				valor, ok := currentPCB.CPU_reg[instruccionDecodificada[2]].(string)
				if !ok {
					log.Fatalf("Error: el valor en el registro no es de tipo string")
				}
				requestBody := BodyRequestEscribir{
					Direccion_fisica: direcsFisicas[i].direccion_fisica,
					// revisar error
					Valor_a_escribir: valor,
					Desplazamiento:   5, /*Lo hardcodeo, Seguro no necesite este dato*/
				}

				SolicitarEscritura(w, r, requestBody)
			}

		}
		currentPCB.PC++
		//----------------------------------------------------------------------------

		// MOV_IN (Registro Datos, Registro Dirección): Lee el valor
		// de memoria correspondiente a la Dirección Lógica que se encuentra
		// en el Registro Dirección y lo almacena en el Registro Datos.

	case "MOV_IN":

		tamanio := int(unsafe.Sizeof(instruccionDecodificada[2]))

		direcsFisicas := ObtenerDireccionesFisicas(PasarAInt(instruccionDecodificada[1]), tamanio, int(currentPCB.PID))

		cantDirecciones := len(direcsFisicas)

		for i := 0; i < cantDirecciones; i++ {
			requestBody := BodyRequestLeer{
				Direccion_fisica: direcsFisicas[i].direccion_fisica,
				Tamanio:          tamanio,
			}
			datos := SolicitarLectura(w, r, requestBody)
			currentPCB.CPU_reg[instruccionDecodificada[1]] = datos

		}

		currentPCB.PC++
		//-----------------------------------------------------------------------------
		//COPY_STRING (Tamaño): Toma del string apuntado por el registro SI y
		//copia la cantidad de bytes indicadas en el parámetro tamaño a la
		//posición de memoria apuntada por el registro DI.

	case "COPY_STRING":
		tamanio := PasarAInt(instruccionDecodificada[1])

		direc_logica, ok := currentPCB.CPU_reg["DI"].(int)
		if !ok {
			log.Fatalf("Error: el valor en el registro no es de tipo string")
		}

		direcsFisicas := ObtenerDireccionesFisicas(direc_logica, tamanio, int(currentPCB.PID))
		cantDirecciones := len(direcsFisicas)

		datos_a_esribir, ok := currentPCB.CPU_reg["SI"].(string)
		if !ok {
			log.Fatalf("Error: el valor en el registro no es de tipo string")
		}

		for i := 0; i < cantDirecciones; i++ {
			requestBody := BodyRequestEscribir{
				Direccion_fisica: direcsFisicas[i].direccion_fisica,
				Valor_a_escribir: datos_a_esribir,
				Desplazamiento:   5, /*Lo hardcodeo, Seguro no necesite este dato*/
			}

			SolicitarEscritura(w, r, requestBody)
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
