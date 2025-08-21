package middleware

import (
	"crypto/subtle"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/redhat-data-and-ai/usernaut/pkg/config"
	
)

func BasicAuth(cfg *config.AppConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !cfg.APIServer.Auth.Enabled {
			c.Next()
			return
		}

		username, password, ok := c.Request.BasicAuth()
		if !ok || username == "" || password == "" {
			c.Header("WWW-Authenticate", `Basic realm="Usernaut"`)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		authorized := false
		for _, u := range cfg.APIServer.Auth.BasicUsers {
			if subtle.ConstantTimeCompare([]byte(username), []byte(u.Username)) == 1 &&
				subtle.ConstantTimeCompare([]byte(password), []byte(u.Password)) == 1 {
				authorized = true
				break
			}
		}

		if !authorized {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		c.Set("clientId", username)
		c.Next()
	}
}
