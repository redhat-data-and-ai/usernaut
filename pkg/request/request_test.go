package request

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockClient is a mock implementation of heimdall.Doer
// to be used in tests.
type MockClient struct {
	mock.Mock
}

const (
	exampleURL = "http://example.com"
)

func (m *MockClient) Do(req *http.Request) (*http.Response, error) {
	args := m.Called(req)
	return args.Get(0).(*http.Response), args.Error(1)
}

func TestNewRequest(t *testing.T) {
	ctx := context.Background()
	method := http.MethodGet
	url := exampleURL
	body := []byte("test body")

	req, err := NewRequest(ctx, method, url, body)

	assert.NoError(t, err)
	assert.NotNil(t, req)
	assert.Equal(t, method, req.(*Requester).request.Method)
	assert.Equal(t, url, req.(*Requester).request.URL.String())
	bodyContent, _ := io.ReadAll(req.(*Requester).request.Body)
	assert.Equal(t, body, bodyContent)
}

func TestSetHeaders(t *testing.T) {
	ctx := context.Background()
	method := http.MethodGet
	url := exampleURL
	body := []byte("test body")

	req, _ := NewRequest(ctx, method, url, body)
	headers := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": "Bearer token",
	}

	req.SetHeaders(headers)

	for key, value := range headers {
		assert.Equal(t, value, req.GetHeaders().Get(key))
	}
}

func TestMakeRequest(t *testing.T) {
	ctx := context.Background()
	method := http.MethodGet
	url := exampleURL
	body := []byte("test body")

	req, _ := NewRequest(ctx, method, url, body)

	mockClient := new(MockClient)
	response := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString("response body")),
	}

	mockClient.On("Do", req.(*Requester).request).Return(response, nil)

	responseBody, statusCode, err := req.MakeRequest(mockClient, "TestMethod", "TestService")

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, statusCode)
	assert.Equal(t, "response body", string(responseBody))

	mockClient.AssertExpectations(t)
}
