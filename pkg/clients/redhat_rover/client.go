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

package redhatrover

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gojek/heimdall/v7"
	"github.com/redhat-data-and-ai/usernaut/pkg/request"
	"github.com/redhat-data-and-ai/usernaut/pkg/request/httpclient"
	"github.com/redhat-data-and-ai/usernaut/pkg/utils"
)

type RoverClient struct {
	client             heimdall.Doer
	serviceAccountName string
	url                string
}

type RoverConfig struct {
	URL                string `json:"url"`
	PrivateKeyPath     string `json:"private_key_path"`
	CertPath           string `json:"cert_path"`
	ServiceAccountName string `json:"service_account_name"`
}

func NewClient(roverAppConfig map[string]interface{},
	connectionPoolConfig httpclient.ConnectionPoolConfig,
	hystrixResiliencyConfig httpclient.HystrixResiliencyConfig) (*RoverClient, error) {

	// parse the rover configuration from the provided maps
	roverConfig := RoverConfig{}
	if err := utils.MapToStruct(roverAppConfig, &roverConfig); err != nil {
		return nil, err
	}

	// return an error if any of the required fields are missing
	// URL, CertPath, or PrivateKeyPath
	if roverConfig.CertPath == "" || roverConfig.PrivateKeyPath == "" ||
		roverConfig.URL == "" || roverConfig.ServiceAccountName == "" {
		return nil, fmt.Errorf(
			"rover configuration is missing required fields: URL, CertPath, PrivateKeyPath, or ServiceAccountName")
	}

	connectionPoolConfig.CertPath = roverConfig.CertPath
	connectionPoolConfig.PrivateKeyPath = roverConfig.PrivateKeyPath

	client, err := httpclient.InitializeClient(
		"redhat_rover",
		connectionPoolConfig,
		hystrixResiliencyConfig,
		heimdall.NewRetrier(heimdall.NewConstantBackoff(100*time.Millisecond, 50*time.Millisecond)), 3,
		nil)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize http client: %w", err)
	}

	return &RoverClient{
		client:             client,
		url:                roverConfig.URL,
		serviceAccountName: roverConfig.ServiceAccountName,
	}, nil
}

func (rC *RoverClient) sendRequest(ctx context.Context, url string, method string, body interface{},
	headers map[string]string, methodName string) ([]byte, int, error) {

	// Validate URL and method
	if url == "" {
		return nil, 0, fmt.Errorf("URL cannot be empty")
	}
	if method == "" {
		return nil, 0, fmt.Errorf("HTTP method cannot be empty")
	}

	// For DELETE requests, we don't want to send "null" in request body
	var requestBody []byte
	if body != nil {
		var err error
		requestBody, err = json.Marshal(body)
		if err != nil {
			return nil, 0, err
		}
	}

	req, err := request.NewRequest(ctx, method, url, requestBody)
	if err != nil {
		return nil, 0, err
	}
	req.SetHeaders(headers)

	return req.MakeRequest(rC.client, methodName, "redhat_rover")
}
