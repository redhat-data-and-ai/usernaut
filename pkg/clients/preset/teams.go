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

	"github.com/redhat-data-and-ai/usernaut/pkg/common/structs"
	"github.com/redhat-data-and-ai/usernaut/pkg/logger"
	"github.com/sirupsen/logrus"
)

// FetchAllTeams retrieves all SCIM groups from Preset
func (pc *PresetClient) FetchAllTeams(ctx context.Context) (map[string]structs.Team, error) {
	log := logger.Logger(ctx).WithFields(logrus.Fields{
		"service": "preset",
	})
	log.Info("fetching all SCIM groups from Preset")

	teams := make(map[string]structs.Team)

	for startIndex := 1; ; startIndex += scimPageSize {
		reqURL := fmt.Sprintf("%s/Groups?startIndex=%d&count=%d", pc.scimURL(), startIndex, scimPageSize)
		response, _, err := pc.sendRequest(ctx, reqURL, http.MethodGet, nil)
		if err != nil {
			log.WithError(err).Error("failed to fetch SCIM groups from Preset")
			return nil, fmt.Errorf("failed to fetch SCIM groups from Preset: %w", err)
		}

		var scimResp scimGroupsResponse
		if err := json.Unmarshal(response, &scimResp); err != nil {
			log.WithError(err).Error("failed to parse SCIM groups response")
			return nil, fmt.Errorf("failed to parse SCIM groups response: %w", err)
		}

		if len(scimResp.Resources) == 0 {
			break
		}

		for _, g := range scimResp.Resources {
			team := scimGroupToTeam(&g)
			teams[team.ID] = *team
		}

		if startIndex+scimPageSize > scimResp.TotalResults {
			break
		}
	}

	log.WithFields(logrus.Fields{
		"team_count": len(teams),
	}).Info("successfully fetched SCIM groups from Preset")
	return teams, nil
}

// FetchTeamDetails retrieves a specific SCIM group by ID
func (pc *PresetClient) FetchTeamDetails(ctx context.Context, teamID string) (*structs.Team, error) {
	log := logger.Logger(ctx).WithFields(logrus.Fields{
		"service": "preset",
		"teamID":  teamID,
	})
	log.Info("fetching SCIM group details from Preset")

	group, err := pc.fetchSCIMGroup(ctx, teamID)
	if err != nil {
		log.WithError(err).Error("failed to fetch SCIM group from Preset")
		return nil, fmt.Errorf("failed to fetch SCIM group from Preset: %w", err)
	}

	log.Info("successfully fetched SCIM group details from Preset")
	return scimGroupToTeam(group), nil
}

// CreateTeam creates a new SCIM group in Preset.
// If a group with the same displayName already exists, returns the existing group.
func (pc *PresetClient) CreateTeam(ctx context.Context, team *structs.Team) (*structs.Team, error) {
	log := logger.Logger(ctx).WithFields(logrus.Fields{
		"service":   "preset",
		"team_name": team.Name,
	})
	log.Info("creating SCIM group in Preset")

	if existing, err := pc.findGroupByDisplayName(ctx, team.Name); err == nil {
		log.WithFields(logrus.Fields{
			"team_id": existing.ID,
		}).Info("SCIM group already exists in Preset")
		return existing, nil
	}

	reqURL := fmt.Sprintf("%s/Groups", pc.scimURL())
	reqBody := scimGroupCreateRequest{
		Schemas:     []string{scimGroupSchema},
		DisplayName: team.Name,
	}

	response, statusCode, err := pc.sendRequest(ctx, reqURL, http.MethodPost, reqBody)
	if err != nil {
		if statusCode == http.StatusConflict {
			return pc.requireGroupByDisplayName(ctx, team.Name, "SCIM group conflict but lookup failed")
		}
		log.WithError(err).Error("failed to create SCIM group in Preset")
		return nil, fmt.Errorf("failed to create SCIM group in Preset: %w", err)
	}

	var createdGroup scimGroup
	if err := json.Unmarshal(response, &createdGroup); err != nil {
		log.WithError(err).Error("failed to parse created SCIM group response")
		return nil, fmt.Errorf("failed to parse created SCIM group response: %w", err)
	}

	log.WithFields(logrus.Fields{
		"team_id": createdGroup.ID,
	}).Info("successfully created SCIM group in Preset")
	return scimGroupToTeam(&createdGroup), nil
}

// DeleteTeamByID deletes a SCIM group by its ID
func (pc *PresetClient) DeleteTeamByID(ctx context.Context, teamID string) error {
	log := logger.Logger(ctx).WithFields(logrus.Fields{
		"service": "preset",
		"teamID":  teamID,
	})
	log.Info("deleting SCIM group from Preset")

	reqURL := fmt.Sprintf("%s/Groups/%s", pc.scimURL(), teamID)
	_, statusCode, err := pc.sendRequest(ctx, reqURL, http.MethodDelete, nil)
	if err != nil {
		if statusCode == http.StatusNotFound {
			log.Info("SCIM group does not exist in Preset, nothing to delete")
			return nil
		}
		log.WithError(err).Error("failed to delete SCIM group from Preset")
		return fmt.Errorf("failed to delete SCIM group from Preset: %w", err)
	}

	log.Info("successfully deleted SCIM group from Preset")
	return nil
}

func (pc *PresetClient) findGroupByDisplayName(ctx context.Context, displayName string) (*structs.Team, error) {
	log := logger.Logger(ctx).WithFields(logrus.Fields{
		"service":   "preset",
		"team_name": displayName,
	})

	filter := fmt.Sprintf(`displayName eq "%s"`, displayName)
	response, err := pc.querySCIMByFilter(ctx, "Groups", filter, "team_name", displayName)
	if err != nil {
		return nil, err
	}

	var scimResp scimGroupsResponse
	if err := json.Unmarshal(response, &scimResp); err != nil {
		log.WithError(err).Error("failed to parse SCIM group lookup response")
		return nil, err
	}

	if len(scimResp.Resources) == 0 {
		log.Warn("SCIM group not found by displayName")
		return nil, fmt.Errorf("SCIM group not found: %s", displayName)
	}

	return scimGroupToTeam(&scimResp.Resources[0]), nil
}

func (pc *PresetClient) requireGroupByDisplayName(ctx context.Context, displayName, msg string) (*structs.Team, error) {
	team, err := pc.findGroupByDisplayName(ctx, displayName)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", msg, err)
	}
	return team, nil
}
