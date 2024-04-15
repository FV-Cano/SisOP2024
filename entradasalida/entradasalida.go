package main

import (
	"log"

	datasend "github.com/sisoputnfrba/tp-golang/utils/data-send"

	cfg "github.com/sisoputnfrba/tp-golang/utils/config"
)

type T_ConfigIO struct {
	Port               int    `json:"port"`
	Type               string `json:"type"`
	Unit_work_time     int    `json:"unit_work_time"`
	Ip_kernel          string `json:"ip_kernel"`
	Port_kernel        int    `json:"port_kernel"`
	Ip_memory          string `json:"ip_memory"`
	Port_memory        int    `json:"port_memory"`
	Dialfs_path        string `json:"dialfs_path"`
	Dialfs_block_size  int    `json:"dialfs_block_size"`
	Dialfs_block_count int    `json:"dialfs_block_count"`
	Message            string `json:"message"`
}

var configio T_ConfigIO

func main() {
	err := cfg.ConfigInit("config_io.json", &configio)
	if err != nil {
		log.Fatalf("Error al cargar la configuracion %v", err)
	}

	log.Printf(("Cantidad de bloques de DialFS: %d"), configio.Dialfs_block_count)


	/*if configio == nil {
		log.Fatalf("No se pudo cargar la configuración")
	}*/

	Comunicacion(configio.Message, configio.Ip_kernel, configio.Port_kernel)
	Comunicacion(configio.Message, configio.Ip_memory, configio.Port_memory)

}

func Comunicacion(mensaje string, ip string, puerto int) {
	log.Println(mensaje)
	datasend.EnviarMensaje(ip, puerto, mensaje)
	datasend.GenerarYEnviarPaquete(ip, puerto)

}
