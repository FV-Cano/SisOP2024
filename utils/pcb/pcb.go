package pcb

// Estructura PCB que comparten tanto el kernel como el CPU

type T_PCB struct {
	PID 			uint32 	`json:"pid"`
	PC 				uint32 	`json:"pc"`
	Quantum 		uint32 	`json:"quantum"`
	CPU_reg 		[8]int 	`json:"cpu_reg"`	
	// TODO: "Estructura que contendrá los valores de los registros de uso general de la CPU". Tiene los valores enteros de AX, BX, CX, DX, etc., o es una estructura que se crea a partir de otra que represente los registros de la CPU?  
	State 			string 	`json:"state"`
	EvictionReason 	string `json:"eviction_reason"`
}

var pcbPrueba T_PCB

