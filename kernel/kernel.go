package main

func main() {
	
}

// Resources 			list `json:"resources"`
// Resource_instances 	list `json:"resource_instances"`
// TODO check list type

var configkernel T_ConfigKernel

func main() {
    err := cfg.ConfigInit("config_kernel.json", &configkernel)
	if err != nil {
		log.Fatalf("Error al cargar la configuracion %v", err)
	}

	log.Printf(("Algoritmo de planificacion: %s"), configkernel.Planning_algorithm)
}