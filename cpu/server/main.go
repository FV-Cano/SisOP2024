package main

import (
	"log"
	"net/http"
	"server/utils"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/paquetes", utils.RecibirPaquetes)
	mux.HandleFunc("/mensaje", utils.RecibirMensaje)

	log.Println("Servidor corriendo en el puerto 8003")

	err := http.ListenAndServe(":8003", mux)
	if err != nil {
		panic(err)
	}
}
