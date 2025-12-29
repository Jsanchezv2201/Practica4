package main

import (
	"testing"
	"time"
)

// TestSimulacion ejecuta las comparativas solicitadas en el enunciado
// Ejecutar con: go test -v taller_test.go taller.go
func TestSimulacion(t *testing.T) {
	// IMPORTANTE: Asegúrate de que servidor.go esté ejecutándose en otra terminal
	// antes de lanzar los tests, para que la simulación reciba cambios de estado.
	
	tests := []struct {
		nombre string
		cfg    ConfigSimulacion
	}{
		{
			nombre: "Test 1: Carga Equilibrada (6 Plazas, 3 Mecánicos)",
			cfg: ConfigSimulacion{
				NumPlazas:    6,
				NumMecanicos: 3,
				CochesA:      10, CochesB: 10, CochesC: 10,
			},
		},
		{
			nombre: "Test 2: Carga Mecánica Alta (4 Plazas, 4 Mecánicos)",
			cfg: ConfigSimulacion{
				NumPlazas:    4,
				NumMecanicos: 4,
				CochesA:      20, CochesB: 5, CochesC: 5,
			},
		},
		{
			nombre: "Test 3: Carga Carrocería Alta (6 Plazas, 3 Mecánicos)",
			cfg: ConfigSimulacion{
				NumPlazas:    6,
				NumMecanicos: 3,
				CochesA:      5, CochesB: 5, CochesC: 20,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.nombre, func(t *testing.T) {
			t.Logf("Iniciando %s...", tt.nombre)
			start := time.Now()
			
			// Llamada a la lógica refactorizada en taller.go
			EjecutarSimulacion(tt.cfg)
			
			duracion := time.Since(start)
			t.Logf("Finalizado %s en %v", tt.nombre, duracion)
			
			// Pausa entre tests para limpiar sockets y dar tiempo al servidor
			time.Sleep(2 * time.Second)
		})
	}
}