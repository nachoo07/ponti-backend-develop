package pkgdotenv

import (
	"errors"
	"fmt"

	pkgutils "github.com/alphacodinggroup/ponti-backend/pkg/utils"
	"github.com/joho/godotenv"
)

// core loader (privada)
func loadEnvFiles(filePaths []string, overload bool) error {
	if len(filePaths) == 0 {
		return errors.New("no environment file paths provided")
	}

	foundFiles, err := pkgutils.FilesFinder(filePaths...)
	if err != nil {
		return fmt.Errorf("fatal error: failed to find configuration files: %w", err)
	}

	if len(foundFiles) == 0 {
		return errors.New("no environment files found to load")
	}

	if overload {
		if err := godotenv.Overload(foundFiles...); err != nil {
			return fmt.Errorf("error loading environment files: %w", err)
		}
	} else {
		if err := godotenv.Load(foundFiles...); err != nil {
			return fmt.Errorf("error loading environment files: %w", err)
		}
	}

	fmt.Printf("Searching files: %v\n", filePaths)
	fmt.Printf("Found files: %v\n", foundFiles)

	return nil
}

// Pública: no sobrescribe
func LoadConfig(filePaths ...string) error {
	return loadEnvFiles(filePaths, false)
}

// Pública: sobrescribe siempre
func OverloadConfig(filePaths ...string) error {
	return loadEnvFiles(filePaths, true)
}
