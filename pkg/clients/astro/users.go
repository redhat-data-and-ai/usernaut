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

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/redhat-data-and-ai/usernaut/pkg/common/structs"
	"github.com/redhat-data-and-ai/usernaut/pkg/logger"
	"github.com/sirupsen/logrus"
)

// astroUserToStruct converts an AstroUser to a structs.User
func astroUserToStruct(user AstroUser) *structs.User {
	// Extract first and last name from full name if available
	firstName := ""
	lastName := ""
	if user.FullName != "" {
		parts := strings.SplitN(user.FullName, " ", 2)
		firstName = parts[0]
		if len(parts) > 1 {
			lastName = parts[1]
		}
	}

	return &structs.User{
		ID:          user.ID,
		UserName:    strings.ToLower(user.Username),
		Email:       strings.ToLower(user.Username), // In Astro, username is the email
		FirstName:   firstName,
		LastName:    lastName,
		DisplayName: user.FullName,
		Role:        user.OrganizationRole,
	}
}

// FetchAllUsers fetches all users from Astro using REST API with pagination
// Returns 2 maps: 1st map keyed by ID, 2nd map keyed by email
func (c *AstroClient) FetchAllUsers(ctx context.Context) (map[string]*structs.User,
	map[string]*structs.User, error) {
	log := logger.Logger(ctx).WithField("service", "astro")

	log.Info("fetching all users")
	resultByID := make(map[string]*structs.User)
	resultByEmail := make(map[string]*structs.User)

	baseEndpoint := c.getOrganizationEndpoint() + "/users"

	err := c.fetchAllWithPagination(ctx, baseEndpoint, func(resp []byte) (int, error) {
		var usersResp AstroUsersResponse
		if err := json.Unmarshal(resp, &usersResp); err != nil {
			return 0, fmt.Errorf("failed to parse users response: %w", err)
		}

		for _, user := range usersResp.Users {
			structUser := astroUserToStruct(user)
			resultByID[structUser.ID] = structUser
			if structUser.Email != "" {
				resultByEmail[structUser.Email] = structUser
			}
		}
		return len(usersResp.Users), nil
	})

	if err != nil {
		log.WithError(err).Error("error fetching list of users")
		return nil, nil, err
	}

	log.WithField("total_user_count", len(resultByID)).Info("found users")
	return resultByID, resultByEmail, nil
}

// FetchUserDetails fetches details for a specific user using REST API
func (c *AstroClient) FetchUserDetails(ctx context.Context, userID string) (*structs.User, error) {
	log := logger.Logger(ctx).WithFields(logrus.Fields{
		"service": "astro",
		"userID":  userID,
	})
	log.Info("fetching user details by ID")

	endpoint := fmt.Sprintf("%s/users/%s", c.getOrganizationEndpoint(), userID)
	resp, status, err := c.makeRequest(ctx, endpoint, http.MethodGet, nil)
	if err != nil {
		return nil, err
	}

	if status == http.StatusNotFound {
		return nil, fmt.Errorf("user with ID %s not found", userID)
	}

	if status != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch user %s: status %d, body: %s", userID, status, string(resp))
	}

	var user AstroUser
	if err := json.Unmarshal(resp, &user); err != nil {
		return nil, fmt.Errorf("failed to parse user response: %w", err)
	}

	log.Info("successfully fetched user details")
	return astroUserToStruct(user), nil
}

// CreateUser creates a new user in Astro by sending an invitation
// In Astro, users are created via invitations
func (c *AstroClient) CreateUser(ctx context.Context, user *structs.User) (*structs.User, error) {
	log := logger.Logger(ctx).WithFields(logrus.Fields{
		"service": "astro",
		"user":    user.Email,
	})

	log.Info("creating user (sending invitation)")
	endpoint := c.getOrganizationEndpoint() + "/invites"

	if user.Email == "" {
		return nil, fmt.Errorf("email is required for Astro user creation")
	}

	payload := CreateInviteRequest{
		InviteeEmail: user.Email,
		Role:         DefaultOrganizationRole,
	}

	resp, status, err := c.makeRequest(ctx, endpoint, http.MethodPost, payload)
	if err != nil {
		log.WithError(err).Error("error creating user invitation")
		return nil, err
	}

	// Handle conflict - user might already exist
	if status == http.StatusConflict {
		log.Info("user already exists, fetching existing user")
		// Try to find the user in the existing users list
		_, usersByEmail, fetchErr := c.FetchAllUsers(ctx)
		if fetchErr != nil {
			return nil, fetchErr
		}
		if existingUser, ok := usersByEmail[strings.ToLower(user.Email)]; ok {
			return existingUser, nil
		}
		return nil, fmt.Errorf("user conflict but couldn't find existing user: %s", user.Email)
	}

	if status != http.StatusOK && status != http.StatusCreated {
		return nil, fmt.Errorf("failed to create user invitation, status: %s, body: %s",
			http.StatusText(status), string(resp))
	}

	var inviteResp CreateInviteResponse
	if err := json.Unmarshal(resp, &inviteResp); err != nil {
		return nil, fmt.Errorf("failed to parse invite response: %w", err)
	}

	log.WithField("userID", inviteResp.UserID).Info("user invitation sent successfully")

	// Return the created user with the ID from the response
	return &structs.User{
		ID:       inviteResp.UserID,
		Email:    user.Email,
		UserName: user.Email,
		Role:     DefaultOrganizationRole,
	}, nil
}

// DeleteUser removes a user from Astro by setting their organization role to null
// Astro doesn't have a direct delete endpoint; instead, we remove the org-level role
func (c *AstroClient) DeleteUser(ctx context.Context, userID string) error {
	log := logger.Logger(ctx).WithFields(logrus.Fields{
		"service": "astro",
		"userID":  userID,
	})

	log.Info("deleting user (removing organization role)")
	endpoint := fmt.Sprintf("%s/users/%s/roles", c.getOrganizationEndpoint(), userID)

	// Set organizationRole to null to remove the user
	payload := UpdateUserRoleRequest{
		OrganizationRole: nil,
	}

	resp, status, err := c.makeRequest(ctx, endpoint, http.MethodPost, payload)
	if err != nil {
		log.WithError(err).Error("error deleting user")
		return fmt.Errorf("failed to delete user: %w", err)
	}

	if status != http.StatusOK && status != http.StatusNoContent {
		return fmt.Errorf("failed to delete user, status: %s, body: %s",
			http.StatusText(status), string(resp))
	}

	log.Info("user deleted successfully")
	return nil
}
