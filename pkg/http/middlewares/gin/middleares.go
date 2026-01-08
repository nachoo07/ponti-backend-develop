package pkgmwr

import (
	"github.com/gin-gonic/gin"

	pkgutils "github.com/alphacodinggroup/ponti-backend/pkg/utils"
)

type Middlewares struct {
	global     []gin.HandlerFunc
	validation []gin.HandlerFunc
	protected  []gin.HandlerFunc
}

func NewDefaultMiddlewares() *Middlewares {
	cfg := pkgutils.NewConfigFromEnv()
	global := []gin.HandlerFunc{
		ErrorHandling(),
		RequestAndResponseLogger(HttpLoggingOptions{
			LogLevel:       "info",
			IncludeHeaders: true,
			IncludeBody:    false,
			ExcludedPaths:  []string{"/health", "/ping", "/swagger/spec", "/swagger/ui/index.html"},
		}),
	}
	validation := []gin.HandlerFunc{
		RequireUserIDHeader(),
		RequireAPIKey(),
	}
	protected := []gin.HandlerFunc{
		RequireJWT(cfg),
	}
	return &Middlewares{
		global:     global,
		validation: validation,
		protected:  protected,
	}
}
func (m *Middlewares) GetGlobal() []gin.HandlerFunc     { return m.global }
func (m *Middlewares) GetValidation() []gin.HandlerFunc { return m.validation }
func (m *Middlewares) GetProtected() []gin.HandlerFunc  { return m.protected }
