package pkgast

import (
	"github.com/spf13/viper"

	defs "github.com/alphacodinggroup/ponti-backend/pkg/repo-tools/ast/defs"
)

// Bootstrap inicializa y valida la configuración del AST parser.
func Bootstrap() (defs.Service, error) {
	config := newConfig(
		viper.GetString("AST_ANALYZE_PATH"),
	)

	if err := config.Validate(); err != nil {
		return nil, err
	}

	return newService(config)
}
