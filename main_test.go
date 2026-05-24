package main

import (
	"testing" // herramienta de Go para hacer tests
	"time"
)

// TestFormatearFecha prueba que la fecha se formatea correctamente
func TestFormatearFecha(t *testing.T) {
	// Creo una fecha de prueba: 1 de Junio 2023
	fecha, _ := time.Parse("2006-01-02", "2023-06-01")

	// Llamo a mi función
	resultado := formatearFecha(fecha)

	// Verifico que el resultado sea correcto
	esperado := "JUN 1"
	if resultado != esperado {
		t.Errorf("Se esperaba %s pero se obtuvo %s", esperado, resultado)
	}
}

// TestFormatearMes prueba que el mes se formatea correctamente
func TestFormatearMes(t *testing.T) {
	// Creo una fecha de prueba: 1 de Junio 2023
	fecha, _ := time.Parse("2006-01-02", "2023-06-01")

	// Llamo a mi función
	resultado := formatearMes(fecha)

	// Verifico que el resultado sea correcto
	esperado := "JUN 2023"
	if resultado != esperado {
		t.Errorf("Se esperaba %s pero se obtuvo %s", esperado, resultado)
	}
}