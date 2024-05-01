package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	cpu_api "github.com/sisoputnfrba/tp-golang/cpu/API"
	"github.com/sisoputnfrba/tp-golang/cpu/globals"
	cfg "github.com/sisoputnfrba/tp-golang/utils/config"
	logger "github.com/sisoputnfrba/tp-golang/utils/log"
	server "github.com/sisoputnfrba/tp-golang/utils/server-Functions"
)

func main() {
	// Iniciar loggers
	logger.ConfigurarLogger("cpu.log")
	logger.LogfileCreate("cpu_debug.log")


	// *** CONFIGURACION ***
	err := cfg.ConfigInit("config-cpu.json", &globals.Configcpu)
	if err != nil {
		log.Fatalf("Error al cargar la configuracion %v", err)
	}
	log.Println("Configuracion CPU cargada")

	// *** SERVIDOR ***
	go server.ServerStart(globals.Configcpu.Port)

	// *** CLIENTE ***
	
	
		
	// Fetch()
	
	//select {}
}

func RegisteredModuleRoutes() http.Handler {
	moduleHandler := &server.ModuleHandler{
		RouteHandlers: map[string]http.HandlerFunc{
			"POST /pcb-recv": cpu_api.PCB_recv,
		},
	}
	return moduleHandler
}

// TODO: Mover fetch a un paquete de ciclo de instrucción
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