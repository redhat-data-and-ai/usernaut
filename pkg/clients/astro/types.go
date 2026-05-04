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

package astro

import "github.com/gojek/heimdall/v7"

// AstroConfig holds the configuration for Astro client
type AstroConfig struct {
	APIToken       string
	OrganizationID string
	BaseURL        string
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
	Status           string `json:"status,omitempty"`
	FullName         string `json:"fullName,omitempty"`
	AvatarURL        string `json:"avatarUrl,omitempty"`
	CreatedAt        string `json:"createdAt,omitempty"`
	UpdatedAt        string `json:"updatedAt,omitempty"`
	OrganizationRole string `json:"organizationRole,omitempty"`
}

// AstroUsersResponse represents the response from list users API
type AstroUsersResponse struct {
	Users      []AstroUser `json:"users"`
	Offset     int         `json:"offset,omitempty"`
	Limit      int         `json:"limit,omitempty"`
	TotalCount int         `json:"totalCount,omitempty"`
}

// AstroTeam represents a team object from Astro API response
type AstroTeam struct {
	ID               string           `json:"id"`
	Name             string           `json:"name"`
	OrganizationID   string           `json:"organizationId,omitempty"`
	OrganizationRole string           `json:"organizationRole,omitempty"`
	Description      string           `json:"description,omitempty"`
	IsIdpManaged     bool             `json:"isIdpManaged,omitempty"`
	WorkspaceRoles   []WorkspaceRole  `json:"workspaceRoles,omitempty"`
	DeploymentRoles  []DeploymentRole `json:"deploymentRoles,omitempty"`
	RolesCount       int              `json:"rolesCount,omitempty"`
	CreatedAt        string           `json:"createdAt,omitempty"`
	UpdatedAt        string           `json:"updatedAt,omitempty"`
}

// WorkspaceRole represents a workspace role assignment
type WorkspaceRole struct {
	WorkspaceID string `json:"workspaceId"`
	Role        string `json:"role"`
}

// DeploymentRole represents a deployment role assignment
type DeploymentRole struct {
	DeploymentID string `json:"deploymentId"`
	Role         string `json:"role"`
}

// AstroTeamsResponse represents the response from list teams API
type AstroTeamsResponse struct {
	Teams      []AstroTeam `json:"teams"`
	Offset     int         `json:"offset,omitempty"`
	Limit      int         `json:"limit,omitempty"`
	TotalCount int         `json:"totalCount,omitempty"`
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
	Offset      int               `json:"offset,omitempty"`
	Limit       int               `json:"limit,omitempty"`
	TotalCount  int               `json:"totalCount,omitempty"`
}

// CreateTeamRequest represents the request body for creating a team
type CreateTeamRequest struct {
	Name             string   `json:"name"`
	Description      string   `json:"description,omitempty"`
	MemberIDs        []string `json:"memberIds,omitempty"`
	OrganizationRole string   `json:"organizationRole,omitempty"`
}

// CreateInviteRequest represents the request body for inviting a user
type CreateInviteRequest struct {
	InviteeEmail string `json:"inviteeEmail"`
	Role         string `json:"role"`
}

// CreateInviteResponse represents the response from invite user API
type CreateInviteResponse struct {
	UserID string `json:"userId"`
	Invite struct {
		ID           string `json:"id"`
		InviteeEmail string `json:"inviteeEmail"`
	} `json:"invite,omitempty"`
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
	DefaultPageLimit = 100

	// Default organization role for new users and teams
	DefaultOrganizationRole = "ORGANIZATION_MEMBER"

	// Default base URL for Astro API
	DefaultBaseURL = "https://api.astronomer.io"
)
