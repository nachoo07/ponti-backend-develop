package pkcresty

import "fmt"

// Logger es una interfaz simple de logging.
type Logger interface {
	Info(args ...any)
	Error(args ...any)
}

// SimpleLogger es una implementación básica que utiliza fmt.Println.
type SimpleLogger struct{}

// Info imprime mensajes informativos.
func (l *SimpleLogger) Info(args ...any) {
	fmt.Println("[INFO]", fmt.Sprint(args...))
}

// Error imprime mensajes de error.
func (l *SimpleLogger) Error(args ...any) {
	fmt.Println("[ERROR]", fmt.Sprint(args...))
}
