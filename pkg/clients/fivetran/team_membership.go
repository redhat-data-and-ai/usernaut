package fivetran

import (
	"context"

	"github.com/redhat-data-and-ai/usernaut/pkg/common/structs"
	"github.com/redhat-data-and-ai/usernaut/pkg/logger"
	"github.com/sirupsen/logrus"
)

func (fc *FivetranClient) FetchTeamMembersByTeamID(
	ctx context.Context,
	teamID string) (map[string]*structs.User, error) {

	log := logger.Logger(ctx).WithFields(logrus.Fields{
		"service": "fivetran",
		"teamID":  teamID,
	})
	log.Info("fetching team members by team ID")

	teamMembers := make(map[string]*structs.User, 0)
	resp, err := fc.fivetranClient.NewTeamUserMembershipsList().
		TeamId(teamID).
		Do(ctx)
	if err != nil {
		log.WithError(err).Error("error fetching team members by team ID")
		return nil, err
	}
	for _, item := range resp.Data.Items {
		teamMembers[item.UserId] = &structs.User{
			ID:   item.UserId,
			Role: item.Role,
		}
	}

	cursor := resp.Data.NextCursor
	for len(cursor) != 0 {
		resp, err := fc.fivetranClient.NewTeamUserMembershipsList().
			TeamId(teamID).
			Cursor(cursor).
			Do(ctx)
		if err != nil {
			log.WithField("response", resp.Code).WithError(err).Error("error fetching list of team members")
			return nil, err
		}
		for _, item := range resp.Data.Items {
			teamMembers[item.UserId] = &structs.User{
				ID:   item.UserId,
				Role: item.Role,
			}
		}
		cursor = resp.Data.NextCursor
	}

	return teamMembers, nil

}

func (fc *FivetranClient) AddUserToTeam(ctx context.Context, teamID, userID string) error {
	log := logger.Logger(ctx).WithFields(logrus.Fields{
		"service": "fivetran",
		"userID":  userID,
		"teamID":  teamID,
	})

	log.Info("adding user to the team")
	resp, err := fc.fivetranClient.NewTeamUserMembershipCreate().
		Role("Team Member").
		TeamId(teamID).
		UserId(userID).
		Do(ctx)
	if err != nil {
		log.WithField("response", resp.CommonResponse).WithError(err).Error("error adding user to the team")
		return err

	}
	log.Info("user added to the team")
	return nil
}

func (fc *FivetranClient) RemoveUserFromTeam(ctx context.Context, teamID, userID string) error {
	log := logger.Logger(ctx).WithFields(logrus.Fields{
		"service": "fivetran",
		"userID":  userID,
		"teamID":  teamID,
	})

	log.Info("removing user from the team")
	resp, err := fc.fivetranClient.NewTeamUserMembershipDelete().
		TeamId(teamID).
		UserId(userID).
		Do(ctx)
	if err != nil {
		log.WithField("response", resp).WithError(err).Error("error removing user from the team")
		return err

	}
	log.Info("user removed from the team")
	return nil
}
