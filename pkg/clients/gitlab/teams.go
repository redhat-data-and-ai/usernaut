package gitlab

import (
	"context"
	"fmt"
	"strconv"

	"github.com/redhat-data-and-ai/usernaut/pkg/common/structs"
	"github.com/redhat-data-and-ai/usernaut/pkg/logger"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

func (g *GitlabClient) FetchAllTeams(ctx context.Context) (map[string]structs.Team, error) {
	// TODO: Implement using g.client.Groups.ListGroups
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
	log := logger.Logger(ctx)
	log.Info("Create GitLab team")

	if g.ldapSync {
		// Check if team exists in external teams (e.g., from Rover)
		if g.externalTeams != nil {
			if externalTeam, exists := g.externalTeams[team.Name]; exists {
				log.Info("Found team in external teams")
				return externalTeam, nil
			}
		}
	}

	// Create regular GitLab group (non-LDAP mode)

	createGroupOptions := &gitlab.CreateGroupOptions{
		Name: &team.Name,
		Path: &team.Name,
	}

	group, _, err := g.gitlabClient.Groups.CreateGroup(createGroupOptions)
	if err != nil {
		// Try to find existing group if creation failed
		log.WithField("error", err).Warn("Group creation failed, trying to fetch existing group")
		groups, _, fetchErr := g.gitlabClient.Groups.ListGroups(&gitlab.ListGroupsOptions{
			Search: &team.Name,
		})
		if fetchErr == nil && len(groups) > 0 {
			for _, existingGroup := range groups {
				if existingGroup.Name == team.Name {
					return &structs.Team{
						ID:   fmt.Sprintf("%d", existingGroup.ID),
						Name: existingGroup.Name,
					}, nil
				}
			}
		}
		return nil, err
	}

	return &structs.Team{
		ID:   fmt.Sprintf("%d", group.ID),
		Name: group.Name,
	}, nil
}

func (g *GitlabClient) DeleteTeamByID(ctx context.Context, teamID string) error {
	// TODO: Implement using g.client.Groups.DeleteGroup
	return nil
}
