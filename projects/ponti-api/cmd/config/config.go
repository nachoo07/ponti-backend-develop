package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	envs "github.com/alphacodinggroup/ponti-backend/pkg/config/godotenv"
)

type AppConfig struct {
	AppName     string
	Version     string
	Environment string
	APIVersion  string
	MaxRetries  int
}

type Config struct {
	App AppConfig
}

type configLoader struct {
	config *Config
}

func NewConfigLoader() (Loader, error) {
	// Ruta al archivo .env
	envPath := "/projects/ponti-api/.env"

	// Cargar el archivo .env
	if err := envs.LoadConfig(envPath); err != nil {
		return nil, fmt.Errorf("error loading configuration from %s: %w", envPath, err)
	}

	// Parsear variables de entorno para AppConfig
	appConfig := AppConfig{
		AppName:     getEnv("APP_NAME", "ponti-api"),
		Version:     getEnv("APP_VERSION", "1.0"),
		Environment: getEnv("APP_ENV", "dev"),
		APIVersion:  getEnv("API_VERSION", "v1"),
		MaxRetries:  getEnvInt("APP_MAX_RETRIES", 5),
	}

	// Agrupar todas las configuraciones
	cfg := &Config{
		App: appConfig,
	}

	// Validar configuraciones
	if err := validateConfig(cfg); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &configLoader{config: cfg}, nil
}

// getEnv obtiene una variable de entorno o retorna un valor por defecto si no está establecida.
func getEnv(key, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}

// getEnvInt obtiene una variable de entorno, la convierte a int o retorna un valor por defecto si no está establecida o falla la conversión.
func getEnvInt(key string, defaultVal int) int {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultVal
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		fmt.Printf("Warning: could not convert %s to int. Using default value %d.\n", key, defaultVal)
		return defaultVal
	}
	return value
}

// getEnvDuration obtiene una variable de entorno, la convierte a time.Duration (en minutos) o retorna un valor por defecto si no está establecida o falla la conversión.
func getEnvDuration(key string, defaultMinutes int) time.Duration {
	minutes := getEnvInt(key, defaultMinutes)
	return time.Duration(minutes) * time.Minute
}

// validateConfig valida que las configuraciones críticas estén presentes y sean válidas.
func validateConfig(cfg *Config) error {
	// Validaciones para AppConfig
	if cfg.App.AppName == "" {
		return fmt.Errorf("APP_NAME is required")
	}
	if cfg.App.Version == "" {
		return fmt.Errorf("APP_VERSION is required")
	}
	if cfg.App.Environment == "" {
		return fmt.Errorf("APP_ENV is required")
	}
	if cfg.App.APIVersion == "" {
		return fmt.Errorf("API_VERSION is required")
	}

	// Añade más validaciones según sea necesario
	return nil
}

// Métodos de la interfaz Loader para obtener configuraciones.

// GetAppConfig retorna la configuración de la aplicación.
func (cl *configLoader) GetAppConfig() AppConfig {
	return cl.config.App
}
