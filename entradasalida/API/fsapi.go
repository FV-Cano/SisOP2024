package IO_api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/sisoputnfrba/tp-golang/entradasalida/globals"
	ioutils "github.com/sisoputnfrba/tp-golang/entradasalida/utils"
)

/**
 * InicializarFS: inicializa el sistema de archivos (crea el directorio y el archivo de bitmap) o carga el sistema de archivos desde disco (bitmap.dat y bloques.dat)
 */
func InicializarFS() {
	// * Define el nombre del directorio y los archivos
	dirName := "dialfs"
	bitmapFile := "bitmap.dat"
	blocksFile := "bloques.dat"

	// * Crea el directorio si no existe
	if _, err := os.Stat(dirName); os.IsNotExist(err) {
		err = os.Mkdir(dirName, 0755)
		if err != nil {
			log.Fatalf("Failed creating directory: %s", err)
		}
	}

	// * Crea el archivo de bitmap si no existe
	if _, err := os.Stat(dirName + "/" + bitmapFile); os.IsNotExist(err) {
		// Carga elf bitmap en globals
		globals.CurrentBitMap = ioutils.NewBitMap(globals.ConfigIO.Dialfs_block_count)
		ioutils.ActualizarBitmap()

	} else {
		// Carga el contenido del archivo
		file, err := os.Open(dirName + "/" + bitmapFile)
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
	if _, err := os.Stat(dirName + "/" + blocksFile); os.IsNotExist(err) {
		// Carga los bloques en globals
		globals.Blocks = make([]byte, globals.ConfigIO.Dialfs_block_count*globals.ConfigIO.Dialfs_block_size)
		ioutils.ActualizarBloques()

	} else {
		// Carga el contenido del archivo
		file, err := os.Open(dirName + "/" + blocksFile)
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
		archivo := ioutils.LeerArchivoEnStruct(dirName + "/" + file.Name())
		globals.Fcbs[file.Name()] = *archivo
	}
}

// * Funciones para ciclo de instrucción

/**
 * CreateFile: se define un nuevo archivo y se lo posiciona en el sistema de archivos (se crea el FCB y se lo agrega al directorio)
 */
func CreateFile(pid int, nombreArchivo string) {
	//Se crea el archivo metadata, con el size en 0 y 1 bloque asignado
	//Archivos de metadata en el módulo FS cargados en alguna estructura para que les sea fácil acceder

	bloqueInicial := ioutils.CalcularBloqueLibre()
	ioutils.Set(bloqueInicial)

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
	ioutils.CrearModificarArchivo(nombreArchivo, metadataJson)

	// Agrego el FCB al directorio
	globals.Fcbs[nombreArchivo] = metadata
	log.Printf("PID: %d - Crear Archivo: %s", pid, nombreArchivo)
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

	archivo := ioutils.LeerArchivoEnStruct(nombreArchivo) //TODO: VER QUE NO LO LEA SIEMPRE
	sizeArchivo := archivo.Size
	sizeArchivoEnBloques := int(math.Ceil(float64(sizeArchivo) / float64(globals.ConfigIO.Dialfs_block_size)))
	bloqueInicial := archivo.InitialBlock

	for i := 0; i < sizeArchivoEnBloques; i++ {
		ioutils.Clear(bloqueInicial + i)
	}
	log.Printf("PID: %d - Eliminar Archivo: %s", pid, nombreArchivo)
	return nil
}

/*
*
  - ReadFile: Lee un archivo del sistema de archivos
    // Leer un bloque de fs implica escribirlo en memoria
    (Interfaz, Nombre Archivo, Registro Dirección, Registro Tamaño, Registro Puntero Archivo):
    Esta instrucción solicita al Kernel que mediante la interfaz seleccionada, se lea desde el
    archivo a partir del valor del Registro Puntero Archivo la cantidad de bytes indicada por Registro
    Tamaño y se escriban en la Memoria a partir de la dirección lógica indicada en el Registro Dirección.
*/
func ReadFile(pid int, nombreArchivo string, direccion []globals.DireccionTamanio, tamanio int, puntero int) {
	var contenidoALeer []byte

	primerByteArchivo := globals.Fcbs[nombreArchivo].InitialBlock * globals.ConfigIO.Dialfs_block_size
	posicionPuntero := primerByteArchivo + puntero
	limite := tamanio + posicionPuntero
	tamanioTotalArchivo := primerByteArchivo + globals.Fcbs[nombreArchivo].Size

	// verificar tamaño del archivo a leer valido
	if limite > tamanioTotalArchivo {
		log.Println("El tamaño a leer es superior a el correspondiente del archivo")
	} else {
		contenidoALeer = ioutils.ReadFs(nombreArchivo, posicionPuntero, limite)
		IO_DIALFS_READ(pid, direccion, string(contenidoALeer))

	}

	log.Printf("PID: %d - Leer Archivo: %s - Tamaño a Leer: %d - Puntero Archivo: %d", pid, nombreArchivo, tamanio, puntero)

}

/*
*
  - WriteFile: Escribe un archivo del sistema de archivos
    IO_FS_WRITE (Interfaz, Nombre Archivo, Registro Dirección, Registro Tamaño, Registro Puntero Archivo):
    Esta instrucción solicita al Kernel que mediante la interfaz seleccionada, se lea desde Memoria
    la cantidad de bytes indicadas por el Registro Tamaño a partir de la dirección lógica que se encuentra
    en el Registro Dirección y se escriban en el archivo a partir del valor del Registro Puntero Archivo.
*/
func WriteFile(pid int, nombreArchivo string, direccion []globals.DireccionTamanio, tamanio int, puntero int) {
	// Escritura de un bloque de fs implica leerlo de memoria para luego escribirlo en fs

	primerByteArchivo := globals.Fcbs[nombreArchivo].InitialBlock * globals.ConfigIO.Dialfs_block_size
	punteroEnArchivo := primerByteArchivo + puntero
	ultimoByteArchivo := globals.Fcbs[nombreArchivo].Size + primerByteArchivo
	leerDeMemoria := IO_DIALFS_WRITE(pid, direccion)
	cantidadBytesLeidos := len(leerDeMemoria)
	cantidadBytesAEscribir := punteroEnArchivo + cantidadBytesLeidos
	cantidadBytesDisponibles := ultimoByteArchivo - punteroEnArchivo

	if cantidadBytesAEscribir > cantidadBytesDisponibles {
		TruncateFile(pid, nombreArchivo, puntero+cantidadBytesLeidos)
	}

	ioutils.WriteFs(leerDeMemoria, punteroEnArchivo)

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
	bloqueFinalInicial := bloqueInicial + tamArchivoOriginalEnBloques //- 1 //es decir el final del archivo actual //TODO: revisar si es -1 o no

	// Chequeamos si el archivo tiene que crecer o achicarse
	// Si el archivo crece
	if tamanio > archivo.Size {
		tamanioATruncarEnBytes := tamanio - archivo.Size
		tamanioATruncarEnBloques := int(math.Ceil(float64(tamanioATruncarEnBytes) / float64(globals.ConfigIO.Dialfs_block_size)))

		bloquesLibresAlFinalDelArchivo := ioutils.CalcularBloquesLibreAPartirDe(bloqueFinalInicial)

		// Hay bloques libres al final del archivo
		if bloquesLibresAlFinalDelArchivo >= tamanioATruncarEnBloques {
			//le alcanza como está, seteamos los bloques finales
			ioutils.OcuparBloquesDesde(bloqueFinalInicial, tamanioATruncarEnBloques)
			archivo.Size = tamanio

			archivoMarshallado, err := json.Marshal(archivo)
			if err != nil {
				log.Fatalf("Failed to marshal metadata: %s", err)
			}

			ioutils.CrearModificarArchivo(nombreArchivo, archivoMarshallado)

			// No hay bloques libres al final del archivo
		} else if bloquesLibresAlFinalDelArchivo < tamanioATruncarEnBloques {
			// Buscar si entra en algún lugar el archivo completo
			bloqueEnElQueEntro := ioutils.EntraEnDisco(tamArchivoOriginalEnBloques)

			// Entra en algún lugar el archivo completo todo junto
			if bloqueEnElQueEntro != (-1) {
				// Limpiar el bitmap
				ioutils.LiberarBloquesDesde(bloqueInicial, tamArchivoOriginalEnBloques)
				archivo.InitialBlock = bloqueEnElQueEntro
				archivo.Size = tamanio
				// Setear en el nuevo lugar
				ioutils.OcuparBloquesDesde(bloqueEnElQueEntro, tamanioFinalEnBloques)

				contenidoArchivo := ioutils.ReadFs(nombreArchivo, 0, -1)
				ioutils.WriteFs(contenidoArchivo, bloqueEnElQueEntro*globals.ConfigIO.Dialfs_block_size)

				archivoMarshallado, err := json.Marshal(archivo)
				if err != nil {
					log.Fatalf("Failed to marshal metadata: %s", err)
				}

				ioutils.CrearModificarArchivo(nombreArchivo, archivoMarshallado)

				// No entra el archivo pero alcanza la cantidad de bloques libres
			} else if ioutils.ContadorDeEspaciosLibres() >= tamanioFinalEnBloques {
				//Si no entra en ningún lugar, compactar
				contenidoArchivo := ioutils.ReadFs(nombreArchivo, 0, -1)
				Compactar()
				primerBloqueLibre := ioutils.CalcularBloqueLibre()
				ioutils.WriteFs(contenidoArchivo, primerBloqueLibre*globals.ConfigIO.Dialfs_block_size)

				archivo.InitialBlock = primerBloqueLibre
				archivo.Size = tamanio

				archivoMarshallado, err := json.Marshal(archivo)
				if err != nil {
					log.Fatalf("Failed to marshal metadata: %s", err)
				}
				ioutils.CrearModificarArchivo(nombreArchivo, archivoMarshallado)
			} else {
				log.Print("No hay espacio suficiente para el archivo.")
			}
		}

		// Si el archivo se achica
	} else if tamanio < archivo.Size {
		tamanioATruncarEnBytes := archivo.Size - tamanio
		tamanioATruncarEnBloques := int(math.Ceil(float64(tamanioATruncarEnBytes) / float64(globals.ConfigIO.Dialfs_block_size)))
		// Liberar los bloques que ya no se usan
		ioutils.LiberarBloque(bloqueFinalInicial, tamanioATruncarEnBloques)
		archivo.Size = tamanio

		archivoMarshallado, err := json.Marshal(archivo)
		if err != nil {
			log.Fatalf("Failed to marshal metadata: %s", err)
		}
		ioutils.CrearModificarArchivo(nombreArchivo, archivoMarshallado)
	}

	log.Printf("PID: %d - Truncar Archivo: %s - Tamaño: %d", pid, nombreArchivo, tamanio)
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
		ioutils.Clear(i)
	}

	globals.Blocks = bloquesDeArchivos

	for i := range bloquesDeArchivos {
		ioutils.Set(i)
	}
	ioutils.ActualizarBloques()
	ioutils.ActualizarBitmap()

	time.Sleep(time.Duration(globals.ConfigIO.Dialfs_compaction_delay))
}

