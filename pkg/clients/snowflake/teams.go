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

	"github.com/redhat-data-and-ai/usernaut/pkg/common/structs"
)

// FetchAllTeams fetches all roles from Snowflake using REST API
func (c *SnowflakeClient) FetchAllTeams(ctx context.Context) (map[string]structs.Team, error) {
	endpoint := "/api/v2/roles"
	resp, _, status, err := c.sendRequest(ctx, endpoint, http.MethodGet, nil)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch roles, status: %s, body: %s", http.StatusText(status), string(resp))
	}

	// Parse the response - expecting an array of roles
	var roles []map[string]interface{}
	if err := json.Unmarshal(resp, &roles); err != nil {
		return nil, fmt.Errorf("failed to parse roles response: %w", err)
	}

	// Extract roles from the response
	teams := make(map[string]structs.Team)
	for _, roleMap := range roles {
		if name, ok := roleMap["name"].(string); ok {
			team := structs.Team{
				ID:   name,
				Name: name,
			}
			teams[name] = team
		}
	}

	return teams, nil
}

// CreateTeam creates a new role in Snowflake using REST API
func (c *SnowflakeClient) CreateTeam(ctx context.Context, team *structs.Team) (*structs.Team, error) {
	endpoint := "/api/v2/roles"

	// Create payload for role creation
	payload := map[string]interface{}{
		"name": team.Name,
	}

	resp, _, status, err := c.sendRequest(ctx, endpoint, http.MethodPost, payload)
	if err != nil {
		return nil, err
	}

	// Check for successful creation
	if status != http.StatusOK && status != http.StatusCreated {
		return nil, fmt.Errorf("failed to create role, status: %s, body: %s", http.StatusText(status), string(resp))
	}

	// Return the created team using the request data since Snowflake API
	// returns minimal information in create response
	createdTeam := &structs.Team{
		ID:   team.Name,
		Name: team.Name,
	}

	return createdTeam, nil
}

// FetchTeamDetails fetches details for a specific team/role by filtering from all teams
func (c *SnowflakeClient) FetchTeamDetails(ctx context.Context, teamID string) (*structs.Team, error) {
	// Since REST API doesn't support GET /api/v2/roles/{name}, we fetch all teams and filter
	teams, err := c.FetchAllTeams(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch teams to get team details: %w", err)
	}

	// Find the specific team
	if team, exists := teams[teamID]; exists {
		return &team, nil
	}

	return nil, fmt.Errorf("team with ID %s not found", teamID)
}

// DeleteTeamByID deletes a role in Snowflake using REST API
func (c *SnowflakeClient) DeleteTeamByID(ctx context.Context, teamID string) error {
	endpoint := fmt.Sprintf("/api/v2/roles/%s", teamID)

	resp, _, status, err := c.sendRequest(ctx, endpoint, http.MethodDelete, nil)
	if err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}

	// Check for successful deletion
	if status != http.StatusOK && status != http.StatusNoContent {
		return fmt.Errorf("failed to delete role, status: %s, body: %s", http.StatusText(status), string(resp))
	}

	return nil
}
