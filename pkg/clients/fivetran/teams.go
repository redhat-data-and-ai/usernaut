package fivetran

import (
	"context"

	"github.com/fivetran/go-fivetran/teams"
	"github.com/redhat-data-and-ai/usernaut/pkg/common/structs"
	"github.com/redhat-data-and-ai/usernaut/pkg/logger"
	"github.com/sirupsen/logrus"
)

// populateTeamsMap populates the provided teams map with team data from the response items
func populateTeamsMap(teamsMap map[string]structs.Team, items []teams.TeamData) {
	for _, team := range items {
		teamsMap[team.Name] = structs.Team{
			ID:          team.Id,
			Name:        team.Name,
			Description: team.Description,
			Role:        team.Role,
		}
	}
}

func (fc *FivetranClient) FetchAllTeams(ctx context.Context) (map[string]structs.Team, error) {
	log := logger.Logger(ctx).WithField("service", "fivetran")

	log.Info("fetching all the teams")

	teams := make(map[string]structs.Team)
	var cursor string

	for {
		req := fc.fivetranClient.NewTeamsList()
		if cursor != "" {
			req.Cursor(cursor)
		}

		resp, err := req.Do(ctx)
		if err != nil {
			log.WithError(err).Error("error fetching list of teams")
			return nil, err
		}

		populateTeamsMap(teams, resp.Data.Items)

		if resp.Data.NextCursor == "" {
			break
		}
		cursor = resp.Data.NextCursor
	}

	log.WithFields(logrus.Fields{
		"total_teams_count": len(teams),
	}).Info("found teams")

	return teams, nil
}

func (fc *FivetranClient) CreateTeam(ctx context.Context, team *structs.Team) (*structs.Team, error) {
	log := logger.Logger(ctx).WithFields(logrus.Fields{
		"service": "fivetran",
		"req":     team,
	})

	if team.Role == "" {
		team.Role = AccountReviewerRole
	}

	log.Info("creating team")
	resp, err := fc.fivetranClient.NewTeamsCreate().
		Name(team.Name).
		Role(team.Role).
		Description(team.Description).
		Do(ctx)

	if err != nil {
		log.WithError(err).WithField("response", resp).Error("error creating the team")
		return nil, err
	}

	return &structs.Team{
		Name:        resp.Data.Name,
		ID:          resp.Data.Id,
		Description: resp.Data.Description,
		Role:        resp.Data.Role,
	}, nil
}

func (fc *FivetranClient) UpdateTeam(ctx context.Context, g *UpdateTeam) (*structs.Team, error) {
	log := logger.Logger(ctx).WithFields(logrus.Fields{
		"service": "fivetran",
		"req":     g,
	})

	if g.NewRole == "" {
		g.NewRole = AccountReviewerRole
	}

	log.Info("updating team")
	resp, err := fc.fivetranClient.NewTeamsModify().
		TeamId(g.ExistingTeamID).
		Role(g.NewRole).
		Name(g.NewTeamName).
		Description(g.NewDescription).
		Do(ctx)

	if err != nil {
		log.WithError(err).WithField("response", resp).Error("error updating the team")
		return nil, err
	}

	return &structs.Team{
		Name:        resp.Data.Name,
		ID:          resp.Data.Id,
		Description: resp.Data.Description,
		Role:        resp.Data.Role,
	}, nil
}

func (fc *FivetranClient) FetchTeamDetails(ctx context.Context, teamID string) (*structs.Team, error) {
	log := logger.Logger(ctx).WithFields(logrus.Fields{
		"service": "fivetran",
		"teamID":  teamID,
	})

	log.Info("fetching team details")

	resp, err := fc.fivetranClient.NewTeamsDetails().
		TeamId(teamID).
		Do(ctx)
	if err != nil {
		log.WithField("responseCode", resp.Code).WithError(err).Error("error fetching team details")
		return &structs.Team{}, err
	}

	log.Info("successfully fetched team details")

	return &structs.Team{
		ID:          resp.Data.Id,
		Name:        resp.Data.Name,
		Description: resp.Data.Description,
		Role:        resp.Data.Role,
	}, nil
}

func (fc *FivetranClient) DeleteTeamByID(ctx context.Context, teamID string) error {
	log := logger.Logger(ctx).WithFields(logrus.Fields{
		"service": "fivetran",
		"teamID":  teamID,
	})

	log.Info("deleting the team")
	resp, err := fc.fivetranClient.NewTeamsDelete().TeamId(teamID).Do(ctx)
	if err != nil {
		log.WithField("response", resp).WithError(err).Error("error deleting the team")
		return err
	}

	log.WithField("response", resp).Info("team deleted successfully")
	return nil

}
