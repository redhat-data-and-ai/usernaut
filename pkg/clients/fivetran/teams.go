package fivetran

import (
	"context"

	"github.com/fivetran/go-fivetran/teams"
	"github.com/redhat-data-and-ai/usernaut/pkg/logger"
	"github.com/sirupsen/logrus"
)

func (fc *FivetranClient) FetchAllTeams(ctx context.Context) (map[string]teams.TeamData, error) {
	log := logger.Logger(ctx).WithField("service", "fivetran")

	log.Info("fetching all the teams")

	resp, err := fc.fivetranClient.NewTeamsList().Do(ctx)
	if err != nil {
		log.WithError(err).Error("error fetching list of teams")
		return nil, err
	}

	log.WithField("total_teams_count", len(resp.Data.Items)).Info("found teams")

	teams := make(map[string]teams.TeamData, 0)
	for _, g := range resp.Data.Items {
		teams[g.Name] = g
	}
	return teams, nil
}

func (fc *FivetranClient) CreateTeam(ctx context.Context, g *Team) (*Team, error) {
	log := logger.Logger(ctx).WithFields(logrus.Fields{
		"service": "fivetran",
		"req":     g,
	})

	if g.Role == "" {
		g.Role = AccountReviewerRole
	}

	log.Info("creating team")
	resp, err := fc.fivetranClient.NewTeamsCreate().
		Name(g.TeamName).
		Role(g.Role).
		Description(g.Description).
		Do(ctx)

	if err != nil {
		log.WithError(err).WithField("response", resp).Error("error creating the team")
		return nil, err
	}

	return &Team{
		TeamName:    resp.Data.Name,
		TeamID:      resp.Data.Id,
		Description: resp.Data.Description,
		Role:        resp.Data.Role,
	}, nil
}

func (fc *FivetranClient) UpdateTeam(ctx context.Context, g *UpdateTeam) (*Team, error) {
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
		log.WithError(err).WithField("response", resp).Error("error creating the team")
		return nil, err
	}

	return &Team{
		TeamName:    resp.Data.Name,
		TeamID:      resp.Data.Id,
		Description: resp.Data.Description,
		Role:        resp.Data.Role,
	}, nil
}

func (fc *FivetranClient) FetchTeamDetails(ctx context.Context, teamID string) (teams.TeamData, error) {
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
		return teams.TeamData{}, err
	}

	log.Info("successfully fetched team details")

	return resp.Data, nil
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
