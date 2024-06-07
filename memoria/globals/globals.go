package globals

import "sync"

// Global variables:
var InstruccionesProceso = make(map[int][]string)

// Global semaphores
var (
	// * Mutex
		InstructionsMutex 			sync.Mutex
	// * Binarios
		// Binary  					= make (chan bool, 1)
	// * Contadores
		// Contador 				= make (chan int, 10)
)

type T_ConfigMemory struct {
	Port              int    `json:"port"`
	Memory_size       int    `json:"memory_size"`
	Page_size         int    `json:"page_size"`
	Instructions_path string `json:"instructions_path"`
	Delay_response    int    `json:"delay_response"`
}

type Frame *int

var Frames *int //chequear si se vuela, por ahora lo dejamos

//Tabla de páginas (donde a cada página(indice) le corresponde un frame)
var TablaPaginas []Frame 

//Diccionario para idenficiar a que proceso pertenece cada TablaPaginas
var  Tablas_de_paginas map[int][]Frame //ver nombre

var Configmemory *T_ConfigMemory

// Inicializo la memoria
var User_Memory = make([]byte,Configmemory.Memory_size) // de 0 a 15 corresponde a una página, marco compuesto por 16 bytes (posiciones)




