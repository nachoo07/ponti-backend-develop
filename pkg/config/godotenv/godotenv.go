package pkgenvs

import (
	"errors"
	"fmt"

	"github.com/joho/godotenv"

	pkgutils "github.com/alphacodinggroup/ponti-backend/pkg/utils"
)

func LoadConfig(filePaths ...string) error {
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

	if err := godotenv.Load(foundFiles...); err != nil {
		return fmt.Errorf("error loading environment files: %w", err)
	}

	return nil
}
