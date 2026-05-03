package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestEmailRegexAcceptsValidEmail(t *testing.T) {
	assert.True(t, emailRegex.MatchString("user.name+tag@example.com"))
}

func TestGetUserGroupsValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		email        string
		expectedBody string
	}{
		{
			name:         "empty email returns required error",
			email:        "",
			expectedBody: `{"error":"email parameter is required"}`,
		},
		{
			name:         "not an email returns invalid format",
			email:        "not-an-email",
			expectedBody: `{"error":"invalid email format"}`,
		},
		{
			name:         "missing local part returns invalid format",
			email:        "@nodomain",
			expectedBody: `{"error":"invalid email format"}`,
		},
		{
			name:         "missing domain returns invalid format",
			email:        "foo@",
			expectedBody: `{"error":"invalid email format"}`,
		},
		{
			name:         "plain value returns invalid format",
			email:        "plain",
			expectedBody: `{"error":"invalid email format"}`,
		},
		{
			name:         "redis pattern probe returns invalid format",
			email:        "*@*",
			expectedBody: `{"error":"invalid email format"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/user/"+tt.email+"/groups", nil)
			c.Params = gin.Params{{Key: "email", Value: tt.email}}

			h := &Handlers{}
			h.GetUserGroups(c)

			assert.Equal(t, http.StatusBadRequest, w.Code)
			assert.JSONEq(t, tt.expectedBody, w.Body.String())
		})
	}
}
