package cicloInstruccion

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/sisoputnfrba/tp-golang/cpu/globals"
	"github.com/sisoputnfrba/tp-golang/utils/device"
	"github.com/sisoputnfrba/tp-golang/utils/pcb"
)

/**
 * Delimitador: Función que separa la instrucción en sus partes
 * @param instActual: Instrucción a separar
 * @return instruccionDecodificada: Instrucción separada
**/
func Delimitador(instActual string) []string {
	delimitador := " "
	i := 0

	instruccionDecodificadaConComillas := strings.Split(instActual, delimitador)
	instruccionDecodificada := instruccionDecodificadaConComillas

	largoInstruccion := len (instruccionDecodificadaConComillas) 
	for i < largoInstruccion {
		instruccionDecodificada[i] = strings.Trim(instruccionDecodificadaConComillas[i], `"`)
		i++
	}
	
	return instruccionDecodificada
}

func Fetch(currentPCB *pcb.T_PCB) string {
	// CPU pasa a memoria el PID y el PC, y memoria le devuelve la instrucción
	// (después de identificar en el diccionario la key: PID,
	// va a buscar en la lista de instrucciones de ese proceso, la instrucción en la posición
	// pc y nos va a devolver esa instrucción)
	// GET /instrucciones	
	globals.PCBMutex.Lock()
	pid := currentPCB.PID
	pc := currentPCB.PC
	globals.PCBMutex.Unlock()
	
	cliente := &http.Client{}
	url := fmt.Sprintf("http://%s:%d/instrucciones", globals.Configcpu.IP_memory,globals.Configcpu.Port_memory)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal("Error al hacer el request")
	}
	q := req.URL.Query()
	q.Add("pid", strconv.Itoa(int(pid)))
	q.Add("pc", strconv.Itoa(int(pc)))
	req.URL.RawQuery = q.Encode()

	req.Header.Set("Content-Type", "application/json")
	respuesta, err := cliente.Do(req)
	if err != nil {
		log.Fatal("Error al hacer el request")
	}

	if respuesta.StatusCode != http.StatusOK {
		log.Fatal("Error en el estado de la respuesta")
	}

	instruccion, err := io.ReadAll(respuesta.Body)
	if err != nil {
		log.Fatal("Error al leer el cuerpo de la respuesta")
	}
	
	instruccion1 := string(instruccion)
		
	log.Printf("PID: %d - FETCH - Program Counter: %d", pid, pc)

	return instruccion1
}

