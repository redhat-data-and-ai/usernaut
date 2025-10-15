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
	"github.com/sirupsen/logrus"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

func (g *GitlabClient) FetchTeamMembersByTeamID(ctx context.Context, teamID string) (map[string]*structs.User, error) {
	log := logger.Logger(ctx).WithFields(logrus.Fields{
		"service": "gitlab",
		"teamID":  teamID,
	})
	log.Info("fetching team members by team ID")

	teamMembers := make(map[string]*structs.User)
	members, _, err := g.gitlabClient.Groups.ListAllGroupMembers(teamID, nil)
	if err != nil {
		return nil, err
	}
	for _, m := range members {
		teamMembers[fmt.Sprintf("%d", m.ID)] = &structs.User{
			ID:       fmt.Sprintf("%d", m.ID),
			Email:    m.PublicEmail,
			UserName: m.Username,
		}
	}
	return teamMembers, nil
}

func (g *GitlabClient) AddUserToTeam(ctx context.Context, teamID string, userIDs []string) error {
	log := logger.Logger(ctx).WithFields(logrus.Fields{
		"service": "gitlab",
		"teamID":  teamID,
		"userIDs": userIDs,
	})
	log.Info("adding users to team")

	if g.ldapSync || len(userIDs) == 0 {
		return nil
	}

	accessLevel := gitlab.DeveloperPermissions
	for _, userID := range userIDs {
		userIDInt, convErr := strconv.Atoi(userID)
		if convErr != nil {
			return convErr
		}
		addMemberOpts := &gitlab.AddGroupMemberOptions{
			UserID:      &userIDInt,
			AccessLevel: &accessLevel,
		}
		_, resp, err := g.gitlabClient.GroupMembers.AddGroupMember(teamID, addMemberOpts)
		if err != nil {
			return err
		}
		if resp.StatusCode != http.StatusCreated {
			return fmt.Errorf("failed to add user %s to team %s, status: %s", userID, teamID, resp.Status)
		}
	}
	return nil
}

func (g *GitlabClient) RemoveUserFromTeam(ctx context.Context, teamID string, userIDs []string) error {
	log := logger.Logger(ctx).WithFields(logrus.Fields{
		"service": "gitlab",
		"teamID":  teamID,
		"userIDs": userIDs,
	})
	log.Info("removing users from team")

	if g.ldapSync || len(userIDs) == 0 {
		return nil
	}

	for _, userID := range userIDs {
		userIDInt, convErr := strconv.Atoi(userID)
		if convErr != nil {
			return convErr
		}
		resp, err := g.gitlabClient.GroupMembers.RemoveGroupMember(teamID, userIDInt, nil)
		if err != nil {
			return err
		}
		if resp.StatusCode != http.StatusNoContent {
			return fmt.Errorf("failed to remove user %s from team %s, status: %s", userID, teamID, resp.Status)
		}
	}
	return nil
}
