/*
ESTE ES EL ÚNICO ARCHIVO QUE SE PUEDE MODIFICAR
*/

package main

import (
	"bufio"
	"container/heap"
	"fmt"
	"log"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ==================== CONFIGURACIÓN ====================

// ConfigSimulacion permite configurar los parámetros para cada test
type ConfigSimulacion struct {
	NumPlazas    int
	NumMecanicos int
	CochesA      int // Mecánica
	CochesB      int // Eléctrica
	CochesC      int // Carrocería
}

const (
	NumLimpieza = 2
	NumRevision = 2

	TiempoCategoriaA = 5 * time.Second
	TiempoCategoriaB = 3 * time.Second
	TiempoCategoriaC = 1 * time.Second
)

// ==================== ESTRUCTURAS DE DATOS ====================

type Coche struct {
	ID             int
	TipoIncidencia string
	PrioridadBase  int // 0: Alta, 1: Media, 2: Baja
	TiempoPorFase  time.Duration
	TiempoInicio   time.Time
}

// ItemCola para la Priority Queue
type ItemCola struct {
	Coche          *Coche
	PrioridadActual int // Puede cambiar dinámicamente según el estado del servidor
	Timestamp       time.Time
	Index           int
}

// ColaPrioridad implementa heap.Interface
type ColaPrioridad []*ItemCola

func (cp ColaPrioridad) Len() int { return len(cp) }
func (cp ColaPrioridad) Less(i, j int) bool {
	// Menor valor = Mayor prioridad
	if cp[i].PrioridadActual != cp[j].PrioridadActual {
		return cp[i].PrioridadActual < cp[j].PrioridadActual
	}
	// FIFO para prioridades iguales
	return cp[i].Timestamp.Before(cp[j].Timestamp)
}
func (cp ColaPrioridad) Swap(i, j int) {
	cp[i], cp[j] = cp[j], cp[i]
	cp[i].Index = i
	cp[j].Index = j
}
func (cp *ColaPrioridad) Push(x interface{}) {
	n := len(*cp)
	item := x.(*ItemCola)
	item.Index = n
	*cp = append(*cp, item)
}
func (cp *ColaPrioridad) Pop() interface{} {
	old := *cp
	n := len(old)
	item := old[n-1]
	item.Index = -1
	*cp = old[0 : n-1]
	return item
}

// GestorRecurso maneja un recurso con prioridad dinámica
type GestorRecurso struct {
	Capacidad int
	Ocupados  int
	Cola      *ColaPrioridad
	Mu        sync.Mutex
	Cond      *sync.Cond
	TallerRef *Taller // Referencia para consultar estado al reordenar
}

func NuevoGestorRecurso(capacidad int, taller *Taller) *GestorRecurso {
	g := &GestorRecurso{
		Capacidad: capacidad,
		Ocupados:  0,
		Cola:      &ColaPrioridad{},
		TallerRef: taller,
	}
	g.Cond = sync.NewCond(&g.Mu)
	heap.Init(g.Cola)
	return g
}

// CalcularPrioridad determina la prioridad dinámica
func (g *GestorRecurso) CalcularPrioridad(coche *Coche) int {
	estado := g.TallerRef.GetEstado()
	
	// Si el estado otorga prioridad especial (4, 5, 6)
	// Asignamos -1 para que sea más importante que la prioridad 0 (Alta)
	switch estado {
	case 4: // Prioridad Mecánica (A)
		if coche.TipoIncidencia == "mecánica" { return -1 }
	case 5: // Prioridad Eléctrica (B)
		if coche.TipoIncidencia == "eléctrica" { return -1 }
	case 6: // Prioridad Carrocería (C)
		if coche.TipoIncidencia == "carrocería" { return -1 }
	}
	return coche.PrioridadBase
}

func (g *GestorRecurso) Solicitar(coche *Coche) {
	g.Mu.Lock()
	
	// Calculamos prioridad al entrar
	prio := g.CalcularPrioridad(coche)
	
	item := &ItemCola{
		Coche:           coche,
		PrioridadActual: prio,
		Timestamp:       time.Now(),
	}
	
	heap.Push(g.Cola, item)
	
	// Bucle de espera (Monitor)
	for {
		// Condición: Hay espacio Y soy el primero de la cola
		if g.Ocupados < g.Capacidad && g.Cola.Len() > 0 {
			primero := (*g.Cola)[0]
			if primero.Coche.ID == coche.ID {
				heap.Pop(g.Cola) // Salgo de la cola
				g.Ocupados++     // Ocupo el recurso
				g.Mu.Unlock()
				return
			}
		}
		g.Cond.Wait()
	}
}

func (g *GestorRecurso) Liberar() {
	g.Mu.Lock()
	g.Ocupados--
	g.Cond.Broadcast() // Despertar a todos para que el nuevo primero verifique
	g.Mu.Unlock()
}

// ReordenarCola se llama cuando cambia el estado del servidor
func (g *GestorRecurso) ReordenarCola() {
	g.Mu.Lock()
	defer g.Mu.Unlock()
	
	if g.Cola.Len() == 0 {
		return
	}

	// Recalcular prioridades de todos los que esperan
	for _, item := range *g.Cola {
		item.PrioridadActual = g.CalcularPrioridad(item.Coche)
	}
	
	// Reconstruir el heap con las nuevas prioridades
	heap.Init(g.Cola)
	
	// Avisar a los hilos esperando para que verifiquen si ahora son los primeros
	g.Cond.Broadcast()
}

// ==================== TALLER (CONTROLADOR) ====================

type Taller struct {
	plazas       *GestorRecurso
	mecanicos    *GestorRecurso
	limpieza     *GestorRecurso
	revision     *GestorRecurso
	wg           sync.WaitGroup
	tiempoInicio time.Time
	estado       int
	estadoMu     sync.RWMutex
	condEntrada  *sync.Cond
}

func NuevoTaller(cfg ConfigSimulacion) *Taller {
	t := &Taller{
		tiempoInicio: time.Now(),
		estado:       0,
	}
	// Inicializamos el Cond usando el Mutex de estado
	t.condEntrada = sync.NewCond(t.estadoMu.RLocker()) // <--- NUEVO
	
	t.plazas = NuevoGestorRecurso(cfg.NumPlazas, t)
	t.mecanicos = NuevoGestorRecurso(cfg.NumMecanicos, t)
	t.limpieza = NuevoGestorRecurso(NumLimpieza, t)
	t.revision = NuevoGestorRecurso(NumRevision, t)
	
	return t
}

func (t *Taller) GetEstado() int {
	t.estadoMu.RLock()
	defer t.estadoMu.RUnlock()
	return t.estado
}

func (t *Taller) SetEstado(nuevoEstado int) {
	t.estadoMu.Lock()
	t.estado = nuevoEstado
	t.estadoMu.Unlock()

	// 1. Despertar a los coches en la puerta
	t.condEntrada.Broadcast()

	// 2. Reordenar colas internas
	t.plazas.ReordenarCola()
	t.mecanicos.ReordenarCola()
	t.limpieza.ReordenarCola()
	t.revision.ReordenarCola()
	
	// Obtener descripción para el log
	desc := getDescripcionEstado(nuevoEstado)
	fmt.Printf("\n>>> CAMBIO DE ESTADO: %d [%s] (Colas reordenadas)\n", nuevoEstado, desc)
}

// puedeEntrar valida si un coche nuevo puede ingresar al sistema
func (t *Taller) puedeEntrar(coche *Coche) bool {
	estado := t.GetEstado()
	switch estado {
	case 0, 9: return false // Inactivo / Cerrado
	case 1: return coche.TipoIncidencia == "mecánica"
	case 2: return coche.TipoIncidencia == "eléctrica"
	case 3: return coche.TipoIncidencia == "carrocería"
	default: return true // 4-8 permiten entrada general
	}
}

// validarEntradaInternal verifica la entrada SIN adquirir lock (se usa dentro del loop)
func (t *Taller) validarEntradaInternal(coche *Coche) bool {
	switch t.estado {
	case 0, 9: return false // Inactivo / Cerrado
	case 1: return coche.TipoIncidencia == "mecánica"
	case 2: return coche.TipoIncidencia == "eléctrica"
	case 3: return coche.TipoIncidencia == "carrocería"
	default: return true // 4-8 permiten entrada general
	}
}

func (t *Taller) procesarCoche(coche *Coche) {
	defer t.wg.Done()

	// --- CAMBIO PRINCIPAL AQUÍ ---
	// En lugar de irse si no puede entrar, espera.
	t.estadoMu.RLock()
	for !t.validarEntradaInternal(coche) {
		// Log opcional para depuración si quieres verlos esperando
		// fmt.Printf("Coche %d esperando cambio de estado...\n", coche.ID)
		t.condEntrada.Wait()
	}
	t.estadoMu.RUnlock()
	// -----------------------------

	// Secuencia de pasos (Igual que antes)
	// 1. Plaza
	t.logEvento(coche, 1, "Esperando plaza")
	t.plazas.Solicitar(coche)
	t.logEvento(coche, 1, "Ocupando plaza")
	time.Sleep(coche.TiempoPorFase) // Simula documentación
	t.plazas.Liberar()

	// 2. Mecánico
	t.logEvento(coche, 2, "Esperando mecánico")
	t.mecanicos.Solicitar(coche)
	t.logEvento(coche, 2, "Siendo reparado")
	time.Sleep(coche.TiempoPorFase)
	t.mecanicos.Liberar()

	// 3. Limpieza
	t.logEvento(coche, 3, "Esperando limpieza")
	t.limpieza.Solicitar(coche)
	t.logEvento(coche, 3, "Siendo limpiado")
	time.Sleep(coche.TiempoPorFase)
	t.limpieza.Liberar()

	// 4. Revisión
	t.logEvento(coche, 4, "Esperando revisión")
	t.revision.Solicitar(coche)
	t.logEvento(coche, 4, "Siendo revisado")
	time.Sleep(coche.TiempoPorFase)
	t.revision.Liberar()
	
	// Salida
	t.logEvento(coche, 5, "Terminado")
}

// logEvento cumple estrictamente con el formato solicitado:
// Tiempo (Tiempo Ejecución Programa) Coche (N) Incidencia (Tipo) Fase (Fase Actual) Estado (Estado Fase)
func (t *Taller) logEvento(coche *Coche, fase int, estadoFase string) {
	fmt.Printf("Tiempo %-10v Coche %-4d Incidencia %-10s Fase %-2d Estado %s\n",
		time.Since(t.tiempoInicio).Round(time.Millisecond),
		coche.ID,
		coche.TipoIncidencia,
		fase,
		estadoFase)
}

// ==================== GENERADOR Y MAIN ====================

func generarCoches(cfg ConfigSimulacion) []*Coche {
	var coches []*Coche
	id := 1

	crear := func(n int, tipo string, prioBase int, tBase time.Duration) {
		for i := 0; i < n; i++ {
			// Variación del tiempo (enunciado [cite: 39])
			variacion := 0.5 + rand.Float64() // Factor entre 0.5 y 1.5
			tiempoFinal := time.Duration(float64(tBase.Milliseconds())*variacion) * time.Millisecond
			
			coches = append(coches, &Coche{
				ID:             id,
				TipoIncidencia: tipo,
				PrioridadBase:  prioBase,
				TiempoPorFase:  tiempoFinal,
			})
			id++
		}
	}

	crear(cfg.CochesA, "mecánica", 0, TiempoCategoriaA)   // Prio Alta (0)
	crear(cfg.CochesB, "eléctrica", 1, TiempoCategoriaB)  // Prio Media (1)
	crear(cfg.CochesC, "carrocería", 2, TiempoCategoriaC) // Prio Baja (2)

	rand.Shuffle(len(coches), func(i, j int) {
		coches[i], coches[j] = coches[j], coches[i]
	})
	return coches
}

func manejarRed(taller *Taller) {
	conn, err := net.Dial("tcp", "localhost:8000")
	if err != nil {
		log.Println("No se pudo conectar al servidor (modo sin red o servidor apagado)")
		return
	}
	defer conn.Close()

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		msg := scanner.Text()
		// Intentar convertir mensaje a número de estado
		if val, err := strconv.Atoi(strings.TrimSpace(msg)); err == nil {
			taller.SetEstado(val)
		}
	}
}

