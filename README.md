# 🔥 Flint IDE

> El IDE de C/C++ que simplemente funciona.

[![Release](https://img.shields.io/github/v/release/EmanuelFonseca2023/flint-ide?style=flat-square&color=E86C2C)](https://github.com/EmanuelFonseca2023/flint-ide/releases/latest)
[![License](https://img.shields.io/github/license/EmanuelFonseca2023/flint-ide?style=flat-square)](LICENSE)
[![Platform](https://img.shields.io/badge/platform-Windows%20%7C%20macOS%20%7C%20Linux-informational?style=flat-square)](https://github.com/EmanuelFonseca2023/flint-ide/releases)
[![Built with Wails](https://img.shields.io/badge/built%20with-Wails%20v2-blueviolet?style=flat-square)](https://wails.io)

Flint es un IDE ligero y pedagógico para aprender C y C++. Está diseñado para estudiantes universitarios que necesitan escribir, compilar y ejecutar código sin configurar nada — el equivalente moderno y estable de Dev-C++ y Code::Blocks.

---

## ✨ Características

- **Compilar y Ejecutar con un botón** — sin terminales, sin configuración
- **Editor Monaco** — el mismo motor de VS Code, con resaltado de sintaxis C++17
- **`cin` interactivo** — el input del usuario funciona en tiempo real dentro del IDE
- **Errores humanizados** — los mensajes crípticos de GCC se traducen a español claro
- **Highlight de línea con error** — haz clic en el error y vas directo a la línea
- **Detección automática del compilador** — detecta GCC o Clang al arrancar sin configurar nada
- **Paneles redimensionables** — arrastra el separador para ajustar editor, salida y errores
- **Guardar / Abrir** — diálogos nativos del sistema con `Cmd/Ctrl+S` y `Cmd/Ctrl+O`
- **Tema oscuro estable** — no se rompe, no parpadea

---

## 📦 Instalación

### Windows
1. Descarga `Flint.exe` desde [Releases](https://github.com/EmanuelFonseca2023/flint-ide/releases/latest)
2. Ejecuta el instalador — incluye el compilador GCC (WinLibs), no necesitas instalar nada más
3. Abre Flint y empieza a programar

### macOS
1. Descarga `Flint-macOS-arm64.dmg` (Apple Silicon) o `Flint-macOS-amd64.dmg` (Intel)
2. Abre el `.dmg` y arrastra Flint a Aplicaciones
3. Asegúrate de tener Xcode Command Line Tools instalado:
   ```bash
   xcode-select --install
   ```

### Linux
Descarga el formato que prefieras desde [Releases](https://github.com/EmanuelFonseca2023/flint-ide/releases/latest):

| Formato | Comando de instalación |
|---|---|
| `.deb` (Ubuntu/Debian) | `sudo dpkg -i flint_*.deb` |
| `.rpm` (Fedora/RHEL) | `sudo rpm -i flint-*.rpm` |
| `.tar.gz` (portable) | Extraer y ejecutar `./Flint` |

Asegúrate de tener GCC instalado: `sudo apt install g++` o equivalente.

---

## ⌨️ Atajos de teclado

| Atajo | Acción |
|---|---|
| `Cmd/Ctrl + Enter` | Compilar y Ejecutar |
| `Cmd/Ctrl + S` | Guardar |
| `Cmd/Ctrl + Shift + S` | Guardar como |
| `Cmd/Ctrl + O` | Abrir archivo |
| `Enter` (en panel de input) | Enviar input al programa |

---

## 🛠️ Compilar desde código fuente

### Requisitos
- [Go 1.21+](https://go.dev/dl/)
- [Node.js 20+](https://nodejs.org/)
- [Wails v2](https://wails.io/docs/gettingstarted/installation)
- GCC / Clang

```bash
# Instalar Wails
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# Clonar el repo
git clone https://github.com/EmanuelFonseca2023/flint-ide.git
cd flint-ide

# Instalar dependencias del frontend
cd frontend && npm install && cd ..

# Modo desarrollo (hot-reload)
wails dev

# Build de producción
wails build
```

---

## 🏗️ Arquitectura

```
flint/
├── main.go              # Entry point Wails
├── app.go               # Bindings Go ↔ Frontend
├── compiler/
│   ├── detector.go      # Detecta g++/clang al arrancar
│   ├── compiler.go      # Compila el código fuente
│   ├── parser.go        # Humaniza errores de GCC
│   └── runner.go        # Ejecuta el binario con pipes interactivos
└── frontend/
    └── src/
        ├── main.ts      # Lógica del IDE
        ├── style.css    # Estilos
        └── index.html   # Shell del IDE
```

**Stack:** Go + TypeScript · [Wails v2](https://wails.io) · [Monaco Editor](https://microsoft.github.io/monaco-editor/) · GCC/Clang

---

## 🎯 Scope de v1

Flint v1 es intencionalmente simple. Está diseñado para el contenido típico de los primeros semestres de programación:

✅ Un archivo `.cpp` a la vez  
✅ Punteros, arreglos, structs  
✅ POO básica (clases, constructores, herencia simple)  
✅ `cin` / `cout` interactivo  

❌ No tiene debugger visual  
❌ No tiene autocompletado inteligente (LSP)  
❌ No tiene gestor de proyectos multi-archivo  
❌ No tiene terminal expuesto  

---

## 📄 Licencia

MIT © [Emanuel Fonseca](https://github.com/EmanuelFonseca2023)