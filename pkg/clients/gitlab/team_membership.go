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
	"strconv"

	"github.com/redhat-data-and-ai/usernaut/pkg/common/structs"
	"github.com/redhat-data-and-ai/usernaut/pkg/logger"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

func (g *GitlabClient) FetchTeamMembersByTeamID(ctx context.Context, teamID string) (map[string]*structs.User, error) {
	teamMembers := make(map[string]*structs.User)

	// Default: use GitLab SDK API to fetch the team members
	groupIDInt, err := strconv.Atoi(teamID)
	if err != nil {
		return nil, err
	}
	members, _, err := g.gitlabClient.Groups.ListAllGroupMembers(groupIDInt, nil)
	if err != nil {
		return nil, err
	}
	for _, m := range members {
		teamMembers[fmt.Sprintf("%d", m.ID)] = &structs.User{
			ID:    fmt.Sprintf("%d", m.ID),
			Email: m.Email,
		}
	}
	return teamMembers, nil
}

func (g *GitlabClient) AddUserToTeam(ctx context.Context, teamID string, userIDs []string) error {
	logger.Logger(ctx).Info("Add users to GitLab team")

	if g.ldapSync {
		return nil
	}

	if len(userIDs) == 0 {
		return nil
	}

	groupIDInt, err := strconv.Atoi(teamID)
	if err != nil {
		return err
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
		_, _, err = g.gitlabClient.GroupMembers.AddGroupMember(groupIDInt, addMemberOpts)
		if err != nil {
			return err
		}
	}
	return nil
}

func (g *GitlabClient) RemoveUserFromTeam(ctx context.Context, teamID string, userIDs []string) error {
	if g.ldapSync {
		return nil
	}

	return nil
}
