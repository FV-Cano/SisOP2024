package cicloinstruccion

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
)

func Fetch(){

	pc := strconv.Itoa(5) // acá debería ir pcb.T_PCB.PC
	
	cliente := &http.Client{}
	req, err := http.NewRequest("GET", "http://localhost:8002/instrucciones" , nil)
	if err != nil {
		return 
	}
	q := req.URL.Query()
	q.Add("name", pc)
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

	fmt.Println(string(bodyBytes))
}

