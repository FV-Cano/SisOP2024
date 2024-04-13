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

	go Comunicacion(globals.ClientConfig.Message, globals.ClientConfig.Ip_Kernel, globals.ClientConfig.Port)
	go Comunicacion(globals.ClientConfig.Message, globals.ClientConfig.Ip_Memory, globals.ClientConfig.Port_Memory)

}

func Comunicacion(mensaje string, ip string, puerto int) {
	log.Println(mensaje)
	utils.EnviarMensaje(ip, puerto, mensaje)
	utils.LeerConsola()
	utils.GenerarYEnviarPaquete(ip, puerto)
}
