package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

// Direccion representa la dirección de un medidor
type Direccion struct {
	MeterId string `json:"meter_id"`
	Address string `json:"address"`
}

// iniciarServidorDirecciones inicia el servidor de direcciones
func iniciarServidorDirecciones() {
	enrutador := mux.NewRouter()
	enrutador.HandleFunc("/address", obtenerDireccion).Methods("GET")
	fmt.Println("Servidor de direcciones iniciado en http://localhost:8082")
	http.ListenAndServe(":8082", enrutador)
}

// obtenerDireccion responde con la dirección de un medidor
func obtenerDireccion(w http.ResponseWriter, r *http.Request) {
	meterId := r.URL.Query().Get("meter_id")

	direcciones := map[string]string{
		"1": "Calle 13 # 41-50",
		"2": "Carrera 64 # 2-02",
		"3": "Carrera 14 # 5-05",
	}

	address, existe := direcciones[meterId]
	if !existe {
		address = "Dirección no encontrada"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Direccion{
		MeterId: meterId,
		Address: address,
	})
}
