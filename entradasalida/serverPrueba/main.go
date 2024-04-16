package main

import (
	"net/http"

	datareceive "github.com/sisoputnfrba/tp-golang/utils/data-receive"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/paquetes", datareceive.RecibirPaquetes)
	mux.HandleFunc("/mensaje", datareceive.RecibirMensaje)

	//panic("no implementado!")
	err := http.ListenAndServe(":8001", mux)
	if err != nil {
		panic(err)
	}

}
