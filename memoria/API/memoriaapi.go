package memoria_api

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/sisoputnfrba/tp-golang/memoria/globals"
)

type GetInstructions_BRQ struct {
	Path string `json:"path"`
	Pid  uint32 `json:"pid"`
	Pc   uint32  `json:"pc"`
}

func AbrirArchivo(filePath string) *os.File {
	file, err := os.Open(filePath) //El paquete os provee el método ReadFile el cual recibe como argumento el nombre de un archivo el cual se encargará de leer. Al completar la lectura, retorna un slice de bytes, de forma que si se desea leer, tiene que ser convertido primero a una cadena de tipo string
	if err != nil {
		log.Fatal(err)
	}
	log.Print("Se leyó el archivo")
	return file
}

func InstruccionActual(w http.ResponseWriter, r *http.Request) {
	// Traer query path (PID y PC)
	// Acá lo que hacemos es, según el PID que nos indican buscar la lista de instrucciones
	// que tiene asociada (una vez que encontramos esa lista tenemos que devolver el
	// elemento según el PC)
	fmt.Println("HOLAAAA")

	pid := r.PathValue("pid")
	pc := r.PathValue("pc")

	fmt.Println("A VER LABURÁ: ", BuscarInstruccionMap(PasarAInt(pc), PasarAInt(pid)))

	respuesta, err := json.Marshal((BuscarInstruccionMap(PasarAInt(pc), PasarAInt(pid))))
	if err != nil {
		http.Error(w, "Error al codificar los datos como JSON", http.StatusInternalServerError)
		return
	}
	

	log.Printf("La instruccion buscada fue: %s", BuscarInstruccionMap(PasarAInt(pc), PasarAInt(pid)))

	w.WriteHeader(http.StatusOK)
	w.Write(respuesta)
}

func BuscarInstruccionMap(pc int, pid int) string {
	//globals.InstructionsMutex.Lock()
	//defer globals.InstructionsMutex.Unlock()

	resultado := globals.InstruccionesProceso[pid][pc]
	return resultado
}

func PasarAInt(cadena string) int {
	num, err := strconv.Atoi(cadena)

	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("Número:", num)
	}
	return num
}

// Esta funcion de abajo no sabemos si la necesitamos
func CargarInstrucciones(w http.ResponseWriter, r *http.Request) {
	var request GetInstructions_BRQ
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	pathInstrucciones := strings.Trim(request.Path, "\"")
	pid := request.Pid
	pc := request.Pc
	
	var instrucciones []string
	//Lee linea por linea el archivo
	file := AbrirArchivo(globals.Configmemory.Instructions_path + pathInstrucciones)
    defer file.Close()

    scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		// Agregar cada línea al slice de strings
		instrucciones = append(instrucciones, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
		
	globals.InstructionsMutex.Lock()
	defer globals.InstructionsMutex.Unlock()
	globals.InstruccionesProceso[int(pid)] = instrucciones

	log.Printf("Instrucciones cargadas para el PID %d ", pid)

	respuesta, err := json.Marshal((BuscarInstruccionMap(int(pc), int(pid))))
	if err != nil {
		http.Error(w, "Error al codificar los datos como JSON", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(respuesta)
}