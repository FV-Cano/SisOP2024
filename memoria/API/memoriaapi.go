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
	globals.Tablas_de_paginas[int(pid)] = []globals.Frame{}
	log.Printf("Tabla cargada para el PID %d ", pid)

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
	cantPaginas := tamanio / globals.Configmemory.Page_size
	// agregar a la tabla de páginas del proceso la cantidad de páginas que se le asignaron
	globals.Tablas_de_paginas[int(pid)] = make(globals.TablaPaginas, cantPaginas)
	/*
	   make(globals.TablaPaginas, cantPaginas) crea una nueva tabla de páginas con una cantidad específica de páginas (cantPaginas).
	   Cada página en la tabla es un Frame.
	*/
	ModificarTamanioProceso(cantPaginasActual, cantPaginas, pid)
	log.Printf("Tabla de páginas del PID %d redimensionada a %d páginas", pid, cantPaginas)
	return "OK"
}

func ModificarTamanioProceso(tamanioProcesoActual int, tamanioProcesoNuevo int, pid int) {
	if tamanioProcesoActual < tamanioProcesoNuevo { //ampliar proceso
		var diferenciaEnPaginas = tamanioProcesoNuevo - tamanioProcesoActual
		log.Printf("PID: %d - Tamanio Actual: %d - Tamanio a Ampliar: %d", pid, tamanioProcesoActual, tamanioProcesoNuevo) // verificar si en el último parámetro va diferenciaEnPaginas
		AmpliarProceso(diferenciaEnPaginas, pid)

	} else { // reducir proceso
		var diferenciaEnPaginas = tamanioProcesoActual - tamanioProcesoNuevo
		log.Printf("PID: %d - Tamanio Actual: %d - Tamanio a Reducir: %d", pid, tamanioProcesoActual, tamanioProcesoNuevo) // verificar si en el último parámetro va diferenciaEnPaginas
		ReducirProceso(diferenciaEnPaginas, pid)
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

func ReducirProceso(diferenciaEnPaginas int, pid int) {
	for diferenciaEnPaginas > 0 {
		//obtener el marco que le corresponde a la página
		marco := BuscarMarco(pid, diferenciaEnPaginas)
		//marcar marco como desocupado
		globals.Tablas_de_paginas[pid] = append(globals.Tablas_de_paginas[pid][:diferenciaEnPaginas], globals.Tablas_de_paginas[pid][diferenciaEnPaginas+1:]...)
		Clear(marco)
		diferenciaEnPaginas--
	}
}

// --------------------------------------------------------------------------------------//
// ACCESO A TABLA DE PAGINAS: PETICION DESDE CPU (GET)
// Busca el marco que pertenece al proceso y a la página que envía CPU, dentro del diccionario
func BuscarMarco(pid int, pagina int) int {
	resultado := globals.Tablas_de_paginas[pid][pagina]
	return int(resultado)
}

func EnviarMarco(w http.ResponseWriter, r *http.Request) {
	//Ante cada peticion de CPU, dado un pid y una página, enviar frame a CPU

	queryParams := r.URL.Query()
	pid := queryParams.Get("pid")
	pagina := queryParams.Get("pagina")
	respuesta, err := json.Marshal(BuscarMarco(PasarAInt(pid), PasarAInt(pagina)))
	if err != nil {
		http.Error(w, "Error al codificar los datos como JSON", http.StatusInternalServerError)
		return
	}
	log.Printf("PID: %d - Pagina %d - Marco %d", PasarAInt(pid), PasarAInt(pagina), BuscarMarco(PasarAInt(pid), PasarAInt(pagina)))
	w.WriteHeader(http.StatusOK)
	w.Write(respuesta)
}

//--------------------------------------------------------------------------------------//
//FINALIZACION DE PROCESO: PETICION DESDE KERNEL (PATCH)

func FinalizarProceso(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	pid := queryParams.Get("pid")
	ReducirProceso(len(globals.Tablas_de_paginas[PasarAInt(pid)]), PasarAInt(pid))
	w.WriteHeader(http.StatusOK)
}

// --------------------------------------------------------------------------------------//
// ACCESO A ESPACIO DE USUARIO: Esta petición puede venir tanto de la CPU como de un Módulo de Interfaz de I/O
type DireccionTamanio struct {
	DireccionFisica int
	Tamanio         int
}

// le va a llegar la lista de struct de direccionfisica y tamanio
// por cada struct va a leer la memoria en el tamaño que le pide y devolver el contenido
func LeerMemoria(w http.ResponseWriter, r *http.Request) {
	var request []DireccionTamanio
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	respuesta, err := json.Marshal(LeerDeMemoria(request))
	if err != nil {
		http.Error(w, "Error al codificar los datos como JSON", http.StatusInternalServerError)
		return
	}

	time.Sleep(time.Duration(globals.Configmemory.Delay_response) * time.Millisecond) //nos dan los milisegundos o lo dejamos así?

	w.WriteHeader(http.StatusOK)
	w.Write(respuesta)
}

// le va a llegar la lista de struct de direccionfisica y tamanio (O LE LLEGA DE A UNA? ES DECIR DE A UNA PETICION)
// por cada struct va a leer la memoria en el tamaño que le pide y devolver el contenido
func LeerDeMemoria(direccionesTamanios []DireccionTamanio) string {
	/*Ante un pedido de lectura, devolver el valor que se encuentra a partir de la dirección física pedida.*/
	var contenido []byte
	for _, dt := range direccionesTamanios {
		if (dt.DireccionFisica + dt.Tamanio) <= len(globals.User_Memory) {
			contenido = append(contenido, globals.User_Memory[dt.DireccionFisica:dt.DireccionFisica+dt.Tamanio]...)
		} else {
			return "Error: dirección fuera de rango"
		}
	}
	return string(contenido)
}

type BodyRequestEscribir struct {
	DireccionesTamanios []DireccionTamanio `json:"direcciones_tamanios"`
	Valor_a_escribir    string             `json:"valor_a_escribir"`
	Pid                 int                `json:"pid"`
}

func EscribirMemoria(w http.ResponseWriter, r *http.Request) {
	var request BodyRequestEscribir
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	respuesta, err := json.Marshal(EscribirEnMemoria(request.DireccionesTamanios, request.Valor_a_escribir, request.Pid))
	if err != nil {
		http.Error(w, "Error al codificar los datos como JSON", http.StatusInternalServerError)
		return
	}

	time.Sleep(time.Duration(globals.Configmemory.Delay_response) * time.Millisecond) //nos dan los milisegundos o lo dejamos así?

	w.WriteHeader(http.StatusOK)
	w.Write(respuesta)
}

// le va a llegar la lista de struct de direccionfisica y tamanio (O LE LLEGA DE A UNA? ES DECIR DE A UNA PETICION)
// por cada struct va a ESCRIBIR la memoria en el tamaño que le pide
func EscribirEnMemoria(direccionesTamanios []DireccionTamanio, valor_a_escribir string, pid int) string { //TODO: tenemos que validar que al proceso le corresponda escribir ahí o ya la validación la hizo cpu al traducir la dirección?
	/*Ante un pedido de escritura, escribir lo indicado a partir de la dirección física pedida.
	  En caso satisfactorio se responderá un mensaje de ‘OK’.*/
	var tamanioTotal int
	for _, dt := range direccionesTamanios {
		tamanioTotal += dt.Tamanio
	}
	bytesValor := []byte(valor_a_escribir)
	if len(bytesValor) > globals.Configmemory.Page_size {
		return "Error: dirección o tamanio fuera de rango"
	}
	copy(globals.User_Memory[dt.DireccionFisica:], bytesValor)

	return "OK"
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
	respuesta, err := json.Marshal(len(globals.Tablas_de_paginas[PasarAInt(pid)]))
	if err != nil {
		http.Error(w, "Error al codificar los datos como JSON", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(respuesta)
}
