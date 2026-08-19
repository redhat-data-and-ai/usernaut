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
	"net/url"
	"strings"

	"github.com/redhat-data-and-ai/usernaut/pkg/common/structs"
	"github.com/redhat-data-and-ai/usernaut/pkg/logger"
	"github.com/sirupsen/logrus"
)

// FetchAllTeams retrieves all SCIM groups from Preset
func (pc *PresetClient) FetchAllTeams(ctx context.Context) (map[string]structs.Team, error) {
	log := logger.Logger(ctx).WithField("service", "preset")
	log.Info("fetching all SCIM groups from Preset")

	teams := make(map[string]structs.Team)
	startIndex := 1
	count := 100

	for {
		reqURL := fmt.Sprintf("%s/Groups?startIndex=%d&count=%d", pc.scimURL(), startIndex, count)
		response, _, err := pc.sendRequest(ctx, reqURL, http.MethodGet, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch SCIM groups from Preset: %w", err)
		}

		var scimResp scimGroupsResponse
		if err := json.Unmarshal(response, &scimResp); err != nil {
			return nil, fmt.Errorf("failed to parse SCIM groups response: %w", err)
		}

		if len(scimResp.Resources) == 0 {
			break
		}

		for _, g := range scimResp.Resources {
			teams[g.ID] = structs.Team{
				ID:   g.ID,
				Name: g.DisplayName,
			}
		}

		if startIndex+count > scimResp.TotalResults {
			break
		}
		startIndex += count
	}

	log.WithField("team_count", len(teams)).Info("successfully fetched SCIM groups from Preset")
	return teams, nil
}

// FetchTeamDetails retrieves a specific SCIM group by ID
func (pc *PresetClient) FetchTeamDetails(ctx context.Context, teamID string) (*structs.Team, error) {
	log := logger.Logger(ctx).WithFields(logrus.Fields{
		"service": "preset",
		"teamID":  teamID,
	})
	log.Info("fetching SCIM group details from Preset")

	reqURL := fmt.Sprintf("%s/Groups/%s", pc.scimURL(), teamID)
	response, _, err := pc.sendRequest(ctx, reqURL, http.MethodGet, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch SCIM group from Preset: %w", err)
	}

	var group scimGroup
	if err := json.Unmarshal(response, &group); err != nil {
		return nil, fmt.Errorf("failed to parse SCIM group response: %w", err)
	}

	log.Info("successfully fetched SCIM group details from Preset")
	return &structs.Team{
		ID:   group.ID,
		Name: group.DisplayName,
	}, nil
}

// CreateTeam creates a new SCIM group in Preset.
// If a group with the same displayName already exists, returns the existing group.
func (pc *PresetClient) CreateTeam(ctx context.Context, team *structs.Team) (*structs.Team, error) {
	log := logger.Logger(ctx).WithFields(logrus.Fields{
		"service":   "preset",
		"team_name": team.Name,
	})
	log.Info("creating SCIM group in Preset")

	existing, err := pc.findGroupByDisplayName(ctx, team.Name)
	if err == nil && existing != nil {
		log.WithField("team_id", existing.ID).Info("SCIM group already exists in Preset")
		return existing, nil
	}

	reqURL := fmt.Sprintf("%s/Groups", pc.scimURL())
	reqBody := scimGroupCreateRequest{
		Schemas:     []string{"urn:ietf:params:scim:schemas:core:2.0:Group"},
		DisplayName: team.Name,
	}

	response, statusCode, err := pc.sendRequest(ctx, reqURL, http.MethodPost, reqBody)
	if err != nil {
		if statusCode == http.StatusConflict {
			log.Info("SCIM group already exists (conflict), fetching details")
			existing, findErr := pc.findGroupByDisplayName(ctx, team.Name)
			if findErr == nil && existing != nil {
				return existing, nil
			}
		}
		return nil, fmt.Errorf("failed to create SCIM group in Preset: %w", err)
	}

	var createdGroup scimGroup
	if err := json.Unmarshal(response, &createdGroup); err != nil {
		return nil, fmt.Errorf("failed to parse created SCIM group response: %w", err)
	}

	log.WithField("team_id", createdGroup.ID).Info("successfully created SCIM group in Preset")
	return &structs.Team{
		ID:   createdGroup.ID,
		Name: createdGroup.DisplayName,
	}, nil
}

// DeleteTeamByID deletes a SCIM group by its ID
func (pc *PresetClient) DeleteTeamByID(ctx context.Context, teamID string) error {
	log := logger.Logger(ctx).WithFields(logrus.Fields{
		"service": "preset",
		"teamID":  teamID,
	})
	log.Info("deleting SCIM group from Preset")

	reqURL := fmt.Sprintf("%s/Groups/%s", pc.scimURL(), teamID)
	_, _, err := pc.sendRequest(ctx, reqURL, http.MethodDelete, nil)
	if err != nil {
		if strings.Contains(err.Error(), "404") {
			log.Info("SCIM group does not exist in Preset, nothing to delete")
			return nil
		}
		return fmt.Errorf("failed to delete SCIM group from Preset: %w", err)
	}

	log.Info("successfully deleted SCIM group from Preset")
	return nil
}

// findGroupByDisplayName searches for a SCIM group by its displayName
func (pc *PresetClient) findGroupByDisplayName(ctx context.Context, displayName string) (*structs.Team, error) {
	filter := fmt.Sprintf(`displayName eq "%s"`, displayName)
	reqURL := fmt.Sprintf("%s/Groups?filter=%s", pc.scimURL(), url.QueryEscape(filter))
	response, _, err := pc.sendRequest(ctx, reqURL, http.MethodGet, nil)
	if err != nil {
		return nil, err
	}

	var scimResp scimGroupsResponse
	if err := json.Unmarshal(response, &scimResp); err != nil {
		return nil, err
	}

	if len(scimResp.Resources) == 0 {
		return nil, fmt.Errorf("SCIM group not found: %s", displayName)
	}

	g := scimResp.Resources[0]
	return &structs.Team{
		ID:   g.ID,
		Name: g.DisplayName,
	}, nil
}
