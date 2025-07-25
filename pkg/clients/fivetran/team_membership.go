package fivetran

import (
	"context"
	"fmt"
	"sync"

	"github.com/redhat-data-and-ai/usernaut/pkg/common/structs"
	"github.com/redhat-data-and-ai/usernaut/pkg/logger"
	"github.com/sirupsen/logrus"
)

// maxConcurrentUsers defines the max number of concurrent operations allowed when interaction with API
const maxConcurrentUsers = 10

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

func (fc *FivetranClient) AddUserToTeam(ctx context.Context, teamID string, userIDs []string) error {
	log := logger.Logger(ctx).WithFields(logrus.Fields{
		"service":    "fivetran",
		"teamID":     teamID,
		"user_count": len(userIDs),
	})

	log.Info("adding users to the team")

	var wg sync.WaitGroup
	errch := make(chan error, len(userIDs)) // this is an error channel
	sem := make(chan struct{}, maxConcurrentUsers)

	for _, id := range userIDs {
		wg.Add(1)
		sem <- struct{}{}

		go func(uid string, log logrus.FieldLogger) {
			defer wg.Done()
			defer func() { <-sem }()
			slog := log.WithField("userID", uid)

			slog.Info("adding user to fivetran team ")
			resp, err := fc.fivetranClient.
				NewTeamUserMembershipCreate().
				TeamId(teamID).
				UserId(uid).
				Role("Team Member").
				Do(ctx)

			if err != nil {
				slog.WithField("response", resp.CommonResponse).WithError(err).
					Error("Error adding user to team")
				errch <- fmt.Errorf("%s: %w", uid, err)
				return
			}
			slog.Info("added users to the team successfully")
		}(id, log)
	}

	wg.Wait()
	close(errch)

	allErrors := make([]error, 0, len(userIDs))
	for err := range errch {
		allErrors = append(allErrors, err)
	}
	if len(allErrors) > 0 {
		return fmt.Errorf("multiple errors occurred: %v", allErrors)
	}
	return nil
}

func (fc *FivetranClient) RemoveUserFromTeam(ctx context.Context, teamID string, userIDs []string) error {
	log := logger.Logger(ctx).WithFields(logrus.Fields{
		"service":    "fivetran",
		"teamID":     teamID,
		"user_count": len(userIDs),
	})

	log.Info("removing users from the team")
	var wg sync.WaitGroup
	errch := make(chan error, len(userIDs))
	sem := make(chan struct{}, maxConcurrentUsers)

	for _, id := range userIDs {
		wg.Add(1)
		sem <- struct{}{}

		go func(uid string, log logrus.FieldLogger) {
			defer wg.Done()
			defer func() { <-sem }()

			slog := log.WithField("userID", uid)
			slog.Info("removing user from the team")
			resp, err := fc.fivetranClient.NewTeamUserMembershipDelete().
				TeamId(teamID).
				UserId(uid).
				Do(ctx)
			if err != nil {
				slog.WithField("response", resp).WithError(err).Error("error removing user from the team")
				errch <- fmt.Errorf("%s: %w", uid, err)
				return

			}
			slog.Info("user removed from team successfully")
		}(id, log)

	}

	wg.Wait()
	close(errch)

	allErrors := make([]error, 0, len(userIDs))
	for err := range errch {
		allErrors = append(allErrors, err)
	}
	if len(allErrors) > 0 {
		return fmt.Errorf("multiple errors occurred: %v", allErrors)
	}
	return nil
}
