package main

import (
	"fmt"
	"log"
	"net/http"

	kernel_api "github.com/sisoputnfrba/tp-golang/kernel/API"
	"github.com/sisoputnfrba/tp-golang/kernel/globals"
	resources "github.com/sisoputnfrba/tp-golang/kernel/resources"
	kernelutils "github.com/sisoputnfrba/tp-golang/kernel/utils"
	cfg "github.com/sisoputnfrba/tp-golang/utils/config"
	logger "github.com/sisoputnfrba/tp-golang/utils/log"
	"github.com/sisoputnfrba/tp-golang/utils/server-Functions"
)

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
	// kernelRoutes := RegisteredModuleRoutes()

	// Execution Config
	globals.MultiprogrammingCounter = make (chan int, globals.Configkernel.Multiprogramming)	// Inicializamos el contador de multiprogramación
	globals.STSCounter = make (chan int, globals.Configkernel.Multiprogramming)	// Inicializamos el contador de STS
	resources.InitResourceMap()

	// ! globals.ControlMutex.Lock()
	globals.EmptiedListMutex.Lock() // Bloqueamos la lista de jobs vacía
	globals.LTSPlanBinary <- false
	globals.STSPlanBinary <- false

	// Iniciar servidor
	// go server.ServerStart(globals.Configkernel.Port, kernelRoutes)
	go ServerStart(globals.Configkernel.Port)

	// * Planificación
	go kernelutils.LTS_Plan()
	go kernelutils.STS_Plan()

	select {}		// Deja que la goroutine principal siga corriendo
}

/* func RegisteredModuleRoutes() http.Handler {
	moduleHandler := &server.ModuleHandler{
		RouteHandlers: map[string]http.HandlerFunc{
			// Procesos
			"GET /process": 					kernel_api.ProcessList,
			"PUT /process": 					kernel_api.ProcessInit,
			"GET /process/{pid}":				kernel_api.ProcessState,
			"DELETE /process/{pid}": 			kernel_api.ProcessDelete,
			// Planificación
			"PUT /plani": 						kernel_api.PlanificationStart,
			"DELETE /plani": 					kernel_api.PlanificationStop,
			// I/O
			"POST /io-handshake": 				kernel_api.GetIOInterface,
			"POST /io-interface": 				kernel_api.ExisteInterfaz,
			"POST /iodata-gensleep":			kernel_api.RecvData_gensleep,
			"POST /iodata-stdin":				kernel_api.RecvData_stdin,
			"POST /iodata-stdout":				kernel_api.RecvData_stdout,
			"POST /iodata-dialfs":				kernel_api.RecvData_dialfs,
			"POST /io-return-pcb":				kernel_api.RecvPCB_IO,
			// Recursos
			"GET /resource-info":				resources.GETResourcesInstances,
			"GET /resourceblocked":				resources.GETResourceBlockedJobs,
		},
	}
	return moduleHandler
} */

func ServerStart(port int) {
	mux := http.NewServeMux()

	mux.HandleFunc("/paquetes", 				server.RecibirPaquetes)
	mux.HandleFunc("/mensaje", 					server.RecibirMensaje)
	// Procesos
	mux.HandleFunc("GET /process",				kernel_api.ProcessList)
	mux.HandleFunc("PUT /process",				kernel_api.ProcessInit)
	mux.HandleFunc("GET /process/{pid}", 		kernel_api.ProcessState)
	mux.HandleFunc("DELETE /process/{pid}",		kernel_api.ProcessDelete)
	// Planificación
	mux.HandleFunc("PUT /plani", 				kernel_api.PlanificationStart)
	mux.HandleFunc("DELETE /plani",				kernel_api.PlanificationStop)
	// I/O
	mux.HandleFunc("POST /io-handshake", 		kernel_api.GetIOInterface)
	mux.HandleFunc("POST /io-interface", 		kernel_api.ExisteInterfaz)
	mux.HandleFunc("POST /iodata-gensleep",		kernel_api.RecvData_gensleep)
	mux.HandleFunc("POST /iodata-stdin", 		kernel_api.RecvData_stdin)
	mux.HandleFunc("POST /iodata-stdout", 		kernel_api.RecvData_stdout)
	mux.HandleFunc("POST /iodata-dialfs", 		kernel_api.RecvData_dialfs)
	mux.HandleFunc("POST /io-return-pcb", 		kernel_api.RecvPCB_IO)
	// Recursos
	mux.HandleFunc("GET /resource-info", 		resources.GETResourcesInstances)
	mux.HandleFunc("GET /resourceblocked", 		resources.GETResourceBlockedJobs)

	log.Printf("Server listening on port %d\n", port)
	err := http.ListenAndServe(":"+fmt.Sprintf("%v", port), mux)
	if err != nil {
		panic(err)
	}
}

// TODO: Probar finalizar proceso y estado proceso