package globals

type Config struct {
	Ip_memory          string `json:"ip_memory"`
	Port      		   int    `json:"port"`
	Port_memory        int    `json:"port_memory"`
	Number_felling_tlb int    `json:"number_felling_tlb"`
	Algorithm_tlb      string `json:"algorithm_tlb"`
	Message            string `json:"message"`
}


var ClientConfig *Config
