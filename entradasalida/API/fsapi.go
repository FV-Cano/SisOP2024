package IO_api

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"math"
	"os"

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

		// Convierte el contenido a JSON
		contenidoBitmap, err := json.Marshal(globals.CurrentBitMap)
		if err != nil {
			log.Fatalf("Failed to marshal contenido: %s", err)
		}
		
		// Carga el contenido en el archivo
		CrearModificarArchivo(bitmapFile, contenidoBitmap)
		
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
		globals.Blocks = make([]byte, globals.ConfigIO.Dialfs_block_count * globals.ConfigIO.Dialfs_block_size)

		// Convierte el contenido a JSON
		contenidoBloques, err := json.Marshal(globals.Blocks)
		if err != nil {
			log.Fatalf("Failed to marshal contenido: %s", err)
		}
		
		// Carga el contenido en el archivo
		CrearModificarArchivo(blocksFile, contenidoBloques)

    } else {
		// Carga el contenido del archivo
		file, err := os.Open(blocksFile)
		if err != nil {
			log.Fatalf("Failed opening file: %s", err)
		}
		defer file.Close()

		// Lee el contenido del archivo
		contenidoBloques := make([]byte, globals.ConfigIO.Dialfs_block_count * globals.ConfigIO.Dialfs_block_size)
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
		if len(globals.Blocks) != globals.ConfigIO.Dialfs_block_count * globals.ConfigIO.Dialfs_block_size {
			log.Fatalf("Blocks slice size is incorrect")
		}
	}
}

/**
 * CrearModificarArchivo: carga un archivo en el sistema de archivos

 * @param nombreArchivo: nombre del archivo a cargar
 * @param contenido: contenido del archivo a cargar (en bytes)
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
func CreateFile(nombreArchivo string) {
	//Se crea el archivo metadata, con el size en 0 y 1 bloque asignado
	//Archivos de metadata en el módulo FS cargados en alguna estructura para que les sea fácil acceder

	bloqueInicial := CalcularBloqueLibre()

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
}

func CalcularBloqueLibre() int {
	var i = 0
	for  i < globals.ConfigIO.Dialfs_block_count { 
		if IsNotSet(i) {
			Set(i)
			break
		}
		i++
	}
	return i
}

/**
 * DeleteFile: elimina un archivo del sistema de archivos y su FCB asociado (incluye liberar los bloques de datos)
*/
func DeleteFile(nombreArchivo string) error {
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
	archivo := LeerArchivoEnStruct(nombreArchivo)
	sizeArchivo := archivo.Size 
	sizeArchivoEnBloques := int(math.Ceil(float64(sizeArchivo) / float64(globals.ConfigIO.Dialfs_block_size)))
	bloqueInicial := archivo.InitialBlock
	
	for i := 0; i < sizeArchivoEnBloques; i++ {
		Clear(bloqueInicial + i)
	}

	return nil
}

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
func ReadFile(nombreArchivo string, direccion int, tamanio int, puntero int) {
	// Leer un bloque de fs implica escribirlo en memoria
}

/**
 * WriteFile: Escribe un archivo del sistema de archivos
*/
func WriteFile(nombreArchivo string, direccion int, tamanio int, puntero int) {
	// Escritura de un bloque de fs implica leerlo de memoria para luego escribirlo en fs
}

/**
 * TruncateFile: Trunca un archivo del sistema de archivos (puede incluir compactar el archivo)
*/
func TruncateFile(nombreArchivo string, tamanio int) {
//	


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