// logger.go
package pkgrabbit

import "log"


// defaultLogger utiliza el logger por defecto del paquete "log".
type defaultLogger struct{}

func (l *defaultLogger) Printf(format string, v ...any) {
	log.Printf(format, v...)
}
