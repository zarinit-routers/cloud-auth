package middleware

import (
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"
)

// CORS middleware для обработки CORS заголовков
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Разрешаем запросы с любых локальных адресов для development
		allowedOrigins := []string{
			"http://localhost:3000",
			"http://127.0.0.1:3000",
			"http://localhost:3001",
			"http://127.0.0.1:3001",
		}

		allowOrigin := ""

		if slices.Contains(allowedOrigins, origin) {
			allowOrigin = origin
		}

		if allowOrigin == "" && origin != "" {
			// Для development разрешаем любые origin (осторожно! только для разработки)
			allowOrigin = origin
		}

		c.Writer.Header().Set("Access-Control-Allow-Origin", allowOrigin)
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length, Content-Type")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
