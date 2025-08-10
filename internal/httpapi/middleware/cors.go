package middleware

import (
	"github.com/gin-gonic/gin"
)



func CORS(cfg *config.APIServerConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		allowed := false
		for _, allowedOrigin := range cfg.CORS.AllowedOrigins {
			if origin == allowedOrigin {
				allowed = true
				break
			}
		}
		if allowed {
			c.Header("Access-Control-Allow-Origin", origin)
		}
		c.Header("Access-Control-Allow-Methods", strings.Join(cfg.CORS.AllowedMethods, ", "))
		c.Header("Access-Control-Allow-Headers", strings.Join(cfg.CORS.AllowedHeaders, ", "))
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
