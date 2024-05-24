package cicloInstruccion

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/sisoputnfrba/tp-golang/cpu/globals"
	"github.com/sisoputnfrba/tp-golang/cpu/operaciones"
	"github.com/sisoputnfrba/tp-golang/utils/pcb"
	"github.com/sisoputnfrba/tp-golang/utils/semaphores"
)

func Delimitador(instActual string) []string {
	delimitador := " "
	i:=0

	instruccionDecodificadaConComillas := strings.Split(instActual, delimitador)
	instruccionDecodificada := instruccionDecodificadaConComillas

	
	largoInstruccion := len (instruccionDecodificadaConComillas) 
	for i < largoInstruccion {
	instruccionDecodificada[i]= strings.Trim(instruccionDecodificadaConComillas[i], `"` )
	i++
	}
	
	return instruccionDecodificada
}

func Fetch(currentPCB *pcb.T_PCB) string {
	//CPU pasa a memoria el PID y el PC, y memoria le devuelve la instrucción
	//(después de identificar en el diccionario la key:PID,
	//va a buscar en la lista de instrucciones de ese proceso, la instrucción en la posición
	//pc y nos va a devolver esa instrucción)
	// GET /instrucciones
	fmt.Println("LABURASTESSSS?")
	
	semaphores.PCBMutex.Lock()
	pid := currentPCB.PID
	pc := currentPCB.PC
	semaphores.PCBMutex.Unlock()
	
	cliente := &http.Client{}
	url := fmt.Sprintf("http://%s:%d/instrucciones", globals.Configcpu.IP_memory,globals.Configcpu.Port_memory)
	
	fmt.Println("Hasta ahora tenemos: ", url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "errorisimo"
	}
	q := req.URL.Query()
	q.Add("pid", strconv.Itoa(int(pid)))
	q.Add("pc", strconv.Itoa(int(pc)))
	req.URL.RawQuery = q.Encode()

	fmt.Println("Hola que taaal: ", q)

	req.Header.Set("Content-Type", "application/json")
	respuesta, err := cliente.Do(req)
	if err != nil {
		return "horror"
	}

	fmt.Println("El cliente laburó: ", respuesta.StatusCode)

	// Verificar el código de estado de la respuesta
	if respuesta.StatusCode != http.StatusOK {
		return "herror"
	}

	instruccion, err := io.ReadAll(respuesta.Body)
	if err != nil {
		return "orror"
	}

	fmt.Println(string(instruccion))
/*
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "error"
	}

	req.Header.Set("Content-Type", "application/json")
	respuesta, err := cliente.Do(req)
	if err != nil {
		return "error"
	}

	fmt.Println("El cliente laburó: ", respuesta.StatusCode)

	// Verificar el código de estado de la respuesta
	if respuesta.StatusCode != http.StatusOK {
		return "error DE STATUS"
	}

	instruccionEnBytes, err := io.ReadAll(respuesta.Body)
	if err != nil {
		return "error REBELDE"
	}
	fmt.Println("RECIBIO EL CUERPO")*/
	
	instruccion1 := string(instruccion)
	
	fmt.Println("SE PASO A STRING LA INSTRUCCION")
	
	log.Printf("PID: %d - FETCH - Program Counter: %d", pid, pc)

	return instruccion1
}

