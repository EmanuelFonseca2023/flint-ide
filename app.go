package main

import (
	"context"
	"flint/compiler"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx         context.Context
	compiler    *compiler.Compiler
	currentFile string
	runner      *compiler.Runner
	tmpDir      string
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	c, err := compiler.Detect()
	if err != nil {
		a.compiler = nil
		return
	}
	a.compiler = c
}

func (a *App) CompilerInfo() compiler.Info {
	if a.compiler == nil {
		return compiler.Info{Found: false}
	}
	return a.compiler.Info()
}

// Compile compila en goroutine y emite eventos al frontend.
// Retorna inmediatamente para no bloquear el botón Detener.
func (a *App) Compile(sourceCode string) {
	// Matar proceso anterior
	if a.runner != nil && a.runner.IsRunning() {
		a.runner.Kill()
	}
	if a.tmpDir != "" {
		os.RemoveAll(a.tmpDir)
		a.tmpDir = ""
	}

	if a.compiler == nil {
		runtime.EventsEmit(a.ctx, "compile:error", []compiler.CodeError{{
			Message: "No se encontró compilador. Instala Xcode Command Line Tools.",
		}})
		return
	}

	// Compilar en goroutine para no bloquear
	go func() {
		runtime.EventsEmit(a.ctx, "compile:start", nil)

		result := a.compiler.Compile(sourceCode)

		if !result.Success {
			runtime.EventsEmit(a.ctx, "compile:error", result.Errors)
			return
		}

		a.tmpDir = result.TmpDir
		runtime.EventsEmit(a.ctx, "compile:ok", nil)

		// Arrancar proceso
		runner, err := compiler.StartProcess(result.BinPath, func(event compiler.IOEvent) {
			runtime.EventsEmit(a.ctx, "process:output", event)
		})

		if err != nil {
			os.RemoveAll(a.tmpDir)
			runtime.EventsEmit(a.ctx, "compile:error", []compiler.CodeError{{
				Message: "Error al ejecutar el programa: " + err.Error(),
			}})
			return
		}

		a.runner = runner
	}()
}

func (a *App) SendInput(text string) error {
	if a.runner == nil || !a.runner.IsRunning() {
		return nil
	}
	return a.runner.SendInput(text)
}

func (a *App) KillProcess() {
	if a.runner != nil {
		a.runner.Kill()
	}
}

func (a *App) IsRunning() bool {
	return a.runner != nil && a.runner.IsRunning()
}

func (a *App) OpenFile() (string, error) {
	path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Abrir archivo",
		Filters: []runtime.FileFilter{
			{DisplayName: "C / C++", Pattern: "*.cpp;*.cc;*.cxx;*.c;*.h;*.hpp"},
		},
	})
	if err != nil || path == "" {
		return "", err
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	a.currentFile = path
	return string(content), nil
}

func (a *App) SaveFile(content string) error {
	if a.currentFile == "" {
		return a.SaveFileAs(content)
	}
	return os.WriteFile(a.currentFile, []byte(content), 0644)
}

func (a *App) SaveFileAs(content string) error {
	path, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "Guardar como",
		DefaultFilename: "main.cpp",
		Filters: []runtime.FileFilter{
			{DisplayName: "C / C++", Pattern: "*.cpp;*.cc;*.cxx;*.c;*.h;*.hpp"},
		},
	})
	if err != nil || path == "" {
		return err
	}
	a.currentFile = path
	return os.WriteFile(path, []byte(content), 0644)
}

func (a *App) CurrentFileName() string {
	if a.currentFile == "" {
		return "main.cpp"
	}
	return filepath.Base(a.currentFile)
}
