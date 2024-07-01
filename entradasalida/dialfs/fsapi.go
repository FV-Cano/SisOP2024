package FS_api

/**
 * CreateFile: se define un nuevo archivo y se lo posiciona en el sistema de archivos (se crea el FCB y se lo agrega al directorio)
 */
func CreateFile(nombreArchivo string) {
	//Se crea el archivo metadata, con el size en 0 y 1 bloque asignado
	//Archivos de metadata en el módulo FS cargados en alguna estructura para que les sea fácil acceder
}

/**
 * DeleteFile: elimina un archivo del sistema de archivos y su FCB asociado (incluye liberar los bloques de datos)
*/
func DeleteFile(nombreArchivo string) {

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

}