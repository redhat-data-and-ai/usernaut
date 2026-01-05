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

package atlan

import (
	atlansdk "github.com/atlanhq/atlan-go/atlan/assets"
	"github.com/gojek/heimdall/v7"
)

// AtlanClient is the HTTP client for Atlan API
type AtlanClient struct {
	client                heimdall.Doer
	sdkClient             *atlansdk.AtlanClient
	url                   string
	apiToken              string
	identityProviderAlias string
	// assetTransferUsername is the username to transfer asset ownership to when deleting a user.
	// Atlan requires ownership transfer before user deletion.
	assetTransferUsername string
	defaultPersona        string
	ssoSync               bool
	ldapSync              bool
	ssoGroupName          string
}

// AtlanConfig holds the configuration needed to connect to Atlan
type AtlanConfig struct {
	URL                   string `json:"url"`
	APIToken              string `json:"api_token"`
	IdentityProviderAlias string `json:"identity_provider_alias"`
	// AssetTransferUsername is the username to transfer asset ownership to when deleting a user
	AssetTransferUsername string `json:"asset_transfer_username"`
	DefaultPersona        string `json:"default_persona"`
}

// AtlanUser represents a user in Atlan's API response
type AtlanUser struct {
	ID          string `json:"id"`
	Username    string `json:"username"`
	Email       string `json:"email"`
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
	DisplayName string `json:"name"`
	Enabled     bool   `json:"enabled"`
}

// AtlanUsersResponse represents the response from Atlan's users API
type AtlanUsersResponse struct {
	TotalRecord  int         `json:"totalRecord"`
	FilterRecord int         `json:"filterRecord"`
	Records      []AtlanUser `json:"records"`
}

// AtlanGroup represents a group in Atlan's API response
type AtlanGroup struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// AtlanGroupsResponse represents the response from Atlan's groups API
type AtlanGroupsResponse struct {
	Records      []AtlanGroup `json:"records"`
	TotalRecord  int          `json:"totalRecord"`
	FilterRecord int          `json:"filterRecord"`
}

// AtlanGroupCreateResponse represents the response when creating a group
type AtlanGroupCreateResponse struct {
	Group string `json:"group"`
}

// AtlanGroupMember represents a member in Atlan's group members API response
type AtlanGroupMember struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

// AtlanGroupMembersResponse represents the response from Atlan's group members API
type AtlanGroupMembersResponse struct {
	Records      []AtlanGroupMember `json:"records"`
	TotalRecord  int                `json:"totalRecord"`
	FilterRecord int                `json:"filterRecord"`
}

// SSOGroupMappingConfig holds the configuration for SSO group mapping
type SSOGroupMappingConfig struct {
	SyncMode                string `json:"syncMode"`
	Attributes              string `json:"attributes"`
	AreAttributeValuesRegex string `json:"are.attribute.values.regex"`
	AttributeName           string `json:"attribute.name"`
	Group                   string `json:"group"`
	AttributeValue          string `json:"attribute.value"`
}

// SSOGroupMapping represents an SSO group mapping in Atlan
type SSOGroupMapping struct {
	ID                     string                `json:"id,omitempty"`
	Name                   string                `json:"name"`
	IdentityProviderMapper string                `json:"identityProviderMapper"`
	IdentityProviderAlias  string                `json:"identityProviderAlias"`
	Config                 SSOGroupMappingConfig `json:"config"`
}
