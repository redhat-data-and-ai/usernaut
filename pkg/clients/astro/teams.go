/*
Copyright 2026.

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
	"net/url"
	"strings"

	"github.com/redhat-data-and-ai/usernaut/pkg/common/structs"
	"github.com/redhat-data-and-ai/usernaut/pkg/logger"
	"github.com/sirupsen/logrus"
)

// astroTeamToStruct converts an AstroTeam to a structs.Team
func astroTeamToStruct(team AstroTeam) structs.Team {
	return structs.Team{
		ID:          team.ID,
		Name:        team.Name,
		Description: team.Description,
		Role:        team.OrganizationRole,
	}
}

// FetchAllTeams fetches all teams from Astro using REST API with pagination
func (c *AstroClient) FetchAllTeams(ctx context.Context) (map[string]structs.Team, error) {
	log := logger.Logger(ctx).WithField("service", "astro")

	log.Info("fetching all teams")
	teams := make(map[string]structs.Team)

	baseEndpoint := "/teams"

	err := c.fetchAllWithPagination(ctx, baseEndpoint, func(resp []byte) (int, error) {
		var teamsResp AstroTeamsResponse
		if err := json.Unmarshal(resp, &teamsResp); err != nil {
			return 0, fmt.Errorf("failed to parse teams response: %w", err)
		}

		for _, team := range teamsResp.Teams {
			structTeam := astroTeamToStruct(team)
			teams[team.ID] = structTeam
		}
		return len(teamsResp.Teams), nil
	})

	if err != nil {
		log.WithError(err).Error("error fetching list of teams")
		return nil, err
	}

	log.WithField("total_teams_count", len(teams)).Info("found teams")
	return teams, nil
}

// FetchTeamDetails fetches details for a specific team
func (c *AstroClient) FetchTeamDetails(ctx context.Context, teamID string) (*structs.Team, error) {
	log := logger.Logger(ctx).WithFields(logrus.Fields{
		"service": "astro",
		"teamID":  teamID,
	})

	log.Info("fetching team details")

	endpoint := fmt.Sprintf("/teams/%s", teamID)
	resp, status, err := c.makeRequest(ctx, endpoint, http.MethodGet, nil)
	if err != nil {
		return nil, err
	}

	if status == http.StatusNotFound {
		return nil, fmt.Errorf("team with ID %s not found", teamID)
	}

	if status != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch team %s: status %d, body: %s", teamID, status, string(resp))
	}

	var team AstroTeam
	if err := json.Unmarshal(resp, &team); err != nil {
		return nil, fmt.Errorf("failed to parse team response: %w", err)
	}

	log.Info("successfully fetched team details")
	result := astroTeamToStruct(team)
	return &result, nil
}

// fetchTeamByName looks up a team by name using the Astro names filter.
func (c *AstroClient) fetchTeamByName(ctx context.Context, name string) (*structs.Team, error) {
	log := logger.Logger(ctx).WithFields(logrus.Fields{
		"service":  "astro",
		"teamName": name,
	})

	endpoint := fmt.Sprintf("/teams?names=%s&limit=1", url.QueryEscape(name))
	resp, status, err := c.makeRequest(ctx, endpoint, http.MethodGet, nil)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch team by name %s: status %s, body: %s",
			name, http.StatusText(status), string(resp))
	}

	var teamsResp AstroTeamsResponse
	if err := json.Unmarshal(resp, &teamsResp); err != nil {
		return nil, fmt.Errorf("failed to parse teams response: %w", err)
	}

	for _, team := range teamsResp.Teams {
		if team.Name == name {
			result := astroTeamToStruct(team)
			log.WithField("teamID", result.ID).Info("found existing team by name")
			return &result, nil
		}
	}

	return nil, fmt.Errorf("team not found by name: %s", name)
}

// CreateTeam creates a new team in Astro
func (c *AstroClient) CreateTeam(ctx context.Context, team *structs.Team) (*structs.Team, error) {
	log := logger.Logger(ctx).WithFields(logrus.Fields{
		"service":  "astro",
		"teamName": team.Name,
	})

	log.Info("creating team")
	endpoint := "/teams"

	payload := CreateTeamRequest{
		Name:             team.Name,
		Description:      team.Description,
		OrganizationRole: DefaultOrganizationRole,
	}

	resp, status, err := c.makeRequest(ctx, endpoint, http.MethodPost, payload)
	if err != nil {
		log.WithError(err).Error("error creating team")
		return nil, err
	}

	// Astro returns 400 with "already exists" for duplicate team names.
	if status == http.StatusConflict ||
		(status == http.StatusBadRequest && strings.Contains(strings.ToLower(string(resp)), "already exists")) {
		log.Info("team already exists, fetching existing team by name")
		return c.fetchTeamByName(ctx, team.Name)
	}

	if status != http.StatusOK && status != http.StatusCreated {
		return nil, fmt.Errorf("failed to create team, status: %s, body: %s",
			http.StatusText(status), string(resp))
	}

	var createdTeam AstroTeam
	if err := json.Unmarshal(resp, &createdTeam); err != nil {
		return nil, fmt.Errorf("failed to parse create team response: %w", err)
	}

	log.WithField("teamID", createdTeam.ID).Info("team created successfully")

	result := astroTeamToStruct(createdTeam)
	return &result, nil
}

// DeleteTeamByID deletes a team from Astro
func (c *AstroClient) DeleteTeamByID(ctx context.Context, teamID string) error {
	log := logger.Logger(ctx).WithFields(logrus.Fields{
		"service": "astro",
		"teamID":  teamID,
	})

	log.Info("deleting team")
	endpoint := fmt.Sprintf("/teams/%s", teamID)

	resp, status, err := c.makeRequest(ctx, endpoint, http.MethodDelete, nil)
	if err != nil {
		log.WithError(err).Error("error deleting team")
		return fmt.Errorf("failed to delete team: %w", err)
	}

	if status != http.StatusOK && status != http.StatusNoContent {
		return fmt.Errorf("failed to delete team, status: %s, body: %s",
			http.StatusText(status), string(resp))
	}

	log.Info("team deleted successfully")
	return nil
}
