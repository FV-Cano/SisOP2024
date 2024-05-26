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

	select {}
}

func RegisteredModuleRoutes() http.Handler {
	moduleHandler := &server.ModuleHandler{
		RouteHandlers: map[string]http.HandlerFunc{
			"POST /io-gen-sleep": 	IO_api.IOGenSleep,
		},
	}
	return moduleHandler
}
