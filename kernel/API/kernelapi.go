package kernel_api

import (
	"encoding/json"
	"net/http"
)

/* Glossary:

- BRQ: Body Request
- BRS: Body Response

*/

type ProcessStart_BRQ struct {
	Path string `json:"path"`
}

type ProcessStart_BRS struct {
	Pid int `json:"pid"`
}

func ProcessInit(w http.ResponseWriter, r *http.Request) {
	var request ProcessStart_BRQ
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// No sé que hago con el path

	var respBody ProcessStart_BRS = ProcessStart_BRS{Pid: 1} // TODO: Implementar la creación de un proceso, valor hardcodeado, incrementar en 1 por cada proceso creado

	response, err := json.Marshal(respBody)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

// -----------------

func HelloWorld(w http.ResponseWriter, r *http.Request) {

	respuesta, err := json.Marshal("Hola! Como andas?")
	if err != nil {
		http.Error(w, "Error al codificar los datos como JSON", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(respuesta)
}