package main

import (
	"log"
	"net/http"

	cpu_api "github.com/sisoputnfrba/tp-golang/cpu/API"
	cicloinstruccion "github.com/sisoputnfrba/tp-golang/cpu/cicloInstruccion"
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

	cicloinstruccion.DecodeAndExecute()

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
