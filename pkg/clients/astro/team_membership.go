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

// FetchTeamMembersByTeamID fetches team members for a given team ID with pagination support
func (c *AstroClient) FetchTeamMembersByTeamID(ctx context.Context,
	teamID string) (map[string]*structs.User, error) {
	log := logger.Logger(ctx).WithFields(logrus.Fields{
		"service": "astro",
		"teamID":  teamID,
	})
	log.Info("fetching team members by team ID")

	members := make(map[string]*structs.User)
	baseEndpoint := fmt.Sprintf("%s/teams/%s/members", c.getOrganizationEndpoint(), teamID)

	err := c.fetchAllWithPagination(ctx, baseEndpoint, func(resp []byte) (int, error) {
		var membersResp AstroTeamMembersResponse
		if err := json.Unmarshal(resp, &membersResp); err != nil {
			return 0, fmt.Errorf("failed to parse team members response: %w", err)
		}

		for _, member := range membersResp.TeamMembers {
			// Extract first and last name from full name if available
			firstName := ""
			lastName := ""
			if member.FullName != "" {
				parts := strings.SplitN(member.FullName, " ", 2)
				firstName = parts[0]
				if len(parts) > 1 {
					lastName = parts[1]
				}
			}

			members[member.UserID] = &structs.User{
				ID:          member.UserID,
				UserName:    strings.ToLower(member.Username),
				Email:       strings.ToLower(member.Username), // In Astro, username is the email
				FirstName:   firstName,
				LastName:    lastName,
				DisplayName: member.FullName,
			}
		}
		return len(membersResp.TeamMembers), nil
	})

	if err != nil {
		log.WithError(err).Error("error fetching team members by team ID")
		return nil, err
	}

	log.WithField("member_count", len(members)).Info("found team members")
	return members, nil
}

// AddUserToTeam adds users to a team
func (c *AstroClient) AddUserToTeam(ctx context.Context, teamID string, userIDs []string) error {
	log := logger.Logger(ctx).WithFields(logrus.Fields{
		"service":    "astro",
		"teamID":     teamID,
		"user_count": len(userIDs),
	})
	log.Info("adding users to team")

	if len(userIDs) == 0 {
		log.Info("no users to add")
		return nil
	}

	endpoint := fmt.Sprintf("%s/teams/%s/members", c.getOrganizationEndpoint(), teamID)

	payload := AddTeamMembersRequest{
		MemberIDs: userIDs,
	}

	resp, status, err := c.makeRequest(ctx, endpoint, http.MethodPost, payload)
	if err != nil {
		log.WithError(err).Error("error adding users to team")
		return fmt.Errorf("failed to add users to team %s: %w", teamID, err)
	}

	if status != http.StatusOK && status != http.StatusCreated && status != http.StatusNoContent {
		return fmt.Errorf("failed to add users to team %s, status: %s, body: %s",
			teamID, http.StatusText(status), string(resp))
	}

	log.Info("users added to team successfully")
	return nil
}

// RemoveUserFromTeam removes users from a team
func (c *AstroClient) RemoveUserFromTeam(ctx context.Context, teamID string, userIDs []string) error {
	log := logger.Logger(ctx).WithFields(logrus.Fields{
		"service":    "astro",
		"teamID":     teamID,
		"user_count": len(userIDs),
	})
	log.Info("removing users from team")

	if len(userIDs) == 0 {
		log.Info("no users to remove")
		return nil
	}

	// Astro requires removing users one at a time
	for _, userID := range userIDs {
		endpoint := fmt.Sprintf("%s/teams/%s/members/%s", c.getOrganizationEndpoint(), teamID, userID)

		resp, status, err := c.makeRequest(ctx, endpoint, http.MethodDelete, nil)
		if err != nil {
			log.WithError(err).WithField("userID", userID).Error("error removing user from team")
			return fmt.Errorf("failed to remove user %s from team %s: %w", userID, teamID, err)
		}

		// 404 is acceptable - user might already be removed
		if status != http.StatusOK && status != http.StatusNoContent && status != http.StatusNotFound {
			return fmt.Errorf("failed to remove user %s from team %s, status: %s, body: %s",
				userID, teamID, http.StatusText(status), string(resp))
		}

		log.WithField("userID", userID).Debug("user removed from team")
	}

	log.Info("users removed from team successfully")
	return nil
}

// ReconcileGroupParams reconciles backend-specific parameters for a group/team.
// For Astro, this could be used to manage workspace/deployment role assignments in the future.
func (c *AstroClient) ReconcileGroupParams(
	ctx context.Context, teamID string, groupParams structs.TeamParams) error {
	log := logger.Logger(ctx).WithFields(logrus.Fields{
		"service": "astro",
		"teamID":  teamID,
	})

	log.Debug("ReconcileGroupParams called - no implementation needed for Astro at this time")
	return nil
}
