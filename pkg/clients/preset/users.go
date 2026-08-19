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

	"github.com/redhat-data-and-ai/usernaut/pkg/common/structs"
	"github.com/redhat-data-and-ai/usernaut/pkg/logger"
	"github.com/sirupsen/logrus"
)

// FetchAllUsers retrieves all users via SCIM and returns maps keyed by email and SCIM ID
func (pc *PresetClient) FetchAllUsers(ctx context.Context) (map[string]*structs.User, map[string]*structs.User, error) {
	log := logger.Logger(ctx).WithFields(logrus.Fields{
		"service": "preset",
	})
	log.Info("fetching all SCIM users from Preset")

	userEmailMap := make(map[string]*structs.User)
	userIDMap := make(map[string]*structs.User)

	for startIndex := 1; ; startIndex += scimPageSize {
		reqURL := fmt.Sprintf("%s/Users?startIndex=%d&count=%d", pc.scimURL(), startIndex, scimPageSize)
		response, _, err := pc.sendRequest(ctx, reqURL, http.MethodGet, nil)
		if err != nil {
			log.WithError(err).Error("failed to fetch SCIM users from Preset")
			return nil, nil, fmt.Errorf("failed to fetch SCIM users from Preset: %w", err)
		}

		var scimResp scimUsersResponse
		if err := json.Unmarshal(response, &scimResp); err != nil {
			log.WithError(err).Error("failed to parse SCIM users response")
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

		if startIndex+scimPageSize > scimResp.TotalResults {
			break
		}
	}

	log.WithFields(logrus.Fields{
		"total_user_count": len(userIDMap),
	}).Info("successfully fetched SCIM users from Preset")
	return userEmailMap, userIDMap, nil
}

// FetchUserDetails retrieves details of a specific user by SCIM ID
func (pc *PresetClient) FetchUserDetails(ctx context.Context, userID string) (*structs.User, error) {
	log := logger.Logger(ctx).WithFields(logrus.Fields{
		"service": "preset",
		"userID":  userID,
	})
	log.Info("fetching user details from Preset")

	user, err := pc.fetchSCIMUser(ctx, userID)
	if err != nil {
		log.WithError(err).Error("failed to fetch user details from Preset")
		return nil, fmt.Errorf("failed to fetch user details from Preset: %w", err)
	}

	log.Info("successfully fetched user details from Preset")
	return scimUserToStruct(user), nil
}

// CreateUser provisions a user via SCIM. If the user already exists, returns their existing details.
func (pc *PresetClient) CreateUser(ctx context.Context, u *structs.User) (*structs.User, error) {
	log := logger.Logger(ctx).WithFields(logrus.Fields{
		"service":  "preset",
		"username": u.UserName,
		"email":    u.Email,
	})
	log.Info("creating SCIM user in Preset")

	if existing, err := pc.findUserByEmail(ctx, u.Email); err == nil {
		log.WithFields(logrus.Fields{
			"user_id": existing.ID,
		}).Info("SCIM user already exists in Preset")
		return existing, nil
	}

	reqURL := fmt.Sprintf("%s/Users", pc.scimURL())
	reqBody := scimUserCreateRequest{
		Schemas:  []string{scimUserSchema},
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
			return pc.requireUserByEmail(ctx, u.Email, "SCIM user conflict but lookup failed")
		}
		log.WithError(err).Error("failed to create SCIM user in Preset")
		return nil, fmt.Errorf("failed to create SCIM user in Preset: %w", err)
	}

	var createdUser scimUser
	if err := json.Unmarshal(response, &createdUser); err != nil {
		log.WithError(err).Warn("failed to parse SCIM user creation response")
		return pc.requireUserByEmail(ctx, u.Email, "failed to parse SCIM user creation response")
	}

	log.WithFields(logrus.Fields{
		"user_id": createdUser.ID,
	}).Info("successfully created SCIM user in Preset")
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
	_, statusCode, err := pc.sendRequest(ctx, reqURL, http.MethodDelete, nil)
	if err != nil {
		if statusCode == http.StatusNotFound {
			log.Info("SCIM user does not exist in Preset, nothing to delete")
			return nil
		}
		log.WithError(err).Error("failed to delete SCIM user from Preset")
		return fmt.Errorf("failed to delete SCIM user from Preset: %w", err)
	}

	log.Info("successfully deleted SCIM user from Preset")
	return nil
}

func (pc *PresetClient) findUserByEmail(ctx context.Context, email string) (*structs.User, error) {
	log := logger.Logger(ctx).WithFields(logrus.Fields{
		"service": "preset",
		"email":   email,
	})

	filter := fmt.Sprintf(`userName eq "%s"`, email)
	response, err := pc.querySCIMByFilter(ctx, "Users", filter, "email", email)
	if err != nil {
		return nil, err
	}

	var scimResp scimUsersResponse
	if err := json.Unmarshal(response, &scimResp); err != nil {
		log.WithError(err).Error("failed to parse SCIM user lookup response")
		return nil, err
	}

	if len(scimResp.Resources) == 0 {
		log.Warn("SCIM user not found by email")
		return nil, fmt.Errorf("user not found: %s", email)
	}

	return scimUserToStruct(&scimResp.Resources[0]), nil
}

func (pc *PresetClient) requireUserByEmail(ctx context.Context, email, msg string) (*structs.User, error) {
	user, err := pc.findUserByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", msg, err)
	}
	return user, nil
}
