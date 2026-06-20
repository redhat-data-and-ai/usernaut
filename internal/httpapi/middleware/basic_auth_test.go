package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/redhat-data-and-ai/usernaut/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestBasicAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		users          []config.BasicUser
		username       string
		password       string
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "correct username and password authorizes request",
			users: []config.BasicUser{
				{Username: "test-user", Password: "test-password"},
			},
			username:       "test-user",
			password:       "test-password",
			expectedStatus: http.StatusOK,
			expectedBody:   "test-user",
		},
		{
			name: "wrong password returns unauthorized",
			users: []config.BasicUser{
				{Username: "test-user", Password: "test-password"},
			},
			username:       "test-user",
			password:       "wrong-password",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "wrong username returns unauthorized",
			users: []config.BasicUser{
				{Username: "test-user", Password: "test-password"},
			},
			username:       "wrong-user",
			password:       "test-password",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "empty username and password returns unauthorized",
			users: []config.BasicUser{
				{Username: "test-user", Password: "test-password"},
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "matching second configured credential authorizes request",
			users: []config.BasicUser{
				{Username: "first-user", Password: "first-password"},
				{Username: "second-user", Password: "second-password"},
			},
			username:       "second-user",
			password:       "second-password",
			expectedStatus: http.StatusOK,
			expectedBody:   "second-user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.AppConfig{
				APIServer: config.APIServerConfig{
					Auth: config.AuthConfig{
						Enabled:    true,
						BasicUsers: tt.users,
					},
				},
			}

			router := gin.New()
			router.Use(BasicAuth(cfg))
			router.GET("/protected", func(c *gin.Context) {
				clientID, _ := c.Get("clientId")
				c.String(http.StatusOK, clientID.(string))
			})

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			req.SetBasicAuth(tt.username, tt.password)

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, tt.expectedBody, w.Body.String())
		})
	}
}
