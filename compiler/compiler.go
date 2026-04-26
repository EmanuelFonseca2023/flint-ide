package compiler

import (
	"os"
	"path/filepath"
)

type RunResult struct {
	Success bool
	Output  string
	Errors  []CodeError
	BinPath string // ruta al binario compilado, para ejecutar luego
	TmpDir  string // para limpieza posterior
}

type CodeError struct {
	Line    int
	Column  int
	Message string
	Raw     string
}

// Compile compila el código fuente y devuelve la ruta al binario.
// No ejecuta nada — eso lo hace el App con StartProcess.
func (c *Compiler) Compile(sourceCode string) RunResult {
	tmpDir, err := os.MkdirTemp("", "flint-*")
	if err != nil {
		return errorResult("No se pudo crear directorio temporal")
	}

	srcPath := filepath.Join(tmpDir, "main.cpp")
	if err := os.WriteFile(srcPath, []byte(sourceCode), 0644); err != nil {
		os.RemoveAll(tmpDir)
		return errorResult("No se pudo escribir el archivo fuente")
	}

	outPath := filepath.Join(tmpDir, "main")
	if os.PathSeparator == '\\' {
		outPath += ".exe"
	}
	result := compileOnly(c.path, srcPath, outPath)
	result.BinPath = outPath
	result.TmpDir = tmpDir
	return result
}

// CompileAndRun sigue existiendo para compatibilidad futura
// pero ahora lo usamos solo en tests o flujos sin interactividad
func (c *Compiler) CompileAndRun(sourceCode string) RunResult {
	result := c.Compile(sourceCode)
	if !result.Success {
		return result
	}
	defer os.RemoveAll(result.TmpDir)
	return result
}

func errorResult(msg string) RunResult {
	return RunResult{
		Success: false,
		Errors:  []CodeError{{Message: msg}},
	}
}

func isWindowsCompat() bool {
	return os.PathSeparator == '\\'
}
