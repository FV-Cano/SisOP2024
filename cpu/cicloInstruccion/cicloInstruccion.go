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

type T_Instruccion struct {
	instruccion string
	parametro1 string
	parametro2 string
}

var instruccionActual T_Instruccion

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
	parametro1 := Convertir(instruccionActual.parametro1)
	//var Convertir(instruccionActual.parametro2)  = Convertir(instruccion[2]) 
	
	
	switch instruccionActual.instruccion {	
		//case "IO_GEN_SLEEP": operaciones.IO_GEN_SLEEP(instruccionActual.parametro1, instruccionActual.parametro2)
		case "JNZ": 		 operaciones.JNZ(&parametro1, 4) //mandamos un 4 para probar
		case "SET": 		 operaciones.SET(&parametro1,Convertir(instruccionActual.parametro2) )
		case "SUM": 		 operaciones.SUM(&parametro1,Convertir(instruccionActual.parametro2) )
		case "SUB": 		 operaciones.SUB(&parametro1,Convertir(instruccionActual.parametro2) )
		
	}

}

func Delimitador() []string {
	var instruccion  =  Fetch()
	delimitador := ""
	instruccionDecodificada := strings.Split(instruccion, delimitador)
	return instruccionDecodificada
}


func Convertir(parametro string) uint32 {
	
	registro, err := strconv.Atoi(parametro)
	if err != nil {
		log.Fatal(err)
	}
	
	return uint32(registro)
}

func CargarStruct() {
	instruccionDecodificada := Delimitador()
	
	instruccionActual.instruccion = instruccionDecodificada[0]
	instruccionActual.parametro1 = instruccionDecodificada[1]
	instruccionActual.parametro2 = instruccionDecodificada[2]
	
}