func DecodeAndExecute(currentPCB *pcb.T_PCB) {
	// ? Semaforo?
	instActual := Fetch(currentPCB)
	instruccionDecodificada := Delimitador(instActual)

	fmt.Println("Intruc actual: ", instActual)
	fmt.Println("Intruc decod: ", instruccionDecodificada)

	parametros := currentPCB.CPU_reg
	fmt.Println("DALE QUE LLEGO: ", parametros)
	fmt.Println("ABER: ", instruccionDecodificada[1])
	
	reg1 := parametros[instruccionDecodificada[1]]
	fmt.Println("C DECODIFICO: ", reg1)
	
	tipoReg1 := reflect.TypeOf(reg1).String()
	
	reg1Uint8 := Convertir[uint8](tipoReg1, reg1)
	reg1Uint32 := Convertir[uint32](tipoReg1, reg1)

	//currentPCB.PC++
	//fmt.Println("PC AUMENTADO BRO")

	log.Printf("PID: %d - Ejecutando: %s - %s", currentPCB.PID, instruccionDecodificada[0], instruccionDecodificada[1:])
	fmt.Println("LA instruccion es" ,instruccionDecodificada[0])
	fmt.Println("Los parametros son" ,instruccionDecodificada[1:])
	
	switch instruccionDecodificada[0] {
		case "IO_GEN_SLEEP":
			tiempo_esp, err := strconv.Atoi(instruccionDecodificada[2])
			if err != nil {
				log.Fatal("Error al convertir el tiempo de espera a entero")
			}
			cond, err := existeInterfazGen()
			if err != nil {
				log.Fatal("Error al verificar la existencia de la interfaz genérica")
			}
			if cond {
				currentPCB.EvictionReason = "BLOCKED_IO"
				pcb.EvictionFlag = true
				comunicarTiempoEspera(tiempo_esp)
			} 
		case "JNZ":
			if tipoReg1 == "uint8" {
				operaciones.JNZ(reg1Uint8, Convertir[uint8](tipoReg1, instruccionDecodificada[2]))
				currentPCB.PC++
				} else {
					operaciones.JNZ(reg1Uint32, Convertir[uint32](tipoReg1, instruccionDecodificada[2]))
					currentPCB.PC++
				}
			
		case "SET":
			fmt.Println("ALOHEISHON SET")

			if tipoReg1 == "uint8" {
				operaciones.SET(&reg1Uint8, Convertir[uint8](tipoReg1, instruccionDecodificada[2]))
				currentPCB.PC++
				fmt.Println(currentPCB.PC)
				fmt.Println("PC AUMENTADO BRO1")
			} else {
				operaciones.SET(&reg1Uint32, Convertir[uint32](tipoReg1, instruccionDecodificada[2]))
				currentPCB.PC++
				fmt.Println("PC AUMENTADO BRO2")
				fmt.Println(currentPCB.PC)
			}
			
		case "SUM":
			reg2 := parametros[instruccionDecodificada[2]]
			tipoReg2 := reflect.TypeOf(reg2).String()
			reg2Uint8 := reg2.(uint8)
			reg2Uint32 := reg2.(uint32)
			
			if (tipoReg1 == "uint8" && tipoReg2 == "uint8")  {
				operaciones.SUM(&reg1Uint8, reg2Uint8)
				currentPCB.PC++
				fmt.Println("PC AUMENTADO BRO3")
			} else if (tipoReg1 == "uint32" && tipoReg2 == "uint32"){
				operaciones.SUM(&reg1Uint32, reg2Uint32)
				currentPCB.PC++
				fmt.Println("PC AUMENTADO BRO")
			} else if (tipoReg1 == "uint32" && tipoReg2 == "uint8"){
				operaciones.SUM(&reg1Uint32, reg2Uint8)
				currentPCB.PC++
				fmt.Println("PC AUMENTADO BRO")
			} else {
				operaciones.SUM(&reg2Uint8, reg2Uint32)
				currentPCB.PC++
				fmt.Println("PC AUMENTADO BRO")
			}
				
		case "SUB":
			reg2 := parametros[instruccionDecodificada[2]]
			tipoReg2 := reflect.TypeOf(reg2).String()
			reg2Uint8 := reg2.(uint8)
			reg2Uint32 := reg2.(uint32)
			
			if (tipoReg1 == "uint8" && tipoReg2 == "uint8")  {
				operaciones.SUB(&reg1Uint8, reg2Uint8)
				currentPCB.PC++
				fmt.Println("PC AUMENTADO BRO44")

			} else if (tipoReg1 == "uint32" && tipoReg2 == "uint32"){
				operaciones.SUB(&reg1Uint32, reg2Uint32)
				currentPCB.PC++
				fmt.Println("PC AUMENTADO BRO33")

			} else if (tipoReg1 == "uint32" && tipoReg2 == "uint8"){
				operaciones.SUB(&reg1Uint32, reg2Uint8)
				currentPCB.PC++
				fmt.Println("PC AUMENTADO BRO77")

			} else {
				operaciones.SUB(&reg2Uint8, reg2Uint32)
				currentPCB.PC++
				fmt.Println("PC AUMENTADO BRO99")

			}
		//Placeholder
		case "EXIT":
			currentPCB.EvictionReason = "EXIT"
			pcb.EvictionFlag = true
	}
}

type Uint interface {~uint8 | ~uint32}
func Convertir[T Uint](tipo string, parametro interface {}) T {

	if parametro == "" {
		log.Fatal("La cadena de texto está vacía")
	}

	switch tipo {
	
	case "uint8":
		valor := parametro.(uint8)
		log.Println("Conversion realizada UINT8")
		return T(valor)
	case "uint32":
		valor := parametro.(uint32)
		log.Println("Conversion realizada UINT32")
		return T(valor)
	}
	return T(0)
}

func existeInterfazGen() (bool, error) {
	jsonData, err := json.Marshal("GENERICA")
	if err != nil {
		return false, fmt.Errorf("failed to encode interface: %v", err)
	}
	
	url := fmt.Sprintf("http://%s:%d/io-gen-interface", globals.Configcpu.IP_kernel, globals.Configcpu.Port_kernel)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return false, fmt.Errorf("POST request failed. Failed to send interface: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("unexpected response status: %s", resp.Status)
	}

	var response bool
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return false, fmt.Errorf("failed to decode response: %v", err)
	}

	return response, nil
}

func comunicarTiempoEspera(tiempo_esp int) error {
	jsonData, err := json.Marshal(tiempo_esp)
	if err != nil {
		return fmt.Errorf("failed to encode interface: %v", err)
	}
	
	url := fmt.Sprintf("http://%s:%d/tiempoBloq", globals.Configcpu.IP_kernel, globals.Configcpu.Port_kernel)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("POST request failed. Failed to send interface: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected response status: %s", resp.Status)
	}

	return nil
}