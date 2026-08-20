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

	"github.com/redhat-data-and-ai/usernaut/pkg/common/structs"
	"github.com/redhat-data-and-ai/usernaut/pkg/logger"
	"github.com/sirupsen/logrus"
)

const scimPageSize = 100

const (
	scimUserSchema  = "urn:ietf:params:scim:schemas:core:2.0:User"
	scimGroupSchema = "urn:ietf:params:scim:schemas:core:2.0:Group"
	scimPatchSchema = "urn:ietf:params:scim:api:messages:2.0:PatchOp"
)

func scimUserToStruct(su *scimUser) *structs.User {
	email := ""
	if len(su.Emails) > 0 {
		for _, e := range su.Emails {
			if e.Primary {
				email = e.Value
				break
			}
		}
		if email == "" {
			email = su.Emails[0].Value
		}
	}

	return &structs.User{
		ID:          su.ID,
		Email:       email,
		UserName:    su.UserName,
		DisplayName: su.DisplayName,
	}
}

func scimGroupToTeam(g *scimGroup) *structs.Team {
	return &structs.Team{
		ID:   g.ID,
		Name: g.DisplayName,
	}
}

func (pc *PresetClient) fetchSCIMUser(ctx context.Context, userID string) (*scimUser, error) {
	reqURL := fmt.Sprintf("%s/Users/%s", pc.scimURL(), userID)
	response, _, err := pc.sendRequest(ctx, reqURL, http.MethodGet, nil)
	if err != nil {
		return nil, err
	}

	var user scimUser
	if err := json.Unmarshal(response, &user); err != nil {
		return nil, fmt.Errorf("failed to parse SCIM user response: %w", err)
	}
	return &user, nil
}

func (pc *PresetClient) fetchSCIMGroup(ctx context.Context, teamID string) (*scimGroup, error) {
	reqURL := fmt.Sprintf("%s/Groups/%s", pc.scimURL(), teamID)
	response, _, err := pc.sendRequest(ctx, reqURL, http.MethodGet, nil)
	if err != nil {
		return nil, err
	}

	var group scimGroup
	if err := json.Unmarshal(response, &group); err != nil {
		return nil, fmt.Errorf("failed to parse SCIM group response: %w", err)
	}
	return &group, nil
}

func scimMembersToUserMap(group *scimGroup) map[string]*structs.User {
	members := make(map[string]*structs.User, len(group.Members))
	for _, m := range group.Members {
		members[m.Value] = &structs.User{
			ID:          m.Value,
			DisplayName: m.Display,
		}
	}
	return members
}

func (pc *PresetClient) querySCIMByFilter(
	ctx context.Context, resource string, filter string, fieldKey string, fieldValue string,
) ([]byte, error) {
	log := logger.Logger(ctx).WithFields(logrus.Fields{
		"service": "preset",
		fieldKey:  fieldValue,
	})

	reqURL := fmt.Sprintf("%s/%s?filter=%s", pc.scimURL(), resource, url.QueryEscape(filter))
	response, _, err := pc.sendRequest(ctx, reqURL, http.MethodGet, nil)
	if err != nil {
		log.WithError(err).Errorf("failed to query SCIM %s by filter", resource)
		return nil, err
	}
	return response, nil
}
