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
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/redhat-data-and-ai/usernaut/pkg/common/structs"
	"github.com/redhat-data-and-ai/usernaut/pkg/logger"
	"github.com/sirupsen/logrus"
)

// FetchAllUsers retrieves all users via SCIM and returns maps keyed by email and SCIM ID
func (pc *PresetClient) FetchAllUsers(ctx context.Context) (map[string]*structs.User, map[string]*structs.User, error) {
	log := logger.Logger(ctx).WithField("service", "preset")
	log.Info("fetching all SCIM users from Preset")

	userEmailMap := make(map[string]*structs.User)
	userIDMap := make(map[string]*structs.User)

	startIndex := 1
	count := 100

	for {
		reqURL := fmt.Sprintf("%s/Users?startIndex=%d&count=%d", pc.scimURL(), startIndex, count)
		response, _, err := pc.sendRequest(ctx, reqURL, http.MethodGet, nil)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to fetch SCIM users from Preset: %w", err)
		}

		var scimResp scimUsersResponse
		if err := json.Unmarshal(response, &scimResp); err != nil {
			return nil, nil, fmt.Errorf("failed to parse SCIM users response: %w", err)
		}

		if len(scimResp.Resources) == 0 {
			break
		}

		for _, su := range scimResp.Resources {
			user := scimUserToStruct(&su)
			userIDMap[user.ID] = user
			if user.Email != "" {
				userEmailMap[user.Email] = user
			}
		}

		if startIndex+count > scimResp.TotalResults {
			break
		}
		startIndex += count
	}

	log.WithField("total_user_count", len(userIDMap)).Info("successfully fetched SCIM users from Preset")
	return userEmailMap, userIDMap, nil
}

// FetchUserDetails retrieves details of a specific user by SCIM ID
func (pc *PresetClient) FetchUserDetails(ctx context.Context, userID string) (*structs.User, error) {
	log := logger.Logger(ctx).WithFields(logrus.Fields{
		"service": "preset",
		"userID":  userID,
	})
	log.Info("fetching user details from Preset")

	reqURL := fmt.Sprintf("%s/Users/%s", pc.scimURL(), userID)
	response, _, err := pc.sendRequest(ctx, reqURL, http.MethodGet, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user details from Preset: %w", err)
	}

	var su scimUser
	if err := json.Unmarshal(response, &su); err != nil {
		return nil, fmt.Errorf("failed to parse user details response: %w", err)
	}

	log.Info("successfully fetched user details from Preset")
	return scimUserToStruct(&su), nil
}

// CreateUser provisions a user via SCIM. If the user already exists, returns their existing details.
func (pc *PresetClient) CreateUser(ctx context.Context, u *structs.User) (*structs.User, error) {
	log := logger.Logger(ctx).WithFields(logrus.Fields{
		"service":  "preset",
		"username": u.UserName,
		"email":    u.Email,
	})
	log.Info("creating SCIM user in Preset")

	existing, err := pc.findUserByEmail(ctx, u.Email)
	if err == nil && existing != nil {
		log.WithField("user_id", existing.ID).Info("SCIM user already exists in Preset")
		return existing, nil
	}

	reqURL := fmt.Sprintf("%s/Users", pc.scimURL())
	reqBody := scimUserCreateRequest{
		Schemas:  []string{"urn:ietf:params:scim:schemas:core:2.0:User"},
		UserName: u.Email,
		Emails: []scimEmailValue{
			{Value: u.Email, Primary: true},
		},
		Name: scimName{
			GivenName:  u.FirstName,
			FamilyName: u.LastName,
		},
		Active: true,
	}

	response, statusCode, err := pc.sendRequest(ctx, reqURL, http.MethodPost, reqBody)
	if err != nil {
		if statusCode == http.StatusConflict {
			log.Info("SCIM user already exists (conflict), fetching details")
			existing, findErr := pc.findUserByEmail(ctx, u.Email)
			if findErr == nil && existing != nil {
				return existing, nil
			}
		}
		return nil, fmt.Errorf("failed to create SCIM user in Preset: %w", err)
	}

	var createdUser scimUser
	if err := json.Unmarshal(response, &createdUser); err != nil {
		log.Warn("failed to parse SCIM user creation response, looking up by email")
		existing, findErr := pc.findUserByEmail(ctx, u.Email)
		if findErr == nil && existing != nil {
			return existing, nil
		}
		return &structs.User{
			ID:       u.Email,
			Email:    u.Email,
			UserName: u.UserName,
		}, nil
	}

	log.WithField("user_id", createdUser.ID).Info("successfully created SCIM user in Preset")
	return scimUserToStruct(&createdUser), nil
}

// DeleteUser removes a user via SCIM
func (pc *PresetClient) DeleteUser(ctx context.Context, userID string) error {
	log := logger.Logger(ctx).WithFields(logrus.Fields{
		"service": "preset",
		"userID":  userID,
	})
	log.Info("deleting SCIM user from Preset")

	reqURL := fmt.Sprintf("%s/Users/%s", pc.scimURL(), userID)
	_, _, err := pc.sendRequest(ctx, reqURL, http.MethodDelete, nil)
	if err != nil {
		if strings.Contains(err.Error(), "404") {
			log.Info("SCIM user does not exist in Preset, nothing to delete")
			return nil
		}
		return fmt.Errorf("failed to delete SCIM user from Preset: %w", err)
	}

	log.Info("successfully deleted SCIM user from Preset")
	return nil
}

// findUserByEmail searches for a user by email using SCIM filter
func (pc *PresetClient) findUserByEmail(ctx context.Context, email string) (*structs.User, error) {
	filter := fmt.Sprintf(`userName eq "%s"`, email)
	reqURL := fmt.Sprintf("%s/Users?filter=%s", pc.scimURL(), url.QueryEscape(filter))
	response, _, err := pc.sendRequest(ctx, reqURL, http.MethodGet, nil)
	if err != nil {
		return nil, err
	}

	var scimResp scimUsersResponse
	if err := json.Unmarshal(response, &scimResp); err != nil {
		return nil, err
	}

	if len(scimResp.Resources) == 0 {
		return nil, fmt.Errorf("user not found: %s", email)
	}

	return scimUserToStruct(&scimResp.Resources[0]), nil
}

func scimUserToStruct(su *scimUser) *structs.User {
	email := ""
	if len(su.Emails) > 0 {
		for _, e := range su.Emails {
			if e.Primary {
				email = e.Value
				break
			}
		}
		if email == "" {
			email = su.Emails[0].Value
		}
	}

	return &structs.User{
		ID:          su.ID,
		Email:       email,
		UserName:    su.UserName,
		DisplayName: su.DisplayName,
	}
}
