package cicloinstruccion

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/sisoputnfrba/tp-golang/cpu/operaciones"
)

func Fetch() string {

	pc := strconv.Itoa(5) // acá debería ir pcb.T_PCB.PC
	
	cliente := &http.Client{}
	req, err := http.NewRequest("GET", "http://localhost:8002/instrucciones" , nil)
	if err != nil {
		 fmt.Printf("Error") //revisar
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

	return string(bodyBytes)
}

func Decode(){
	var instruccion  =  Delimitador() 
	switch instruccion[0] {	
		case "IO_GEN_SLEEP": operaciones.IO_GEN_SLEEP((instruccion[2]),instruccion[2])
		case "SET": 		 operaciones.SET(instruccion[1],instruccion[2])
		case "SUM": 		 operaciones.SUM(instruccion[1],instruccion[2])
		case "SUB": 		 operaciones.SUB(instruccion[1],instruccion[2])
		case "JNZ": 		 operaciones.JNZ(instruccion[1],instruccion[2])
	}

}

func Delimitador() []string {
	var instruccion  =  Fetch()
	delimitador := ""
	instruccionDecodificada := strings.Split(instruccion, delimitador)
	return instruccionDecodificada
}