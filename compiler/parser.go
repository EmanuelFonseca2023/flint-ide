package compiler

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
)

// ParseGCCErrors convierte stderr de GCC en errores humanizados.
// Formato GCC: archivo.cpp:línea:col: tipo: mensaje
func ParseGCCErrors(stderr, srcPath string) []CodeError {
	var errors []CodeError
	filename := filepath.Base(srcPath)

	for _, line := range strings.Split(stderr, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parsed, ok := parseGCCLine(line, filename)
		if !ok {
			continue
		}

		parsed.Message = humanize(parsed.Raw)
		errors = append(errors, parsed)
	}

	return errors
}

func parseGCCLine(line, filename string) (CodeError, bool) {
	// Buscar el patrón: main.cpp:10:5: error: mensaje
	// También acepta rutas absolutas
	parts := strings.SplitN(line, ":", 5)
	if len(parts) < 4 {
		return CodeError{}, false
	}

	// parts[0] puede ser la ruta completa, chequeamos que contenga nuestro archivo
	if !strings.Contains(parts[0], filename) && !strings.HasSuffix(parts[0], ".cpp") {
		return CodeError{}, false
	}

	lineNum, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return CodeError{}, false
	}

	col, _ := strconv.Atoi(strings.TrimSpace(parts[2])) // columna es opcional

	// parts[3] = " error" o " warning" o " note"
	kind := strings.TrimSpace(parts[3])
	if kind != "error" && kind != "warning" {
		return CodeError{}, false // ignorar "note" por ahora
	}

	rawMsg := ""
	if len(parts) == 5 {
		rawMsg = strings.TrimSpace(parts[4])
	}

	return CodeError{
		Line:   lineNum,
		Column: col,
		Raw:    rawMsg,
	}, true
}

// humanize convierte mensajes crípticos de GCC a mensajes entendibles.
// Esta tabla crece con el tiempo según los errores más comunes en estudiantes.
func humanize(raw string) string {
	// Tabla de traducciones — orden importa, más específico primero
	translations := []struct {
		contains string
		human    string
	}{
		// Errores de declaración
		{"was not declared in this scope", "Variable o función no declarada. ¿La escribiste correctamente? ¿Está en el scope correcto?"},
		{"undeclared identifier", "Identificador no declarado. ¿Olvidaste declarar esta variable?"},
		{"expected ';'", "Falta un punto y coma (;) antes de esta línea."},
		{"expected '}'", "Falta una llave de cierre (}). Revisa que cada { tenga su }."},
		{"expected '{'", "Falta una llave de apertura ({)."},
		{"expected ')' before", "Falta un paréntesis de cierre. Revisa tus paréntesis."},
		{"expected primary-expression", "Expresión inválida. Revisa la sintaxis en esta línea."},

		// Tipos
		{"cannot convert", "Tipos incompatibles. Estás asignando un tipo a una variable de otro tipo."},
		{"invalid conversion from", "Conversión inválida entre tipos. Usa un cast si es intencional."},
		{"no matching function", "No hay una función que acepte esos argumentos. Revisa los tipos que estás pasando."},
		{"too many arguments", "Estás pasando más argumentos de los que la función acepta."},
		{"too few arguments", "Estás pasando menos argumentos de los que la función necesita."},

		// Punteros
		{"invalid use of incomplete type", "Estás usando un tipo que no está completamente definido."},
		{"dereferencing pointer to incomplete type", "Intentas usar un puntero a un tipo incompleto."},
		{"null pointer dereference", "Dereferenciando un puntero nulo — causaría un crash."},
		{"is not a pointer", "Estás usando * en algo que no es un puntero."},

		// Clases / OOP
		{"has no member named", "Ese atributo o método no existe en la clase. ¿Lo declaraste en el .h o en la clase?"},
		{"is private within this context", "Estás intentando acceder a un miembro privado desde fuera de la clase."},
		{"undefined reference to", "La función existe declarada pero no implementada, o falta enlazar una librería."},
		{"multiple definition of", "Estás definiendo la misma función o variable más de una vez."},
		{"redefinition of", "Estás redefiniendo algo que ya fue definido arriba."},

		// Return
		{"control reaches end of non-void function", "La función debe retornar un valor pero no todos los caminos tienen return."},
		{"return-type is 'int'", "La función main debe retornar int. Agrega return 0; al final."},
	}

	rawLower := strings.ToLower(raw)
	for _, t := range translations {
		if strings.Contains(rawLower, strings.ToLower(t.contains)) {
			return fmt.Sprintf("%s\n    → GCC: %s", t.human, raw)
		}
	}

	// Sin traducción conocida — mostrar el original con contexto
	return fmt.Sprintf("Error de compilación: %s", raw)
}
