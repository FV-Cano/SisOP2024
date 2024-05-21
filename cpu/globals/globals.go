package globals

var Configcpu *T_CPU
type T_CPU struct {
	Port               	int    `json:"port"`
	IP_memory          	string `json:"ip_memory"`
	Port_memory        	int    `json:"port_memory"`
	Number_felling_tlb 	int    `json:"number_felling_tlb"`
	Algorithm_tlb      	string `json:"algorithm_tlb"`
}