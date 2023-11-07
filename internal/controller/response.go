package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type ResponseData struct {
	Code ReturnCode  `json:"code"`
	Msg  interface{} `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

func ResponseSuccess(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, &ResponseData{
		Code: CodeSuccess,
		Msg:  CodeSuccess.Message(),
		Data: data,
	})
}

func ResponseError(c *gin.Context, code ReturnCode) {
	c.JSON(http.StatusOK, &ResponseData{
		Code: code,
		Msg:  code.Message(),
		Data: nil,
	})
}

func ResponseErrorWithMessage(c *gin.Context, code ReturnCode, message interface{}) {
	c.JSON(http.StatusOK, &ResponseData{
		Code: code,
		Msg:  message,
		Data: nil,
	})
}
