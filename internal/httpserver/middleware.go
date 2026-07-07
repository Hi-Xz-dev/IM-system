package httpserver

import (
	"IM-system/internal/logger"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// RequestLogger() 不是中间件本身；
// RequestLogger() 返回的那个匿名函数才是真正的中间件。
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now() //记录开始时间
		c.Next()
		cost := time.Since(start) //整个请求处理链耗时
		logger.Log.Info(
			"http request",
			"client_ip", c.ClientIP(),
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"cost", cost,
		)
	}
}

func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger.Log.Error(
					"panic recovered",
					"panic", err,
					"method", c.Request.Method,
					"path", c.Request.URL.Path,
					"status", http.StatusInternalServerError,
				)

				c.JSON(http.StatusInternalServerError, Fail("internal server error"))
			}
		}()

		c.Next()

	}
}
