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
	"unsafe"

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

type BodyRequestLeer struct {
	DireccionesTamanios []globals.DireccionTamanio `json:"direcciones_tamanios"`
}

type BodyRequestEscribir struct {
	DireccionesTamanios []globals.DireccionTamanio `json:"direcciones_tamanios"`
	Valor_a_escribir    string                     `json:"valor_a_escribir"`
	Pid                 int                        `json:"pid"`
}

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

	if instruccionDecodificada[0] == "EXIT" {
		currentPCB.EvictionReason = "EXIT"
		pcb.EvictionFlag = true

		log.Printf("PID: %d - Ejecutando: %s", currentPCB.PID, instruccionDecodificada[0])
	} else {
		log.Printf("PID: %d - Ejecutando: %s - %s", currentPCB.PID, instruccionDecodificada[0], instruccionDecodificada[1:])
	}

	switch instruccionDecodificada[0] {
	/*case "IO_GEN_SLEEP":
	cond, err := HallarInterfaz(instruccionDecodificada[1], "GENERICA")
	if err != nil {
		log.Print("La interfaz no existe o no acepta operaciones de IO Genéricas")
	}
	tiempo_esp, err := strconv.Atoi(instruccionDecodificada[2])
	if err != nil {
		log.Fatal("Error al convertir el tiempo de espera a entero")
	}
	if cond {
		currentPCB.EvictionReason = "BLOCKED_IO_GEN"
		ComunicarTiempoEspera(instruccionDecodificada[1], tiempo_esp)
	} else {
		currentPCB.EvictionReason = "EXIT"
	}


	/* tiempo_esp, err := strconv.Atoi(instruccionDecodificada[2])
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
	currentPCB.PC++ // Ver si aumenta siempre */

	case "IO_STDIN_READ":
		interfazEncontrada, err := HallarInterfaz(instruccionDecodificada[1], "STDIN")
		if err != nil {
			log.Print("Error al verificar la existencia de la interfaz STDIN")
			currentPCB.EvictionReason = "NOT_FOUND_IO"
		} else {
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

			// Mandar a leer a la interfaz (a través de kernel)
			url := fmt.Sprintf("http://%s:%d/io-stdin-read", globals.Configcpu.IP_kernel, globals.Configcpu.Port_kernel)

			bodyStdin, err := json.Marshal(struct {
				direccionesFisicas []globals.DireccionTamanio
				interfaz           globals.InterfaceController
				tamanio            int
			}{direccionesFisicas, interfazEncontrada, dataSizeInt})
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
			tipoActualReg2 := reflect.TypeOf(currentPCB.CPU_reg[instruccionDecodificada[3]]).String()
			memoryAddressInt := int(Convertir[uint32](tipoActualReg2, memoryAddress))

			// Obtener la cantidad de datos a leer desde el registro
			dataSize := currentPCB.CPU_reg[instruccionDecodificada[3]]
			tipoActualReg3 := reflect.TypeOf(currentPCB.CPU_reg[instruccionDecodificada[3]]).String()
			dataSizeInt := int(Convertir[uint32](tipoActualReg3, dataSize))

			// ? Chequear como lo implementaron en mmu
			direccionesFisicas := mmu.ObtenerDireccionesFisicas(memoryAddressInt, dataSizeInt, int(currentPCB.PID))

			// Mandar a escribir a la interfaz (a través de kernel)
			url := fmt.Sprintf("http://%s:%d/io-stdout-write", globals.Configcpu.IP_kernel, globals.Configcpu.Port_kernel)

			bodyStdout, err := json.Marshal(struct {
				direccionesFisicas []globals.DireccionTamanio
				interfaz           globals.InterfaceController
			}{direccionesFisicas, interfazEncontrada})
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
		currentPCB.EvictionReason = "WAIT"

	case "SIGNAL":
		currentPCB.RequestedResource = instruccionDecodificada[1]
		currentPCB.EvictionReason = "SIGNAL"

	case "MOV_OUT":
		// MOV_OUT (Registro Dirección, Registro Origen): Mueve el contenido del Registro Origen al Registro Dirección (DL).

		//(Registro Dirección, Registro Datos): Lee el valor del Registro Datos y lo escribe en la dirección física de memoria
		//obtenida a partir de la Dirección Lógica almacenada en el Registro Dirección.
		tamanio := int(unsafe.Sizeof(currentPCB.CPU_reg[instruccionDecodificada[2]])) //ver de usar el switch que tenemos en globals
		direc_log := Convertir[uint32]("uint32", currentPCB.CPU_reg[instruccionDecodificada[1]])

		fmt.Println("LA INST DECODIFICADA 1 ES", instruccionDecodificada[1])
		fmt.Println("LA INST DECODIFICADA 2 ES", instruccionDecodificada[2])

		/*direc_logica, ok := currentPCB.CPU_reg[instruccionDecodificada[1]].(int)
		if !ok {
			fmt.Printf("El tipo de valor es %T\n", direc_logica)
			log.Fatalf("Error: el valor en el registro no es de tipo int")
		}
*/
		direcsFisicas := mmu.ObtenerDireccionesFisicas(int(direc_log), tamanio, int(currentPCB.PID))

		//valor := int(Convertir[uint32]("uint32", currentPCB.CPU_reg[instruccionDecodificada[2]]))

		valor, ok := currentPCB.CPU_reg[instruccionDecodificada[2]].(string)
		if !ok {
			fmt.Printf("El tipo de valor es %T\n", valor)
			log.Fatalf("Error: el valor en el registro no es de tipo string")
		}

		solicitudesmemoria.SolicitarEscritura(direcsFisicas, valor, int(currentPCB.PID)) //([direccion fisica y tamanio], valorAEscribir, pid

		currentPCB.PC++

		//----------------------------------------------------------------------------

		// MOV_IN (Registro Datos, Registro Dirección): Lee el valor
		// de memoria correspondiente a la Dirección Lógica que se encuentra
		// en el Registro Dirección y lo almacena en el Registro Datos.

	case "MOV_IN":
		// MOV_IN (Registro Destino, Registro Dirección): Mueve el contenido del Registro Dirección (DL) al Registro Destino.

		tamanio := int(unsafe.Sizeof(currentPCB.CPU_reg[instruccionDecodificada[1]]))
		direc_log := Convertir[uint32]("uint32", currentPCB.CPU_reg[instruccionDecodificada[2]])
		/*direc_logica, ok := currentPCB.CPU_reg[instruccionDecodificada[2]].(int)
		if !ok {
			log.Fatalf("Error: el valor en el registro no es de tipo string")
		}*/

		fmt.Println("El valor de la direc logica es", int(direc_log))

		direcsFisicas := mmu.ObtenerDireccionesFisicas(int(direc_log), tamanio, int(currentPCB.PID))

		datos := solicitudesmemoria.SolicitarLectura(direcsFisicas)
		currentPCB.CPU_reg[instruccionDecodificada[1]] = datos

		currentPCB.PC++
		//-----------------------------------------------------------------------------
		//COPY_STRING (Tamaño): Toma del string apuntado por el registro SI y
		//copia la cantidad de bytes indicadas en el parámetro tamaño a la
		//posición de memoria apuntada por el registro DI.

	case "COPY_STRING":
		// COPY_STRING (Longitud): Copia la cantidad de bytes indicadas por la Longitud desde el Registro SI (que apunta a un string) al Registro Destino DI (que apunta a una posicion de memoria).
		tamanio := globals.PasarAInt(instruccionDecodificada[1])

		direc_logicaSI, ok := currentPCB.CPU_reg["SI"].(int)
		if !ok {
			log.Fatalf("Error: el valor en el registro no es de tipo string")
		}

		direcsFisicasSI := mmu.ObtenerDireccionesFisicas(direc_logicaSI, tamanio, int(currentPCB.PID))
		datos := solicitudesmemoria.SolicitarLectura(direcsFisicasSI)

		direc_logicaDI, ok := currentPCB.CPU_reg["DI"].(int)
		if !ok {
			log.Fatalf("Error: el valor en el registro no es de tipo string")
		}

		direcsFisicasDI := mmu.ObtenerDireccionesFisicas(direc_logicaDI, tamanio, int(currentPCB.PID))

		solicitudesmemoria.SolicitarEscritura(direcsFisicasDI, datos, int(currentPCB.PID)) //([direccion fisica y tamanio], valorAEscribir, pid)

		currentPCB.PC++

	//RESIZE (Tamaño)
	case "RESIZE":

		tamanio := globals.PasarAInt(instruccionDecodificada[1])
		fmt.Println("MIRA EL TAMNIOOO: ", tamanio)

		fmt.Println("el resize devuelve", solicitudesmemoria.Resize(tamanio))
		if solicitudesmemoria.Resize(tamanio) != "\"OK\"" {
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

func HallarInterfaz(nombre string, tipo string) (globals.InterfaceController, error) {
	interf := SearchInterface{
		Name: nombre,
		Type: tipo,
	}

	log.Println("Interfaz a buscar: ", interf)

	jsonData, err := json.Marshal(interf)
	if err != nil {
		return globals.InterfaceController{}, fmt.Errorf("failed to encode interface: %v", err)
	}

	url := fmt.Sprintf("http://%s:%d/io-interface", globals.Configcpu.IP_kernel, globals.Configcpu.Port_kernel)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return globals.InterfaceController{}, fmt.Errorf("POST request failed. Failed to send interface: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return globals.InterfaceController{}, fmt.Errorf("unexpected response status: %s", resp.Status)
	}

	var response globals.InterfaceController
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return globals.InterfaceController{}, fmt.Errorf("failed to decode response: %v", err)
	}

	return response, nil
}

type Interfac_Time struct {
	Name  string `json:"name"`
	WTime int    `json:"wtime"`
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