func DecodeAndExecute(currentPCB *pcb.T_PCB) {
	instActual := Fetch(currentPCB)
	instruccionDecodificada := Delimitador(instActual)
	
	if (instruccionDecodificada[0] == "EXIT"){
		currentPCB.EvictionReason = "EXIT"
		pcb.EvictionFlag = true

		log.Printf("PID: %d - Ejecutando: %s", currentPCB.PID, instruccionDecodificada[0])
	} else {
		log.Printf("PID: %d - Ejecutando: %s - %s", currentPCB.PID, instruccionDecodificada[0], instruccionDecodificada[1:])
	}

	switch instruccionDecodificada[0] {
		case "IO_GEN_SLEEP":
			tiempo_esp, err := strconv.Atoi(instruccionDecodificada[2])
			if err != nil {
				log.Fatal("Error al convertir el tiempo de espera a entero")
			}
			_, err = HallarInterfaz(instruccionDecodificada[1], "GENERICA")
			if err != nil {
				log.Print("Error al verificar la existencia de la interfaz genérica")
				currentPCB.EvictionReason = "NOT_FOUND_IO"
			} else {
				currentPCB.EvictionReason = "BLOCKED_IO_GEN"
				ComunicarTiempoEspera(instruccionDecodificada[1], tiempo_esp)
			}
			pcb.EvictionFlag = true
			currentPCB.PC++ // Ver si aumenta siempre

		case "IO_STDIN_READ":
			interfazEncontrada, err := HallarInterfaz(instruccionDecodificada[1], "STDIN")
			if err != nil {
				log.Print("Error al verificar la existencia de la interfaz STDIN")
				currentPCB.EvictionReason = "NOT_FOUND_IO"
			} else {		
				// Obtener la dirección de memoria desde el registro
				memoryAddress := currentPCB.CPU_reg[instruccionDecodificada[2]]
		
				// Obtener la cantidad de datos a leer desde el registro
				dataSize := currentPCB.CPU_reg[instruccionDecodificada[3]]
				
				// ? Chequear como lo implementaron en mmu
				direccionesFisicas := ObtenerDireccionFisica(memoryAddress, dataSize, currentPCB.PID)

				// Mandar a leer a la interfaz (a través de kernel)				
				url := fmt.Sprintf("http://%s:%d/io-stdin-read", globals.Configcpu.IP_kernel, globals.Configcpu.Port_kernel)

				bodyStdin, err := json.Marshal(struct {
					direccionesFisicas []T_DireccionFisica
					interfaz device.T_IOInterface
				} {direccionesFisicas, interfazEncontrada})
				if err != nil {
					log.Printf("Failed to encode adresses: %v", err)
				}

				response, err := http.Post(url, "application/json", bytes.NewBuffer(bodyStdin))
				if err != nil {
					log.Printf("Failed to send adresses: %v", err)
				}

				if response.StatusCode != http.StatusOK {
					log.Printf("Unexpected response status: %s", response.Status)
				}

				currentPCB.EvictionReason = "BLOCKED_IO_STD"
			}
			pcb.EvictionFlag = true
			currentPCB.PC++

		case "IO_STDOUT_WRITE":
			interfazEncontrada, err := HallarInterfaz(instruccionDecodificada[1], "STDOUT")
			if err != nil {
				log.Print("Error al verificar la existencia de la interfaz STDOUT")
				currentPCB.EvictionReason = "NOT_FOUND_IO"
			} else {		
				// Obtener la dirección de memoria desde el registro
				memoryAddress := currentPCB.CPU_reg[instruccionDecodificada[2]]
		
				// Obtener la cantidad de datos a leer desde el registro
				dataSize := currentPCB.CPU_reg[instruccionDecodificada[3]]
				
				// ? Chequear como lo implementaron en mmu
				direccionesFisicas := ObtenerDireccionFisica(memoryAddress, dataSize, currentPCB.PID)

				// Mandar a escribir a la interfaz (a través de kernel)				
				url := fmt.Sprintf("http://%s:%d/io-stdout-write", globals.Configcpu.IP_kernel, globals.Configcpu.Port_kernel)

				bodyStdout, err := json.Marshal(struct {
					direccionesFisicas []T_DireccionFisica
					interfaz device.T_IOInterface
				} {direccionesFisicas, interfazEncontrada})
				if err != nil {
					log.Printf("Failed to encode adresses: %v", err)
				}

				response, err := http.Post(url, "application/json", bytes.NewBuffer(bodyStdout))
				if err != nil {
					log.Printf("Failed to send adresses: %v", err)
				}

				if response.StatusCode != http.StatusOK {
					log.Printf("Unexpected response status: %s", response.Status)
				}

				currentPCB.EvictionReason = "BLOCKED_IO_STD"
			}
			pcb.EvictionFlag = true
			currentPCB.PC++

		case "JNZ":
			if currentPCB.CPU_reg[instruccionDecodificada[1]] != 0 {
				currentPCB.PC = ConvertirUint32(instruccionDecodificada[2])
			} else {
				currentPCB.PC++
			}
			
		case "SET":
			tipoReg := pcb.TipoReg(instruccionDecodificada[1])
			valor := instruccionDecodificada[2]
			
			if (tipoReg == "uint32") {
				currentPCB.CPU_reg[instruccionDecodificada[1]] = ConvertirUint32(valor)
			} else {
				currentPCB.CPU_reg[instruccionDecodificada[1]] = ConvertirUint8(valor)
			}
			currentPCB.PC++
			
		case "SUM":
			tipoReg1 := pcb.TipoReg(instruccionDecodificada[1])
			valorReg2 := currentPCB.CPU_reg[instruccionDecodificada[2]]
			
			tipoActualReg1 := reflect.TypeOf(currentPCB.CPU_reg[instruccionDecodificada[1]]).String()
			tipoActualReg2 := reflect.TypeOf(valorReg2).String()

			if (tipoReg1 == "uint32") {
				currentPCB.CPU_reg[instruccionDecodificada[1]] = Convertir[uint32](tipoActualReg1, currentPCB.CPU_reg[instruccionDecodificada[1]]) + Convertir[uint32](tipoActualReg2, valorReg2)

			} else {
				currentPCB.CPU_reg[instruccionDecodificada[1]] = Convertir[uint8](tipoActualReg1, currentPCB.CPU_reg[instruccionDecodificada[1]]) + Convertir[uint8](tipoActualReg2, valorReg2)
			}
			currentPCB.PC++
				
		case "SUB":
			//SUB (Registro Destino, Registro Origen): Resta al Registro Destino 
			//el Registro Origen y deja el resultado en el Registro Destino.
			tipoReg1 := pcb.TipoReg(instruccionDecodificada[1])
			valorReg2 := currentPCB.CPU_reg[instruccionDecodificada[2]]
			
			tipoActualReg1 := reflect.TypeOf(currentPCB.CPU_reg[instruccionDecodificada[1]]).String()
			tipoActualReg2 := reflect.TypeOf(valorReg2).String()

			if (tipoReg1 == "uint32") {
				currentPCB.CPU_reg[instruccionDecodificada[1]] = Convertir[uint32](tipoActualReg1, currentPCB.CPU_reg[instruccionDecodificada[1]]) - Convertir[uint32](tipoActualReg2, valorReg2)

			} else {
				currentPCB.CPU_reg[instruccionDecodificada[1]] = Convertir[uint8](tipoActualReg1, currentPCB.CPU_reg[instruccionDecodificada[1]]) - Convertir[uint8](tipoActualReg2, valorReg2)
			}
			currentPCB.PC++
	}

}

