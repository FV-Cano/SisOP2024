package slice

/**
 * RemoveAtIndex: Remueve un elemento de un slice en base a su índice.

 * @param slice: Slice de cualquier tipo.
 * @param index: Índice del elemento a remover.
 * @return []T: Slice sin el elemento removido.
 */
func RemoveAtIndex[T any](slice []T, index int) []T {
	return append(slice[:index], slice[index+1:]...)
}

/**
 * Pop: Remueve el último elemento de un slice

 * @param slice: Slice de cualquier tipo.
 * @return []T: Slice sin el último elemento.
 * @return T: Último elemento del slice.
*/
func Pop[T any](slice []T) ([]T, T) {
	last := slice[len(slice)-1]
	return slice[:len(slice)-1], last
}

/**
 * Push: Agrega un elemento al final de un slice

 * @param slice: Slice de cualquier tipo.
 * @param elem: Elemento a agregar.
 * @return []T: Slice con el elemento agregado.
*/
func Push[T any](slice []T, elem T) []T {
	return append(slice, elem)
}