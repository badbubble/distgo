package router

import (
	"distgo/internal/controller"

	"github.com/gin-gonic/gin"
)

func SetupRouter(mode string) (error, *gin.Engine) {
	if mode == gin.ReleaseMode {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	err := r.SetTrustedProxies([]string{"127.0.0.1"})
	if err != nil {
		return err, nil
	}
	v1 := r.Group("/api/v1")
	{
		v1.POST("/compile", controller.CompileHandler)
	}
	return nil, r
}
