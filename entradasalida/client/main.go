package main

import (
	"client/globals"
	"client/utils"
	"log"
)


func main() {

	utils.ConfigurarLogger()
	
	globals.ClientConfig = utils.IniciarConfiguracion("config.json")
	
	if globals.ClientConfig == nil {
		log.Fatalf("No se pudo cargar la configuración")
	}

	// KERNEL
	
	log.Println(globals.ClientConfig.Message)
	utils.EnviarMensaje(globals.ClientConfig.Ip_Kernel, globals.ClientConfig.Port_Kernel, globals.ClientConfig.Message)
	utils.LeerConsola()
	utils.GenerarYEnviarPaquete(globals.ClientConfig.Ip_Kernel, globals.ClientConfig.Port_Kernel)
	
	
	//utils.EnviarPaquete(globals.ClientConfig.Ip, globals.ClientConfig.Puerto,)

	// MEMORIA

	log.Println(globals.ClientConfig.Message)
	utils.EnviarMensaje(globals.ClientConfig.Ip_Memory, globals.ClientConfig.Port_Memory, globals.ClientConfig.Message)
	utils.LeerConsola()
	utils.GenerarYEnviarPaquete(globals.ClientConfig.Ip_Memory, globals.ClientConfig.Port_Memory)
	//utils.EnviarPaquete(globals.ClientConfig.Ip, globals.ClientConfig.Puerto,)
	
}
