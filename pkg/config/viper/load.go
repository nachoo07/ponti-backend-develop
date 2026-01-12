package pkgviper

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"

	pkgutils "github.com/alphacodinggroup/ponti-backend/pkg/utils"
)

// Load loads multiple configuration files using Viper.
// It merges all found configuration files into Viper's configuration.
// Returns an error if no files are provided, no files are found, or all loading attempts fail.
func LoadConfig(filePaths ...string) error {
	if len(filePaths) == 0 {
		return errors.New("no file paths provided")
	}

	// Find and filter existing files using FilesFinder
	foundFiles, err := pkgutils.FilesFinder(filePaths...)
	if err != nil {
		return fmt.Errorf("fatal error: failed to find configuration files: %w", err)
	}

	if len(foundFiles) == 0 {
		return errors.New("no configuration files found to load")
	}

	// Configure Viper to read environment variables
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

	// If no file was successfully loaded, return an error
	if !successfullyLoaded {
		return fmt.Errorf("failed to load any configuration files: %v", loadErrors)
	}

	// If some files failed to load, print the errors
	if len(loadErrors) > 0 {
		fmt.Printf("Some configuration files failed to load:\n%s\n", strings.Join(loadErrors, "\n"))
	}

	return nil
}

// configureViper sets up Viper to load environment variables
func configureViper() {
	viper.SetEnvPrefix("")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
}

// loadViperConfig loads and merges a configuration file into Viper
func loadViperConfig(configFilePath string) error {
	fileNameWithoutExt, fileExtension, err := pkgutils.FileNameAndExtension(configFilePath)
	if err != nil {
		return fmt.Errorf("invalid file '%s': %w", configFilePath, err)
	}

	viper.SetConfigName(fileNameWithoutExt)
	viper.SetConfigType(fileExtension)

	dir := filepath.Dir(configFilePath)
	viper.AddConfigPath(dir)

	// Use MergeInConfig to merge multiple configurations
	if err := viper.MergeInConfig(); err != nil {
		return fmt.Errorf("error reading config: %w", err)
	}

	return nil
}
