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

	// Parse response to get created role details
	var response map[string]interface{}
	if err := json.Unmarshal(resp, &response); err != nil {
		return nil, fmt.Errorf("failed to parse create role response: %w", err)
	}

	// Return the created team
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
