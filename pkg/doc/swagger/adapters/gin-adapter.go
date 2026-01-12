package ginadapter

import (
	gin "github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	pkgswagger "github.com/alphacodinggroup/ponti-backend/pkg/doc/swagger"
)

// SetupSwagger configura Swagger en un router Gin
func SetupSwagger(engine *gin.Engine, service pkgswagger.Service) error {
	// Primero configura las rutas base usando el servicio Swagger
	addRoute := func(config pkgswagger.HandlerConfig) {
		handler := gin.WrapH(config.Handler)
		engine.Handle(config.Method, config.Path, handler)
	}

	if err := service.Setup(addRoute); err != nil {
		return err
	}

	// Usa una ruta más específica
	engine.GET("/api-docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler,
		ginSwagger.URL("/api-docs/doc.json"),
		ginSwagger.DefaultModelsExpandDepth(-1),
		ginSwagger.DocExpansion("none"),
	))

	return nil
}
