package mmu

import (
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/sisoputnfrba/tp-golang/cpu/globals"
)

//hago este archivo para que no se rompa nada si hago pull de cpu, pero va para cpu.api!!!
//peticion para RESIZE de memoria (DESDE CPU A MEMORIA)
func Resize(tamaño int) string{
	cliente := &http.Client{}
	url := fmt.Sprintf("http://%s:%d/resize", globals.Configcpu.IP_memory,globals.Configcpu.Port_memory)
	req, err := http.NewRequest("PATCH", url, nil)
	if err != nil {
		return "Error al hacer el request"
	}
	tamañoEnString := strconv.Itoa(tamaño)
	pid := strconv.Itoa(int(globals.CurrentJob.PID))

	q := req.URL.Query()
	q.Add("tamaño", tamañoEnString)
	q.Add("pid", pid)
	req.URL.RawQuery = q.Encode()

	req.Header.Set("Content-Type", "application/json")
	respuesta, err := cliente.Do(req)
	if err != nil {
		return "Error al hacer el request"
	}
	// Verificar el código de estado de la respuesta
	if respuesta.StatusCode != http.StatusOK {
		return "Error en el estado de la respuesta"
	}
	bodyBytes, err := io.ReadAll(respuesta.Body)
	if err != nil {
		return "Error al leer el cuerpo de la respuesta"
	}
	return string(bodyBytes)
}
//TODO: En caso de que la respuesta de la memoria sea Out of Memory, se deberá devolver el contexto de ejecución al Kernel informando de esta situación
