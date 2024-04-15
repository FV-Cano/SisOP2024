package main

import (
	"log"
	"net/http"

	cfg "github.com/sisoputnfrba/tp-golang/utils/config"
	datareceive "github.com/sisoputnfrba/tp-golang/utils/data-receive"
	datasend "github.com/sisoputnfrba/tp-golang/utils/data-send"
)

type T_CPU struct {
	Port               int    `json:"port"`
	IP_memory          string `json:"ip_memory"`
	Port_memory        int    `json:"port_memory"`
	Number_felling_tlb int    `json:"number_felling_tlb"`
	Algorithm_tlb      string `json:"algorithm_tlb"`
	Message            string `json:"message"`
}

var configcpu T_CPU

func main() {
	// *** CONFIGURACION ***
	err := cfg.ConfigInit("config-cpu.json", &configcpu)
	if err != nil {
		log.Fatalf("Error al cargar la configuracion %v", err)
	}
	// log.Printf(("Algoritmo de reemplazo de TLB: %s"), configcpu.Algorithm_tlb)

	log.Println("Configuracion cargada")

	// ? loggeamos el valor de la config
	// log.Println(configcpu.Message)

	// *** SERVIDOR ***
	go inicializarServidor()

	// *** CLIENTE ***
	// ! ADVERTENCIA: Antes de continuar, tenemos que asegurarnos que el servidor esté corriendo para poder conectarnos a él

	//TODO PONER configcpu.Port_memory
	log.Println("Enviando mensaje al servidor")
	datasend.EnviarMensaje(configcpu.IP_memory, configcpu.Port, configcpu.Message)

	// Generamos un paquete y lo enviamos al servidor
	// datasend.GenerarYEnviarPaquete(configcpu.IP_memory, configcpu.Port)

	select {}
}

func inicializarServidor() {
	mux := http.NewServeMux()

	mux.HandleFunc("/paquetes", datareceive.RecibirPaquetes)
	mux.HandleFunc("/mensaje", datareceive.RecibirMensaje)

	log.Println("Servidor corriendo en el puerto 8003")

	err := http.ListenAndServe(":8003", mux)
	if err != nil {
		panic(err)
	}
}
