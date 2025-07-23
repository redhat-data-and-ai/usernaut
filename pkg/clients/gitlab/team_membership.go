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

func (g *GitlabClient) AddUserToTeam(ctx context.Context, teamID, userID string) error {
	if g.ldapSync {
		return nil
	}

	log := logger.Logger(ctx)
	log.Info("Add user to GitLab team")

	groupIDInt, err := strconv.Atoi(teamID)
	if err != nil {
		return err
	}
	userIDInt, err := strconv.Atoi(userID)
	if err != nil {
		return err
	}
	accessLevel := gitlab.DeveloperPermissions
	addMemberOpts := &gitlab.AddGroupMemberOptions{
		UserID:      &userIDInt,
		AccessLevel: &accessLevel,
	}
	_, _, err = g.gitlabClient.GroupMembers.AddGroupMember(groupIDInt, addMemberOpts)
	if err != nil {
		return err
	}
	return nil
}

func (g *GitlabClient) RemoveUserFromTeam(ctx context.Context, teamID, userID string) error {
	// TODO: Implement using g.client.GroupMembers.RemoveGroupMember
	return nil
}
