package cicloInstruccion

import (
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
	instruccionDecodificada := strings.Split(instActual, delimitador)
	return instruccionDecodificada
}

func Fetch(currentPCB pcb.T_PCB) string {
	//CPU pasa a memoria el PID y el PC, y memoria le devuelve la instrucción
	//(después de identificar en el diccionario la key:PID,
	//va a buscar en la lista de instrucciones de ese proceso, la instrucción en la posición
	//pc y nos va a devolver esa instrucción)
	// GET /instrucciones/{pid}/{pc}
	semaphores.PCBMutex.Lock()
	pid := currentPCB.PID
	pc := currentPCB.PC
	semaphores.PCBMutex.Unlock()
	
	url := fmt.Sprintf("http://%s:%d/instrucciones/%d/%d", globals.Configcpu.IP_memory,globals.Configcpu.Port_memory, pid, pc)
	
	cliente := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "error"
	}

	req.Header.Set("Content-Type", "application/json")
	respuesta, err := cliente.Do(req)
	if err != nil {
		return "error"
	}

	// Verificar el código de estado de la respuesta
	if respuesta.StatusCode != http.StatusOK {
		return "error"
	}

	instruccionEnBytes, err := io.ReadAll(respuesta.Body)
	if err != nil {
		return "error"
	}

	instruccion := string(instruccionEnBytes)

	log.Print(instruccion)

	return instruccion
}

func DecodeAndExecute(currentPCB pcb.T_PCB) {
	// ? Semaforo?
	instActual := Fetch(currentPCB)

	instruccionDecodificada := Delimitador(instActual)

	semaphores.PCBMutex.Lock()
	parametros := currentPCB.CPU_reg
	defer semaphores.PCBMutex.Unlock()

	reg1 := parametros[instruccionDecodificada[1]]
	tipoReg1 := reflect.TypeOf(reg1).String()
	reg1Uint8 := reg1.(uint8)
	reg1Uint32 := reg1.(uint32)

	semaphores.PCBMutex.Lock()
	currentPCB.PC++
	defer semaphores.PCBMutex.Unlock()

	switch instruccionDecodificada[0] {
		case "IO_GEN_SLEEP": 
		//operaciones.IO_GEN_SLEEP(instruccionActual.parametro1, instruccionActual.parametro2)
		case "JNZ":
			globals.OperationMutex.Lock()
			defer globals.OperationMutex.Unlock()
			if tipoReg1 == "uint8" {
					operaciones.JNZ(reg1Uint8, Convertir[uint8](tipoReg1, instruccionDecodificada[2]))
			} else {
					operaciones.JNZ(reg1Uint32, Convertir[uint32](tipoReg1, instruccionDecodificada[2]))
			}
			
		case "SET":
			globals.OperationMutex.Lock()
			defer globals.OperationMutex.Unlock()
			if tipoReg1 == "uint8" {
				operaciones.SET(&reg1Uint8, Convertir[uint8](tipoReg1, instruccionDecodificada[2]))
			} else {
				operaciones.SET(&reg1Uint32, Convertir[uint32](tipoReg1, instruccionDecodificada[2]))
			}
			
		case "SUM":
			reg2 := parametros[instruccionDecodificada[2]]
			tipoReg2 := reflect.TypeOf(reg2).String()
			reg2Uint8 := reg2.(uint8)
			reg2Uint32 := reg2.(uint32)
			
			globals.OperationMutex.Lock()
			defer globals.OperationMutex.Unlock()
			if (tipoReg1 == "uint8" && tipoReg2 == "uint8")  {
				operaciones.SUM(&reg1Uint8, reg2Uint8)
			} else if (tipoReg1 == "uint32" && tipoReg2 == "uint32"){
				operaciones.SUM(&reg1Uint32, reg2Uint32)
			} else if (tipoReg1 == "uint32" && tipoReg2 == "uint8"){
				operaciones.SUM(&reg1Uint32, reg2Uint8)
			} else {
				operaciones.SUM(&reg2Uint8, reg2Uint32)
			}
				
		case "SUB":
			reg2 := parametros[instruccionDecodificada[2]]
			tipoReg2 := reflect.TypeOf(reg2).String()
			reg2Uint8 := reg2.(uint8)
			reg2Uint32 := reg2.(uint32)
			
			globals.OperationMutex.Lock()
			defer globals.OperationMutex.Unlock()
			if (tipoReg1 == "uint8" && tipoReg2 == "uint8")  {
				operaciones.SUB(&reg1Uint8, reg2Uint8)
			} else if (tipoReg1 == "uint32" && tipoReg2 == "uint32"){
				operaciones.SUB(&reg1Uint32, reg2Uint32)
			} else if (tipoReg1 == "uint32" && tipoReg2 == "uint8"){
				operaciones.SUB(&reg1Uint32, reg2Uint8)
			} else {
				operaciones.SUB(&reg2Uint8, reg2Uint32)
			}
		//Placeholder
		case "EXIT":
			semaphores.PCBMutex.Lock()
			defer semaphores.PCBMutex.Unlock()
			currentPCB.EvictionReason = "EXIT"

			globals.EvictionMutex.Lock()
			defer globals.EvictionMutex.Unlock()
			pcb.EvictionFlag = true
	}
}

type Uint interface {~uint8 | ~uint32}
func Convertir[T Uint](tipo string, parametro string) T {

	if parametro == "" {
		log.Fatal("La cadena de texto está vacía")
	}
	var valor uint64
	var err error

	switch tipo {
	
	case "uint8":
		valor, err = strconv.ParseUint(parametro, 10, 8)
		if err != nil {
			log.Fatal(err)
		}
	case "uint32":
		valor, err = strconv.ParseUint(parametro, 10, 32)
		if err != nil {
			log.Fatal(err)
		}
	}

	log.Println("Conversion realizada")
	return T(valor)
}