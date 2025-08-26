/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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
