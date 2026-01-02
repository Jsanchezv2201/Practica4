package main

import (
	"testing"
	"time"
)

// TestSimulacion ejecuta las 6 comparativas solicitadas (3 escenarios de carga x 2 configuraciones de recursos)
// Ejecutar con: go test -v taller_test.go taller.go
func TestSimulacion(t *testing.T) {
	// IMPORTANTE: Recuerda ejecutar 'go run mutua.go' en otra terminal para enviar estados.
	
	tests := []struct {
		nombre string
		cfg    ConfigSimulacion
	}{
		// --- ESCENARIO 1: CARGA EQUILIBRADA (10A, 10B, 10C) ---
		{
			nombre: "Test 1.1: Equilibrada - Config Base (6 Plazas, 3 Mec)",
			cfg: ConfigSimulacion{
				NumPlazas: 6, NumMecanicos: 3,
				CochesA: 10, CochesB: 10, CochesC: 10,
			},
		},
		{
			nombre: "Test 1.2: Equilibrada - Config Ajustada (4 Plazas, 4 Mec)",
			cfg: ConfigSimulacion{
				NumPlazas: 4, NumMecanicos: 4,
				CochesA: 10, CochesB: 10, CochesC: 10,
			},
		},

		// --- ESCENARIO 2: CARGA MECÁNICA ALTA (20A, 5B, 5C) ---
		{
			nombre: "Test 2.1: Mecánica Alta - Config Base (6 Plazas, 3 Mec)",
			cfg: ConfigSimulacion{
				NumPlazas: 6, NumMecanicos: 3,
				CochesA: 20, CochesB: 5, CochesC: 5,
			},
		},
		{
			nombre: "Test 2.2: Mecánica Alta - Config Ajustada (4 Plazas, 4 Mec)",
			cfg: ConfigSimulacion{
				NumPlazas: 4, NumMecanicos: 4,
				CochesA: 20, CochesB: 5, CochesC: 5,
			},
		},

		// --- ESCENARIO 3: CARGA CARROCERÍA ALTA (5A, 5B, 20C) ---
		{
			nombre: "Test 3.1: Carrocería Alta - Config Base (6 Plazas, 3 Mec)",
			cfg: ConfigSimulacion{
				NumPlazas: 6, NumMecanicos: 3,
				CochesA: 5, CochesB: 5, CochesC: 20,
			},
		},
		{
			nombre: "Test 3.2: Carrocería Alta - Config Ajustada (4 Plazas, 4 Mec)",
			cfg: ConfigSimulacion{
				NumPlazas: 4, NumMecanicos: 4,
				CochesA: 5, CochesB: 5, CochesC: 20,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.nombre, func(t *testing.T) {
			t.Logf(">>> Iniciando %s...", tt.nombre)
			start := time.Now()
			
			EjecutarSimulacion(tt.cfg)
			
			duracion := time.Since(start)
			t.Logf("<<< Finalizado %s en %v", tt.nombre, duracion)
			
			// Pausa técnica para permitir que el servidor procese desconexiones y no saturar sockets
			time.Sleep(2 * time.Second)
		})
	}
}