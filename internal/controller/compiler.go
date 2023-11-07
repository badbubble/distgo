package controller

import "github.com/gin-gonic/gin"

func CompileHandler(c *gin.Context) {
	ResponseSuccess(c, "you are successful access this api.")
}
