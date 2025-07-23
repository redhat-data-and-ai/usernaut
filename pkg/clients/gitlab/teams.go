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
	"time"

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
		groups, resp, err := g.gitlabClient.Groups.ListSubGroups(g.gitlabConfig.ParentGroupId, opt)
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
		ParentID:   &g.gitlabConfig.ParentGroupId,
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
		ldapLink, err := g.addToLdapGroup(group.ID)
		if err != nil {
			log.WithError(err).Error("failed to add group to LDAP", "groupID", group.ID)
		} else {
			log.Info("ldap link added successfully", ldapLink)
		}

		// Initiate LDAP sync
		statusCode, err := g.initiateSync(ctx)
		if err != nil {
			log.WithError(err).Error("failed to initiate LDAP sync", "groupID", group.ID)
		} else {
			log.Infof("ldap sync initiated successfully with status: %d", statusCode)
		}
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

	// 1. Initiate Soft Delete
	resp, err := g.gitlabClient.Groups.DeleteGroup(teamID, &gitlab.DeleteGroupOptions{})
	if err != nil {
		return fmt.Errorf("failed to initiate soft delete: %w", err)
	}
	log.Infof("team %v soft-deleted with status: %s", teamID, resp.Status)

	// 2. Poll until pending deletion status is confirmed
	groupFullPath, err := g.pollForPendingDeletion(ctx, teamID, 5, 5*time.Second)
	if err != nil {
		return err
	}

	// 3. Perform Hard Delete
	permanentlyRemove := true
	deleteGroupOpts := &gitlab.DeleteGroupOptions{
		PermanentlyRemove: &permanentlyRemove,
		FullPath:          &groupFullPath,
	}
	resp, err = g.gitlabClient.Groups.DeleteGroup(teamID, deleteGroupOpts)
	if err != nil {
		return err
	}
	log.Infof("team %v hard-deleted with status: %s", teamID, resp.Status)
	return nil
}

func (g *GitlabClient) addToLdapGroup(groupID int) (*gitlab.LDAPGroupLink, error) {
	accessLevel := gitlab.DeveloperPermissions
	ldapLink, _, err := g.gitlabClient.Groups.AddGroupLDAPLink(groupID, &gitlab.AddGroupLDAPLinkOptions{
		GroupAccess: &accessLevel,
		CN:          &g.cn,
		Provider:    &ldapProvider,
	})
	if err != nil {
		return nil, err
	}
	return ldapLink, nil
}

func (g *GitlabClient) initiateSync(ctx context.Context) (int, error) {
	log := logger.Logger(ctx).WithField("service", "gitlab")
	log.Info("initiating LDAP sync")

	resp, statusCode, err := g.sendLdapSyncRequest(ctx)
	if err != nil {
		return 0, err
	}
	if statusCode != http.StatusAccepted {
		return 0, fmt.Errorf("ldap synchronization request failed with status: %s", string(resp))
	}
	return statusCode, nil
}

func (g *GitlabClient) pollForPendingDeletion(ctx context.Context,
	teamID string,
	maxAttempts int,
	interval time.Duration) (string, error) {
	log := logger.Logger(ctx).WithFields(logrus.Fields{
		"service": "gitlab",
		"teamID":  teamID,
	})
	for i := 0; i < maxAttempts; i++ {
		group, resp, err := g.gitlabClient.Groups.GetGroup(teamID, &gitlab.GetGroupOptions{})
		if err != nil {
			if resp.StatusCode == 404 {
				log.Infof("Group %v not found", teamID)
				return "", nil
			}
			log.Infof("Error checking group status (attempt %d/%d): %v\n", i+1, maxAttempts, err)
		}
		if group.MarkedForDeletionOn != nil {
			log.Infof("Group %s is now marked for deletion on %s.", group.Name, group.MarkedForDeletionOn.String())
			return group.FullPath, nil
		}

		log.Infof("Group %s not yet marked for deletion. Retrying in %v.", group.Name, interval)
		time.Sleep(interval)
	}

	return "", fmt.Errorf("timeout: Group %v was not marked for deletion after %d attempts", teamID, maxAttempts)
}
