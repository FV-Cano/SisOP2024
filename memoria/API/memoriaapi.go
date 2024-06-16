package memoria_api

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/sisoputnfrba/tp-golang/memoria/globals"
)

type GetInstructions_BRQ struct {
	Path string `json:"path"`
	Pid  uint32 `json:"pid"`
	Pc   uint32 `json:"pc"`
}

type BitMap []int

func AbrirArchivo(filePath string) *os.File {
	file, err := os.Open(filePath) //El paquete nos provee el método ReadFile el cual recibe como argumento el nombre de un archivo el cual se encargará de leer. Al completar la lectura, retorna un slice de bytes, de forma que si se desea leer, tiene que ser convertido primero a una cadena de tipo string
	if err != nil {
		log.Fatal(err)
	}
	return file
}

func InstruccionActual(w http.ResponseWriter, r *http.Request) {

	queryParams := r.URL.Query()
	pid := queryParams.Get("pid")
	pc := queryParams.Get("pc")

	respuesta, err := json.Marshal((BuscarInstruccionMap(PasarAInt(pc), PasarAInt(pid))))
	if err != nil {
		http.Error(w, "Error al codificar los datos como JSON", http.StatusInternalServerError)
		return
	}

	log.Printf("La instruccion buscada fue: %s", BuscarInstruccionMap(PasarAInt(pc), PasarAInt(pid)))

	time.Sleep(time.Duration(globals.Configmemory.Delay_response))

	w.WriteHeader(http.StatusOK)
	w.Write(respuesta)

}

func BuscarInstruccionMap(pc int, pid int) string {
	resultado := globals.InstruccionesProceso[pid][pc]
	return resultado
}


