package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/kelseyhightower/envconfig"

	envvars "github.com/alphacodinggroup/ponti-backend/pkg/config/godotenv"
)

// Config agrupa todas las configuraciones de la aplicación.
type Config struct {
	App            App            // General variables
	API            API            // API configuration
	HTTPServer     HTTPServer     // HTTP server configuration
	Debugger       Debugger       // debugger configuration
	DB             DB             // database configuration
	WordsSuggester WordsSuggester // suggester configuration
	Migrations     Migrations     // migrations configuration
	Deploy         Deploy         // deployment configuration
}

// LoadConfig carga la configuración desde variables de entorno y archivos .env.
func LoadConfig() (*Config, error) {

	if os.Getenv("GO_ENVIRONMENT") == "" {
		// 1) SIEMPRE cargar primero el base para poblar las variables
		if err := envvars.OverloadConfig("projects/ponti-api/.env"); err != nil {
			return nil, fmt.Errorf("no se pudo cargar el archivo .env base: %w", err)
		}

		// 2) Ahora que el entorno ya tiene DEPLOY_ENV (si estaba en .env), lo leemos
		platform := strings.ToLower(os.Getenv("DEPLOY_PLATFORM"))
		env := strings.ToLower(os.Getenv("DEPLOY_ENV"))
		// No forzar a minúsculas el path del proyecto (puede romper el path en Linux)
		root := os.Getenv("DEPLOY_PROJECT_ROOT")
		if env != "" {
			envFile := fmt.Sprintf(root+"/.env.%s.%s", platform, env)
			if _, err := os.Stat(envFile); err == nil {
				// El archivo existe, cargar y sobrescribir variables
				if err := envvars.OverloadConfig(envFile); err != nil {
					return nil, fmt.Errorf("error al sobrescribir configuración desde el archivo .env %v: %w", envFile, err)
				}
			} else if os.IsNotExist(err) {
				// El archivo no existe, solo advertir (sin error)
				fmt.Printf("Advertencia: el archivo %s no existe, se omite override\n", envFile)
			} else {
				// Otro error al intentar acceder al archivo
				return nil, fmt.Errorf("error al verificar la existencia de %v: %w", envFile, err)
			}
		}
	}

	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, fmt.Errorf("no se pudo procesar la configuración: %w", err)
	}

	// 4) Valores derivados
	cfg.API.BaseURL = fmt.Sprintf("api/%s", cfg.API.APIVersion())

	// 5) Validación final
	validate := validator.New()
	if err := validate.Struct(&cfg); err != nil {
		return nil, fmt.Errorf("configuración inválida: %w", err)
	}

	return &cfg, nil
}
