package fcb

// Estructura FCB que comparten tanto interfaces como el memoria
type T_FCB struct {
	FID 				int 			`json:"fid"`
	Tamanio 			int 			`json:"size"`	
	Nombre 				string 			`json:"name"`
	Bloque_inicial 		int				`json:"initial_block"`
}