func IO_DIALFS_READ(pid int, direccionesFisicas []globals.DireccionTamanio, contenido string) {

	// Le pido a memoria que me guarde los datos
	url := fmt.Sprintf("http://%s:%d/write", globals.ConfigIO.Ip_memory, globals.ConfigIO.Port_memory)

	bodyWrite, err := json.Marshal(struct {
		DireccionesTamanios []globals.DireccionTamanio `json:"direcciones_tamanios"`
		Valor_a_escribir    string                     `json:"valor_a_escribir"`
		Pid                 int                        `json:"pid"`
	}{direccionesFisicas, contenido, pid})
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

func IO_DIALFS_WRITE(pid int, direccionesFisicas []globals.DireccionTamanio) []byte {

	url := fmt.Sprintf("http://%s:%d/read", globals.ConfigIO.Ip_memory, globals.ConfigIO.Port_memory)

	bodyRead, err := json.Marshal(BodyRequestLeer{
		DireccionesTamanios: direccionesFisicas,
		Pid:                 pid,
	})
	if err != nil {
		return nil
	}

	datosLeidos, err := http.Post(url, "application/json", bytes.NewBuffer(bodyRead))
	if err != nil {
		log.Printf("Failed to receive data: %v", err)
	}

	if datosLeidos.StatusCode != http.StatusOK {
		log.Printf("Unexpected response status: %s", datosLeidos.Status)
	}

	var response [][]byte
	err = json.NewDecoder(datosLeidos.Body).Decode(&response)
	if err != nil {
		return nil
	}

	var bytesConcatenados []byte
	for _, sliceBytes := range response {
		bytesConcatenados = append(bytesConcatenados, sliceBytes...)
	}

	return bytesConcatenados
}
