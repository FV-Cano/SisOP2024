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

		fmt.Println("FS - Directorio CREADO")
	}

	// * Crea el archivo de bitmap si no existe
	if _, err := os.Stat(dirName + "/" + bitmapFile); os.IsNotExist(err) {
		// Carga elf bitmap en globals
		globals.CurrentBitMap = ioutils.NewBitMap(globals.ConfigIO.Dialfs_block_count)
		ioutils.ActualizarBitmap()

		fmt.Println("FS - Bitmap.dat CREADO")

	} else {
		// Carga el contenido del archivo
		file, err := os.Open(dirName + "/" + bitmapFile)
		if err != nil {
			log.Fatalf("Failed opening file: %s", err)
		}
		defer file.Close()

		// Lee el contenido del archivo
		contenidoBitmap, err := os.ReadFile(dirName + "/" + bitmapFile)
		if err != nil {
			log.Fatalf("Failed to read file: %s", err)
		}

		fmt.Println("Quiero leer el contenido del bitmap: ", contenidoBitmap)
		fmt.Println("El contenido del bitmap es: ", string(contenidoBitmap))

		var bitmapAux globals.T_Bitmap
		// Convierte el contenido a JSON y lo carga en globals
		err = json.Unmarshal(contenidoBitmap, &bitmapAux)
		if err != nil {
			log.Fatalf("Failed to unmarshal contenido: %s", err)
		}

		globals.CurrentBitMap = bitmapAux.BitMap

		// Verifica que el tamaño del bitmap sea el correcto
		if len(globals.CurrentBitMap) != globals.ConfigIO.Dialfs_block_count {
			log.Fatalf("Bitmap size is incorrect")
		}

		fmt.Println("FS - Bitmap.dat LEIDO")
	}

	// * Crea el archivo de bloques si no existe
	if _, err := os.Stat(dirName + "/" + blocksFile); os.IsNotExist(err) {
		// Carga los bloques en globals
		globals.Blocks = make([]byte, globals.ConfigIO.Dialfs_block_count*globals.ConfigIO.Dialfs_block_size)
		ioutils.ActualizarBloques()

		fmt.Println("FS - Bloques.dat CREADO")

	} else {
		// Carga el contenido del archivo
		file, err := os.Open(dirName + "/" + blocksFile)
		if err != nil {
			log.Fatalf("Failed opening file: %s", err)
		}
		defer file.Close()

		// Lee el contenido del archivo
		contenidoBloques, err := os.ReadFile(dirName + "/" + blocksFile)
		if err != nil {
			log.Fatalf("Failed to read file: %s", err)
		}

		var bloquesAux globals.T_Blocks

		// Convierte el contenido a JSON y lo carga en globals
		err = json.Unmarshal(contenidoBloques, &bloquesAux)
		if err != nil {
			log.Fatalf("Failed to unmarshal contenido: %s", err)
		}

		globals.Blocks = bloquesAux.Blocks

		// Verifica que el tamaño del slice de bloques sea el correcto
		if len(globals.Blocks) != globals.ConfigIO.Dialfs_block_count*globals.ConfigIO.Dialfs_block_size {
			log.Fatalf("Blocks slice size is incorrect")
		}

		fmt.Println("FS - Bloques.dat LEIDO")
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

		fmt.Println("FS - Archivo CARGADO: ", file.Name())
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
	ioutils.Set(bloqueInicial - 1)

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
	nombreArchivo = "dialfs/" + nombreArchivo

	// Paso 1: Verificar si el archivo existe
	if _, err := os.Stat(nombreArchivo); os.IsNotExist(err) {
		// El archivo no existe
		return errors.New("el archivo no existe")
	}

	// Leer la información del archivo antes de eliminarlo
	archivo := ioutils.LeerArchivoEnStruct(nombreArchivo)

	sizeArchivo := archivo.Size
	sizeArchivoEnBloques := int(math.Max(1, math.Ceil(float64(sizeArchivo)/float64(globals.ConfigIO.Dialfs_block_size))))

	bloqueInicial := archivo.InitialBlock

	// Paso 2: Eliminar el archivo del sistema de archivos
	err := os.Remove(nombreArchivo)
	if err != nil {
		// Error al intentar eliminar el archivo
		return err
	}

	// Paso 3: Eliminar el FCB asociado y liberar los bloques de datos
	delete(globals.Fcbs, nombreArchivo)

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
func ReadFile(pid int, nombreArchivo string, direccionesFisicas []globals.DireccionTamanio, tamanioLeer int, puntero int) {
	var contenidoALeer []byte

	posBloqueInicial := globals.Fcbs[nombreArchivo].InitialBlock - 1
	primerByteArchivo := posBloqueInicial * globals.ConfigIO.Dialfs_block_size

	posicionPuntero := primerByteArchivo + puntero
	limiteLeer := tamanioLeer + posicionPuntero

	limiteArchivo := primerByteArchivo + globals.Fcbs[nombreArchivo].Size

	// Verificar tamaño del archivo a leer valido
	if limiteLeer > limiteArchivo {
		log.Println("El tamaño a leer es superior a el correspondiente del archivo")
	} else {
		contenidoALeer = ioutils.ReadFs(nombreArchivo, puntero, tamanioLeer)
		IO_DIALFS_READ(pid, direccionesFisicas, string(contenidoALeer))
	}

	log.Printf("PID: %d - Leer Archivo: %s - Tamaño a Leer: %d - Puntero Archivo: %d", pid, nombreArchivo, tamanioLeer, puntero)
}

/*
*
  - WriteFile: Escribe un archivo del sistema de archivos
    IO_FS_WRITE (Interfaz, Nombre Archivo, Registro Dirección, Registro Tamaño, Registro Puntero Archivo):
    Esta instrucción solicita al Kernel que mediante la interfaz seleccionada, se lea desde Memoria
    la cantidad de bytes indicadas por el Registro Tamaño a partir de la dirección lógica que se encuentra
    en el Registro Dirección y se escriban en el archivo a partir del valor del Registro Puntero Archivo.
  - Escritura de un bloque de fs implica leerlo de memoria para luego escribirlo en fs
*/
func WriteFile(pid int, nombreArchivo string, direccion []globals.DireccionTamanio, tamanio int, puntero int) {
	archivo := globals.Fcbs[nombreArchivo]

	bloqueInicial := archivo.InitialBlock
	posBloqueInicial := bloqueInicial - 1
	primerByteArchivo := posBloqueInicial * globals.ConfigIO.Dialfs_block_size
	ultimoByteArchivo := primerByteArchivo + archivo.Size

	posicionPuntero := primerByteArchivo + puntero

	leidoEnMemoria := IO_DIALFS_WRITE(pid, direccion)
	cantidadBytesLeidos := len(leidoEnMemoria)

	posInicialByteAEscribir := posicionPuntero + cantidadBytesLeidos
	
	// TODO: Revisar
	cantidadBytesFinales := math.Max(float64(archivo.Size), float64(cantidadBytesLeidos + puntero))

	// 1 2 3 4					archivo original 4 bytes
	// x    					puntero en el byte 2
	// 8 			 			nuevo contenido 4 bytes
	// 8 2 3 4					archivo final 5 bytes
	
	//cantidadBytesDisponibles := ultimoByteArchivo - posicionPuntero

	/* if posInicialByteAEscribir > cantidadBytesDisponibles {
		TruncateFile(pid, nombreArchivo, puntero + cantidadBytesLeidos)
	} */

	if posInicialByteAEscribir > ultimoByteArchivo {
		TruncateFile(pid, nombreArchivo, int(cantidadBytesFinales))
	}

	fmt.Println("CONTENIDO A ESCRIBIR EN FS: ", leidoEnMemoria)

	ioutils.WriteFs(leidoEnMemoria, posicionPuntero)

	log.Printf("PID: %d - Escribir Archivo: %s - Tamaño a Escribir: %d - Puntero Archivo: %d", pid, nombreArchivo, tamanio, puntero)
}

/**
 * TruncateFile: Trunca un archivo del sistema de archivos (puede incluir compactar el archivo)
 */
func TruncateFile(pid int, nombreArchivo string, tamanioDeseado int) { //revisar si tiene que devolver un msje
	archivo := globals.Fcbs[nombreArchivo]

	bloqueInicial := archivo.InitialBlock
	posBloqueInicial := bloqueInicial - 1

	tamOriginalEnBloques := int(math.Max(1, math.Ceil(float64(archivo.Size)/float64(globals.ConfigIO.Dialfs_block_size))))
	tamFinalEnBloques := int(math.Ceil(float64(tamanioDeseado) / float64(globals.ConfigIO.Dialfs_block_size)))

	fmt.Println("El tamaño original del archivo en bloques es ", tamOriginalEnBloques)

	// ? bloqueFinalInicial: El final del archivo actual, bloque donde tiene que arrancar el siguiente archivo
	// TODO: revisar si es -1 o no
	bloqueFinalInicial := bloqueInicial + tamOriginalEnBloques
	//posBloqueFinalInicial := bloqueFinalInicial - 1

	fmt.Println("Bloque inicial archivo actual", bloqueInicial)
	fmt.Println("Bloque final archivo actual", bloqueFinalInicial)

	// Chequeamos si el archivo tiene que crecer o achicarse
	// Si el archivo crece
	if tamanioDeseado > archivo.Size {
		tamanioATruncarEnBytes := tamanioDeseado - archivo.Size
		tamanioATruncarEnBloques := int(math.Ceil(float64(tamanioATruncarEnBytes) / float64(globals.ConfigIO.Dialfs_block_size)))

		if tamanioDeseado < globals.ConfigIO.Dialfs_block_size && archivo.Size < globals.ConfigIO.Dialfs_block_size {
			tamanioATruncarEnBloques = 0
		}

		fmt.Println("Tamaño a agrandar el archivo en bloques", tamanioATruncarEnBloques)

		bloquesLibresAlFinalDelArchivo, _:= ioutils.CalcularBloquesLibreAPartirDe(bloqueFinalInicial, tamanioATruncarEnBloques)

		// Hay bloques libres al final del archivo
		if bloquesLibresAlFinalDelArchivo >= tamanioATruncarEnBloques {
			fmt.Println("FS - Hay bloques libres al final del archivo")

			ioutils.OcuparBloquesDesde(bloqueFinalInicial, tamanioATruncarEnBloques)
			archivo.Size = tamanioDeseado

			archivoMarshallado, err := json.Marshal(archivo)
			if err != nil {
				log.Fatalf("Failed to marshal metadata: %s", err)
			}

			ioutils.CrearModificarArchivo(nombreArchivo, archivoMarshallado)

			// No hay bloques libres al final del archivo
		} else if bloquesLibresAlFinalDelArchivo < tamanioATruncarEnBloques {
			// Buscar si entra en algún lugar el archivo completo
			posbloqueEnElQueEntro := ioutils.EntraEnDisco(tamOriginalEnBloques)
			bloqueEnElQueEntro := posbloqueEnElQueEntro + 1

			fmt.Println("EL ARCHIVO ENTRA EN EL BLOQUE ", bloqueEnElQueEntro)

			// Entra en algún lugar el archivo completo todo junto
			if bloqueEnElQueEntro != -1 {
				fmt.Println("FS - Entra en algún lugar el archivo completo")
				// Limpiar el bitmap
				ioutils.LiberarBloquesDesde(posBloqueInicial+1, tamOriginalEnBloques)
				archivo.InitialBlock = bloqueEnElQueEntro
				archivo.Size = tamanioDeseado
				// Setear en el nuevo lugar
				ioutils.OcuparBloquesDesde(posbloqueEnElQueEntro, tamFinalEnBloques)

				contenidoArchivo := ioutils.ReadFs(nombreArchivo, 0, -1)

				byteInicioDestino := bloqueEnElQueEntro * globals.ConfigIO.Dialfs_block_size
				ioutils.WriteFs(contenidoArchivo, byteInicioDestino)

				archivoMarshallado, err := json.Marshal(archivo)
				if err != nil {
					log.Fatalf("Failed to marshal metadata: %s", err)
				}

				ioutils.CrearModificarArchivo(nombreArchivo, archivoMarshallado)

				// No entra el archivo pero alcanza la cantidad de bloques libres
			} else if ioutils.ContadorDeEspaciosLibres() >= tamFinalEnBloques {
				fmt.Println("FS - No entra el archivo pero alcanza la cantidad de bloques libres")
				//Si no entra en ningún lugar, compactar
				contenidoArchivo := ioutils.ReadFs(nombreArchivo, 0, -1)
				Compactar()
				primerBloqueLibre := ioutils.CalcularBloqueLibre()
				ioutils.WriteFs(contenidoArchivo, primerBloqueLibre*globals.ConfigIO.Dialfs_block_size)

				archivo.InitialBlock = primerBloqueLibre
				archivo.Size = tamanioDeseado

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
	} else if tamanioDeseado < archivo.Size {
		tamanioATruncarEnBytes := archivo.Size - tamanioDeseado
		tamanioATruncarEnBloques := int(math.Ceil(float64(tamanioATruncarEnBytes) / float64(globals.ConfigIO.Dialfs_block_size)))
		// Liberar los bloques que ya no se usan
		ioutils.LiberarBloque(bloqueFinalInicial, tamanioATruncarEnBloques)
		archivo.Size = tamanioDeseado

		archivoMarshallado, err := json.Marshal(archivo)
		if err != nil {
			log.Fatalf("Failed to marshal metadata: %s", err)
		}
		ioutils.CrearModificarArchivo(nombreArchivo, archivoMarshallado)
	}

	log.Printf("PID: %d - Truncar Archivo: %s - Tamaño: %d", pid, nombreArchivo, tamanioDeseado)
}

func Compactar() {
	var bloquesDeArchivos []byte
	c := 0

	for _, metadata := range globals.Fcbs {

		tamanioArchivoEnBloques := int(math.Max(1, math.Ceil(float64(metadata.Size)/float64(globals.ConfigIO.Dialfs_block_size))))

		limite := metadata.InitialBlock + tamanioArchivoEnBloques
		nuevoBloqueInicial := c
		for i := metadata.InitialBlock; i < limite; i++ {
			bloquesDeArchivos = append(bloquesDeArchivos, globals.Blocks[i])
			c++
		}

		metadata.InitialBlock = nuevoBloqueInicial + 1
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

	fmt.Println("IO DIALFS WRITE BODY READ")

	datosLeidos, err := http.Post(url, "application/json", bytes.NewBuffer(bodyRead))
	if err != nil {
		log.Printf("Failed to receive data: %v", err)
	}

	if datosLeidos.StatusCode != http.StatusOK {
		log.Printf("Unexpected response status: %s", datosLeidos.Status)
	}

	var response BodyADevolver
	err = json.NewDecoder(datosLeidos.Body).Decode(&response)
	if err != nil {
		return []byte("error al deserializar la respuesta")
	}

	var bytesConcatenados []byte
	for _, sliceBytes := range response.Contenido {
		bytesConcatenados = append(bytesConcatenados, sliceBytes...)
	}

	fmt.Println("IO_DIALFS_WRITE DATOS CONC: ", bytesConcatenados)

	return bytesConcatenados
}
