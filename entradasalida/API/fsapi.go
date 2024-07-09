package IO_api

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"math"
	"os"
	"strings"

	"github.com/sisoputnfrba/tp-golang/entradasalida/globals"
)

type BitMap []int

/**
 * InicializarFS: inicializa el sistema de archivos (crea el directorio y el archivo de bitmap) o carga el sistema de archivos desde disco (bitmap.dat y bloques.dat)
 */
func InicializarFS() {
	// * Define el nombre del directorio y los archivos
	dirName := "dialfs"
	bitmapFile := dirName + "/bitmap.dat"
	blocksFile := dirName + "/bloques.dat"

	// * Crea el directorio si no existe
	if _, err := os.Stat(dirName); os.IsNotExist(err) {
		err = os.Mkdir(dirName, 0755)
		if err != nil {
			log.Fatalf("Failed creating directory: %s", err)
		}
	}

	// * Crea el archivo de bitmap si no existe
	if _, err := os.Stat(bitmapFile); os.IsNotExist(err) {
		// Carga el bitmap en globals
		globals.CurrentBitMap = NewBitMap(globals.ConfigIO.Dialfs_block_count)
		ActualizarBitmap()

	} else {
		// Carga el contenido del archivo
		file, err := os.Open(bitmapFile)
		if err != nil {
			log.Fatalf("Failed opening file: %s", err)
		}
		defer file.Close()

		// Lee el contenido del archivo
		contenidoBitmap := make([]byte, globals.ConfigIO.Dialfs_block_count)
		_, err = file.Read(contenidoBitmap)
		if err != nil {
			log.Fatalf("Failed reading file: %s", err)
		}

		// Convierte el contenido a JSON y lo carga en globals
		err = json.Unmarshal(contenidoBitmap, &globals.CurrentBitMap)
		if err != nil {
			log.Fatalf("Failed to unmarshal contenido: %s", err)
		}

		// Verifica que el tamaño del bitmap sea el correcto
		if len(globals.CurrentBitMap) != globals.ConfigIO.Dialfs_block_count {
			log.Fatalf("Bitmap size is incorrect")
		}
	}

	// * Crea el archivo de bloques si no existe
	if _, err := os.Stat(blocksFile); os.IsNotExist(err) {
		// Carga los bloques en globals
		globals.Blocks = make([]byte, globals.ConfigIO.Dialfs_block_count*globals.ConfigIO.Dialfs_block_size)
		ActualizarBloques()

	} else {
		// Carga el contenido del archivo
		file, err := os.Open(blocksFile)
		if err != nil {
			log.Fatalf("Failed opening file: %s", err)
		}
		defer file.Close()

		// Lee el contenido del archivo
		contenidoBloques := make([]byte, globals.ConfigIO.Dialfs_block_count*globals.ConfigIO.Dialfs_block_size)
		_, err = file.Read(contenidoBloques)
		if err != nil {
			log.Fatalf("Failed reading file: %s", err)
		}

		// Convierte el contenido a JSON y lo carga en globals
		err = json.Unmarshal(contenidoBloques, &globals.Blocks)
		if err != nil {
			log.Fatalf("Failed to unmarshal contenido: %s", err)
		}

		// Verifica que el tamaño del slice de bloques sea el correcto
		if len(globals.Blocks) != globals.ConfigIO.Dialfs_block_count*globals.ConfigIO.Dialfs_block_size {
			log.Fatalf("Blocks slice size is incorrect")
		}
	}

	// * Carga los archivos metadata del directorio si existen
	globals.Fcbs = make(map[string]globals.Metadata)
	files, err := os.ReadDir(dirName)
	if err != nil {
		log.Fatalf("Failed reading directory: %s", err)
	}
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".txt") {
			continue
		}
		archivo := LeerArchivoEnStruct(dirName + "/" + file.Name())
		globals.Fcbs[file.Name()] = *archivo
	}
}

