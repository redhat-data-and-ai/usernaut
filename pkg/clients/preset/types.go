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

package preset

import (
	"github.com/gojek/heimdall/v7"
)

// PresetClient is the HTTP client for Preset SCIM API
type PresetClient struct {
	client    heimdall.Doer
	baseURL   string
	scimToken string
	teamSlug  string
}

// PresetConfig holds the configuration needed to connect to Preset
type PresetConfig struct {
	SCIMToken string `json:"scim_token"`
	BaseURL   string `json:"base_url"`
	TeamSlug  string `json:"team_slug"`
}

// scimUser represents a user in SCIM format
type scimUser struct {
	ID       string `json:"id"`
	UserName string `json:"userName"`
	Emails   []struct {
		Value   string `json:"value"`
		Primary bool   `json:"primary"`
	} `json:"emails"`
	DisplayName string `json:"displayName"`
	Active      bool   `json:"active"`
}

// scimUsersResponse represents the SCIM Users list response
type scimUsersResponse struct {
	TotalResults int        `json:"totalResults"`
	Resources    []scimUser `json:"Resources"`
}

// scimUserCreateRequest represents the request body to create a SCIM user
type scimUserCreateRequest struct {
	Schemas  []string         `json:"schemas"`
	UserName string           `json:"userName"`
	Emails   []scimEmailValue `json:"emails"`
	Name     scimName         `json:"name"`
	Active   bool             `json:"active"`
}

// scimEmailValue represents an email entry in SCIM format
type scimEmailValue struct {
	Value   string `json:"value"`
	Primary bool   `json:"primary"`
}

// scimName represents the name component of a SCIM user
type scimName struct {
	GivenName  string `json:"givenName"`
	FamilyName string `json:"familyName"`
}

// scimGroup represents a group in SCIM format
type scimGroup struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
	Members     []struct {
		Value   string `json:"value"`
		Display string `json:"display"`
	} `json:"members"`
}

// scimGroupsResponse represents the SCIM Groups list response
type scimGroupsResponse struct {
	TotalResults int         `json:"totalResults"`
	Resources    []scimGroup `json:"Resources"`
}

// scimGroupCreateRequest represents the request body to create a SCIM group
type scimGroupCreateRequest struct {
	Schemas     []string `json:"schemas"`
	DisplayName string   `json:"displayName"`
}

// scimPatchOperation represents a single SCIM PATCH operation
type scimPatchOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value"`
}

// scimPatchRequest represents a SCIM PATCH request body
type scimPatchRequest struct {
	Schemas    []string             `json:"schemas"`
	Operations []scimPatchOperation `json:"Operations"`
}

// scimMemberValue represents a member reference in SCIM PATCH operations
type scimMemberValue struct {
	Value string `json:"value"`
}
