package solicitudesmemoria

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/sisoputnfrba/tp-golang/cpu/globals"
)

// Peticion para RESIZE de memoria (DESDE CPU A MEMORIA)
func Resize(tamanio int) {
	cliente := &http.Client{}
	url := fmt.Sprintf("http://%s:%d/resize", globals.Configcpu.IP_memory,globals.Configcpu.Port_memory)
	req, err := http.NewRequest("PATCH", url, nil)
	if err != nil {
		return 
	}

	q := req.URL.Query()
	q.Add("tamanio", "tamanio")
	req.URL.RawQuery = q.Encode()

	req.Header.Set("Content-Type", "application/json")
	respuesta, err := cliente.Do(req)
	if err != nil {
		return 
	}

	// Verificar el código de estado de la respuesta
	if respuesta.StatusCode != http.StatusOK {
		return 
	}

	bodyBytes, err := io.ReadAll(respuesta.Body)
	if err != nil {
		return 	
	}
	//En caso de que la respuesta de la memoria sea Out of Memory, se deberá devolver el contexto de ejecución al Kernel informando de esta situación
	// Y Avisar que el error es por out of memory
	var respuestaResize = string(bodyBytes)
	if respuestaResize != "OK" {
		globals.CurrentJob.EvictionReason = "OUT OF MEMORY"
		// falta algo mas?
	} 
}


type BodyRequestEscribir struct {
	DireccionesTamanios []globals.DireccionTamanio `json:"direcciones_tamanios"`
	Valor_a_escribir    string             `json:"valor_a_escribir"`
	Pid                 int                `json:"pid"`
}

func SolicitarEscritura(direccionesTamanios []globals.DireccionTamanio, valorAEscribir string, pid int) {
	body, err := json.Marshal(BodyRequestEscribir{
		DireccionesTamanios : direccionesTamanios,
		Valor_a_escribir    : valorAEscribir,
		Pid                 : pid,
	})          
	if err != nil {
		return
	}

	cliente := &http.Client{}
	url := fmt.Sprintf("http://%s:%d/write", globals.Configcpu.IP_memory, globals.Configcpu.Port_memory)
	escribirEnMemoria, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
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


type BodyRequestLeer struct {
	DireccionesTamanios []globals.DireccionTamanio `json:"direcciones_tamanios"`
}

// LE SOLICITO A MEMORIA LEER Y DEVOLVER LO QUE ESTÉ EN LA DIREC FISICA INDICADA
func SolicitarLectura(direccionesFisicas []globals.DireccionTamanio) string {

	jsonDirecYTamanio, err := json.Marshal(BodyRequestLeer{
		DireccionesTamanios: direccionesFisicas,
	})
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
		return "Error al realizar la lectura"	
	}

	bodyBytes, err := io.ReadAll(respuesta.Body)
	if err != nil {
		return "error"
	}

	return string(bodyBytes)
}