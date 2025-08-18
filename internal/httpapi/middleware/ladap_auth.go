package middleware

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	goldap "github.com/go-ldap/ldap/v3"
	"github.com/redhat-data-and-ai/usernaut/pkg/config"
)

func LDAPBasicAuth(cfg *config.AppConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !cfg.APIServer.Auth.Enabled {
			c.Next()
			return
		}
		username, password, ok := c.Request.BasicAuth()
		if !ok || username == "" || password == "" {
			c.Header("authenticate", `Basic = "usernaut"`)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		userDN := fmt.Sprintf(cfg.LDAP.UserDN, username)

		conn, err := goldap.DialURL(cfg.LDAP.Server, goldap.DialWithDialer(&net.Dialer{Timeout: 10 * time.Second}))

		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authentication failed"})
			return
		}
		defer conn.Close()

		if strings.HasPrefix(strings.ToLower(cfg.LDAP.Server), "ldap://") {
			_ = conn.StartTLS(nil)
		}
		if err := conn.Bind(userDN, password); err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}
		c.Set("userId", username)
		c.Set("userDN", userDN)
		c.Next()
	}
}
