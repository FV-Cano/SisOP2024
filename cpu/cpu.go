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