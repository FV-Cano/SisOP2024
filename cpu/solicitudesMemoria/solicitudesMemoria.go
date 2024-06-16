package solicitudesmemoria

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/sisoputnfrba/tp-golang/cpu/globals"
)

// Peticion para RESIZE de memoria (DESDE CPU A MEMORIA)
func Resize(tamanio int) string {
	cliente := &http.Client{}
	url := fmt.Sprintf("http://%s:%d/resize", globals.Configcpu.IP_memory, globals.Configcpu.Port_memory)
	req, err := http.NewRequest("PATCH", url, nil)
	if err != nil {
		return "error"
	}

	q := req.URL.Query()
	tamanioEnString := strconv.Itoa(tamanio)
	q.Add("tamanio", tamanioEnString)
	q.Add("pid", strconv.Itoa(int(globals.CurrentJob.PID)))
	req.URL.RawQuery = q.Encode()

	req.Header.Set("Content-Type", "application/json")
	respuesta, err := cliente.Do(req)
	if err != nil {
		return "error"
	}

	// Verificar el código de estado de la respuesta
	if respuesta.StatusCode != http.StatusOK {
		return "Error al realizar la petición de resize"
	}

	bodyBytes, err := io.ReadAll(respuesta.Body)
	if err != nil {
		return "error"
	}
	//En caso de que la respuesta de la memoria sea Out of Memory, se deberá devolver el contexto de ejecución al Kernel informando de esta situación
	// Y Avisar que el error es por out of memory
	return string(bodyBytes)
}

type BodyRequestEscribir struct {
	DireccionesTamanios []globals.DireccionTamanio `json:"direcciones_tamanios"`
	Valor_a_escribir    string                     `json:"valor_a_escribir"`
	Pid                 int                        `json:"pid"`
}

func SolicitarEscritura(direccionesTamanios []globals.DireccionTamanio, valorAEscribir string, pid int) {
	body, err := json.Marshal(BodyRequestEscribir{
		DireccionesTamanios: direccionesTamanios,
		Valor_a_escribir:    valorAEscribir,
		Pid:                 pid,
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

	respuestaSinComillas := strings.Trim(respuestaEnString, `"`)

	fmt.Println("Respuesta de memoria: ", respuestaSinComillas)

	if respuestaSinComillas != "OK" {
		fmt.Println("Se produjo un error al escribir", respuestaSinComillas)
	} else {
		fmt.Println("Se realizó la escritura correctamente", respuestaSinComillas)
    
	if respuestaEnString != "\"OK\"" {
		fmt.Println("Se realizó la escritura correctamente", respuestaEnString)
	} else {
		log.Printf("PID: %d - Acción: ESCRIBIR - %s - Valor: %s", pid, DireccionesFisicasAString(direccionesTamanios) ,valorAEscribir)

	}
}
//Hacemos esta funcion para que quede prolijo loguearla en el log xd
func DireccionesFisicasAString(direccionesFisicas []globals.DireccionTamanio) string {
	var direccionesString string
	for i, direc := range direccionesFisicas {
		direccionesString += fmt.Sprintf("Dirección física número %d: %d - Tamaño: %d\n", i, direc.DireccionFisica, direc.Tamanio)
	}
	return direccionesString
}

type BodyRequestLeer struct {
	DireccionesTamanios []globals.DireccionTamanio `json:"direcciones_tamanios"`
}

// LE SOLICITO A MEMORIA LEER Y DEVOLVER LO QUE ESTÉ EN LA DIREC FISICA INDICADA
func SolicitarLectura(direccionesFisicas []globals.DireccionTamanio) []byte {

	/* jsonDirecYTamanio, err := json.Marshal(BodyRequestLeer{
		DireccionesTamanios: direccionesFisicas,
	}) */
	jsonDirecYTamanio, err := json.Marshal(direccionesFisicas)
	if err != nil {
		return []byte("error")
	}

	cliente := &http.Client{}
	url := fmt.Sprintf("http://%s:%d/read", globals.Configcpu.IP_memory, globals.Configcpu.Port_memory)
	leerMemoria, err := http.NewRequest("GET", url, bytes.NewBuffer(jsonDirecYTamanio))
	if err != nil {
		return []byte("error")
	}

	fmt.Println("Solicito lectura de memoria")

	leerMemoria.Header.Set("Content-Type", "application/json")
	respuesta, err := cliente.Do(leerMemoria)
	if err != nil {
		return []byte("error")
	}

	fmt.Println("Recibí respuesta de memoria: ", respuesta.Body)

	if respuesta.StatusCode != http.StatusOK {
		return []byte("Error al realizar la lectura")
	}

	bodyBytes, err := io.ReadAll(respuesta.Body)
	if err != nil {
		return []byte("error")
	}
	contenidoLeido := string(bodyBytes)
	//TODO nosotras no le pasamos el PID cuando lee, emtomses se lo pasamos para poder loguear?
	//log.Printf("PID: %d - Acción: LEER - Dirección Física: %s - Valor: %s", pid, DireccionesFisicasAString(direccionesTamanios) ,contenidoLeido) 

	return bodyBytes
}
