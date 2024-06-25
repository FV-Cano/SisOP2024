package IO_api

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/sisoputnfrba/tp-golang/entradasalida/globals"
	"github.com/sisoputnfrba/tp-golang/utils/device"
	"github.com/sisoputnfrba/tp-golang/utils/generics"
	"github.com/sisoputnfrba/tp-golang/utils/pcb"
)

type CantUnidadesTrabajo struct {
	Unidades int `json:"cantUnidades"`
}

func HandshakeKernel(nombre string) error {
	genInterface := device.T_IOInterface{
		InterfaceName: nombre,
		InterfaceType: globals.ConfigIO.Type,
		InterfaceIP: globals.ConfigIO.Ip,
		InterfacePort: globals.ConfigIO.Port,
	}

	jsonData, err := json.Marshal(genInterface)
	if err != nil {
		return fmt.Errorf("failed to encode interface: %v", err)
	}

	url := fmt.Sprintf("http://%s:%d/io-handshake", globals.ConfigIO.Ip_kernel, globals.ConfigIO.Port_kernel)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("POST request failed. Failed to send interface: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected response status: %s", resp.Status)
	}

	log.Println("Handshake con Kernel exitoso")

	return nil
}

// * Hay que declarar los tipos de body que se van a recibir desde kernel porque por alguna razón no se puede crear un struct type dentro de una función con un tipo creado por uno mismo, están todos en globals

func InterfaceQueuePCB(w http.ResponseWriter, r *http.Request) {
	switch globals.ConfigIO.Type {
	case "GENERICA":
		var decodedStruct globals.GenSleep

		err := json.NewDecoder(r.Body).Decode(&decodedStruct)
		if err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		log.Print("Nueva PCB ID: ", decodedStruct.Pcb.PID, " para usar Interfaz")
		globals.Generic_QueueChannel <- decodedStruct
	case "STDIN":
		var decodedStruct globals.StdinRead
		
		err := json.NewDecoder(r.Body).Decode(&decodedStruct)
		if err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
		
		log.Print("Nueva PCB ID: ", decodedStruct.Pcb.PID, " para usar Interfaz")
		globals.Stdin_QueueChannel <- decodedStruct
	case "STDOUT":
		var decodedStruct globals.StdoutWrite

		err := json.NewDecoder(r.Body).Decode(&decodedStruct)
		if err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		log.Print("Nueva PCB ID: ", decodedStruct.Pcb.PID, " para usar Interfaz")
		globals.Stdout_QueueChannel <- decodedStruct
	}
}

func IOWork() {
	switch globals.ConfigIO.Type {
	case "GENERICA":
		var interfaceToWork globals.GenSleep
		for {
			interfaceToWork = <- globals.Generic_QueueChannel

			IO_GEN_SLEEP(interfaceToWork.TimeToSleep, interfaceToWork.Pcb)
			log.Println("Fin de bloqueo")
			returnPCB(interfaceToWork.Pcb)
		}
	case "STDIN":
		var interfaceToWork globals.StdinRead
		for {
			interfaceToWork = <- globals.Stdin_QueueChannel

			IO_STDIN_READ(interfaceToWork.Pcb, interfaceToWork.DireccionesFisicas)
			log.Println("Fin de bloqueo")
			returnPCB(interfaceToWork.Pcb)
		}
	case "STDOUT":
		var interfaceToWork globals.StdoutWrite
		for {
			interfaceToWork = <- globals.Stdout_QueueChannel

			IO_STDOUT_WRITE(interfaceToWork.Pcb, interfaceToWork.DireccionesFisicas)
			log.Println("Fin de bloqueo")
			returnPCB(interfaceToWork.Pcb)
		}
	}
}

func returnPCB(pcb pcb.T_PCB) {
	generics.DoRequest("POST", fmt.Sprintf("http://%s:%d/io-return-pcb", globals.ConfigIO.Ip_kernel, globals.ConfigIO.Port_kernel), pcb, nil)
}
	

// ------------------------- OPERACIONES -------------------------

