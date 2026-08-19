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

// FetchTeamMembersByTeamID retrieves all members of a SCIM group
func (pc *PresetClient) FetchTeamMembersByTeamID(ctx context.Context, teamID string) (map[string]*structs.User, error) {
	log := logger.Logger(ctx).WithFields(logrus.Fields{
		"service": "preset",
		"teamID":  teamID,
	})
	log.Info("fetching SCIM group members from Preset")

	reqURL := fmt.Sprintf("%s/Groups/%s", pc.scimURL(), teamID)
	response, _, err := pc.sendRequest(ctx, reqURL, http.MethodGet, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch SCIM group members from Preset: %w", err)
	}

	var group scimGroup
	if err := json.Unmarshal(response, &group); err != nil {
		return nil, fmt.Errorf("failed to parse SCIM group response: %w", err)
	}

	members := make(map[string]*structs.User)
	for _, m := range group.Members {
		members[m.Value] = &structs.User{
			ID:          m.Value,
			DisplayName: m.Display,
		}
	}

	log.WithField("member_count", len(members)).Info("fetched SCIM group members from Preset")
	return members, nil
}

// AddUserToTeam adds users to a SCIM group via PATCH operation
func (pc *PresetClient) AddUserToTeam(ctx context.Context, teamID string, userIDs []string) error {
	log := logger.Logger(ctx).WithFields(logrus.Fields{
		"service":    "preset",
		"teamID":     teamID,
		"user_count": len(userIDs),
	})

	if len(userIDs) == 0 {
		return nil
	}

	log.Info("adding users to SCIM group in Preset")

	members := make([]scimMemberValue, 0, len(userIDs))
	for _, uid := range userIDs {
		members = append(members, scimMemberValue{Value: uid})
	}

	patchReq := scimPatchRequest{
		Schemas: []string{"urn:ietf:params:scim:api:messages:2.0:PatchOp"},
		Operations: []scimPatchOperation{
			{
				Op:    "add",
				Path:  "members",
				Value: members,
			},
		},
	}

	reqURL := fmt.Sprintf("%s/Groups/%s", pc.scimURL(), teamID)
	_, _, err := pc.sendRequest(ctx, reqURL, http.MethodPatch, patchReq)
	if err != nil {
		return fmt.Errorf("failed to add users to SCIM group in Preset: %w", err)
	}

	log.Info("successfully added users to SCIM group in Preset")
	return nil
}

// RemoveUserFromTeam removes users from a SCIM group via PATCH operation.
// Each user is removed individually using the path filter format required by Preset.
func (pc *PresetClient) RemoveUserFromTeam(ctx context.Context, teamID string, userIDs []string) error {
	log := logger.Logger(ctx).WithFields(logrus.Fields{
		"service":    "preset",
		"teamID":     teamID,
		"user_count": len(userIDs),
	})

	if len(userIDs) == 0 {
		return nil
	}

	log.Info("removing users from SCIM group in Preset")

	operations := make([]scimPatchOperation, 0, len(userIDs))
	for _, uid := range userIDs {
		operations = append(operations, scimPatchOperation{
			Op:   "remove",
			Path: fmt.Sprintf(`members[value eq "%s"]`, uid),
		})
	}

	patchReq := scimPatchRequest{
		Schemas:    []string{"urn:ietf:params:scim:api:messages:2.0:PatchOp"},
		Operations: operations,
	}

	reqURL := fmt.Sprintf("%s/Groups/%s", pc.scimURL(), teamID)
	_, _, err := pc.sendRequest(ctx, reqURL, http.MethodPatch, patchReq)
	if err != nil {
		return fmt.Errorf("failed to remove users from SCIM group in Preset: %w", err)
	}

	log.Info("successfully removed users from SCIM group in Preset")
	return nil
}
