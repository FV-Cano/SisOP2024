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

	//verificar si estan bien los punteros
	*globals.Frames = globals.Configmemory.Memory_size / globals.Configmemory.Page_size

	globals.User_Memory = make([]globals.T_PageFrame, *globals.Frames)

	//Lo que hago es cargar en cada marco una estructura
	// Esta estructura va a tener un número de pagina y su PID asociado
	// Si quiero saber en que marco está una página, recorro todos los marcos
	// y devuelvo el subíndice (correspondiente a un marco) cuando encuentre la pag

	for i := range *globals.Frames {
		globals.User_Memory[i] = globals.T_PageFrame{
			Data: make([]byte, globals.Configmemory.Page_size),
			PID:  -1, // -1 indica que el marco está libre
		}
	}

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
			"GET /instrucciones":  memoria_api.InstruccionActual,
			"POST /instrucciones": memoria_api.CargarInstrucciones,
		},
	}
	return moduleHandler
}
