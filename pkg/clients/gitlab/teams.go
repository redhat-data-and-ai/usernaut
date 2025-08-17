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

package gitlab

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/redhat-data-and-ai/usernaut/pkg/common/structs"
	"github.com/redhat-data-and-ai/usernaut/pkg/logger"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

func (g *GitlabClient) FetchAllTeams(ctx context.Context) (map[string]structs.Team, error) {
	return map[string]structs.Team{}, nil
}

func (g *GitlabClient) FetchTeamDetails(ctx context.Context, teamID string) (*structs.Team, error) {
	groupIDInt, err := strconv.Atoi(teamID)
	if err != nil {
		return nil, err
	}
	group, _, err := g.gitlabClient.Groups.GetGroup(groupIDInt, &gitlab.GetGroupOptions{})
	if err != nil {
		return nil, err
	}
	return &structs.Team{
		ID:   fmt.Sprintf("%d", group.ID),
		Name: group.Name,
	}, nil
}

func (g *GitlabClient) CreateTeam(ctx context.Context, team *structs.Team) (*structs.Team, error) {
	logger.Logger(ctx).Info("Create GitLab team")

	groupName := team.Name
	visibility := gitlab.PublicVisibility
	createGroupOptions := &gitlab.CreateGroupOptions{
		ParentID:   &g.gitlabConfig.GroupId,
		Name:       &groupName,
		Path:       &groupName,
		Visibility: &visibility,
	}
	group, _, err := g.gitlabClient.Groups.CreateGroup(createGroupOptions)
	if err != nil {
		return nil, err
	}

	if g.ldapSync {
		// Add group to LDAP
		ldapLink, err := g.addToLdapGroup(g.cn, group.ID)
		if err != nil {
			return nil, err
		}
		logger.Logger(ctx).Info("LDAP link added successfully", ldapLink)

		// Initiate LDAP sync
		resp, err := g.initiateSync(ctx, g.gitlabConfig)
		if err != nil {
			return nil, err
		}
		logger.Logger(ctx).Infof("LDAP sync initiated successfully with status: %s", resp.Status)
	}

	return &structs.Team{
		ID:   fmt.Sprintf("%d", group.ID),
		Name: group.Name,
	}, nil
}

func (g *GitlabClient) DeleteTeamByID(ctx context.Context, teamID string) error {
	return nil
}

func (g *GitlabClient) addToLdapGroup(cn string, groupID int) (*gitlab.LDAPGroupLink, error) {
	ldapProvider := "ldapmain"
	accessLevel := gitlab.DeveloperPermissions
	ldapLink, _, err := g.gitlabClient.Groups.AddGroupLDAPLink(groupID, &gitlab.AddGroupLDAPLinkOptions{
		GroupAccess: &accessLevel,
		CN:          &cn,
		Provider:    &ldapProvider,
	})
	if err != nil {
		return nil, err
	}
	return ldapLink, nil
}

func (g *GitlabClient) initiateSync(ctx context.Context, cfg *GitlabConfig) (*http.Response, error) {
	httpClient := &http.Client{}
	url := fmt.Sprintf("%s/groups/%d/ldap_sync", cfg.URL, cfg.GroupId)
	req, err := http.NewRequest(http.MethodPost, url, http.NoBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+cfg.Token)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logger.Logger(ctx).Error("error closing response body", err)
		}
	}()
	if resp.StatusCode != http.StatusAccepted {
		return nil, fmt.Errorf("LDAP synchronization request failed with status: %s", resp.Status)
	}
	return resp, nil
}
