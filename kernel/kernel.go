package main

import (
	"log"
	"net/http"

	kernel_api "github.com/sisoputnfrba/tp-golang/kernel/API"
	cfg "github.com/sisoputnfrba/tp-golang/utils/config"
	logger "github.com/sisoputnfrba/tp-golang/utils/log"
	"github.com/sisoputnfrba/tp-golang/utils/server-Functions"
)

type T_ConfigKernel struct {
	Port 				int 		`json:"port"`
	IP_memory 			string 		`json:"ip_memory"`
	Port_memory 		int 		`json:"port_memory"`
	IP_cpu 				string 		`json:"ip_cpu"`
	Port_cpu 			int 		`json:"port_cpu"`
	Planning_algorithm 	string 		`json:"planning_algorithm"`
	Quantum 			int 		`json:"quantum"`
	Resources 			[]string 	`json:"resources"`
	Resource_instances 	[]int 		`json:"resource_instances"`
	Multiprogramming 	int 		`json:"multiprogramming"`
}

var configkernel T_ConfigKernel

func main() {
	// Iniciar loggers
	logger.ConfigurarLogger()
	logger.LogfileCreate("kernel.log")

	// Inicializamos la config y tomamos valores
	err := cfg.ConfigInit("config_kernel.json", &configkernel)
	if err != nil {
		log.Fatalf("Error al cargar la configuracion %v", err)
	}

	ipMemory := configkernel.IP_memory
	portMemory := configkernel.Port_memory
	ipCpu := configkernel.IP_cpu
	portCpu := configkernel.Port_cpu

	UNUSED(ipMemory, portMemory, ipCpu, portCpu)

	// Handlers
	
	kernelRoutes := RegisteredModuleRoutes()

	//http.ListenAndServe(":8001", nil) -> http.DefaultServerMux vs http.NewServeMux() routing

	// Iniciar servidor

	go server.ServerStart(configkernel.Port, kernelRoutes)

	/* client.EnviarMensaje(ipMemory, portMemory, "Saludo memoria desde Kernel")
	client.EnviarMensaje(ipCpu, portCpu, "Saludo cpu desde Kernel") */

	select {}		// Deja que la goroutine principal siga corriendo (Preguntas)
}

// Literalmente no hace nada, es para evitar el error de compilación de "imported and not used"
func UNUSED(x ...interface{}){}

func RegisteredModuleRoutes() http.Handler {
	moduleHandler := &server.ModuleHandler{
		RouteHandlers: map[string]http.HandlerFunc{
			"PUT /process": kernel_api.ProcessInit,
		},
	}
	return moduleHandler
}


/* func RegisteredModuleRoutes() http.Handler {
	r := mux.NewRouter()
	r.HandleFunc("PUT /process", kernel_api.ProcessInit)
	return r
} */