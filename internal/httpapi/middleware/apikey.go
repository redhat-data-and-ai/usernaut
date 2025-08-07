package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/redhat-data-and-ai/usernaut/pkg/config"
	"github.com/sirupsen/logrus"
)

func APIKeyAuth(cfg *config.AppConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !cfg.APIServer.Auth.Enabled {
			c.Next()
			return
		}

		apiKey := c.GetHeader("X-API-Key")

		if apiKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "API key required",
				"hint":  "Add X-API-Key header",
			})
			c.Abort()
			return
		}

		valid := false
		for _, validKey := range cfg.APIServer.Auth.APIKeys {
			if apiKey == validKey {
				valid = true
				break
			}
		}

		if !valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid API key"})
			c.Abort()
			return
		}

<<<<<<< HEAD
		logrus.Info("API request authenticated")
=======
		logrus.Debug("API request authenticated")
>>>>>>> f6e3bef (API skeleton code to add endpoints as required)
		c.Next()
	}
}
