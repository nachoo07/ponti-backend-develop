package pkgmwr

import "github.com/gin-gonic/gin"

type Middlewares struct {
	Global    []gin.HandlerFunc
	Validated []gin.HandlerFunc
	Protected []gin.HandlerFunc
}
