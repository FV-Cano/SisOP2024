package memoria_api

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/sisoputnfrba/tp-golang/memoria/globals"
)

func AbrirArchivo(filePath string) *os.File {
	file, err := os.Open(filePath) //El paquete os provee el método ReadFile el cual recibe como argumento el nombre de un archivo el cual se encargará de leer. Al completar la lectura, retorna un slice de bytes, de forma que si se desea leer, tiene que ser convertido primero a una cadena de tipo string
	if err != nil {
		log.Fatal(err)
	}

	return file
}

func InstruccionActual(w http.ResponseWriter, r *http.Request) {
	// Traer query path (PID y PC)
	// Acá lo que hacemos es, según el PID que nos indican buscar la lista de instrucciones
	// que tiene asociada (una vez que encontramos esa lista tenemos que devolver el
	// elemento según el PC)
	
	pid := r.PathValue("pid")
	pc := r.PathValue("pc")

	respuesta, err := json.Marshal((BuscarInstruccionMap(PasarAInt(pc), PasarAInt(pid))))
	if err != nil {
		http.Error(w, "Error al codificar los datos como JSON", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(respuesta)
}

func BuscarInstruccionMap(pc int, pid int) string {
	globals.InstructionsMutex.Lock()
	defer globals.InstructionsMutex.Unlock()
	return globals.InstruccionesProceso[pid][pc]
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
	pid := r.PathValue("pid")
	
	pathInstrucciones := globals.Configmemory.Instructions_path + "/" + pid + ".txt"

	var instrucciones []string
	//Lee linea por linea el archivo
	file := AbrirArchivo(pathInstrucciones)
    defer file.Close()

    scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		// Agregar cada línea al slice de strings
		instrucciones = append(instrucciones, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	pidInt := PasarAInt(pid)
	
	globals.InstructionsMutex.Lock()
	defer globals.InstructionsMutex.Unlock()
	globals.InstruccionesProceso[pidInt] = instrucciones

	w.WriteHeader(http.StatusOK)
}