func IO_GEN_SLEEP(sleepTime int, pcb pcb.T_PCB) {
	sleepTimeTotal := sleepTime * globals.ConfigIO.Unit_work_time
	log.Printf("PID: %d - Operacion: <IO_GEN_SLEEP>", pcb.PID)
	log.Printf("Bloqueado por %d segundos\n", sleepTimeTotal)

	time.Sleep(time.Duration(sleepTimeTotal) * time.Second)
}

func IO_STDIN_READ(pcb pcb.T_PCB, direccionesFisicas []globals.DireccionTamanio) {
	// Lee datos de la entrada
	fmt.Println("TE LO MUESTRO", direccionesFisicas)

	fmt.Print("Ingrese datos: ")
	reader := bufio.NewReader(os.Stdin)
	data, _ := reader.ReadString('\n')

	// Le pido a memoria que me guarde los datos
	url := fmt.Sprintf("http://%s:%d/write", globals.ConfigIO.Ip_memory, globals.ConfigIO.Port_memory)

	bodyWrite, err := json.Marshal(struct {
		DireccionesTamanios []globals.DireccionTamanio  `json:"direcciones_tamanios"`
		Valor_a_escribir    string 					    `json:"valor_a_escribir"`
		Pid                 int 						`json:"pid"`
	} {direccionesFisicas, data, int(pcb.PID)})
	if err != nil {
		log.Printf("Failed to encode data: %v", err)
	}

	response, err := http.Post(url, "application/json", bytes.NewBuffer(bodyWrite))
	if err != nil {
		log.Printf("Failed to send data: %v", err)
	}

	if response.StatusCode != http.StatusOK {
		log.Printf("Unexpected response status: %s", response.Status)
	}
}


type BodyRequestLeer struct {
	DireccionesTamanios []globals.DireccionTamanio `json:"direcciones_tamanios"`
	Pid				 	int 					   `json:"pid"`
}
type BodyADevolver struct {
	Contenido [][]byte `json:"contenido"`
}
func IO_STDOUT_WRITE(pcb pcb.T_PCB, direccionesFisicas []globals.DireccionTamanio) {
	url := fmt.Sprintf("http://%s:%d/read", globals.ConfigIO.Ip_memory, globals.ConfigIO.Port_memory)

	bodyRead, err := json.Marshal(BodyRequestLeer{
		DireccionesTamanios: direccionesFisicas,
		Pid:                 int(pcb.PID),
	})
	if err != nil {
		return 
	}
/*
	bodyRead, err := json.Marshal(struct {
		DireccionesTamanios 			[]globals.DireccionTamanio `json:"direcciones_tamanios"`
		Pid 							int  					   `json:"pid"`		
	} {direccionesFisicas, int(pcb.PID)})
	if err != nil {
		log.Printf("Failed to encode data: %v", err)
	}*/

	datosLeidos, err := http.Post(url, "application/json", bytes.NewBuffer(bodyRead))
	if err != nil {
		log.Printf("Failed to receive data: %v", err)
	}

	if datosLeidos.StatusCode != http.StatusOK {
		log.Printf("Unexpected response status: %s", datosLeidos.Status)
	}


	var response BodyADevolver

	err = json.NewDecoder(datosLeidos.Body).Decode(&response) 
	if err != nil {
		return
	}
	fmt.Println("DIRECCCIONES FISICAS: ", direccionesFisicas)

	var bytesConcatenados []byte
    for _, sliceBytes := range response.Contenido {
        bytesConcatenados = append(bytesConcatenados, sliceBytes...)
    }

	// Lee los datos de la respuesta
	/* response, err := io.ReadAll(datosLeidos.Body)
	if err != nil {
		log.Printf("Failed to read response body: %v", err)
	} */

	// Convierto los datos a string
	responseString := string(bytesConcatenados)

	// Consumo una unidad de trabajo
	time.Sleep(time.Duration(globals.ConfigIO.Unit_work_time) * time.Millisecond)

	fmt.Print("Datos leidos: *")
	// Escribo los datos en la salida (los muestro por pantalla)
	writer := bufio.NewWriter(os.Stdout)
	writer.WriteString(responseString)
	writer.Flush()
	fmt.Print("*\n")
}