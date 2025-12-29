# Practica4La estructura que tienes actualmente (todos los archivos dentro de una carpeta `codigo`) es **correcta** y es la forma más sencilla de entregarlo.

Aunque tu editor de texto se queje diciendo "main redeclared", esto es solo visual. Al ejecutar los comandos de Go especificando los archivos, todo funcionará bien.

### 📂 Estructura de Carpetas Final

Tu directorio debería verse exactamente así:

```text
/Practica4
   └── /codigo
        ├── servidor.go      (El original, sin tocar)
        ├── mutua.go         (El original, sin tocar)
        ├── taller.go        (Tu código modificado con la lógica y monitores)
        └── taller_test.go   (El archivo nuevo con los 3 Tests)

```

---

### 🚀 Cómo ejecutarlo todo (Paso a Paso)

Para que te funcione el 10/10 en la evaluación, debes abrir **3 terminales** distintas ubicadas en esa carpeta `codigo` y seguir este orden estricto:

#### 1️⃣ Terminal 1: El Servidor

Este debe ser el primero. Se quedará esperando conexiones.

```bash
go run servidor.go

```

#### 2️⃣ Terminal 2: Los Tests (Tu Taller)

Aquí es donde se ejecuta tu código. Usamos el comando `go test` e incluimos ambos archivos (`taller.go` y `taller_test.go`) para que puedan "verse" entre sí.

```bash
go test -v taller.go taller_test.go

```

*Se quedará "pausado" esperando que la Mutua le diga qué hacer (porque empieza en Estado 0).*

#### 3️⃣ Terminal 3: La Mutua (El Cliente Controlador)

Este envía las órdenes. Ejecútalo en cuanto lances el test.

```bash
go run mutua.go

```

**⚠️ Importante:** Como la `mutua.go` termina rápido (envía 10 mensajes y se cierra), es posible que tengas que **volver a ejecutar `go run mutua.go**` varias veces para que el **Test 2** y el **Test 3** reciban órdenes y terminen.

---

### 📄 Para entregar (El PDF)

Según el enunciado, debes subir un único PDF llamado `Practica_4_TuNombre_SSOO_dist.pdf`. Asegúrate de incluir en él:

1. **Código Fuente:** Copia y pega el contenido de `taller.go` y `taller_test.go` (o pon un enlace a GitHub si el profesor lo permite).
2. **Diagramas:** Pega las imágenes de los diagramas UML (Clases y Secuencia) que generaste con el código PlantUML que te pasé.
3. **Resultados de los Tests:** Copia la salida de la **Terminal 2** donde se ve `PASS: TestSimulacion...` y los tiempos de ejecución.

¡Con esa estructura y esos pasos tienes la práctica terminada! ¿Necesitas ayuda con algo más antes de cerrar?