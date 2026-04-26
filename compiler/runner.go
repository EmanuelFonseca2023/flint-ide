package compiler

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"syscall"
)

type Runner struct {
	cmd      *exec.Cmd
	stdin    io.WriteCloser
	cancelFn context.CancelFunc
	mu       sync.Mutex
	running  bool
}

type IOEvent struct {
	Type string // "stdout" | "stderr" | "exit"
	Data string
}

func StartProcess(binPath string, onOutput func(IOEvent)) (*Runner, error) {
	ctx, cancel := context.WithCancel(context.Background())

	cmd := exec.CommandContext(ctx, binPath)

	// Nuevo process group para poder matar hijos también
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("error creando stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("error creando stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("error creando stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		cancel()
		return nil, fmt.Errorf("error iniciando proceso: %w", err)
	}

	r := &Runner{
		cmd:      cmd,
		stdin:    stdin,
		cancelFn: cancel,
		running:  true,
	}

	// Leer stdout en chunks — no esperar newlines
	// Así "Ingresa un número: " llega aunque no tenga \n
	go func() {
		buf := make([]byte, 256)
		for {
			n, err := stdout.Read(buf)
			if n > 0 {
				onOutput(IOEvent{Type: "stdout", Data: string(buf[:n])})
			}
			if err != nil {
				break
			}
		}
	}()

	// stderr igual — chunks
	go func() {
		buf := make([]byte, 256)
		for {
			n, err := stderr.Read(buf)
			if n > 0 {
				onOutput(IOEvent{Type: "stderr", Data: string(buf[:n])})
			}
			if err != nil {
				break
			}
		}
	}()

	// Esperar que termine
	go func() {
		err := cmd.Wait()
		r.mu.Lock()
		r.running = false
		r.mu.Unlock()

		if err != nil {
			onOutput(IOEvent{Type: "exit", Data: fmt.Sprintf("\n[Proceso terminó: %v]", err)})
		} else {
			onOutput(IOEvent{Type: "exit", Data: "\n[Proceso terminó correctamente]"})
		}
	}()

	return r, nil
}

func (r *Runner) SendInput(text string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.running {
		return fmt.Errorf("el proceso ya no está corriendo")
	}
	_, err := io.WriteString(r.stdin, text+"\n")
	return err
}

func (r *Runner) Kill() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.cmd != nil && r.cmd.Process != nil {
		// Matar el process group completo (incluye procesos hijos)
		pgid, err := syscall.Getpgid(r.cmd.Process.Pid)
		if err == nil {
			syscall.Kill(-pgid, syscall.SIGKILL)
		} else {
			r.cmd.Process.Kill()
		}
	}
	r.cancelFn()
}

func (r *Runner) IsRunning() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.running
}

func compileOnly(compilerPath, srcPath, outPath string) RunResult {
	cmd := exec.Command(compilerPath,
		srcPath,
		"-o", outPath,
		"-std=c++17",
		"-Wall",
		"-Wextra",
		"-g",
	)

	stderrBytes, err := cmd.CombinedOutput()
	stderr := string(stderrBytes)

	if err != nil {
		errors := ParseGCCErrors(stderr, srcPath)
		return RunResult{Success: false, Errors: errors}
	}

	return RunResult{Success: true}
}

func isWindows() bool {
	return os.PathSeparator == '\\'
}