/**
  * CrearModificarArchivo: carga un archivo en el sistema de archivos

  - @param nombreArchivo: nombre del archivo a cargar
  - @param contenido: contenido del archivo a cargar (en bytes)
*/
func CrearModificarArchivo(nombreArchivo string, contenido []byte) {
	var file *os.File

	// Crea un nuevo archivo si no existe
	if _, err := os.Stat(nombreArchivo); os.IsNotExist(err) {
		file, err = os.Create(nombreArchivo)
		if err != nil {
			log.Fatalf("Failed creating file: %s", err)
		}
	} else {
		file, err = os.Open(nombreArchivo)
		if err != nil {
			log.Fatalf("Failed opening file: %s", err)
		}
	}

	// Cierra el archivo al final de la función
	defer file.Close()

	// Escribe el contenido en el archivo
	_, err := file.Write(contenido)
	if err != nil {
		log.Fatalf("Failed writing to file: %s", err)
	}

	// Guarda los cambios en el archivo
	err = file.Sync()
	if err != nil {
		log.Fatalf("Failed syncing file: %s", err)
	}
}

// * Funciones para ciclo de instrucción

/**
 * CreateFile: se define un nuevo archivo y se lo posiciona en el sistema de archivos (se crea el FCB y se lo agrega al directorio)
 */
func CreateFile(pid int, nombreArchivo string) {
	//Se crea el archivo metadata, con el size en 0 y 1 bloque asignado
	//Archivos de metadata en el módulo FS cargados en alguna estructura para que les sea fácil acceder

	bloqueInicial := CalcularBloqueLibre()
	Set(bloqueInicial)

	// Crea la metadata
	metadata := globals.Metadata{
		InitialBlock: bloqueInicial,
		Size:         0,
	}

	// Convierte la metadata a JSON
	metadataJson, err := json.Marshal(metadata)
	if err != nil {
		log.Fatalf("Failed to marshal metadata: %s", err)
	}

	// Crea el archivo
	CrearModificarArchivo(nombreArchivo, metadataJson)

	// Agrego el FCB al directorio
	globals.Fcbs[nombreArchivo] = metadata
	log.Printf("PID: %d - Crear Archivo: %s", pid, nombreArchivo)
}

// TODO: Tirar excepción si no hay bloques libres
func CalcularBloqueLibre() int {
	var i = 0
	for i < globals.ConfigIO.Dialfs_block_count {
		if IsNotSet(i) {
			break
		}
		i++
	}
	return i
}
func LiberarBloquesDesde(numBloque int, tamanioABorrar int) {
	var i = numBloque
	var contador = 0 // Inicializa el contador
	for contador < tamanioABorrar {
		if !IsNotSet(i) {
			Clear(i)
			contador++
		} else {
			break // Rompe el bucle
		}
		i++
	}
	ActualizarBitmap()
}

func LiberarBloque(bloque int, tamanioABorrar int) {
    for i := 0; i < tamanioABorrar; i++ {
        Clear(bloque - i)
    }
	ActualizarBitmap()
}

func OcuparBloquesDesde(numBloque int, tamanioASetear int) {
	var i = numBloque
	var contador = 0                // Inicializa el contador
	for contador < tamanioASetear { // Continúa mientras el contador sea menor que tamanioASetear
		if IsNotSet(i) { // Si el bloque actual no está seteado
			Set(i)     // Setea el bloque
			contador++ // Incrementa el contador
		} else { // Si el bloque ya está seteado
			break // Rompe el bucle
		}
		i++ // Incrementa el índice para revisar el siguiente bloque
	}
	ActualizarBitmap()
}

/**
 * DeleteFile: elimina un archivo del sistema de archivos y su FCB asociado (incluye liberar los bloques de datos)
 */
func DeleteFile(pid int, nombreArchivo string) error {
	// Paso 1: Verificar si el archivo existe
	if _, err := os.Stat(nombreArchivo); os.IsNotExist(err) {
		// El archivo no existe
		return errors.New("el archivo no existe")
	}

	// Paso 2: Eliminar el archivo del sistema de archivos
	err := os.Remove(nombreArchivo)
	if err != nil {
		// Error al intentar eliminar el archivo
		return err
	}

	// Paso 3: Eliminar el FCB asociado y liberar los bloques de datos
	delete(globals.Fcbs, nombreArchivo)

	archivo := LeerArchivoEnStruct(nombreArchivo) //TODO: VER QUE NO LO LEA SIEMPRE
	sizeArchivo := archivo.Size
	sizeArchivoEnBloques := int(math.Ceil(float64(sizeArchivo) / float64(globals.ConfigIO.Dialfs_block_size)))
	bloqueInicial := archivo.InitialBlock

	for i := 0; i < sizeArchivoEnBloques; i++ {
		Clear(bloqueInicial + i)
	}
	log.Printf("PID: %d - Eliminar Archivo: %s", pid, nombreArchivo)
	return nil
}

