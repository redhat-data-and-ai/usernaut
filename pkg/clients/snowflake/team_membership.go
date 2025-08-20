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

// FetchTeamMembersByTeamID fetches team members for a given team ID using the correct REST API endpoint
func (c *SnowflakeClient) FetchTeamMembersByTeamID(ctx context.Context,
	teamID string) (map[string]*structs.User, error) {
	log := logger.Logger(ctx).WithFields(logrus.Fields{
		"service": "snowflake",
		"teamID":  teamID,
	})
	log.Info("fetching team members by team ID")

	// Use the correct endpoint: grants-of (not grants-on)
	endpoint := fmt.Sprintf("/api/v2/roles/%s/grants-of", teamID)

	response, status, err := c.makeRequest(ctx, endpoint, http.MethodGet, nil)
	if err != nil {
		log.WithError(err).Error("error fetching team members by team ID")
		return nil, fmt.Errorf("error making request to fetch team members: %w", err)
	}
	if status != http.StatusOK {
		return nil,
			fmt.Errorf("failed to fetch team members, status: %s, body: %s", http.StatusText(status), string(response))
	}

	var grants []SnowflakeGrant
	if err := json.Unmarshal(response, &grants); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	members := make(map[string]*structs.User)

	for _, grant := range grants {
		// Check if this grant is for a USER (not ROLE)
		if grant.GrantedTo == "USER" && grant.GranteeName != "" {
			members[strings.ToLower(grant.GranteeName)] = &structs.User{
				ID:       strings.ToLower(grant.GranteeName),
				UserName: strings.ToLower(grant.GranteeName),
				Email:    "", // Email not available from grants API
			}
		}
	}

	return members, nil
}

// AddUserToTeam adds users to a team (grants role to users)
func (c *SnowflakeClient) AddUserToTeam(ctx context.Context, teamID string, userIDs []string) error {
	log := logger.Logger(ctx).WithFields(logrus.Fields{
		"service":    "snowflake",
		"teamID":     teamID,
		"user_count": len(userIDs),
	})
	log.Info("adding users to team")

	for _, userID := range userIDs {
		endpoint := fmt.Sprintf("/api/v2/users/%s/grants", userID)

		resp, status, err := c.makeRoleRequest(ctx, teamID, endpoint)
		if err != nil {
			return fmt.Errorf("failed to add user %s to team %s: %w", userID, teamID, err)
		}

		if status != http.StatusOK && status != http.StatusCreated {
			return fmt.Errorf("failed to add user %s to team %s, status: %s, body: %s",
				userID, teamID, http.StatusText(status), string(resp))
		}
	}

	return nil
}

// RemoveUserFromTeam removes users from a team (revokes role from users)
func (c *SnowflakeClient) RemoveUserFromTeam(ctx context.Context, teamID string, userIDs []string) error {
	log := logger.Logger(ctx).WithFields(logrus.Fields{
		"service":    "snowflake",
		"teamID":     teamID,
		"user_count": len(userIDs),
	})
	log.Info("removing users from team")

	for _, userID := range userIDs {
		endpoint := fmt.Sprintf("/api/v2/users/%s/grants:revoke", userID)

		resp, status, err := c.makeRoleRequest(ctx, teamID, endpoint)
		if err != nil {
			return fmt.Errorf("failed to remove user %s from team %s: %w", userID, teamID, err)
		}

		if status != http.StatusOK && status != http.StatusNoContent {
			return fmt.Errorf("failed to remove user %s from team %s, status: %s, body: %s",
				userID, teamID, http.StatusText(status), string(resp))
		}
	}

	return nil
}

// makeRoleRequest sends a role grant/revoke request for a user
func (c *SnowflakeClient) makeRoleRequest(ctx context.Context, teamID, endpoint string) ([]byte, int, error) {
	payload := map[string]interface{}{
		"securable": map[string]string{
			"name": teamID,
		},
		"securable_type": "ROLE",
		"privileges":     []string{},
	}

	return c.makeRequest(ctx, endpoint, http.MethodPost, payload)
}
