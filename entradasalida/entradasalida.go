package main

import (
	"log"
	"net/http"

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
	err := cfg.ConfigInit("config-io.json", &globals.ConfigIO)
	if err != nil {
		log.Fatalf("Error al cargar la configuracion %v", err)
	}
	log.Printf("Configuración IO cargada")

	IORoutes := RegisteredModuleRoutes()

	go server.ServerStart(globals.ConfigIO.Port , IORoutes)

	// Handshake con kernel
	log.Println("Handshake con Kernel")
	IO_api.HandshakeKernel()

	select {}
}

func RegisteredModuleRoutes() http.Handler {
	moduleHandler := &server.ModuleHandler{
		RouteHandlers: map[string]http.HandlerFunc{
			"POST /tiempoBloq": 	IO_api.Resp_TiempoEsperaIO,
			"POST /io-gen-sleep": 	IO_api.IOGenSleep,
		},
	}
	return moduleHandler
}
