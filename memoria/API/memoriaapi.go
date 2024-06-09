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

//--------------------------------------------------------------------------------------//
//AJUSTAR TAMAÑO DE UN PROCESO: CPU le hace la peticion desde estoVaParaCpuApi.go
func Resize(w http.ResponseWriter, r *http.Request) { //hay que hacer un patch ya que vamos a estar modificando un recurso existente (la tabla de páginas)
	queryParams := r.URL.Query()
	tamaño := queryParams.Get("tamaño")
	pid := queryParams.Get("pid")
	respuesta, err := json.Marshal(RealizarResize(PasarAInt(tamaño), PasarAInt(pid))) //devolver error out of memory
	if err != nil {
		http.Error(w, "Error al codificar los datos como JSON", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(respuesta)
}


func RealizarResize(tamaño int, pid int) string {
cantPaginasActual := len(globals.Tablas_de_paginas[int(pid)])
	//ver cuantas paginas tiene el proceso en la tabla
cantPaginas := tamaño / globals.Configmemory.Page_size
//agregar a la tabla de páginas del proceso la cantidad de páginas que se le asignaron
globals.Tablas_de_paginas[int(pid)] = make(globals.TablaPaginas, cantPaginas)
/*make(globals.TablaPaginas, cantPaginas) crea una nueva tabla de páginas con una cantidad específica de páginas (cantPaginas).
Cada página en la tabla es un Frame.*/
ModificarTamañoProceso(cantPaginasActual, cantPaginas, pid)
log.Printf("Tabla de páginas del PID %d redimensionada a %d páginas", pid, cantPaginas)
return "OK"
}

func ModificarTamañoProceso(tamañoProcesoActual int, tamañoProcesoNuevo int, pid int) {
	if tamañoProcesoActual < tamañoProcesoNuevo { //ampliar proceso
		var diferenciaEnPaginas = tamañoProcesoNuevo - tamañoProcesoActual
		log.Printf("PID: %d - Tamaño Actual: %d - Tamaño a Ampliar: %d", pid, tamañoProcesoActual, tamañoProcesoNuevo) // verificar si en el último parámetro va diferenciaEnPaginas
		AmpliarProceso(diferenciaEnPaginas, pid)

	} else { // reducir proceso
		var diferenciaEnPaginas = tamañoProcesoActual - tamañoProcesoNuevo
		log.Printf("PID: %d - Tamaño Actual: %d - Tamaño a Reducir: %d", pid, tamañoProcesoActual, tamañoProcesoNuevo) // verificar si en el último parámetro va diferenciaEnPaginas
		ReducirProceso(diferenciaEnPaginas, pid)
	}	
}

func AmpliarProceso(diferenciaEnPaginas int, pid int) string {
	for pagina := 0; pagina < diferenciaEnPaginas; pagina++ {
		marcoDisponible := false
		for i := 0; i < globals.Frames; i++ { //out of memory si no hay marcos disponibles
			if IsNotSet(i){
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

func ReducirProceso(diferenciaEnPaginas int, pid int){
	for  (diferenciaEnPaginas > 0){
		//obtener el marco que le corresponde a la página
		marco := BuscarMarco(pid, diferenciaEnPaginas)
		//marcar marco como desocupado
		Clear(marco)
		diferenciaEnPaginas--
	}
}
//--------------------------------------------------------------------------------------//
//ACCESO A TABLA DE PAGINAS: PETICION DESDE CPU (GET)
//Busca el marco que pertenece al proceso y a la página que envía CPU, dentro del diccionario
func BuscarMarco(pid int, pagina int) int {
	resultado := globals.Tablas_de_paginas[pid][pagina]
	return int(resultado)
	}

func EnviarMarco(w http.ResponseWriter, r *http.Request){
	//Ante cada peticion de CPU, dado un pid y una página, enviar frame a CPU
	queryParams := r.URL.Query()
	pid := queryParams.Get("pid")
	direccionLogica := queryParams.Get("direccionLogica")
	pagina := PasarAInt(direccionLogica) / globals.Configmemory.Page_size
	respuesta, err := json.Marshal((BuscarMarco(PasarAInt(pid), pagina)))
	if err != nil {
		http.Error(w, "Error al codificar los datos como JSON", http.StatusInternalServerError)
		return
	}
	log.Printf("PID: %d - Pagina %d - Marco %d", PasarAInt(pid), pagina, BuscarMarco(PasarAInt(pid), pagina))
	w.WriteHeader(http.StatusOK)
	w.Write(respuesta)
}
//--------------------------------------------------------------------------------------//
//FINALIZACION DE PROCESO: PETICION DESDE KERNEL (PATCH)

func FinalizarProceso(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	pid := queryParams.Get("pid")
	ReducirProceso(len(globals.Tablas_de_paginas[PasarAInt(pid)]), PasarAInt(pid))
//	respuesta, err := json.Marshal() 
//	if err != nil {
//	http.Error(w, "Error al codificar los datos como JSON", http.StatusInternalServerError)
//	return
//}
	w.WriteHeader(http.StatusOK)
//  w.Write(respuesta)
}
//--------------------------------------------------------------------------------------//
//ACCESO A ESPACIO DE USUARIO: Esta petición puede venir tanto de la CPU como de un Módulo de Interfaz de I/O
type BodyRequestLeer struct {
	Direccion_fisica int `json:"direccion_fisica"`
	Tamaño int `json:"tamaño"`
}

func LeerMemoria(w http.ResponseWriter, r *http.Request) {
	var request BodyRequestLeer
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	respuesta, err := json.Marshal(LeerDeMemoria(request.Direccion_fisica, request.Tamaño))
	if err != nil {
		http.Error(w, "Error al codificar los datos como JSON", http.StatusInternalServerError)
		return
	}
	
	time.Sleep(time.Duration(globals.Configmemory.Delay_response) * time.Millisecond) //nos dan los milisegundos o lo dejamos así?

	w.WriteHeader(http.StatusOK)
	w.Write(respuesta)
}

func LeerDeMemoria(direccion_fisica int, tamaño int) string {
	/*Ante un pedido de lectura, devolver el valor que se encuentra a partir de la dirección física pedida.*/
	if (direccion_fisica + tamaño) < len(globals.User_Memory) { //
	 contenido := globals.User_Memory[direccion_fisica : direccion_fisica+tamaño] //ver si hay que restar bytes o si está ok
	 return string(contenido)
	} else{
		return "Error: dirección fuera de rango"
	}
}
type BodyRequestEscribir struct {
	Direccion_fisica int `json:"direccion_fisica"`
	Valor_a_escribir string `json:"valor_a_escribir"`
	Desplazamiento int `json:"desplazamiento"`
}

func EscribirMemoria(w http.ResponseWriter, r *http.Request) {
	var request BodyRequestEscribir
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	respuesta, err := json.Marshal(EscribirEnMemoria(request.Direccion_fisica, request.Valor_a_escribir, request.Desplazamiento))
	if err != nil {
		http.Error(w, "Error al codificar los datos como JSON", http.StatusInternalServerError)
		return
	}

	time.Sleep(time.Duration(globals.Configmemory.Delay_response) * time.Millisecond) //nos dan los milisegundos o lo dejamos así?

	w.WriteHeader(http.StatusOK)
	w.Write(respuesta)
}

func EscribirEnMemoria(direccion_fisica int, valor string, desplazamiento int) string { //TODO: tenemos que validar que al proceso le corresponda escribir ahí o ya la validación la hizo cpu al traducir la dirección?
	/*Ante un pedido de escritura, escribir lo indicado a partir de la dirección física pedida.
     En caso satisfactorio se responderá un mensaje de ‘OK’.*/
    bytesValor := []byte(valor)
    if  (direccion_fisica + len(bytesValor) > len(bytesValor) - desplazamiento) { //todo: validar si no le alcanza una pagina
        return "Error: dirección o tamaño fuera de rango"
    }
    copy(globals.User_Memory[direccion_fisica:], bytesValor)
    return "OK"
}

//--------------------------------------------------------------------------------------//
// BITMAP AUXILIAR
func NewBitMap(size int) BitMap {
	NewBMAp := make(BitMap,size)
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
  	return  globals.CurrentBitMap[i] == 0
}