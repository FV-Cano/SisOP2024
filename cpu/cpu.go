package main

import (
	"fmt"
	"io"
	"log"
	"net/http"

	client "github.com/sisoputnfrba/tp-golang/utils/client-Functions"
	cfg "github.com/sisoputnfrba/tp-golang/utils/config"
	logger "github.com/sisoputnfrba/tp-golang/utils/log"
	//"github.com/sisoputnfrba/tp-golang/utils/pcb"
	//server "github.com/sisoputnfrba/tp-golang/utils/server-Functions"
)

type T_CPU struct {
	Port               int    `json:"port"`
	IP_memory          string `json:"ip_memory"`
	Port_memory        int    `json:"port_memory"`
	Number_felling_tlb int    `json:"number_felling_tlb"`
	Algorithm_tlb      string `json:"algorithm_tlb"`
}

var configcpu T_CPU

func main() {
	// Iniciar loggers
	logger.ConfigurarLogger("cpu.log")
	logger.LogfileCreate("cpu_debug.log")

	// *** CONFIGURACION ***
	err := cfg.ConfigInit("config-cpu.json", &configcpu)
	if err != nil {
		log.Fatalf("Error al cargar la configuracion %v", err)
	}
	log.Println("Configuracion CPU cargada")

	// *** SERVIDOR ***
	//go server.ServerStart(configcpu.Port)

	// *** CLIENTE ***
	
	log.Println("Enviando mensaje al servidor")

	client.EnviarMensaje(configcpu.IP_memory, configcpu.Port_memory, "Saludo memoria desde CPU")
		/*La cpu empieza a ejecutar y segun el contexto de ejecución va a tener las instrucciones a
	ejecutar, (memoria hace lo suyo), fetch agarra la instrucción y se fija en el program counter
	peticion para que memoria me de la peticion segun pc, ese valor*/
	PeticionMemoria()
	//select {}
}

func PeticionMemoria(){
	pc := string(10)
	cliente := &http.Client{}
	url := fmt.Sprintf("http://localhost:8002/instrucciones/%s", pc)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 
	}

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