//var archivo = LeerArchivoEnStruct(nombreArchivo) //TODO: para que no lo lea siempre

func LeerArchivoEnStruct(nombreArchivo string) *globals.Metadata {
	// Paso 2: Abrir el archivo
	archivo, err := os.Open(nombreArchivo)
	if err != nil {
		return nil
	}
	defer archivo.Close()

	// Paso 3 y 4: Leer y deserializar el contenido del archivo en el struct
	bytes, err := io.ReadAll(archivo)
	if err != nil {
		return nil
	}

	var metadata globals.Metadata
	err = json.Unmarshal(bytes, &metadata)
	if err != nil {
		return nil
	}

	// El archivo se cierra automáticamente gracias a defer
	return &metadata
}

/**
 * ReadFile: Lee un archivo del sistema de archivos
*/
func ReadFile(pid int, nombreArchivo string, direccion []globals.DireccionTamanio, tamanio int, puntero int) {
	// Leer un bloque de fs implica escribirlo en memoria
	//(Interfaz, Nombre Archivo, Registro Dirección, Registro Tamaño, Registro Puntero Archivo): 
	//Esta instrucción solicita al Kernel que mediante la interfaz seleccionada, se lea desde el
	//archivo a partir del valor del Registro Puntero Archivo la cantidad de bytes indicada por Registro
	//Tamaño y se escriban en la Memoria a partir de la dirección lógica indicada en el Registro Dirección.
	
	var contenidoALeer []byte

	primerBloqueArchivo := globals.Fcbs[nombreArchivo].InitialBlock * globals.ConfigIO.Dialfs_block_size
	posicionPuntero := primerBloqueArchivo + puntero 
	limite := tamanio + posicionPuntero
	tamanioTotalArchivo := primerBloqueArchivo + globals.Fcbs[nombreArchivo].Size

	// verificar tamaño del archivo a leer valido
	if(limite > tamanioTotalArchivo){
		Sprintf("El tamaño a leer es superior a el correspondiente del archivo")
	} else {
		contenidoALeer = ReadFs(nombreArchivo, posicionPuntero, limite)
		IO_DIALFS_READ(pid,direccionesFisicas,string(contenidoALeer))	
	}

	log.Printf("PID: %d - Leer Archivo: %s - Tamaño a Leer: %d - Puntero Archivo: %d", pid, nombreArchivo, tamanio, puntero)

}

/**
 * WriteFile: Escribe un archivo del sistema de archivos
 */
func WriteFile(pid int, nombreArchivo string, direccion int, tamanio int, puntero int) {
	// Escritura de un bloque de fs implica leerlo de memoria para luego escribirlo en fs
	log.Printf("PID: %d - Escribir Archivo: %s - Tamaño a Escribir: %d - Puntero Archivo: %d", pid, nombreArchivo, tamanio, puntero)
}

/**
 * TruncateFile: Trunca un archivo del sistema de archivos (puede incluir compactar el archivo)
 */
