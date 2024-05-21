package pcb

// Estructura PCB que comparten tanto el kernel como el CPU

/* type T_CPU_reg struct {
	AX 				uint8 			
	BX 				uint8 			
	CX 				uint8 			
	DX 				uint8 			
	EAX 			uint32			
	EBX				uint32
	ECX 			uint32 			
	EDX 			uint32 			
} */

// Mapa de registros de CPU
var CPU_reg = make(map[string]interface{})


func Init_Registers(){
	CPU_reg["AX"] = uint8(0)
	CPU_reg["BX"] = uint8(0)
	CPU_reg["CX"] = uint8(0)
	CPU_reg["DX"] = uint8(0)
	CPU_reg["EAX"] = uint32(0)
	CPU_reg["EBX"] = uint32(0)
	CPU_reg["ECX"] = uint32(0)
	CPU_reg["EDX"] = uint32(0)

}
type T_PCB struct {
	PID 			uint32 						`json:"pid"`
	PC 				uint32 						`json:"pc"`
	Quantum 		uint32 						`json:"quantum"`
	CPU_reg 		map[string]interface{} 		`json:"cpu_reg"`	
	// TODO: "Estructura que contendrá los valores de los registros de uso general de la CPU". Tiene los valores enteros de AX, BX, CX, DX, etc., o es una estructura que se crea a partir de otra que represente los registros de la CPU?  
	State 			string 						`json:"state"`
	EvictionReason 	string  					`json:"eviction_reason"`
}

// Canal global de finalización de proceso
var Finished chan bool

var InterruptFlag bool

// Interruption codes:
const (
	QUANTUM = 0
)
