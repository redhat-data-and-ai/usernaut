package atlan

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

type AtlanGroup struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type AtlanGroupsResponse struct {
	TotalRecord  int          `json:"totalRecord"`
	FilterRecord int          `json:"filterRecord"`
	Records      []AtlanGroup `json:"records"`
}

func (ac *AtlanClient) FetchAllTeams(ctx context.Context) (map[string]structs.Team, error) {
	log := logger.Logger(ctx).WithField("service", "atlan")
	log.Info("fetching all teams from Atlan")

	url := fmt.Sprintf("%s/api/service/v2/groups?columns=name", ac.url)
	response, statusCode, err := ac.sendRequest(ctx, url, http.MethodGet, nil, nil, "FetchAllTeams")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch teams from Atlan: %w", err)
	}

	if statusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d when fetching teams from Atlan", statusCode)
	}

	var apiResponse AtlanGroupsResponse
	if err := json.Unmarshal(response, &apiResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response from Atlan: %w", err)
	}

	teams := make(map[string]structs.Team)
	for _, group := range apiResponse.Records {
		teams[group.ID] = structs.Team{
			ID:   group.ID,
			Name: group.Name,
		}
	}

	log.WithField("team_count", len(teams)).Info("successfully fetched teams from Atlan")
	return teams, nil
}

func (ac *AtlanClient) CreateTeam(ctx context.Context, team *structs.Team) (*structs.Team, error) {
	log := logger.Logger(ctx).WithFields(logrus.Fields{
		"service":   "atlan",
		"team_name": team.Name,
	})

	log.Info("creating team in Atlan")

	url := fmt.Sprintf("%s/api/service/groups", ac.url)

	// Convert team name to valid internal name as per Atlan guidelines (lowercase, alphanumeric + underscore only)
	internalName := strings.ToLower(strings.ReplaceAll(team.Name, " ", "_"))
	internalName = strings.ReplaceAll(internalName, "-", "_")

	requestBody := map[string]interface{}{
		"group": map[string]interface{}{
			"attributes": map[string]interface{}{
				"alias":     []string{team.Name},
				"isDefault": []string{"false"},
			},
			"name": internalName, // Sanitized internal name
		},
	}

	response, statusCode, err := ac.sendRequest(ctx, url, http.MethodPost, requestBody, nil, "CreateTeam")
	if err != nil {
		return nil, fmt.Errorf("failed to create team in Atlan: %w", err)
	}

	if statusCode != http.StatusCreated && statusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d when creating team in Atlan", statusCode)
	}

	var createdGroup AtlanGroup
	if err := json.Unmarshal(response, &createdGroup); err != nil {
		return nil, fmt.Errorf("failed to parse created team response from Atlan: %w", err)
	}

	log.WithField("team_id", createdGroup.ID).Info("successfully created team in Atlan")
	return &structs.Team{
		ID:   createdGroup.ID,
		Name: createdGroup.Name,
	}, nil
}

func (ac *AtlanClient) DeleteTeamByID(ctx context.Context, teamID string) error {
	log := logger.Logger(ctx).WithFields(logrus.Fields{
		"service": "atlan",
		"team_id": teamID,
	})
	log.Info("deleting team from Atlan")

	url := fmt.Sprintf("%s/api/service/groups/%s", ac.url, teamID)
	_, statusCode, err := ac.sendRequest(ctx, url, http.MethodDelete, nil, nil, "DeleteTeamByID")
	if err != nil {
		return fmt.Errorf("failed to delete team from Atlan: %w", err)
	}

	if statusCode != http.StatusOK && statusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status code %d when deleting team from Atlan", statusCode)
	}

	log.Info("successfully deleted team from Atlan")
	return nil
}

func (ac *AtlanClient) FetchTeamDetails(ctx context.Context, teamID string) (*structs.Team, error) {
	return nil, fmt.Errorf("FetchTeamDetails is not implemented for Atlan")
}
