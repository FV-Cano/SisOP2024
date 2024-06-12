package mmu

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"unsafe"

	"github.com/sisoputnfrba/tp-golang/cpu/cicloInstruccion"
	"github.com/sisoputnfrba/tp-golang/cpu/globals"
	"github.com/sisoputnfrba/tp-golang/utils/pcb"
)

// ENTIENDO QUE UNA VEZ QUE SOLUCIONEMOS LO DE LAS DIRECS FISICAS TODOS

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
		return "OK"
		//hacer esto
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




type BodyRequestLeer struct {
	DireccionesTamanios []DireccionTamanio `json:"direcciones_tamanios"`
}


type BodyRequestEscribir struct {
	DireccionesTamanios []DireccionTamanio `json:"direcciones_tamanios"`
	Valor_a_escribir    string             `json:"valor_a_escribir"`
	Pid                 int                `json:"pid"`
}

func DecodeAndExecute(currentPCB *pcb.T_PCB) {
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
//(Registro Dirección, Registro Datos): Lee el valor del Registro Datos y lo escribe en la dirección física de memoria 
//obtenida a partir de la Dirección Lógica almacenada en el Registro Dirección.
		tamanio := int(unsafe.Sizeof(currentPCB.CPU_reg[instruccionDecodificada[2]])) //ver de usar el switch que tenemos en globals

		direc_logica, ok := currentPCB.CPU_reg[instruccionDecodificada[1]].(int)
		if !ok {
			log.Fatalf("Error: el valor en el registro no es de tipo string")
		}

		direcsFisicas := ObtenerDireccionesFisicas(direc_logica, tamanio, int(currentPCB.PID))

		valor, ok := currentPCB.CPU_reg[instruccionDecodificada[2]].(string)
		if !ok {
					log.Fatalf("Error: el valor en el registro no es de tipo string")
		}

		SolicitarEscritura(direcsFisicas, valor, int(currentPCB.PID)) //([direccion fisica y tamanio], valorAEscribir, pid
			
		currentPCB.PC++

		//----------------------------------------------------------------------------

		// MOV_IN (Registro Datos, Registro Dirección): Lee el valor
		// de memoria correspondiente a la Dirección Lógica que se encuentra
		// en el Registro Dirección y lo almacena en el Registro Datos.

	case "MOV_IN":

		tamanio := int(unsafe.Sizeof(currentPCB.CPU_reg[instruccionDecodificada[1]]))

		direc_logica, ok := currentPCB.CPU_reg[instruccionDecodificada[2]].(int)
		if !ok {
			log.Fatalf("Error: el valor en el registro no es de tipo string")
		}

		direcsFisicas := ObtenerDireccionesFisicas(direc_logica, tamanio, int(currentPCB.PID))

			datos := SolicitarLectura(direcsFisicas)
			currentPCB.CPU_reg[instruccionDecodificada[1]] = datos

		currentPCB.PC++
		//-----------------------------------------------------------------------------
		//COPY_STRING (Tamaño): Toma del string apuntado por el registro SI y
		//copia la cantidad de bytes indicadas en el parámetro tamaño a la
		//posición de memoria apuntada por el registro DI.

	case "COPY_STRING":
		tamanio := PasarAInt(instruccionDecodificada[1])

		direc_logicaSI, ok := currentPCB.CPU_reg["SI"].(int)
		if !ok {
			log.Fatalf("Error: el valor en el registro no es de tipo string")
		}

		direcsFisicasSI := ObtenerDireccionesFisicas(direc_logicaSI, tamanio, int(currentPCB.PID))
		datos := SolicitarLectura(direcsFisicasSI)

		direc_logicaDI, ok := currentPCB.CPU_reg["DI"].(int)
		if !ok {
			log.Fatalf("Error: el valor en el registro no es de tipo string")
		}

		direcsFisicasDI := ObtenerDireccionesFisicas(direc_logicaDI, tamanio, int(currentPCB.PID))

		SolicitarEscritura(direcsFisicasDI, datos, int(currentPCB.PID)) //([direccion fisica y tamanio], valorAEscribir, pid)
		
		currentPCB.PC++

}
}




// -------------------------------------------------------------------------------
// LE SOLICITO A MEMORIA ESCRIBIR EN LA DIRECCION FISICA INDICADA


func SolicitarEscritura(direccionesTamanios []DireccionTamanio  , valorAEscribir string, pid int) {

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

// -------------------------------------------------------------------------------
// LE SOLICITO A MEMORIA LEER Y DEVOLVER LO QUE ESTÉ EN LA DIREC FISICA INDICADA
func SolicitarLectura(direccionesFisicas []DireccionTamanio ) string {

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

func PasarAInt(cadena string) int {
	num, err := strconv.Atoi(cadena)
	if err != nil {
		fmt.Println("Error: ", err)
	}
	return num
}
