package ioutils

import (
	"log"
	"github.com/sisoputnfrba/tp-golang/entradasalida/globals"
	"os"
	"math"
	"encoding/json"
	"io"
)

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
 * ReadFs: Lee un archivo del sistema de archivos

	* @param nombreArchivo: nombre del archivo a leer
	* @param desplazamiento: desplazamiento en bytes desde el inicio del archivo
	* @param tamanio: cantidad de bytes a leer (si es -1, se lee todo el archivo)
	* @return contenido: contenido del archivo leído
 */
 func ReadFs(nombreArchivo string, desplazamiento int, tamanio int) []byte {
	archivo := globals.Fcbs[nombreArchivo]
	var tamanioALeer int
	if tamanio == -1 {
		tamanioALeer = archivo.Size
	} else {
		tamanioALeer = tamanio
	}

	contenido := make([]byte, tamanioALeer)
	byteInicial := (archivo.InitialBlock * globals.ConfigIO.Dialfs_block_size) + desplazamiento

	for i := 0; i < tamanioALeer; i++ {
		contenido[i] = globals.Blocks[byteInicial + i]
	}

	return contenido
}

func WriteFs(contenido []byte, byteInicial int) {
	bloqueInicial := int(math.Ceil(float64(byteInicial) / float64(globals.ConfigIO.Dialfs_block_size)))

	for i := 0; i < len(contenido); i++ {
		globals.Blocks[byteInicial + i] = contenido[i]
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


// * Manejo de BLOQUES
/**
 * ContadorDeEspaciosLibres: cuenta la cantidad de bloques libres TOTAL en el sistema de archivos
 */
func ContadorDeEspaciosLibres() int {
	var contador = 0
	for i := 0; i < globals.ConfigIO.Dialfs_block_count; i++ {
		if globals.CurrentBitMap[i] == 0 {
			contador++
		}
	}
	return contador
}

/**
 * CalcularBloquesLibreAPartirDe: calcula la cantidad de bloques libres a partir de un bloque inicial hasta encontrar uno seteado
 */
func CalcularBloquesLibreAPartirDe(bloqueInicial int) int {
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

/**
 * CalcularBloqueLibre: calcula el primer bloque libre en el sistema de archivos
*/
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

/**
 * LiberarBloquesDesde: libera bloques a partir de un bloque inicial hasta el tamaño a borrar
 */
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

/**
 * LiberarBloque: libera bloques desde el bloque final del archivo hasta el tamaño a borrar
 */
func LiberarBloque(bloque int, tamanioABorrar int) {
    for i := 0; i < tamanioABorrar; i++ {
        Clear(bloque - i)
    }
	ActualizarBitmap()
}

/**
 * OcuparBloquesDesde: ocupa bloques a partir de un bloque inicial hasta el tamaño a setear
 */
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
 * ActualizarBloques: actualiza el archivo de bloques en el sistema de archivos
 */
func ActualizarBloques(){
	CrearModificarArchivo("dialfs/bloques.dat", globals.Blocks)
}


// * Manejo de BITMAP
func NewBitMap(size int) []int {
	NewBMAp := make([]int, size)
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

func ActualizarBitmap() {
	archivoMarshallado, err := json.Marshal(globals.CurrentBitMap)
	if err != nil {
		log.Fatalf("Failed to marshal metadata: %s", err)
	}
	CrearModificarArchivo("dialfs/bitmap.dat", archivoMarshallado)
}