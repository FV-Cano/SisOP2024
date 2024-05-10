package cicloinstruccion

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/sisoputnfrba/tp-golang/cpu/globals"
	"github.com/sisoputnfrba/tp-golang/cpu/operaciones"
	"github.com/sisoputnfrba/tp-golang/utils/pcb"
)

func Delimitador() []string {
	var instruccion = Fetch()
	delimitador := " "
	instruccionDecodificada := strings.Split(instruccion, delimitador)
	return instruccionDecodificada
}

func Fetch() string {

	pc := strconv.Itoa(1) // acá debería ir pcb.T_PCB.PC

	url := fmt.Sprintf("http://%s:%d/instrucciones", globals.Configcpu.IP_memory, globals.Configcpu.Port_memory)

	cliente := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return ("Error")
	}

	q := req.URL.Query()
	q.Add("name", pc)
	req.URL.RawQuery = q.Encode()

	req.Header.Set("Content-Type", "application/json")
	respuesta, err := cliente.Do(req)
	if err != nil {
		fmt.Printf("Error") //revisar
	}

	// Verificar el código de estado de la respuesta
	if respuesta.StatusCode != http.StatusOK {
		fmt.Printf("Error") //revisar
	}

	bodyBytes, err := io.ReadAll(respuesta.Body)
	if err != nil {
		fmt.Printf("Error") //revisar
	}

	//log.Println("Fetch realizado")

	return string(bodyBytes)
}

func DecodeAndExecute() {

	instruccion := Delimitador()
	posRegistro1 := ElegirRegistro(instruccion[1])
	posRegistro2 := ElegirRegistro(instruccion[2])

	switch instruccion[0] {
	//case "IO_GEN_SLEEP": operaciones.IO_GEN_SLEEP(instruccionActual.parametro1, instruccionActual.parametro2)
	case "JNZ":
		operaciones.JNZ(pcb.pcbPrueba.CPU_reg[posRegistro1], Convertir(instruccion[2])) //mandamos un 4 para probar
	case "SET":
		operaciones.SET(&pcb.pcbPrueba.CPU_reg[posRegistro1], Convertir(instruccion[2]))
	case "SUM":
		operaciones.SUM(&pcb.pcbPrueba.CPU_reg[posRegistro1], pcb.pcbPrueba.CPU_reg[posRegistro2])
	case "SUB":
		operaciones.SUB(&pcb.pcbPrueba.CPU_reg[posRegistro1], pcb.pcbPrueba.CPU_reg[posRegistro2])

	}

}

func Convertir(parametro string) uint32 {

	if parametro == "" {
		log.Fatal("La cadena de texto está vacía")
	}

	registro, err := strconv.Atoi(parametro)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Conversion realizada")

	return uint32(registro)

}

func ElegirRegistro(registro string) int {

	switch registro {
	case "AX":   
		return 0
	case "BX":  
		return 1
	case "CX":  
		return 2
	case "DX": 
		return 3
	case "EAX": 
		return 4
	case "EBX": 
		return 5
	case "ECX": 
		return 6
	case "EDX": 
		return 7
	default:
		return -1
	}
}
