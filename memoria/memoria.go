package main

import (
	"log"
	"net/http"

	memoria_api "github.com/sisoputnfrba/tp-golang/memoria/API"
	"github.com/sisoputnfrba/tp-golang/memoria/globals"
	cfg "github.com/sisoputnfrba/tp-golang/utils/config"
	logger "github.com/sisoputnfrba/tp-golang/utils/log"
	"github.com/sisoputnfrba/tp-golang/utils/server-Functions"
)

func main() {
	// Iniciar loggers
	logger.ConfigurarLogger("memory.log")
	logger.LogfileCreate("memory_debug.log")

	// Inicializamos la config
	err := cfg.ConfigInit("config-memory.json", &globals.Configmemory)
	if err != nil {
		log.Fatalf("Error al cargar la configuracion %v", err)
	}
	log.Println("Configuracion MEMORIA cargada")

	// Handlers
	// Iniciar servidor

	// log.Println("Instrucciones leídas por memoria.")
	go server.ServerStart(globals.Configmemory.Port, RegisteredModuleRoutes())
	// log.Println("Instrucciones enviadas a CPU")

	select {}

}

func RegisteredModuleRoutes() http.Handler {
	moduleHandler := &server.ModuleHandler{
		RouteHandlers: map[string]http.HandlerFunc{
			"GET /instrucciones/{pid}":      memoria_api.CargarInstrucciones,
			"GET /instrucciones/{pid}/{pc}": memoria_api.InstruccionActual,
		},
	}
	return moduleHandler
}
