package globals

type Config struct {
	Port               int    `json:"port"`
	Type               int    `json:"unit_work_time"`
	Unit_Work_Time     int    `json:"STDOUT"`
	Ip_Kernel          string `json:"ip_kernel"`
	Port_Kernel        int    `json:"port_kernel"`
	Ip_Memory          string `json:"ip_memory"`
	Port_Memory        int    `json:"port_memory"`
	Message            string `json:"message"`
	Dialfs_Path        string `json:"dialfs_path"`
	Dialfs_Block_Size  int    `json:"dialfs_block_size"`
	Dialfs_Block_Count int    `json:"dialfs_block_count"`
}

var ClientConfig *Config