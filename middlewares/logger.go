package middlewares

import (
	"fmt"
	"github.com/gin-gonic/gin"
)

func myLogger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s %s %s \n", param.ClientIP, param.Method, param.Path)
	})
}
