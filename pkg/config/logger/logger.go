package pkglogger

import (
	"log"

	"go-micro.dev/v4/logger"
)

const (
	red    = "\033[31m"
	yellow = "\033[33m"
	green  = "\033[32m"
	blue   = "\033[34m"
	reset  = "\033[0m"
)

// applyColor aplica un color a un formato de cadena.
func applyColor(color, format string) string {
	return color + format + reset
}

// Funciones de log estándar usando el paquete log por defecto
func Info(format string, v ...any) {
	log.Printf(applyColor(blue, format), v...)
}

func Warn(format string, v ...any) {
	log.Printf(applyColor(yellow, format), v...)
}

func Error(format string, v ...any) {
	log.Printf(applyColor(red, format), v...)
}

// Funciones de log Go-Micro usando el logger de go-micro
func GmInfo(format string, v ...any) {
	logger.Infof(applyColor(blue, format), v...)
}

func GmWarn(format string, v ...any) {
	logger.Warnf(applyColor(yellow, format), v...)
}

func GmError(format string, v ...any) {
	logger.Errorf(applyColor(red, format), v...)
}
