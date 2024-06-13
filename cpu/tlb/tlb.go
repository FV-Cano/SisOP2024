package tlb

type Pagina_marco struct {
	Pagina int
	Marco  int
}

type TLB map[int]Pagina_marco

var CurrentTLB TLB = make(TLB)
