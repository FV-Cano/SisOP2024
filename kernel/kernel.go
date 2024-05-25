package main

import (
	"log"
	"net/http"

	kernel_api "github.com/sisoputnfrba/tp-golang/kernel/API"
	"github.com/sisoputnfrba/tp-golang/kernel/globals"
	kernelutils "github.com/sisoputnfrba/tp-golang/kernel/utils"
	cfg "github.com/sisoputnfrba/tp-golang/utils/config"
	logger "github.com/sisoputnfrba/tp-golang/utils/log"
	"github.com/sisoputnfrba/tp-golang/utils/server-Functions"
)

// ? Handshake IO?

func main() {
	// Iniciar loggers
	logger.ConfigurarLogger("kernel.log")
	logger.LogfileCreate("kernel_debug.log")

	// Inicializamos la config y tomamos valores
	err := cfg.ConfigInit("config_kernel.json", &globals.Configkernel)
	if err != nil {
		log.Fatalf("Error al cargar la configuracion %v", err)
	}
	log.Println("Configuracion KERNEL cargada")

	// Handlers
	kernelRoutes := RegisteredModuleRoutes()

	globals.PlanBinary <- false // ? Default false?

	// Iniciar servidor
	go server.ServerStart(globals.Configkernel.Port, kernelRoutes)

	// * Planificación
	go kernelutils.Plan()

	select {}		// Deja que la goroutine principal siga corriendo
}

// Literalmente no hace nada, es para evitar el error de compilación de "imported and not used"
func UNUSED(x ...interface{}){}

func RegisteredModuleRoutes() http.Handler {
	moduleHandler := &server.ModuleHandler{
		RouteHandlers: map[string]http.HandlerFunc{
			"PUT /process": 			kernel_api.ProcessInit,
			"DELETE /process/{pid}": 	kernel_api.ProcessDelete,
			"GET /process/{pid}": 		kernel_api.ProcessState,
			"PUT /plani": 				kernel_api.PlanificationStart,
			"DELETE /plani": 			kernel_api.PlanificationStop,
			"GET /process": 			kernel_api.ProcessList,
			"POST /io-handshake": 		kernel_api.GetIOInterface,
			"POST /io-gen-interface": 	kernel_api.Resp_ExisteInterfazGen,
			"POST /tiempoBloq":			kernel_api.Resp_TiempoEspera,
			"POST /dispatch":			kernel_api.PCB_recv,
		},
	}
	return moduleHandler
}

// TODO: Probar finalizar proceso y estado proceso
// ?: Preguntar utilización de APIs externas