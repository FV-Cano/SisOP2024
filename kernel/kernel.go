package main

import (
	"log"
	"net/http"

	cfg "github.com/sisoputnfrba/tp-golang/utils/config"
	datareceive "github.com/sisoputnfrba/tp-golang/utils/data-receive"
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
    err := cfg.ConfigInit("config_kernel.json", &configkernel)
	if err != nil {
		log.Fatalf("Error al cargar la configuracion %v", err)
	}

	log.Printf("Config resources: %v", configkernel.Resources)
}

func IO_serverStart() {
	mux := http.NewServeMux()

	mux.HandleFunc("/paquetes", datareceive.RecibirPaquetes)
	mux.HandleFunc("/mensaje", datareceive.RecibirMensaje)

	log.Println("IO server running on http://localhost:8001")
	err := http.ListenAndServe(":8001", mux)
	if err != nil {
		panic(err)
	}
}