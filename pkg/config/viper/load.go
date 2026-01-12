package pkgviper

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"

	pkgutils "github.com/alphacodinggroup/ponti-backend/pkg/utils"
)

// Load carga múltiples archivos de configuración usando Viper.
// Combina todos los archivos de configuración encontrados en la configuración de Viper.
// Retorna un error si no se proporcionan archivos, no se encuentran archivos, o fallan todos los intentos de carga.
func LoadConfig(filePaths ...string) error {
	if len(filePaths) == 0 {
		return errors.New("no file paths provided")
	}

	// Encontrar y filtrar archivos existentes usando FilesFinder
	foundFiles, err := pkgutils.FilesFinder(filePaths...)
	if err != nil {
		return fmt.Errorf("fatal error: failed to find configuration files: %w", err)
	}

	if len(foundFiles) == 0 {
		return errors.New("no configuration files found to load")
	}

	// Configurar Viper para leer variables de entorno
	configureViper()

	var successfullyLoaded bool
	var loadErrors []string

	for _, configFilePath := range foundFiles {
		if err := loadViperConfig(configFilePath); err != nil {
			loadErrors = append(loadErrors, fmt.Sprintf("Failed to load '%s': %v", configFilePath, err))
			continue
		}
		successfullyLoaded = true
		fmt.Printf("Successfully loaded configuration file: %s\n", configFilePath)
	}

	// Si ningún archivo se cargó exitosamente, retornar error
	if !successfullyLoaded {
		return fmt.Errorf("failed to load any configuration files: %v", loadErrors)
	}

	// Si algunos archivos fallaron al cargar, imprimir los errores
	if len(loadErrors) > 0 {
		fmt.Printf("Some configuration files failed to load:\n%s\n", strings.Join(loadErrors, "\n"))
	}

	return nil
}

// configureViper configura Viper para cargar variables de entorno
func configureViper() {
	viper.SetEnvPrefix("")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
}

// loadViperConfig carga y combina un archivo de configuración en Viper
func loadViperConfig(configFilePath string) error {
	fileNameWithoutExt, fileExtension, err := pkgutils.FileNameAndExtension(configFilePath)
	if err != nil {
		return fmt.Errorf("invalid file '%s': %w", configFilePath, err)
	}

	viper.SetConfigName(fileNameWithoutExt)
	viper.SetConfigType(fileExtension)

	dir := filepath.Dir(configFilePath)
	viper.AddConfigPath(dir)

	// Usar MergeInConfig para combinar múltiples configuraciones
	if err := viper.MergeInConfig(); err != nil {
		return fmt.Errorf("error reading config: %w", err)
	}

	return nil
}
