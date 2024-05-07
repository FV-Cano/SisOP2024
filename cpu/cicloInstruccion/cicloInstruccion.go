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
)

type T_Instruccion struct {
	instruccion string
	parametro1  string
	parametro2  string
}

var instruccionActual T_Instruccion

func Delimitador() []string {
	var instruccion = Fetch()
	delimitador := " "
	instruccionDecodificada := strings.Split(instruccion, delimitador)
	return instruccionDecodificada
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

func CargarStruct() {
	instruccionDecodificada := Delimitador()

	instruccionActual.instruccion = instruccionDecodificada[0]
	instruccionActual.parametro1 = instruccionDecodificada[1]
	instruccionActual.parametro2 = instruccionDecodificada[2]

}

func Fetch() string {

	pc := strconv.Itoa(1) // acá debería ir pcb.T_PCB.PC

	url := fmt.Sprintf("http://%s:%d/instrucciones", globals.Configcpu.IP_memory, globals.Configcpu.Port_memory)

    cliente := &http.Client{}
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return("Error")
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
	parametro1 := Convertir(instruccionActual.parametro1)
	//var Convertir(instruccionActual.parametro2)  = Convertir(instruccion[2])

	switch instruccionActual.instruccion {
	//case "IO_GEN_SLEEP": operaciones.IO_GEN_SLEEP(instruccionActual.parametro1, instruccionActual.parametro2)
	case "JNZ":
		operaciones.JNZ(&parametro1, 4) //mandamos un 4 para probar
	case "SET":
		operaciones.SET(&parametro1, Convertir(instruccionActual.parametro2))
	case "SUM":
		operaciones.SUM(&parametro1, Convertir(instruccionActual.parametro2))
	case "SUB":
		operaciones.SUB(&parametro1, Convertir(instruccionActual.parametro2))

	}
	

}
