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

// FetchAllTeams fetches all roles from Snowflake using REST API with proper pagination
func (c *SnowflakeClient) FetchAllTeams(ctx context.Context) (map[string]structs.Team, error) {
	log := logger.Logger(ctx).WithField("service", "snowflake")

	log.Info("fetching all teams")
	teams := make(map[string]structs.Team)

	// Use the generic pagination helper with a closure that processes team pages
	err := c.fetchAllWithPagination(ctx, "/api/v2/roles", func(resp []byte) error {
		return c.processTeamsPage(resp, teams)
	})
	if err != nil {
		log.WithError(err).Error("error fetching list of teams")
		return nil, err
	}

	log.WithField("total_teams_count", len(teams)).Info("found teams")
	return teams, nil
}

func (c *SnowflakeClient) processTeamsPage(resp []byte, teams map[string]structs.Team) error {
	var roles []SnowflakeRole
	if err := json.Unmarshal(resp, &roles); err != nil {
		return fmt.Errorf("failed to parse roles response: %w", err)
	}

	for _, role := range roles {
		team := structs.Team{
			ID:   strings.ToLower(role.Name),
			Name: strings.ToLower(role.Name),
		}
		teams[strings.ToLower(role.Name)] = team
	}

	return nil
}

// CreateTeam creates a new role in Snowflake using REST API
func (c *SnowflakeClient) CreateTeam(ctx context.Context, team *structs.Team) (*structs.Team, error) {
	log := logger.Logger(ctx).WithField("service", "snowflake")

	log.Info("creating team")
	endpoint := "/api/v2/roles"

	payload := map[string]interface{}{
		"name": team.Name,
	}

	resp, _, status, err := c.makeRequestWithPolling(ctx, endpoint, http.MethodPost, payload)
	if err != nil {
		log.WithError(err).Error("error creating team")
		return nil, err
	}

	if status != http.StatusOK && status != http.StatusCreated {
		return nil, fmt.Errorf("failed to create role, status: %s, body: %s", http.StatusText(status), string(resp))
	}

	createdTeam := &structs.Team{
		ID:   strings.ToLower(team.Name),
		Name: strings.ToLower(team.Name),
	}

	return createdTeam, nil
}

// FetchTeamDetails returns basic team information without making API calls
// since the detailed information is not consumed by the reconciliation workflow
func (c *SnowflakeClient) FetchTeamDetails(ctx context.Context, teamID string) (*structs.Team, error) {
	log := logger.Logger(ctx).WithFields(logrus.Fields{
		"service": "snowflake",
		"teamID":  teamID,
	})

	log.Info("fetching team details")
	// Since we're not consuming the detailed information from this function
	// and it's only required for interface, return basic team info
	// without making any API calls
	team := &structs.Team{
		ID:   strings.ToLower(teamID),
		Name: strings.ToLower(teamID),
	}
	log.Info("successfully fetched team details")
	return team, nil
}

// DeleteTeamByID deletes a role in Snowflake using REST API
func (c *SnowflakeClient) DeleteTeamByID(ctx context.Context, teamID string) error {
	log := logger.Logger(ctx).WithFields(logrus.Fields{
		"service": "snowflake",
		"teamID":  teamID,
	})

	log.Info("deleting team")
	endpoint := fmt.Sprintf("/api/v2/roles/%s", teamID)

	resp, _, status, err := c.makeRequestWithPolling(ctx, endpoint, http.MethodDelete, nil)
	if err != nil {
		log.WithError(err).Error("error deleting team")
		return fmt.Errorf("failed to delete role: %w", err)
	}

	if status != http.StatusOK && status != http.StatusNoContent {
		return fmt.Errorf("failed to delete role, status: %s, body: %s", http.StatusText(status), string(resp))
	}

	log.Info("team deleted successfully")
	return nil
}