type Uint interface {~uint8 | ~uint32}
func Convertir[T Uint](tipo string, parametro interface {}) T {

	if parametro == "" {
		log.Fatal("La cadena de texto está vacía")
	}

	switch tipo {
		case "uint8":
			valor := parametro.(uint8)
			return T(valor)
		case "uint32":
			valor := parametro.(uint32)
			return T(valor)
		case "float64":
			valor := parametro.(float64)
			return T(valor)
	}
	return T(0)
}

type SearchInterface struct {
	Name string `json:"name"`
	Type string `json:"type"`
}
func HallarInterfaz(nombre string, tipo string) (device.T_IOInterface, error) {
	interf := SearchInterface{
		Name: nombre, 
		Type: tipo,
	}
	
	log.Println( "Interfaz a buscar: ", interf)

	jsonData, err := json.Marshal(interf)
	if err != nil {
		return device.T_IOInterface{}, fmt.Errorf("failed to encode interface: %v", err)
	}
	
	url := fmt.Sprintf("http://%s:%d/io-interface", globals.Configcpu.IP_kernel, globals.Configcpu.Port_kernel)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return device.T_IOInterface{}, fmt.Errorf("POST request failed. Failed to send interface: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return device.T_IOInterface{}, fmt.Errorf("unexpected response status: %s", resp.Status)
	}

	var response device.T_IOInterface
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return device.T_IOInterface{}, fmt.Errorf("failed to decode response: %v", err)
	}

	return response, nil
}

type Interfac_Time struct {
	Name 	string 	`json:"name"`
	WTime 	int 	`json:"wtime"`
}
func ComunicarTiempoEspera(nombre string, tiempo_esp int) error {
	int_time := Interfac_Time{Name: nombre, WTime: tiempo_esp}

	jsonData, err := json.Marshal(int_time)
	if err != nil {
		return fmt.Errorf("failed to encode interface: %v", err)
	}
	
	url := fmt.Sprintf("http://%s:%d/tiempo-bloq", globals.Configcpu.IP_kernel, globals.Configcpu.Port_kernel)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("POST request failed. Failed to send interface: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected response status: %s", resp.Status)
	}

	return nil
}

func ConvertirUint8(parametro string) uint8 {
	parametroConvertido, err := strconv.Atoi(parametro)
	if err != nil {
		log.Fatal("Error al convertir el parametro a uint8")
	}
	return uint8(parametroConvertido)
}

func ConvertirUint32(parametro string) uint32 {
	parametroConvertido, err := strconv.Atoi(parametro)
	if err != nil {
		log.Fatal("Error al convertir el parametro a uint32")
	}
	return uint32(parametroConvertido)
}
