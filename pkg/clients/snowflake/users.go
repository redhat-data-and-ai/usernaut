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

// FetchAllUsers fetches all users from Snowflake using REST API with proper pagination
// Snowflake pagination works as follows:
// 1. First call /api/v2/users - returns first page + Link header with result ID
// 2. Subsequent calls /api/v2/results/{result_id}?page=N - returns additional pages
// Returns 2 maps: 1st map keyed by ID, 2nd map keyed by email
func (c *SnowflakeClient) FetchAllUsers(ctx context.Context) (map[string]*structs.User,
	map[string]*structs.User, error) {
	log := logger.Logger(ctx).WithField("service", "snowflake")

	log.Info("fetching all users")
	resultByID := make(map[string]*structs.User)
	resultByEmail := make(map[string]*structs.User)

	// Use the generic pagination helper with a closure that processes user pages
	err := c.fetchAllWithPagination(ctx, "/api/v2/users", func(resp []byte) error {
		return c.processUsersPage(resp, resultByID, resultByEmail)
	})
	if err != nil {
		log.WithError(err).Error("error fetching list of users")
		return nil, nil, err
	}

	log.WithFields(logrus.Fields{
		"total_user_count": len(resultByID),
	}).Info("found users")

	return resultByID, resultByEmail, nil
}

func (c *SnowflakeClient) processUsersPage(resp []byte, resultByID map[string]*structs.User,
	resultByEmail map[string]*structs.User) error {
	var users []SnowflakeUser
	if err := json.Unmarshal(resp, &users); err != nil {
		return fmt.Errorf("failed to parse users response: %w", err)
	}

	for _, user := range users {
		structUser := &structs.User{
			ID:          strings.ToLower(user.Name),
			UserName:    strings.ToLower(user.Name),
			Email:       strings.ToLower(user.Email),
			DisplayName: user.DisplayName,
		}

		resultByID[strings.ToLower(user.Name)] = structUser
		if user.Email != "" {
			resultByEmail[strings.ToLower(user.Email)] = structUser
		}
	}

	return nil
}

// CreateUser creates a new user in Snowflake using REST API
func (c *SnowflakeClient) CreateUser(ctx context.Context, user *structs.User) (*structs.User, error) {
	log := logger.Logger(ctx).WithFields(logrus.Fields{
		"service": "snowflake",
		"user":    user,
	})

	log.Info("creating user")
	endpoint := "/api/v2/users"

	if user.Email == "" || user.UserName == "" {
		return nil, fmt.Errorf("email and username are required for Snowflake user creation")
	}

	payload := map[string]interface{}{
		"name":  user.UserName,
		"email": user.Email,
	}

	if user.DisplayName != "" {
		payload["displayName"] = user.DisplayName
	}

	resp, _, status, err := c.makeRequestWithPolling(ctx, endpoint, http.MethodPost, payload)
	if err != nil {
		log.WithError(err).Error("error creating user")
		return nil, err
	}

	if status == http.StatusConflict {
		log.WithField("status", status).Info("user already exists, fetching user details")
		return c.FetchUserDetails(ctx, user.UserName)
	}

	if status != http.StatusOK && status != http.StatusCreated {
		return nil, fmt.Errorf("failed to create user, status: %s, body: %s", http.StatusText(status), string(resp))
	}

	var createdUserResponse SnowflakeUser
	if err := json.Unmarshal(resp, &createdUserResponse); err != nil {
		return nil, fmt.Errorf("failed to parse create user response: %w", err)
	}

	return &structs.User{
		ID:          strings.ToLower(createdUserResponse.Name),
		UserName:    strings.ToLower(createdUserResponse.Name),
		Email:       strings.ToLower(createdUserResponse.Email),
		DisplayName: createdUserResponse.DisplayName,
	}, nil
}

// FetchUserDetails fetches details for a specific user using REST API
func (c *SnowflakeClient) FetchUserDetails(ctx context.Context, userID string) (*structs.User, error) {
	log := logger.Logger(ctx).WithFields(logrus.Fields{
		"service": "snowflake",
		"userID":  userID,
	})
	log.Info("fetching user details by ID")

	endpoint := fmt.Sprintf("/api/v2/users/%s", userID)
	resp, _, status, err := c.makeRequestWithPolling(ctx, endpoint, http.MethodGet, nil)
	if err != nil {
		log.WithError(err).Error("error fetching user details")
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch user details, status: %s, body: %s", http.StatusText(status), string(resp))
	}

	var userResponse SnowflakeUser
	if err := json.Unmarshal(resp, &userResponse); err != nil {
		return nil, fmt.Errorf("failed to parse user response: %w", err)
	}

	log.Info("found user details")
	user := &structs.User{
		ID:          strings.ToLower(userID),
		UserName:    strings.ToLower(userID),
		Email:       strings.ToLower(userResponse.Email),
		DisplayName: userResponse.DisplayName,
	}

	return user, nil
}

// DeleteUser deletes a user from Snowflake using REST API
func (c *SnowflakeClient) DeleteUser(ctx context.Context, userID string) error {
	log := logger.Logger(ctx).WithFields(logrus.Fields{
		"service": "snowflake",
		"userID":  userID,
	})

	log.Debug("deleting user")
	endpoint := fmt.Sprintf("/api/v2/users/%s", userID)

	resp, _, status, err := c.makeRequestWithPolling(ctx, endpoint, http.MethodDelete, nil)
	if err != nil {
		log.WithError(err).Error("error deleting user")
		return fmt.Errorf("failed to delete user: %w", err)
	}

	if status != http.StatusOK && status != http.StatusNoContent {
		return fmt.Errorf("failed to delete user, status: %s, body: %s", http.StatusText(status), string(resp))
	}

	log.WithFields(logrus.Fields{
		"userID": userID,
		"status": status,
	}).Info("user deleted successfully")
	return nil
}