func TruncateFile(pid int, nombreArchivo string, tamanio int) { //revisar si tiene que devolver un msje
	archivo := globals.Fcbs[nombreArchivo]

	tamanioFinalEnBloques := int(math.Ceil(float64(tamanio) / float64(globals.ConfigIO.Dialfs_block_size)))

	bloqueInicial := archivo.InitialBlock
	tamArchivoOriginalEnBloques := int(math.Ceil(float64(archivo.Size) / float64(globals.ConfigIO.Dialfs_block_size)))
	bloqueFinalInicial := bloqueInicial + tamArchivoOriginalEnBloques - 1 //es decir el final del archivo actual //TODO: revisar si es -1 o no

	// Chequeamos si el archivo tiene que crecer o achicarse
	// Si el archivo crece
	if tamanio > archivo.Size {
		tamanioATruncarEnBytes := tamanio - archivo.Size
		tamanioATruncarEnBloques := int(math.Ceil(float64(tamanioATruncarEnBytes) / float64(globals.ConfigIO.Dialfs_block_size)))
	
		bloquesLibresAlFinalDelArchivo := CalcularBloquesLibreAPartirDe(bloqueFinalInicial)

		// Hay bloques libres al final del archivo
		if bloquesLibresAlFinalDelArchivo >= tamanioATruncarEnBloques {
			//le alcanza como está, seteamos los bloques finales
			OcuparBloquesDesde(bloqueFinalInicial, tamanioATruncarEnBloques)
			archivo.Size = tamanio

			archivoMarshallado, err := json.Marshal(archivo)
			if err != nil {
				log.Fatalf("Failed to marshal metadata: %s", err)
			}
			
			CrearModificarArchivo(nombreArchivo, archivoMarshallado)

		// No hay bloques libres al final del archivo
		} else if bloquesLibresAlFinalDelArchivo < tamanioATruncarEnBloques {
			// Buscar si entra en algún lugar el archivo completo
			bloqueEnElQueEntro := EntraEnDisco(tamArchivoOriginalEnBloques)

			// Entra en algún lugar el archivo completo todo junto
			if bloqueEnElQueEntro != (-1) {
				// Limpiar el bitmap
				LiberarBloquesDesde(bloqueInicial, tamArchivoOriginalEnBloques)
				archivo.InitialBlock = bloqueEnElQueEntro
				archivo.Size = tamanio
				// Setear en el nuevo lugar
				OcuparBloquesDesde(bloqueEnElQueEntro, tamanioFinalEnBloques)

				contenidoArchivo := ReadFs(nombreArchivo, 0, -1)
				WriteFs(contenidoArchivo, bloqueEnElQueEntro)
				
				archivoMarshallado, err := json.Marshal(archivo)
				if err != nil {
					log.Fatalf("Failed to marshal metadata: %s", err)
				}

				CrearModificarArchivo(nombreArchivo, archivoMarshallado)
		
			// No entra el archivo pero alcanza la cantidad de bloques libres
			} else if ContadorDeEspaciosLibres() >= tamanioFinalEnBloques {
				//Si no entra en ningún lugar, compactar
				contenidoArchivo := ReadFs(nombreArchivo, 0, -1)
				Compactar()
				primerBloqueLibre := CalcularBloqueLibre()
				WriteFs(contenidoArchivo, primerBloqueLibre)

				archivo.InitialBlock = primerBloqueLibre
				archivo.Size = tamanio

				archivoMarshallado, err := json.Marshal(archivo)
				if err != nil {
					log.Fatalf("Failed to marshal metadata: %s", err)
				}
				CrearModificarArchivo(nombreArchivo, archivoMarshallado)
			} else {
				log.Print("No hay espacio suficiente para el archivo.")
			}
		}

	// Si el archivo se achica
	} else if tamanio < archivo.Size {
		tamanioATruncarEnBytes := archivo.Size - tamanio
		tamanioATruncarEnBloques := int(math.Ceil(float64(tamanioATruncarEnBytes) / float64(globals.ConfigIO.Dialfs_block_size)))
		// Liberar los bloques que ya no se usan
		LiberarBloque(bloqueFinalInicial, tamanioATruncarEnBloques)
		archivo.Size = tamanio

		archivoMarshallado, err := json.Marshal(archivo)
		if err != nil {
			log.Fatalf("Failed to marshal metadata: %s", err)
		}
		CrearModificarArchivo(nombreArchivo, archivoMarshallado)
	}

	log.Printf("PID: %d - Truncar Archivo: %s - Tamaño: %d" , pid, nombreArchivo, tamanio)
}

/**
 * ReadFs: Lee un archivo del sistema de archivos

	* @param nombreArchivo: nombre del archivo a leer
	* @param desplazamiento: desplazamiento en bytes desde el inicio del archivo
	* @param tamanio: cantidad de bytes a leer (si es -1, se lee todo el archivo)
	* @return contenido: contenido del archivo leído
 */
func ReadFs(nombreArchivo string, desplazamiento int, tamanio int) []byte {
	archivo := globals.Fcbs[nombreArchivo]
	tamanioALeer := tamanio == -1 ? archivo.Size : tamanio

	contenido := make([]byte, tamanioALeer)
	byteInicial := (archivo.InitialBlock * globals.ConfigIO.Dialfs_block_size) + desplazamiento

	for i := 0; i < tamanioALeer; i++ {
		contenido[i] = globals.Blocks[byteInicial + i]
	}

	return contenido
}

func WriteFs(contenido []byte, bloqueInicial int) {
	bloqueInicialEnBytes := bloqueInicial * globals.ConfigIO.Dialfs_block_size

	for i := 0; i < len(contenido); i++ {
		globals.Blocks[bloqueInicialEnBytes + i] = contenido[i]
	}

	tamanioFinalEnBloques := int(math.Ceil(float64(len(contenido)) / float64(globals.ConfigIO.Dialfs_block_size)))
	OcuparBloquesDesde(bloqueInicial, tamanioFinalEnBloques)
	ActualizarBloques()
}


