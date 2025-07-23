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

	"github.com/redhat-data-and-ai/usernaut/pkg/common/structs"
	"github.com/redhat-data-and-ai/usernaut/pkg/logger"
	"github.com/sirupsen/logrus"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

func (g *GitlabClient) FetchAllTeams(ctx context.Context) (map[string]structs.Team, error) {
	log := logger.Logger(ctx).WithField("service", "gitlab")
	log.Info("fetching all teams")

	teams := make(map[string]structs.Team)
	opt := &gitlab.ListSubGroupsOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: 100,
			Page:    1,
		},
	}

	for {
		groups, resp, err := g.gitlabClient.Groups.ListSubGroups(g.gitlabConfig.GroupId, opt)
		if err != nil {
			return nil, err
		}

		for _, group := range groups {
			teams[group.Name] = structs.Team{
				ID:   fmt.Sprintf("%d", group.ID),
				Name: group.Name,
			}
		}

		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	log.WithField("total_teams_count", len(teams)).Info("found teams")
	return teams, nil
}

func (g *GitlabClient) FetchTeamDetails(ctx context.Context, teamID string) (*structs.Team, error) {
	log := logger.Logger(ctx).WithFields(logrus.Fields{
		"service": "gitlab",
		"teamID":  teamID,
	})
	log.Info("fetching team details")

	group, _, err := g.gitlabClient.Groups.GetGroup(teamID, &gitlab.GetGroupOptions{})
	if err != nil {
		return nil, err
	}
	return &structs.Team{
		ID:   fmt.Sprintf("%d", group.ID),
		Name: group.Name,
	}, nil
}

func (g *GitlabClient) CreateTeam(ctx context.Context, team *structs.Team) (*structs.Team, error) {
	log := logger.Logger(ctx).WithFields(logrus.Fields{
		"service": "gitlab",
		"team":    team,
	})
	log.Info("creating team")

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
		log.Info("ldap link added successfully", ldapLink)

		// Initiate LDAP sync
		resp, err := g.initiateSync(ctx, g.gitlabConfig)
		if err != nil {
			return nil, err
		}
		log.Infof("ldap sync initiated successfully with status: %s", resp.Status)
	}

	return &structs.Team{
		ID:   fmt.Sprintf("%d", group.ID),
		Name: group.Name,
	}, nil
}

func (g *GitlabClient) DeleteTeamByID(ctx context.Context, teamID string) error {
	log := logger.Logger(ctx).WithFields(logrus.Fields{
		"service": "gitlab",
		"teamID":  teamID,
	})
	log.Info("deleting team")

	// Delete group
	permanentlyRemove := false
	deleteGroupOpts := &gitlab.DeleteGroupOptions{
		PermanentlyRemove: &permanentlyRemove,
	}
	resp, err := g.gitlabClient.Groups.DeleteGroup(teamID, deleteGroupOpts)
	if err != nil {
		return err
	}
	log.Infof("team soft-deleted, permanent deletion will be done after 7 days. With status: %s", resp.Status)
	return nil
}

func (g *GitlabClient) addToLdapGroup(cn string, groupID int) (*gitlab.LDAPGroupLink, error) {
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
	log := logger.Logger(ctx).WithField("service", "gitlab")
	log.Info("initiating LDAP sync")

	url := fmt.Sprintf("%s/groups/%d/ldap_sync", cfg.URL, cfg.GroupId)
	req, err := http.NewRequest(http.MethodPost, url, http.NoBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+cfg.Token)

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.WithError(err).Error("error closing response body")
		}
	}()
	if resp.StatusCode != http.StatusAccepted {
		return nil, fmt.Errorf("ldap synchronization request failed with status: %s", resp.Status)
	}
	return resp, nil
}
