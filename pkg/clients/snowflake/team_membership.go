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
)

// FetchTeamMembersByTeamID fetches team members for a given team ID using the correct REST API endpoint
func (c *SnowflakeClient) FetchTeamMembersByTeamID(ctx context.Context,
	teamID string) (map[string]*structs.User, error) {
	// Use the correct endpoint: grants-of (not grants-on)
	endpoint := fmt.Sprintf("/api/v2/roles/%s/grants-of", teamID)

	response, _, status, err := c.sendRequest(ctx, endpoint, http.MethodGet, nil)
	if err != nil {
		return nil, fmt.Errorf("error making request to fetch team members: %w", err)
	}
	if status != http.StatusOK {
		return nil,
			fmt.Errorf("failed to fetch team members, status: %s, body: %s", http.StatusText(status), string(response))
	}

	var grants []map[string]interface{}
	if err := json.Unmarshal(response, &grants); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	members := make(map[string]*structs.User)

	for _, grant := range grants {
		// Check if this grant is for a USER (not ROLE)
		if grantedTo, ok := grant["granted_to"].(string); ok && grantedTo == "USER" {
			if granteeName, ok := grant["grantee_name"].(string); ok {
				members[strings.ToLower(granteeName)] = &structs.User{
					ID:       strings.ToLower(granteeName),
					UserName: strings.ToLower(granteeName),
					Email:    strings.ToLower(granteeName), // Snowflake typically uses usernames
				}
			}
		}
	}

	return members, nil
}

// AddUserToTeam adds a user to a team (grants role to user)
func (c *SnowflakeClient) AddUserToTeam(ctx context.Context, teamID, userID string) error {
	endpoint := fmt.Sprintf("/api/v2/users/%s/grants", userID)

	// Create payload for granting role to user
	payload := map[string]interface{}{
		"securable": map[string]string{
			"name": teamID,
		},
		"containing_scope": map[string]string{
			"database": "DEFAULT",
		},
		"securable_type": "ROLE",
		"privileges":     []string{},
	}

	resp, _, status, err := c.sendRequest(ctx, endpoint, http.MethodPost, payload)
	if err != nil {
		return err
	}

	// Check for successful grant
	if status != http.StatusOK && status != http.StatusCreated {
		return fmt.Errorf("failed to add user to team, status: %s, body: %s", http.StatusText(status), string(resp))
	}

	return nil
}

// RemoveUserFromTeam removes a user from a team (revokes role from user)
func (c *SnowflakeClient) RemoveUserFromTeam(ctx context.Context, teamID, userID string) error {
	endpoint := fmt.Sprintf("/api/v2/users/%s/grants:revoke", userID)

	// Create payload for revoking role from user
	payload := map[string]interface{}{
		"securable": map[string]string{
			"name": teamID,
		},
		"containing_scope": map[string]string{
			"database": "DEFAULT",
		},
		"securable_type": "ROLE",
		"privileges":     []string{},
	}

	resp, _, status, err := c.sendRequest(ctx, endpoint, http.MethodPost, payload)
	if err != nil {
		return err
	}

	// Check for successful revocation
	if status != http.StatusOK && status != http.StatusNoContent {
		return fmt.Errorf("failed to remove user from team, status: %s, body: %s", http.StatusText(status), string(resp))
	}

	return nil
}
