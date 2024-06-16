package main

import (
	"log"
	"net/http"
	"os"

	IO_api "github.com/sisoputnfrba/tp-golang/entradasalida/API"
	"github.com/sisoputnfrba/tp-golang/entradasalida/globals"
	logger "github.com/sisoputnfrba/tp-golang/utils/log"
	"github.com/sisoputnfrba/tp-golang/utils/server-Functions"

	cfg "github.com/sisoputnfrba/tp-golang/utils/config"
)

// una funcion que codifique las unidades de trabajo en un json

func main() {
	// Iniciar loggers
	logger.ConfigurarLogger("io.log")
	logger.LogfileCreate("io_debug.log")

	// Inicializar config
	err := cfg.ConfigInit(os.Args[2], &globals.ConfigIO)
	if err != nil {
		log.Fatalf("Error al cargar la configuracion %v", err)
	}
	log.Printf("Configuración IO cargada")

	IORoutes := RegisteredModuleRoutes()

	go server.ServerStart(globals.ConfigIO.Port, IORoutes)

	// Handshake con kernel
	log.Println("Handshake con Kernel")
	IO_api.HandshakeKernel(os.Args[1])
	globals.Generic_QueueChannel = make(chan globals.GenSleep, 1)
	globals.Stdin_QueueChannel = make(chan globals.StdinRead, 1)
	globals.Stdout_QueueChannel = make(chan globals.StdoutWrite, 1)

	go IO_api.IOWork()

	select {}
}

func RegisteredModuleRoutes() http.Handler {
	moduleHandler := &server.ModuleHandler{
		RouteHandlers: map[string]http.HandlerFunc{
			"POST /io-operate":	IO_api.InterfaceQueuePCB,
			// "POST /io-gen-sleep": 	IO_api.IOGenSleep, 	Deprecated
			// "POST /io-stdin-read": 	IO_api.IOStdinRead,
			// "POST /io-stdin-write": IO_api.IOStdoutWrite,
		},
	}
	return moduleHandler
}
