package main

import (
	"client/globals"
	"client/utils"
	"log"
)

/*
IDEA:
const Kernel | Memoria

log.Println("Indicar servidor a conectarse")
scanf(nombreServidor)

if Kernel
globals.ClientConfig = utils.IniciarConfiguracion("configKernel.json")
else 
globals.ClientConfig = utils.IniciarConfiguracion("configMemoria.json")*/

func main() {

	utils.ConfigurarLogger()

	
	globals.ClientConfig = utils.IniciarConfiguracion("config.json")
	// validar que la config este cargada correctamente
	if globals.ClientConfig == nil {
		log.Fatalf("No se pudo cargar la configuración")
	}

	// loggeamos el valor de la config
	log.Println(globals.ClientConfig.Mensaje)

	// ADVERTENCIA: Antes de continuar, tenemos que asegurarnos que el servidor esté corriendo para poder conectarnos a él

	// enviar un mensaje al servidor con el valor de la config

	utils.EnviarMensaje(globals.ClientConfig.Ip, globals.ClientConfig.Puerto, globals.ClientConfig.Mensaje)

	// leer de la consola el mensaje
	//utils.LeerConsola()
	utils.LeerConsola()

	// generamos un paquete y lo enviamos al servidor
	utils.GenerarYEnviarPaquete()
	//utils.EnviarPaquete(globals.ClientConfig.Ip, globals.ClientConfig.Puerto,)

	// utils.GenerarYEnviarPaquete()
}
