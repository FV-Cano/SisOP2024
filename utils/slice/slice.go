package slice

/**
 * RemoveAtIndex: Remueve un elemento de un slice en base a su índice.

 * @param slice: Slice de cualquier tipo.
 * @param index: Índice del elemento a remover.
 */
func RemoveAtIndex[T any](slice *[]T, index int) {
	*slice = append((*slice)[:index], (*slice)[index+1:]...)
}

/**
 * Pop: Remueve el último elemento de un slice

 * @param slice: Slice de cualquier tipo.
 * @return T: Último elemento del slice.
*/
func Pop[T any](slice *[]T) T {
	last := (*slice)[len(*slice)-1]
	*slice = (*slice)[:len(*slice)-1]
	return last
}

/**
 * Shift: Remueve el primer elemento de un slice

 * @param slice: Slice de cualquier tipo.
 * @return T: Primer elemento del slice.
*/
func Shift[T any](slice *[]T) T {
	first := (*slice)[0]
	*slice = (*slice)[1:]
	return first
}

/**
 * Push: Agrega un elemento al final de un slice

 * @param slice: Slice de cualquier tipo.
 * @param elem: Elemento a agregar.
*/
func Push[T any](slice *[]T, elem T) {
	*slice = append(*slice, elem)
}
