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

package snowflake

import "github.com/gojek/heimdall/v7"

// SnowflakeConfig holds the configuration for Snowflake client
type SnowflakeConfig struct {
	PAT     string
	BaseURL string
}

// SnowflakeClient is the client for interacting with Snowflake REST API
type SnowflakeClient struct {
	config *SnowflakeConfig
	client heimdall.Doer
}

// SnowflakeUser represents a user object from Snowflake API response
type SnowflakeUser struct {
	Name        string `json:"name"`
	Email       string `json:"email,omitempty"`
	DisplayName string `json:"displayName,omitempty"`
}

// SnowflakeGrant represents a grant object from Snowflake grants API response
type SnowflakeGrant struct {
	GrantedTo   string `json:"granted_to"`
	GranteeName string `json:"grantee_name"`
}

// SnowflakeRole represents a role object from Snowflake roles API response
type SnowflakeRole struct {
	Name string `json:"name"`
}
