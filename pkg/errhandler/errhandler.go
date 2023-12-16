package errhandler

import (
	"net/http"
	"pottogether/pkg/logger"

	"github.com/gin-gonic/gin"
)

func Info(c *gin.Context, err error, msg string) {
	c.JSON(400, gin.H{
		"isSuccess": false,
		"message":   msg + ": " + err.Error(),
	})
	logger.Error(msg + ": " + err.Error())
}

func Error(c *gin.Context, err error, msg string) {
	c.JSON(http.StatusInternalServerError, gin.H{
		"isSuccess": false,
		"message":   msg + ": " + err.Error(),
	})
	logger.Error(msg + ": " + err.Error())
}

func Unauthorized(c *gin.Context, err error, msg string) {
	c.JSON(http.StatusUnauthorized, gin.H{
		"isSuccess": false,
		"message":   msg + ": " + err.Error(),
	})
	logger.Info(msg + ": " + err.Error())
}
