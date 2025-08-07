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

package redhatrover

import (
	"context"
	"fmt"
	"net/http"

	ot "github.com/opentracing/opentracing-go"

	"github.com/redhat-data-and-ai/usernaut/pkg/common/structs"
	"github.com/redhat-data-and-ai/usernaut/pkg/logger"
)

func (rC *RoverClient) FetchAllTeams(ctx context.Context) (map[string]structs.Team, error) {
	// Fetching all the rover groups isn't required as teamName is the unique identifier for teams in Rover.
	return map[string]structs.Team{}, nil
}

func (rC *RoverClient) FetchTeamDetails(ctx context.Context, teamID string) (*structs.Team, error) {
	// Fetching team details is not supported as the teamID is the same as the teamName.
	return nil, fmt.Errorf("fetching team details is not supported")
}

// CreateTeam creates a new team in Rover. If the team already exists, it returns the existing team details.
func (rC *RoverClient) CreateTeam(ctx context.Context, team *structs.Team) (*structs.Team, error) {
	span, ctx := ot.StartSpanFromContext(ctx, "backend.redhatrover.CreateTeam")
	defer span.Finish()

	log := logger.Logger(ctx)
	log.Info("Create Rover team")

	roverGroup := &RoverGroup{
		Name:               team.Name,
		Description:        team.Description,
		MemberApprovalType: MemberApprovalTypeSelfService,
		Owners: []Member{
			{
				ID:   rC.serviceAccountName,
				Type: MemberTypeServiceAccount,
			},
		},
		ContactList: defaultContactEmail,
		Notes:       "Created by Usernaut",
	}

	resp, respCode, err := rC.sendRequest(ctx, rC.url+"/v1/groups",
		http.MethodPost, roverGroup,
		headers, "backend.redhatrover.CreateTeam")
	if err != nil {
		if respCode != http.StatusForbidden {
			return nil, err
		}
	}

	// API return 403 Forbidden if the group already exists.
	if respCode == http.StatusForbidden {
		log.WithField("response", string(resp)).Warn("Rover group already exists, fetching existing group details")
		return &structs.Team{
			ID:          team.Name,
			Name:        team.Name,
			Description: team.Description,
		}, nil
	}

	if respCode != http.StatusCreated {
		log.Error("failed to create rover group")
		return nil, fmt.Errorf("failed to create rover group: %s", string(resp))
	}

	return &structs.Team{
		ID:          team.Name,
		Name:        team.Name,
		Description: team.Description,
	}, nil
}

func (rC *RoverClient) DeleteTeamByID(ctx context.Context, teamID string) error {
	// This will be implemented in the future when Usernaut supports deleting teams.
	return nil
}
