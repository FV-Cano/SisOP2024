package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type Mensaje struct {
	Mensaje string `json:"mensaje"`
}

type Paquete struct {
	Valores []string `json:"valores"`
}

type ModuleHandler struct {
	RouteHandlers map[string]http.HandlerFunc
}

func RecibirPaquetes(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var paquete Paquete
	err := decoder.Decode(&paquete)
	if err != nil {
		log.Printf("error al decodificar mensaje: %s\n", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("error al decodificar mensaje"))
		return
	}

	log.Println("me llego un paquete de un cliente")
	log.Printf("%+v\n", paquete)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func RecibirMensaje(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var mensaje Mensaje
	err := decoder.Decode(&mensaje)
	if err != nil {
		log.Printf("Error al decodificar mensaje: %s\n", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Error al decodificar mensaje"))
		return
	}

	log.Println("Me llego un mensaje de un cliente")
	log.Printf("%+v\n", mensaje)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

/**
 * ServerStart: Inicia un servidor en el puerto especificado y con las rutas especificadas de ser necesario

 * @param port string
 * @param moduleRoutes optional
*/
func ServerStart(port int, moduleRoutes ...http.Handler) {
	var finalHandler http.Handler
	mux := http.NewServeMux()

	mux.HandleFunc("/paquetes", RecibirPaquetes)
	mux.HandleFunc("/mensaje", RecibirMensaje)
	mux.HandleFunc("GET /helloworld", HelloWorld)

	// Combinar rutas comunes y específicas
	finalHandler = mux

	if len(moduleRoutes) > 0 {
		finalHandler = http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for _, moduleRoute := range moduleRoutes {
				moduleRoute.ServeHTTP(w, r)
				return
			}
		mux.ServeHTTP(w, r)
		}))
	}

	log.Printf("Server listening on port %d\n", port)
	err := http.ListenAndServe(":"+fmt.Sprintf("%v", port), finalHandler)
	if err != nil {
		panic(err)
	}
}

/**
 * NewModule: Crea un nuevo módulo con las rutas especificadas

 * @return ModuleHandler
*/
func (m *ModuleHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler, ok := m.RouteHandlers[r.Method+" "+r.URL.Path]
	if !ok {
		http.NotFound(w, r)
		return
	}
	handler(w, r)
}

func HelloWorld(w http.ResponseWriter, r *http.Request) {

	respuesta, err := json.Marshal("Hola! Como andas?")
	if err != nil {
		http.Error(w, "Error al codificar los datos como JSON", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(respuesta)
}