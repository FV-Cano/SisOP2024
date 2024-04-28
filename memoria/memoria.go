package main

import (
	"bufio"
	"log"
	//"path/filepath"

	cfg "github.com/sisoputnfrba/tp-golang/utils/config"
	logger "github.com/sisoputnfrba/tp-golang/utils/log"
	"github.com/sisoputnfrba/tp-golang/utils/server-Functions"

	//"github.com/sisoputnfrba/tp-golang/utils/server-Functions"
	//"github.com/sisoputnfrba/tp-golang/utils/slice"
	"encoding/json"
	"net/http"
	"os"
)

type T_ConfigMemory struct {
	Port 				int 	`json:"port"`
	Memory_size 		int 	`json:"memory_size"`
	Page_size		 	int 	`json:"page_size"`
	Instructions_path 	string 	`json:"instructions_path"`
	Delay_response 		int 	`json:"delay_response"`
}

var configmemory T_ConfigMemory

func main() {
	// Iniciar loggers
	logger.ConfigurarLogger("memory.log")
	logger.LogfileCreate("memory_debug.log")

	// Inicializamos la config
	err := cfg.ConfigInit("config-memory.json", &configmemory)
	if err != nil {
		log.Fatalf("Error al cargar la configuracion %v", err)
	}
	log.Println("Configuracion MEMORIA cargada")
	
	// Handlers
	// Iniciar servidor

	log.Println("Instrucciones leídas por memoria.")
	go server.ServerStart(configmemory.Port,RegisteredModuleRoutes())
	log.Println("Instrucciones enviadas a CPU")	
	
	select {}
	
}

func AbrirArchivo(filePath string)(*os.File){
	file, err := os.Open(filePath) //El paquete os provee el método ReadFile el cual recibe como argumento el nombre de un archivo el cual se encargará de leer. Al completar la lectura, retorna un slice de bytes, de forma que si se desea leer, tiene que ser convertido primero a una cadena de tipo string
	if err != nil {
			log.Fatal(err)
		}
		
	return file
}

func  LeerInstrucciones(filePath string) []string {
	
	var instrucciones []string
	//Lee linea por linea el archivo
	scanner := bufio.NewScanner(AbrirArchivo(filePath))
    for scanner.Scan() {
        // Agregar cada línea al slice de strings
        instrucciones = append(instrucciones, scanner.Text())
    }
	if err := scanner.Err(); err != nil {
        log.Fatal(err)
    }
    return instrucciones
}


func RespuestaServidor(w http.ResponseWriter, r *http.Request) {

	respuesta, err := json.Marshal(LeerInstrucciones(configmemory.Instructions_path))
	if err != nil {
		http.Error(w, "Error al codificar los datos como JSON", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(respuesta)
}

func RegisteredModuleRoutes() http.Handler {
	moduleHandler := &server.ModuleHandler{
		RouteHandlers: map[string]http.HandlerFunc{
			"GET /instrucciones": RespuestaServidor,
		},
	}
	return moduleHandler
}