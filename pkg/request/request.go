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

package request

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"time"

	"github.com/gojek/heimdall/v7"
	"github.com/redhat-data-and-ai/usernaut/pkg/logger"
	"github.com/sirupsen/logrus"

	"github.com/opentracing-contrib/go-stdlib/nethttp"
	ot "github.com/opentracing/opentracing-go"
)

// IRequester exposes Setter for Header and final method to
// make a request: MakeRequest
type IRequester interface {
	GetHeaders() http.Header
	SetHeaders(map[string]string) IRequester
	MakeRequest(heimdall.Doer, string, string) ([]byte, int, error)
	MakeRequestWithHeader(heimdall.Doer, string, string) ([]byte, http.Header, int, error)
}

type Requester struct {
	request *http.Request
}

// NewRequest creates a new Request with the given context, method, URL, and body.
func NewRequest(ctx context.Context, method string, url string, body []byte) (IRequester, error) {
	var err error
	r := &Requester{}
	if r.request, err = http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(body)); err != nil {
		return nil, err
	}
	return r, nil
}

func (r *Requester) GetHeaders() http.Header {
	return r.request.Header
}

func (r *Requester) SetHeaders(headers map[string]string) IRequester {
	for key, value := range headers {
		r.request.Header.Add(key, value)
	}
	return r
}

func (r *Requester) MakeRequest(httpClient heimdall.Doer, methodName string, serviceName string) ([]byte, int, error) {
	response, responseBody, err := r.sendRequest(httpClient, methodName, serviceName)
	if err != nil {
		return nil, http.StatusBadGateway, err
	}

	return responseBody, response.StatusCode, nil
}

func (r *Requester) MakeRequestWithHeader(httpClient heimdall.Doer, methodName string,
	serviceName string) ([]byte, http.Header, int, error) {
	response, responseBody, err := r.sendRequest(httpClient, methodName, serviceName)
	if err != nil {
		return nil, nil, http.StatusBadGateway, err
	}

	return responseBody, response.Header, response.StatusCode, nil
}

// sendRequest contains the common logic for making HTTP requests with logging and tracing
func (r *Requester) sendRequest(httpClient heimdall.Doer, methodName string,
	serviceName string) (*http.Response, []byte, error) {
	// transmit span's TraceContext as HTTP headers to api
	if span := ot.SpanFromContext(r.request.Context()); span != nil {
		_, ok := span.Tracer().(ot.NoopTracer)
		if !ok {
			var ht *nethttp.Tracer
			r.request, ht = nethttp.TraceRequest(ot.GlobalTracer(), r.request)
			defer ht.Finish()
		}
	}

	// Get start time
	start := time.Now()

	log := logger.Logger(r.request.Context())

	log.WithFields(logrus.Fields{
		"service": serviceName,
		"method":  methodName,
		"url":     r.request.URL.String(),
	}).Info("SENDING_HTTP_REQUEST")

	response, err := httpClient.Do(r.request)

	log.WithFields(logrus.Fields{
		"service": serviceName,
		"method":  methodName,
		"url":     r.request.URL.String(),
	}).Info("RECEIVED_HTTP_RESPONSE")

	// Calculate time taken to receive response
	durationMs := float64(time.Since(start).Nanoseconds() / 1000000)

	log.WithFields(logrus.Fields{
		"service":    serviceName,
		"method":     methodName,
		"url":        r.request.URL.String(),
		"durationMs": durationMs,
	}).Info("HTTP_RESPONSE_DURATION")

	if err != nil {
		return nil, nil, err
	}

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, nil, err
	}

	if err := response.Body.Close(); err != nil {
		return nil, nil, err
	}

	return response, responseBody, nil
}
