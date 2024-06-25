package cicloInstruccion

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/sisoputnfrba/tp-golang/cpu/globals"

	mmu "github.com/sisoputnfrba/tp-golang/cpu/mmu"

	solicitudesmemoria "github.com/sisoputnfrba/tp-golang/cpu/solicitudesMemoria"
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

	largoInstruccion := len(instruccionDecodificadaConComillas)
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
	url := fmt.Sprintf("http://%s:%d/instrucciones", globals.Configcpu.IP_memory, globals.Configcpu.Port_memory)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal("Error al crear el request")
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

	if instruccionDecodificada[0] == "EXIT" {
		currentPCB.EvictionReason = "EXIT"
		pcb.EvictionFlag = true

		log.Printf("PID: %d - Ejecutando: %s", currentPCB.PID, instruccionDecodificada[0])
	} else {
		log.Printf("PID: %d - Ejecutando: %s - %s", currentPCB.PID, instruccionDecodificada[0], instruccionDecodificada[1:])
	}

	switch instruccionDecodificada[0] {
	case "IO_GEN_SLEEP":
		cond, err := HallarInterfaz(instruccionDecodificada[1], "GENERICA")
		if err != nil {
			log.Print("La interfaz no existe o no acepta operaciones de IO Genéricas")
			currentPCB.EvictionReason = "EXIT"
		} else {
			tiempo_esp, err := strconv.Atoi(instruccionDecodificada[2])
			if err != nil {
				log.Fatal("Error al convertir el tiempo de espera a entero")
			}
			if cond {
				
				var genSleepBody = struct {
					InterfaceName string
					SleepTime     int
					}{
						InterfaceName: instruccionDecodificada[1],
						SleepTime:     tiempo_esp,
					}
					
					SendIOData(genSleepBody, "iodata-gensleep")
					currentPCB.EvictionReason = "BLOCKED_IO_GEN"
			} else {
				currentPCB.EvictionReason = "EXIT"
			}
		}
		pcb.EvictionFlag = true
		currentPCB.PC++

	case "IO_STDIN_READ":
		cond, err := HallarInterfaz(instruccionDecodificada[1], "STDIN")
		// interfazEncontrada
		if err != nil {
			log.Print("La interfaz no existe o no acepta operaciones de IO de lectura")
			currentPCB.EvictionReason = "EXIT"
		} else {
			if cond {
				// Obtener la dirección de memoria desde el registro
				memoryAddress := currentPCB.CPU_reg[instruccionDecodificada[2]]
				tipoActualReg2 := reflect.TypeOf(currentPCB.CPU_reg[instruccionDecodificada[3]]).String()
				memoryAddressInt := int(Convertir[uint32](tipoActualReg2, memoryAddress))
	
				// Obtener la cantidad de datos a leer desde el registro
				dataSize := currentPCB.CPU_reg[instruccionDecodificada[3]]
				tipoActualReg3 := reflect.TypeOf(currentPCB.CPU_reg[instruccionDecodificada[3]]).String()
				dataSizeInt := int(Convertir[uint32](tipoActualReg3, dataSize))
	
				direccionesFisicas := mmu.ObtenerDireccionesFisicas(memoryAddressInt, dataSizeInt, int(currentPCB.PID))
	
				var stdinreadBody = struct {
					DireccionesFisicas	[]globals.DireccionTamanio
					InterfaceName		string
					Tamanio				int	
				}{
					DireccionesFisicas: direccionesFisicas,
					InterfaceName: 		instruccionDecodificada[1],
					Tamanio: 			dataSizeInt,
				}

				SendIOData(stdinreadBody, "iodata-stdin")
				currentPCB.EvictionReason = "BLOCKED_IO_STDIN"
	
			} else {
				currentPCB.EvictionReason = "EXIT"
			}
		}
		pcb.EvictionFlag = true
		currentPCB.PC++

	case "IO_STDOUT_WRITE":
		cond, err := HallarInterfaz(instruccionDecodificada[1], "STDOUT")
		if err != nil {
			log.Print("La interfaz no existe o no acepta operaciones de IO de escritura")
			currentPCB.EvictionReason = "EXIT"
		} else {
			if cond {
				// Obtener la dirección de memoria desde el registro
				memoryAddress := currentPCB.CPU_reg[instruccionDecodificada[2]]
				tipoActualReg2 := reflect.TypeOf(currentPCB.CPU_reg[instruccionDecodificada[3]]).String()
				memoryAddressInt := int(Convertir[uint32](tipoActualReg2, memoryAddress))
	
				// Obtener la cantidad de datos a leer desde el registro
				dataSize := currentPCB.CPU_reg[instruccionDecodificada[3]]
				tipoActualReg3 := reflect.TypeOf(currentPCB.CPU_reg[instruccionDecodificada[3]]).String()
				dataSizeInt := int(Convertir[uint32](tipoActualReg3, dataSize))
	
				// ? Chequear como lo implementaron en mmu
				direccionesFisicas := mmu.ObtenerDireccionesFisicas(memoryAddressInt, dataSizeInt, int(currentPCB.PID))

				var stdoutBody = struct {
					DireccionesFisicas	[]globals.DireccionTamanio
					InterfaceName		string
				}{
					DireccionesFisicas: direccionesFisicas,
					InterfaceName: 		instruccionDecodificada[1],
				}

				SendIOData(stdoutBody, "iodata-stdout")
				currentPCB.EvictionReason = "BLOCKED_IO_STDOUT"

			} else {
				currentPCB.EvictionReason = "EXIT"
			}
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

		if tipoReg == "uint32" {
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

		if tipoReg1 == "uint32" {
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

		if tipoReg1 == "uint32" {
			currentPCB.CPU_reg[instruccionDecodificada[1]] = Convertir[uint32](tipoActualReg1, currentPCB.CPU_reg[instruccionDecodificada[1]]) - Convertir[uint32](tipoActualReg2, valorReg2)

		} else {
			currentPCB.CPU_reg[instruccionDecodificada[1]] = Convertir[uint8](tipoActualReg1, currentPCB.CPU_reg[instruccionDecodificada[1]]) - Convertir[uint8](tipoActualReg2, valorReg2)
		}
		currentPCB.PC++

	case "WAIT":
		currentPCB.RequestedResource = instruccionDecodificada[1]
		fmt.Print("Requested Resource: ", currentPCB.RequestedResource + "\n") // *Lo hace bien 
		currentPCB.EvictionReason = "WAIT"
		currentPCB.PC++
		pcb.EvictionFlag = true

	case "SIGNAL":
		currentPCB.RequestedResource = instruccionDecodificada[1]
		currentPCB.EvictionReason = "SIGNAL"
		currentPCB.PC++
		pcb.EvictionFlag = true


	case "MOV_OUT":
		// MOV_OUT (Registro Dirección, Registro Origen): Mueve el contenido del Registro Origen al Registro Dirección (DL).

		//(Registro Dirección, Registro Datos): Lee el valor del Registro Datos y lo escribe en la dirección física de memoria
		//obtenida a partir de la Dirección Lógica almacenada en el Registro Dirección.

		// Leer el valor y tamaño del registro de datos (2)
		var tamanio2 int
		tipoReg2 := pcb.TipoReg(instruccionDecodificada[2])
		if (tipoReg2 == "uint32") {
			tamanio2 = 4
		} else if (tipoReg2 == "uint8") {
			tamanio2 = 1
		}
		
		// Leer la dirección lógica del registro de dirección (1)
		valorReg1 := currentPCB.CPU_reg[instruccionDecodificada[1]]
		tipoActualReg1 := reflect.TypeOf(valorReg1).String()

		direc_log := Convertir[uint32](tipoActualReg1, valorReg1)
		
		fmt.Println("LA INST DECODIFICADA 1 ES", instruccionDecodificada[1])
		fmt.Println("LA INST DECODIFICADA 2 ES", instruccionDecodificada[2])

		fmt.Println("ACA LLEGO", instruccionDecodificada[2])

		direcsFisicas := mmu.ObtenerDireccionesFisicas(int(direc_log), tamanio2, int(currentPCB.PID))
		fmt.Println("ACA TAMBIEN LLEGO", direcsFisicas)


		valorReg2 := currentPCB.CPU_reg[instruccionDecodificada[2]]
		tipoActualReg2 := reflect.TypeOf(valorReg2).String()
		
		var valor2EnString string

		if tipoReg2 == "uint32" {
			valor2EnString = fmt.Sprint(Convertir[uint32](tipoActualReg2, valorReg2))
		} else {
			valor2EnString = string(Convertir[uint8](tipoActualReg2, valorReg2))
		}

		fmt.Println("EL STRING ES", valor2EnString)

		solicitudesmemoria.SolicitarEscritura(direcsFisicas, valor2EnString, int(currentPCB.PID)) //([direccion fisica y tamanio], valorAEscribir, pid

	
		currentPCB.PC++

		//----------------------------------------------------------------------------

		// MOV_IN (Registro Datos, Registro Dirección): Lee el valor
		// de memoria correspondiente a la Dirección Lógica que se encuentra
		// en el Registro Dirección y lo almacena en el Registro Datos.

		// valorReg2 = Direccion Logica -> Direccion Fisica
		// valorReg1 registro donde tenemos que guardar el valor que esta en la D fisica

	case "MOV_IN":

		var tamanio int

		valorReg2 := currentPCB.CPU_reg[instruccionDecodificada[2]]
		tipoActualReg2 := reflect.TypeOf(valorReg2).String()
		
		direc_log := Convertir[uint32](tipoActualReg2, valorReg2)
		
		fmt.Println("El valor de la direc logica es", int(direc_log))

		// Obtenemos la direcion fisica del reg direccion
		direcsFisicas := mmu.ObtenerDireccionesFisicas(int(direc_log), tamanio, int(currentPCB.PID))

		fmt.Println("Direcciones fisicas: ", direcsFisicas)
		
		//Obtenemos el valor guardado en las direcciones fisicas
		datos := solicitudesmemoria.SolicitarLectura(direcsFisicas, int(currentPCB.PID))
		fmt.Println("Los datos MOSTRAMELLON son: ", datos)
		
		// Almacenamos lo leido en el registro destino
		tipoReg1 := pcb.TipoReg(instruccionDecodificada[1])

		var datosAAlmacenar uint64
		
	
		if tipoReg1 == "uint32" {

			bigInt := big.NewInt(0).SetBytes(datos)
    		datosAAlmacenar = bigInt.Uint64()

			currentPCB.CPU_reg[instruccionDecodificada[1]] = uint32(datosAAlmacenar)

			fmt.Println("LO GUARDE EN 32: ", currentPCB.CPU_reg[instruccionDecodificada[1]])
		} else {
			datosAAlmacenar = uint64(datos[0])

			currentPCB.CPU_reg[instruccionDecodificada[1]] = uint8(datosAAlmacenar)

			fmt.Println("LO GUARDE EN 8: ", currentPCB.CPU_reg[instruccionDecodificada[1]])
		}
		
		currentPCB.PC++
		
		//-----------------------------------------------------------------------------
		//COPY_STRING (Tamaño): Toma del string apuntado por el registro SI y
		//copia la cantidad de bytes indicadas en el parámetro tamaño a la
		//posición de memoria apuntada por el registro DI.

	case "COPY_STRING":
		tamanio := globals.PasarAInt(instruccionDecodificada[1])
		//Buscar la direccion logica del registro SI
		valorRegSI := currentPCB.CPU_reg["SI"]
		tipoActualRegSI := reflect.TypeOf(valorRegSI).String()
		direc_logicaSI := int(Convertir[uint32](tipoActualRegSI, valorRegSI))
		
		direcsFisicasSI := mmu.ObtenerDireccionesFisicas(direc_logicaSI, tamanio, int(currentPCB.PID))

		// Lee lo que hay en esa direccion fisica pero no todo, lees lo que te pasaron x param
		datos := solicitudesmemoria.SolicitarLectura(direcsFisicasSI, int(currentPCB.PID))

		// Busca la direccion logica del registro DI
		valorRegDI := currentPCB.CPU_reg["DI"]
		tipoActualRegDI := reflect.TypeOf(valorRegDI).String()
		direc_logicaDI := int(Convertir[uint32](tipoActualRegDI, valorRegDI))
		
		// Obtiene la direccion Fisica asociada
		direcsFisicasDI := mmu.ObtenerDireccionesFisicas(direc_logicaDI, tamanio, int(currentPCB.PID))
		
		// Carga en esa direccion fisica lo que leiste antes
		solicitudesmemoria.SolicitarEscritura(direcsFisicasDI, string(datos), int(currentPCB.PID)) //([direccion fisica y tamanio], valorAEscribir, pid)

		currentPCB.PC++

	//RESIZE (Tamaño)
	case "RESIZE":
		tamanio := globals.PasarAInt(instruccionDecodificada[1])
		fmt.Println("MIRA EL TAMNIOOO: ", tamanio)

		respuestaResize := solicitudesmemoria.Resize(tamanio)
		fmt.Println("el resize devuelve", respuestaResize)
		if respuestaResize != "\"OK\"" {
			fmt.Println("ME LAS TOMO DE CPU")
			currentPCB.EvictionReason = "OUT_OF_MEMORY"
			pcb.EvictionFlag = true
		}
		fmt.Println("sigoo en cpu")
		currentPCB.PC++
	}
}

type Uint interface{ ~uint8 | ~uint32 }

func Convertir[T Uint](tipo string, parametro interface{}) T {

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
	case "int":
		valor := parametro.(int)
		return T(valor)
	}
	return T(0)
}

type SearchInterface struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

func HallarInterfaz(nombre string, tipo string) (bool, error) {
	interf := SearchInterface{
		Name: nombre,
		Type: tipo,
	}

	log.Println("Interfaz a buscar: ", interf)

	jsonData, err := json.Marshal(interf)
	if err != nil {
		return false, fmt.Errorf("failed to encode interface: %v", err)
	}

	url := fmt.Sprintf("http://%s:%d/io-interface", globals.Configcpu.IP_kernel, globals.Configcpu.Port_kernel)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return false, fmt.Errorf("POST request failed. Failed to send interface: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("unexpected response status: %s", resp.Status)
	}

	var response bool
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return false, fmt.Errorf("failed to decode response: %v", err)
	}

	return response, nil
}

/*
 SendIOData: Comunica la información necesaria a kernel para el uso de cualquier body de interfaz de entrada/salida

 @param datum: Estructura con la información necesaria para la comunicación (La estructura usada va a depender de la interfaz a utilizar)
 @param endpoint: Endpoint al que se va a enviar la información
	- "iodata-gensleep"
	- "iodata-stdin"
	- "iodata-stdout"
 @return error: Error en caso de que la comunicación falle
**/
func SendIOData(datum interface{}, endpoint string) error {
	jsonData, err := json.Marshal(datum)
	if err != nil {
		return fmt.Errorf("failed to encode interface: %v", err)
	}

	url := fmt.Sprintf("http://%s:%d/%s", globals.Configcpu.IP_kernel, globals.Configcpu.Port_kernel, endpoint)
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
