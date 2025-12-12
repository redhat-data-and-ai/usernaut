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
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/redhat-data-and-ai/usernaut/pkg/common/structs"
	"github.com/redhat-data-and-ai/usernaut/pkg/logger"
	"github.com/sirupsen/logrus"
)

const (
	// paginationLimit is the number of records to fetch per API request
	paginationLimit = 100
)

// FetchAllUsers retrieves all users from Atlan with pagination
// This function fetches users regardless of SSO sync status
func (ac *AtlanClient) FetchAllUsers(ctx context.Context) (map[string]*structs.User, map[string]*structs.User, error) {
	log := logger.Logger(ctx).WithField("service", "atlan")
	log.Info("fetching all users from Atlan")

	userEmailMap := make(map[string]*structs.User)
	userIDMap := make(map[string]*structs.User)

	offset := 0

	for {
		url := fmt.Sprintf("%s/api/service/users?limit=%d&offset=%d", ac.url, paginationLimit, offset)
		response, err := ac.sendRequest(ctx, url, http.MethodGet, nil, "FetchAllUsers")
		if err != nil {
			return nil, nil, fmt.Errorf("failed to fetch users from Atlan: %w", err)
		}

		var apiResponse AtlanUsersResponse
		if err := json.Unmarshal(response, &apiResponse); err != nil {
			return nil, nil, fmt.Errorf("failed to parse users response from Atlan: %w", err)
		}

		for _, user := range apiResponse.Records {
			userStruct := atlanUserToStruct(&user)
			if user.Email != "" {
				userEmailMap[user.Email] = userStruct
			}
			userIDMap[user.ID] = userStruct
		}

		if len(apiResponse.Records) < paginationLimit {
			break
		}
		offset += paginationLimit
	}

	log.WithField("total_user_count", len(userIDMap)).Info("successfully fetched users from Atlan")
	return userEmailMap, userIDMap, nil
}

// FetchUserDetails retrieves details of a specific user by their ID
// This function fetches user details regardless of SSO sync status
func (ac *AtlanClient) FetchUserDetails(ctx context.Context, userID string) (*structs.User, error) {
	log := logger.Logger(ctx).WithFields(logrus.Fields{
		"service": "atlan",
		"userID":  userID,
	})
	log.Info("fetching user details from Atlan")

	url := fmt.Sprintf("%s/api/service/users/%s", ac.url, userID)
	response, err := ac.sendRequest(ctx, url, http.MethodGet, nil, "FetchUserDetails")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user details from Atlan: %w", err)
	}

	var user AtlanUser
	if err := json.Unmarshal(response, &user); err != nil {
		return nil, fmt.Errorf("failed to parse user details response from Atlan: %w", err)
	}

	log.Info("successfully fetched user details from Atlan")
	return atlanUserToStruct(&user), nil
}

// CreateUser creates a new user in Atlan
// If SSO sync is enabled (ssoSync = true), this function skips user creation
// because users are automatically created when they log in via SSO.
// If SSO sync is not enabled, it creates the user manually via the API.
func (ac *AtlanClient) CreateUser(ctx context.Context, u *structs.User) (*structs.User, error) {
	log := logger.Logger(ctx).WithFields(logrus.Fields{
		"service":  "atlan",
		"username": u.UserName,
		"email":    u.Email,
	})
	log.Info("creating user in Atlan")

	// When SSO sync is enabled, users are automatically created when they log in via SSO
	// We skip user creation entirely - the user may or may not exist yet
	if ac.ssoSync {
		log.Info("SSO sync enabled, skipping user creation - user will be created on first SSO login")
		return &structs.User{
			ID:       u.UserName, // Return username as ID placeholder
			UserName: u.UserName,
			Email:    u.Email,
		}, nil
	}

	url := fmt.Sprintf("%s/api/service/users", ac.url)

	requestBody := map[string]interface{}{
		"email":     u.Email,
		"username":  u.UserName,
		"firstName": u.FirstName,
		"lastName":  u.LastName,
		"roleName":  "$guest",
	}

	response, err := ac.sendRequest(ctx, url, http.MethodPost, requestBody, "CreateUser")
	if err != nil {
		return nil, fmt.Errorf("failed to create user in Atlan: %w", err)
	}

	var createdUser AtlanUser
	if err := json.Unmarshal(response, &createdUser); err != nil {
		return nil, fmt.Errorf("failed to parse created user response from Atlan: %w", err)
	}

	log.WithField("user_id", createdUser.ID).Info("successfully created user in Atlan")
	return atlanUserToStruct(&createdUser), nil
}

func (ac *AtlanClient) DeleteUser(ctx context.Context, userID string) error {
	log := logger.Logger(ctx).WithFields(logrus.Fields{
		"service": "atlan",
		"userID":  userID,
	})
	log.Info("deleting user from Atlan")

	if ac.defaultOwnerUserName == "" {
		return fmt.Errorf("default_owner_username is required in atlan config for user deletion")
	}

	if ac.sdkClient == nil {
		return fmt.Errorf("atlan SDK client not initialized")
	}

	// The SDK's RemoveUser expects username, not userID
	// First, fetch the user details to get the username
	userDetails, err := ac.FetchUserDetails(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to fetch user details for deletion: %w", err)
	}

	log.WithField("username", userDetails.UserName).Info("removing user via Atlan SDK")

	// Use the SDK to delete the user - it creates a workflow internally
	_, sdkErr := ac.sdkClient.UserClient.RemoveUser(
		userDetails.UserName,
		ac.defaultOwnerUserName,
		nil, // wfCreatorUserName defaults to transferToUserName
	)
	if sdkErr != nil {
		return fmt.Errorf("failed to delete user from atlan: %w", sdkErr)
	}

	log.Info("successfully initiated user deletion workflow in Atlan")
	return nil
}

// atlanUserToStruct converts an AtlanUser to a structs.User
func atlanUserToStruct(u *AtlanUser) *structs.User {
	displayName := u.DisplayName
	if displayName == "" && (u.FirstName != "" || u.LastName != "") {
		displayName = fmt.Sprintf("%s %s", u.FirstName, u.LastName)
	}

	return &structs.User{
		ID:          u.ID,
		Email:       u.Email,
		UserName:    u.Username,
		FirstName:   u.FirstName,
		LastName:    u.LastName,
		DisplayName: displayName,
	}
}