func EntraEnDisco(tamanioTotalEnBloques int) int {

	for i := 0; i < globals.ConfigIO.Dialfs_block_count; i++ {
		espacioActual := CalcularBloquesLibreAPartirDe(i)

		if espacioActual >= tamanioTotalEnBloques {
			return i
		} else {

			i += espacioActual

		}
	}
	return (-1)
}


func ContadorDeEspaciosLibres() int {
	var contador = 0
	for i := 0; i < globals.ConfigIO.Dialfs_block_count; i++ {
		if globals.CurrentBitMap[i] == 0 {
			contador++
		}
	}
	return contador
}

/*func EncontrarEspacioLibreMasGrande() int {
    maxEspacioLibre := 0
    for i := 0; i < globals.ConfigIO.Dialfs_block_count; i++ {
        espacioActual := CalcularBloquesLibreAPartirDe(i)
        if espacioActual > maxEspacioLibre {
            maxEspacioLibre = espacioActual
        }
        // Omitir los bloques ya considerados en el espacio actual para optimizar
        i += espacioActual
    }
    return maxEspacioLibre
}*/

func ActualizarBitmap() {
	archivoMarshallado, err := json.Marshal(globals.CurrentBitMap)
	if err != nil {
		log.Fatalf("Failed to marshal metadata: %s", err)
	}
	CrearModificarArchivo("dialfs/bitmap.dat", archivoMarshallado)
}

func ActualizarBloques(){
	CrearModificarArchivo("dialfs/bloques.dat", globals.Blocks)
}

func CalcularBloquesLibreAPartirDe(bloqueInicial int) int { //esta funcion va a calcular los bloques libres al final del archivo hasta que encuentre uno seteado y ahí debería cortar
	var i = bloqueInicial
	var contadorLibres = 0 // Inicializamos el contador de bloques libres
	for i < globals.ConfigIO.Dialfs_block_count {
		if IsNotSet(i) { // Si el bloque actual no está seteado (es 0),
			contadorLibres++ // Incrementamos el contador de bloques libres
		} else { // Si encontramos un bloque seteado (es 1),
			break // Terminamos la iteración
		}
		i++ // Pasamos al siguiente bloque
	}
	return contadorLibres // Devolvemos el contador de bloques libres

}

func Compactar() {
	var bloquesDeArchivos []byte
	c := 0

	for _, metadata := range globals.Fcbs {

		tamanioArchivoEnBloques := int(math.Ceil(float64(metadata.Size) / float64(globals.ConfigIO.Dialfs_block_size)))

		limite := metadata.InitialBlock + tamanioArchivoEnBloques
		nuevoBloqueInicial := c
		for i := metadata.InitialBlock; i < limite; i++ {
			bloquesDeArchivos = append(bloquesDeArchivos, globals.Blocks[i])
			c++
		}

		metadata.InitialBlock = nuevoBloqueInicial
	}

	for i := range globals.CurrentBitMap {
		Clear(i)
	}

	globals.Blocks = bloquesDeArchivos

	for i := range bloquesDeArchivos {
		Set(i)
	}
	ActualizarBloques()
	ActualizarBitmap()
}


func IO_DIALFS_READ(int pid, direccionesFisicas []globals.DireccionTamanio, string contenido) {

	// Le pido a memoria que me guarde los datos
	url := fmt.Sprintf("http://%s:%d/write", globals.ConfigIO.Ip_memory, globals.ConfigIO.Port_memory)

	bodyWrite, err := json.Marshal(struct {
		DireccionesTamanios []globals.DireccionTamanio  `json:"direcciones_tamanios"`
		Valor_a_escribir    string 					    `json:"valor_a_escribir"`
		Pid                 int 						`json:"pid"`
	} {direccionesFisicas,contenido,pid})
	if err != nil {
		log.Printf("Failed to encode data: %v", err)
	}

	response, err := http.Post(url, "application/json", bytes.NewBuffer(bodyWrite))
	if err != nil {
		log.Printf("Failed to send data: %v", err)
	}

	if response.StatusCode != http.StatusOK {
		log.Printf("Unexpected response status: %s", response.Status)
	}
}

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
