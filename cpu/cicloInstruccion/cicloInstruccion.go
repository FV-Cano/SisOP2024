package cicloInstruccion

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"reflect"
	"github.com/sisoputnfrba/tp-golang/cpu/globals"
	"github.com/sisoputnfrba/tp-golang/cpu/operaciones"
	"github.com/sisoputnfrba/tp-golang/utils/pcb"
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
	pid := currentPCB.PID
	pc := currentPCB.PC
	
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
	instActual := Fetch(currentPCB)
	instruccionDecodificada := Delimitador(instActual)

	parametros := currentPCB.CPU_reg

	reg1 := parametros[instruccionDecodificada[1]]
	tipoReg1 := reflect.TypeOf(reg1).String()

	currentPCB.PC++

	switch instruccionDecodificada[0] {
		case "IO_GEN_SLEEP": 
		//operaciones.IO_GEN_SLEEP(instruccionActual.parametro1, instruccionActual.parametro2)
		case "JNZ":			
			operaciones.JNZ(reg1, Convertir(tipoReg1, instruccionDecodificada[2]))

		case "SET":
			operaciones.SET(&reg1, Convertir(tipoReg1, instruccionDecodificada[2]))
		
		case "SUM":
			reg2 := parametros[instruccionDecodificada[2]]
			operaciones.SUM(&reg1, reg2)
			
		case "SUB":
			reg2 := parametros[instruccionDecodificada[2]]
			operaciones.SUB(&reg1, reg2)
	}
}

type Uint interface {~uint8 | ~uint32}
func Convertir[T Uint](tipo string, parametro string) T {
	if parametro == "" {
		log.Fatal("La cadena de texto está vacía")
	}

	switch tipo {
	case "uint8":
		valor, err := strconv.ParseUint(parametro, 10, 8)
	case "uint32":
		valor, err := strconv.ParseUint(parametro, 10, 32)
	}
	
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Conversion realizada")

	return valor
}