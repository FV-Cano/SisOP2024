package main

import (
	"log"
	"net/http"

	kernel_api "github.com/sisoputnfrba/tp-golang/kernel/API"
	"github.com/sisoputnfrba/tp-golang/kernel/globals"
	kernelutils "github.com/sisoputnfrba/tp-golang/kernel/utils"
	"github.com/sisoputnfrba/tp-golang/utils/client-Functions"
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

	// Iniciar servidor
	go server.ServerStart(globals.Configkernel.Port, kernelRoutes)

	client.EnviarMensaje(globals.Configkernel.IP_memory, globals.Configkernel.Port_memory, "Saludo memoria desde Kernel")
	client.EnviarMensaje(globals.Configkernel.IP_cpu, globals.Configkernel.Port_cpu, "Saludo cpu desde Kernel")

	// * Planificación
	go kernelutils.Plan()

	select {}		// Deja que la goroutine principal siga corriendo
}

// Literalmente no hace nada, es para evitar el error de compilación de "imported and not used"
func UNUSED(x ...interface{}){}

func RegisteredModuleRoutes() http.Handler {
	moduleHandler := &server.ModuleHandler{
		RouteHandlers: map[string]http.HandlerFunc{
			"PUT /process": kernel_api.ProcessInit,
			"DELETE /process/{pid}": kernel_api.ProcessDelete,
			"GET /process/{pid}": kernel_api.ProcessState,
			"PUT /plani": kernel_api.PlanificationStart,
			"DELETE /plani": kernel_api.PlanificationStop,
			"GET /process": kernel_api.ProcessList,
		},
	}
	return moduleHandler
}

// TODO: Probar finalizar proceso y estado proceso
// ?: Preguntar utilización de APIs externas