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

func (ac *AtlanClient) FetchTeamMembersByTeamID(ctx context.Context, teamID string) (map[string]*structs.User, error) {
	log := logger.Logger(ctx).WithFields(logrus.Fields{
		"service": "atlan",
		"teamID":  teamID,
	})

	if ac.ldapSync {
		log.Info("ldap sync enabled, skipping fetch team members - membership managed via SSO")
		return make(map[string]*structs.User), nil
	}

	log.Info("fetching team members from Atlan")

	teamMembers := make(map[string]*structs.User)

	url := fmt.Sprintf("%s/api/service/groups/%s/members", ac.url, teamID)
	response, statusCode, err := ac.sendRequest(ctx, url, http.MethodGet, nil, nil, "FetchTeamMembersByTeamID")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch team members from Atlan: %w", err)
	}

	if statusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d when fetching team members from Atlan", statusCode)
	}

	var apiResponse AtlanGroupMembersResponse
	if err := json.Unmarshal(response, &apiResponse); err != nil {
		return nil, fmt.Errorf("failed to parse team members response from Atlan: %w", err)
	}

	for _, member := range apiResponse.Records {
		teamMembers[member.ID] = &structs.User{
			ID:       member.ID,
			Email:    member.Email,
			UserName: member.Username,
		}
	}

	log.WithField("member_count", len(teamMembers)).Info("fetched team members from Atlan")
	return teamMembers, nil
}

func (ac *AtlanClient) AddUserToTeam(ctx context.Context, teamID string, userIDs []string) error {
	log := logger.Logger(ctx).WithFields(logrus.Fields{
		"service": "atlan",
		"teamID":  teamID,
		"userIDs": userIDs,
	})

	if ac.ldapSync || len(userIDs) == 0 {
		if ac.ldapSync {
			log.Info("ldap sync enabled, skipping add users - membership managed via SSO")
		}
		return nil
	}

	log.Info("adding users to team in Atlan")

	for _, userID := range userIDs {
		url := fmt.Sprintf("%s/api/service/users/%s/groups", ac.url, userID)
		requestBody := map[string]interface{}{
			"groups": []string{teamID},
		}

		_, statusCode, err := ac.sendRequest(ctx, url, http.MethodPost, requestBody, nil, "AddUserToTeam")
		if err != nil {
			return fmt.Errorf("failed to add user %s to team in Atlan: %w", userID, err)
		}

		if statusCode != http.StatusOK && statusCode != http.StatusCreated && statusCode != http.StatusNoContent {
			return fmt.Errorf("unexpected status code %d when adding user to team in Atlan", statusCode)
		}
	}

	log.Info("added users to team in Atlan")
	return nil
}

func (ac *AtlanClient) RemoveUserFromTeam(ctx context.Context, teamID string, userIDs []string) error {
	log := logger.Logger(ctx).WithFields(logrus.Fields{
		"service": "atlan",
		"teamID":  teamID,
		"userIDs": userIDs,
	})

	if ac.ldapSync || len(userIDs) == 0 {
		if ac.ldapSync {
			log.Info("ldap sync enabled, skipping remove users - membership managed via SSO")
		}
		return nil
	}

	log.Info("removing users from team in Atlan")

	url := fmt.Sprintf("%s/api/service/groups/%s/members/remove", ac.url, teamID)
	requestBody := map[string]interface{}{
		"users": userIDs,
	}

	_, statusCode, err := ac.sendRequest(ctx, url, http.MethodPost, requestBody, nil, "RemoveUserFromTeam")
	if err != nil {
		return fmt.Errorf("failed to remove users from team in Atlan: %w", err)
	}

	if statusCode != http.StatusOK && statusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status code %d when removing users from team in Atlan", statusCode)
	}

	log.Info("removed users from team in Atlan")
	return nil
}
