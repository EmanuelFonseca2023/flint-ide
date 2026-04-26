package compiler

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

type Compiler struct {
	path    string
	version string
}

type Info struct {
	Found   bool
	Path    string
	Version string
}

// Detect busca el compilador según el OS.
// En macOS: busca g++ o clang++
// En Linux: busca g++
// En Windows: busca en la carpeta bundleada primero, luego PATH
func Detect() (*Compiler, error) {
	candidates := detectCandidates()

	for _, candidate := range candidates {
		path, err := exec.LookPath(candidate)
		if err != nil {
			continue
		}

		version, err := getVersion(path)
		if err != nil {
			continue
		}

		return &Compiler{path: path, version: version}, nil
	}

	return nil, fmt.Errorf("no se encontró compilador de C++")
}

func detectCandidates() []string {
	switch runtime.GOOS {
	case "windows":
		// Primero buscar en la carpeta bundleada (relativa al exe)
		// Luego en PATH como fallback
		return []string{
			`.\compiler\bin\g++.exe`, // WinLibs bundleado
			"g++",
		}
	case "darwin":
		return []string{"g++", "clang++"}
	default: // linux
		return []string{"g++", "c++"}
	}
}

func getVersion(path string) (string, error) {
	out, err := exec.Command(path, "--version").Output()
	if err != nil {
		return "", err
	}
	// Primera línea del output de --version
	lines := strings.Split(string(out), "\n")
	if len(lines) == 0 {
		return "", fmt.Errorf("output vacío")
	}
	return strings.TrimSpace(lines[0]), nil
}

func (c *Compiler) Info() Info {
	return Info{
		Found:   true,
		Path:    c.path,
		Version: c.version,
	}
}