func PasarAInt(cadena string) int {
	num, err := strconv.Atoi(cadena)
	if err != nil {
		fmt.Println("Error: ", err)
	}
	return num
}

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

	//acá debemos inicializar vacía la tabla de páginas para el proceso
	if globals.Tablas_de_paginas == nil {
		globals.Tablas_de_paginas = make(map[int]globals.TablaPaginas)
	}
	
	globals.Tablas_de_paginas[int(pid)] = globals.TablaPaginas{}
	log.Printf("PID: %d - Tamaño: %d", pid, len(globals.Tablas_de_paginas[int(pid)]))
	respuesta, err := json.Marshal((BuscarInstruccionMap(int(pc), int(pid))))
	if err != nil {
		http.Error(w, "Error al codificar los datos como JSON", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(respuesta)
}

// --------------------------------------------------------------------------------------//
// AJUSTAR TAMANiO DE UN PROCESO: CPU le hace la peticion desde estoVaParaCpuApi.go
func Resize(w http.ResponseWriter, r *http.Request) { //hay que hacer un patch ya que vamos a estar modificando un recurso existente (la tabla de páginas)
	queryParams := r.URL.Query()
	tamanio := queryParams.Get("tamanio")
	pid := queryParams.Get("pid")
	respuesta, err := json.Marshal(RealizarResize(PasarAInt(tamanio), PasarAInt(pid))) //devolver error out of memory
	if err != nil {
		http.Error(w, "Error al codificar los datos como JSON", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(respuesta)
}

func RealizarResize(tamanio int, pid int) string {
	cantPaginasActual := len(globals.Tablas_de_paginas[int(pid)])
	//ver cuantas paginas tiene el proceso en la tabla
	//cantPaginas := tamanio / globals.Configmemory.Page_size
	cantPaginas := int(math.Ceil(float64(tamanio) / float64(globals.Configmemory.Page_size)))
	// agregar a la tabla de páginas del proceso la cantidad de páginas que se le asignaron
	log.Printf("Tabla de paginas ANTES DE REDIM del PID %d: %v", pid, globals.Tablas_de_paginas[pid])

	globals.Tablas_de_paginas[int(pid)] = make(globals.TablaPaginas, cantPaginas)
	/*
	   make(globals.TablaPaginas, cantPaginas) crea una nueva tabla de páginas con una cantidad específica de páginas (cantPaginas).
	   Cada página en la tabla es un Frame.
	*/

	resultado := ModificarTamanioProceso(cantPaginasActual, cantPaginas, pid)
	log.Printf("Tabla de páginas del PID %d redimensionada a %d páginas", pid, cantPaginas)
	log.Printf("Tabla de paginas del PID %d: %v", pid, globals.Tablas_de_paginas[pid])
	return resultado
}

func ModificarTamanioProceso(tamanioProcesoActual int, tamanioProcesoNuevo int, pid int) string {

	//tamanioMemEnPaginas := globals.Configmemory.Memory_size / globals.Configmemory.Page_size

	if (tamanioProcesoActual <= tamanioProcesoNuevo) { //ampliar proceso
		var diferenciaEnPaginas = tamanioProcesoNuevo - tamanioProcesoActual
		log.Printf("PID: %d - Tamanio Actual: %d - Tamanio a Ampliar: %d", pid, tamanioProcesoActual, tamanioProcesoNuevo) // verificar si en el último parámetro va diferenciaEnPaginas
		fmt.Println("MOSTRAMELON EN PAGINAS EL TAMANIO")
		return AmpliarProceso(diferenciaEnPaginas, pid)
	} else { // reducir proceso
		var diferenciaEnPaginas = tamanioProcesoActual - tamanioProcesoNuevo
		log.Printf("PID: %d - Tamanio Actual: %d - Tamanio a Reducir: %d", pid, tamanioProcesoActual, tamanioProcesoNuevo) // verificar si en el último parámetro va diferenciaEnPaginas
		fmt.Println("MOSTRAMELON EN PAGINAS EL TAMANIO")
		return ReducirProceso(diferenciaEnPaginas, pid)
		
	}
}

func AmpliarProceso(diferenciaEnPaginas int, pid int) string {
	for pagina := 0; pagina < diferenciaEnPaginas; pagina++ {
		marcoDisponible := false
		for i := 0; i < globals.Frames; i++ { //out of memory si no hay marcos disponibles
			if IsNotSet(i) {
				//setear el valor del marco en la tabla de páginas del proceso
				globals.Tablas_de_paginas[pid][pagina] = globals.Frame(i)
				//marcar marco como ocupado
				Set(i)
				marcoDisponible = true
				// Salir del bucle una vez que se ha asignado un marco a la página
				break
			}
		}
		if !marcoDisponible {
			return "out of memory"
		}
	}
	return "OK"

}

func ReducirProceso(diferenciaEnPaginas int, pid int) string {
	for diferenciaEnPaginas > 0 {
		//obtener el marco que le corresponde a la página
		marco := BuscarMarco(pid, diferenciaEnPaginas)
		//marcar marco como desocupado
		globals.Tablas_de_paginas[pid] = append(globals.Tablas_de_paginas[pid][:diferenciaEnPaginas], globals.Tablas_de_paginas[pid][diferenciaEnPaginas+1:]...)
		Clear(marco)
		diferenciaEnPaginas--
		
	}
	return "OK"
}

// --------------------------------------------------------------------------------------//
// ACCESO A TABLA DE PAGINAS: PETICION DESDE CPU (GET)
// Busca el marco que pertenece al proceso y a la página que envía CPU, dentro del diccionario
func BuscarMarco(pid int, pagina int) int {
	fmt.Println("Estoy buscandolon")
	resultado := globals.Tablas_de_paginas[pid][pagina]
	fmt.Println("El resultado es: ", resultado)
	return int(resultado)
}

func EnviarMarco(w http.ResponseWriter, r *http.Request) {
	//Ante cada peticion de CPU, dado un pid y una página, enviar frame a CPU

	queryParams := r.URL.Query()
	pid := queryParams.Get("pid")
	pagina := queryParams.Get("pagina")
	buscarMarco:= BuscarMarco(PasarAInt(pid), PasarAInt(pagina))
	respuesta, err := json.Marshal(buscarMarco)
	if err != nil {
		http.Error(w, "Error al codificar los datos como JSON", http.StatusInternalServerError)
		return
	}
	log.Printf("PID: %d - Pagina %d - Marco %d", PasarAInt(pid), PasarAInt(pagina), buscarMarco)
	w.WriteHeader(http.StatusOK)
	w.Write(respuesta)
}

//--------------------------------------------------------------------------------------//
//FINALIZACION DE PROCESO: PETICION DESDE KERNEL (PATCH) 
//TODO: falta implementar desde el kernel?
func FinalizarProceso(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	pid := queryParams.Get("pid")
	ReducirProceso(len(globals.Tablas_de_paginas[PasarAInt(pid)]), PasarAInt(pid))
	log.Printf("PID: %d - Tamaño reducido: %d", PasarAInt(pid), len(globals.Tablas_de_paginas[PasarAInt(pid)]))
	w.WriteHeader(http.StatusOK)
}

// --------------------------------------------------------------------------------------//
// ACCESO A ESPACIO DE USUARIO: Esta petición puede venir tanto de la CPU como de un Módulo de Interfaz de I/O
type BodyRequestLeer struct {
	DireccionesTamanios []globals.DireccionTamanio `json:"direcciones_tamanios"`
}
// le va a llegar la lista de struct de direccionfisica y tamanio
// por cada struct va a leer la memoria en el tamaño que le pide y devolver el contenido
func LeerMemoria(w http.ResponseWriter, r *http.Request) {
	var request []globals.DireccionTamanio
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Println("Me mandaron a leer :)")

	contenidoLeido := LeerDeMemoria(request)
	fmt.Println("CONTENIDO LEIDO: ", contenidoLeido)

	/* respuesta, err := json.Marshal(contenidoLeido)
	if err != nil {
		http.Error(w, "Error al codificar los datos como JSON", http.StatusInternalServerError)
		return
	} */

	time.Sleep(time.Duration(globals.Configmemory.Delay_response) * time.Millisecond) //nos dan los milisegundos o lo dejamos así?

	w.WriteHeader(http.StatusOK)
	w.Write(contenidoLeido)
}

// le va a llegar la lista de struct de direccionfisica y tamanio (O LE LLEGA DE A UNA? ES DECIR DE A UNA PETICION)
// por cada struct va a leer la memoria en el tamaño que le pide y devolver el contenido
/* func LeerDeMemoria(direccionesTamanios []DireccionTamanio) string {
	//Ante un pedido de lectura, devolver el valor que se encuentra a partir de la dirección física pedida.
	var contenido []byte
	for _, dt := range direccionesTamanios {
		contenido = append(contenido, globals.User_Memory[dt.DireccionFisica:dt.DireccionFisica+dt.Tamanio]...)
		//todo ver que es el mismo problema desde CPU de no pasarle el pid
		//log.Printf("PID: %d - Accion: LEER - Direccion fisica: %d - Tamaño %d", pid, dt.DireccionFisica, dt.Tamanio)
	}
	return string(contenido)
} */
func LeerDeMemoria(direccionesTamanios []globals.DireccionTamanio) []byte {
	/*Ante un pedido de lectura, devolver el valor que se encuentra a partir de la dirección física pedida.*/
	var contenido []byte
	for _, dt := range direccionesTamanios {
		contenido = append(contenido, globals.User_Memory[dt.DireccionFisica:dt.DireccionFisica+dt.Tamanio]...)
	}
	return contenido
}

type BodyRequestEscribir struct {
	DireccionesTamanios []globals.DireccionTamanio
	Valor_a_escribir    string       
	TamanioLimite       int
	Pid                 int      
}

func EscribirMemoria(w http.ResponseWriter, r *http.Request) {
	var request BodyRequestEscribir
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	escribioEnMemoria := EscribirEnMemoria(request.DireccionesTamanios, request.Valor_a_escribir, request.Pid, request.TamanioLimite)

	respuesta, err := json.Marshal(escribioEnMemoria)
	if err != nil {
		http.Error(w, "Error al codificar los datos como JSON", http.StatusInternalServerError)
		return
	}

	time.Sleep(time.Duration(globals.Configmemory.Delay_response) * time.Millisecond) //nos dan los milisegundos o lo dejamos así?

	fmt.Println("LA MEMORIAN QUEDO ASII ", globals.User_Memory)
	
	w.WriteHeader(http.StatusOK)
	w.Write(respuesta)
}

// por cada struct va a ESCRIBIR la memoria en el tamaño que le pide
func EscribirEnMemoria(direccionesTamanios []globals.DireccionTamanio, valor_a_escribir string, pid int, tamanioLimite int) string { //TODO: tenemos que validar que al proceso le corresponda escribir ahí o ya la validación la hizo cpu al traducir la dirección?
	/*Ante un pedido de escritura, escribir lo indicado a partir de la dirección física pedida.
	En caso satisfactorio se responderá un mensaje de ‘OK’.*/
	
	bytesValor := []byte(valor_a_escribir)
	fmt.Println("VALOR EN BYTES: ", bytesValor)
	
	cantEscrita := 0
	
	for _, dt := range direccionesTamanios {
		for cantEscrita < tamanioLimite {
			valorAEscribir := takeAndRemove(dt.Tamanio, &bytesValor)
			fmt.Println("VALOR TRUNCADO: ", valorAEscribir)
			
			copy(globals.User_Memory[dt.DireccionFisica:], valorAEscribir)

			cantEscrita += dt.Tamanio

			log.Printf("PID: %d - Accion: ESCRIBIR - Direccion fisica: %d - Tamaño %d", pid, dt.DireccionFisica, dt.Tamanio)
		}
	}


	/* for _, dt := range direccionesTamanios {
		bytesValor := []byte(valor_a_escribir)
		fmt.Println("VALOR EN BYTES: ", bytesValor)
		valorAEscribir := takeAndRemove(dt.Tamanio, &bytesValor)
		fmt.Println("VALOR TRUNCADO: ", valorAEscribir)
		copy(globals.User_Memory[dt.DireccionFisica:], valorAEscribir)
		log.Printf("PID: %d - Accion: ESCRIBIR - Direccion fisica: %d - Tamaño %d", pid, dt.DireccionFisica, dt.Tamanio)
	} */

	return "OK"
}
	
func takeAndRemove(n int, list *[]byte) []byte {
    if n > len(*list) {
        n = len(*list)
    }
    result := (*list)[:n]
    *list = (*list)[n:]
    return result
}
// --------------------------------------------------------------------------------------//
// BITMAP AUXILIAR
func NewBitMap(size int) BitMap {
	NewBMAp := make(BitMap, size)
	for i := 0; i < size; i++ {
		NewBMAp[i] = 0
	}
	return NewBMAp
}

func Set(i int) {
	globals.CurrentBitMap[i] = 1
}

func Clear(i int) {
	globals.CurrentBitMap[i] = 0
}

func IsNotSet(i int) bool {
	return globals.CurrentBitMap[i] == 0
}

// --------------------------------------------------------------------------------------//
func Page_size(w http.ResponseWriter, r *http.Request) {
	respuesta, err := json.Marshal(globals.Configmemory.Page_size)
	if err != nil {
		http.Error(w, "Error al codificar los datos como JSON", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(respuesta)
}

// --------------------------------------------------------------------------------------//
// PEDIR TAMANIO DE TABLA DE PAGINAS: PETICION DESDE CLIENTE (GET)
func PedirTamTablaPaginas(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	pid := queryParams.Get("pid")

	tableishon := globals.Tablas_de_paginas[PasarAInt(pid)]
	largoTableishon := len(tableishon)

	fmt.Println("TAMANIO DE TABLEISHON: ", largoTableishon)

	respuesta, err := json.Marshal(largoTableishon)
	if err != nil {
		http.Error(w, "Error al codificar los datos como JSON", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(respuesta)
}