// EjecutarSimulacion encapsula la lógica para poder ser testeada
func EjecutarSimulacion(cfg ConfigSimulacion) {
	rand.Seed(time.Now().UnixNano())
	
	fmt.Printf("--- Iniciando Simulación: Plazas %d | Mec %d | A:%d B:%d C:%d ---\n",
		cfg.NumPlazas, cfg.NumMecanicos, cfg.CochesA, cfg.CochesB, cfg.CochesC)

	taller := NuevoTaller(cfg)
	
	// Conexión al servidor en goroutine aparte
	go manejarRed(taller)

	coches := generarCoches(cfg)
	taller.wg.Add(len(coches))

	for _, c := range coches {
		c.TiempoInicio = time.Now()
		go taller.procesarCoche(c)
		// Pequeño retardo aleatorio en llegada
		time.Sleep(time.Duration(rand.Intn(200)) * time.Millisecond)
	}

	taller.wg.Wait()
	fmt.Println("--- Simulación Finalizada ---")
}

func main() {
	// Configuración por defecto (Test 1 del enunciado)
	cfg := ConfigSimulacion{
		NumPlazas:    6,
		NumMecanicos: 3,
		CochesA:      10,
		CochesB:      10,
		CochesC:      10,
	}
	
	// Si hay argumentos, podríamos parsearlos, pero por defecto ejecutamos el caso base
	EjecutarSimulacion(cfg)
}

func getDescripcionEstado(estado int) string {
	switch estado {
	case 0: return "Taller Inactivo"
	case 1: return "Solo Categoría A (Mecánica)"
	case 2: return "Solo Categoría B (Eléctrica)"
	case 3: return "Solo Categoría C (Carrocería)"
	case 4: return "Prioridad Categoría A"
	case 5: return "Prioridad Categoría B"
	case 6: return "Prioridad Categoría C"
	case 7: return "No definido (Mantiene anterior)"
	case 8: return "No definido (Mantiene anterior)"
	case 9: return "Taller Cerrado"
	default: return "Desconocido"
	}
}