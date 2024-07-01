package fcb

// Estructura FCB que comparten tanto interfaces como el memoria
type T_FCB struct {
	Nombre 				string 			`json:"name"`
	BloqueInicial 		int				`json:"initial_block"`
	Tamanio 			int 			`json:"size"`	
}