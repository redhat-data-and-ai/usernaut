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
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/redhat-data-and-ai/usernaut/pkg/config"
)

// compareKey is a per-process random key. It never leaves the process and is
// not used to store anything; it exists so credential comparison happens over
// keyed tags rather than bare digests of the secrets themselves.
var compareKey = func() []byte {
	k := make([]byte, 32)
	if _, err := rand.Read(k); err != nil {
		panic("usernaut: cannot initialise basic-auth compare key: " + err.Error())
	}
	return k
}()

// credentialTag returns a fixed-length keyed tag for s. Comparing tags with
// hmac.Equal is constant time and, unlike comparing the raw values, does not
// leak length. The tag is ephemeral and never persisted.
func credentialTag(s string) []byte {
	mac := hmac.New(sha256.New, compareKey)
	mac.Write([]byte(s))
	return mac.Sum(nil)
}

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
			usernameMatches := hmac.Equal(credentialTag(username), credentialTag(u.Username))
			passwordMatches := hmac.Equal(credentialTag(password), credentialTag(u.Password))
			if usernameMatches && passwordMatches {
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
