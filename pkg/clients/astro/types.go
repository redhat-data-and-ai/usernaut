/*
Copyright 2026.

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

package astro

import "github.com/gojek/heimdall/v7"

// AstroConfig holds the configuration for Astro client
type AstroConfig struct {
	APIToken string
	BaseURL  string
}

// AstroClient is the client for interacting with Astro REST API
type AstroClient struct {
	config *AstroConfig
	client heimdall.Doer
}

// AstroUser represents a user object from Astro API response
type AstroUser struct {
	ID               string `json:"id"`
	Username         string `json:"username"`
	FullName         string `json:"fullName,omitempty"`
	OrganizationRole string `json:"organizationRole,omitempty"`
}

// AstroUsersResponse represents the response from list users API
type AstroUsersResponse struct {
	Users []AstroUser `json:"users"`
}

// AstroTeam represents a team object from Astro API response
type AstroTeam struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	OrganizationRole string `json:"organizationRole,omitempty"`
	Description      string `json:"description,omitempty"`
}

// AstroTeamsResponse represents the response from list teams API
type AstroTeamsResponse struct {
	Teams []AstroTeam `json:"teams"`
}

// AstroTeamMember represents a team member from Astro API response
type AstroTeamMember struct {
	UserID   string `json:"userId"`
	Username string `json:"username,omitempty"`
	FullName string `json:"fullName,omitempty"`
}

// AstroTeamMembersResponse represents the response from list team members API
type AstroTeamMembersResponse struct {
	TeamMembers []AstroTeamMember `json:"teamMembers"`
}

// CreateTeamRequest represents the request body for creating a team
type CreateTeamRequest struct {
	Name             string `json:"name"`
	Description      string `json:"description,omitempty"`
	OrganizationRole string `json:"organizationRole,omitempty"`
}

// CreateInviteRequest represents the request body for inviting a user
type CreateInviteRequest struct {
	InviteeEmail string `json:"inviteeEmail"`
	Role         string `json:"role"`
}

// CreateInviteResponse represents the response from invite user API
type CreateInviteResponse struct {
	UserID string `json:"userId"`
}

// AddTeamMembersRequest represents the request body for adding members to a team
type AddTeamMembersRequest struct {
	MemberIDs []string `json:"memberIds"`
}

// UpdateUserRoleRequest represents the request body for updating user role (used for deletion)
type UpdateUserRoleRequest struct {
	OrganizationRole *string `json:"organizationRole"`
}

// Astro API constants
const (
	// Default pagination limit for Astro API
	DefaultPageLimit = 1000

	// Default organization role for new users and teams
	DefaultOrganizationRole = "ORGANIZATION_MEMBER"

	// Default base URL for Astro API
	DefaultBaseURL = "https://api.astronomer.io"
)
