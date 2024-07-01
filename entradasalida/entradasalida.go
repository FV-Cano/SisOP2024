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
	globals.DialFS_QueueChannel = make(chan globals.DialFSRequest, 1)

	// Si la interfaz es de tipo DialFS se debe inicializar el sistema de archivos
	if globals.ConfigIO.Type == "DialFS" {
		IO_api.InicializarFS()
	}

	go IO_api.IOWork()

	select {}
}

func RegisteredModuleRoutes() http.Handler {
	moduleHandler := &server.ModuleHandler{
		RouteHandlers: map[string]http.HandlerFunc{
			"POST /io-operate":	IO_api.InterfaceQueuePCB,
		},
	}
	return moduleHandler
}
