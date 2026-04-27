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

package snowflake

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"golang.org/x/sync/errgroup"

	"github.com/redhat-data-and-ai/usernaut/pkg/common/structs"
	"github.com/redhat-data-and-ai/usernaut/pkg/logger"
	"github.com/sirupsen/logrus"
)

// FetchTeamMembersByTeamID fetches team members for a given team ID with pagination support
func (c *SnowflakeClient) FetchTeamMembersByTeamID(ctx context.Context,
	teamID string) (map[string]*structs.User, error) {
	log := logger.Logger(ctx).WithFields(logrus.Fields{
		"service": "snowflake",
		"teamID":  teamID,
	})
	log.Info("fetching team members by team ID")

	members := make(map[string]*structs.User)

	// Use the correct endpoint: grants-of (not grants-on)
	endpoint := fmt.Sprintf("/api/v2/roles/%s/grants-of", teamID)

	err := c.fetchAllWithPagination(ctx, endpoint, func(resp []byte) error {
		return c.processGrantsPage(resp, members)
	})
	if err != nil {
		log.WithError(err).Error("error fetching team members by team ID")
		return nil, err
	}

	return members, nil
}

func (c *SnowflakeClient) processGrantsPage(resp []byte, members map[string]*structs.User) error {
	var grants []SnowflakeGrant
	if err := json.Unmarshal(resp, &grants); err != nil {
		return fmt.Errorf("error unmarshaling grants response: %w", err)
	}

	for _, grant := range grants {
		if grant.GrantedTo == "USER" && grant.GranteeName != "" {
			members[strings.ToLower(grant.GranteeName)] = &structs.User{
				ID:       strings.ToLower(grant.GranteeName),
				UserName: strings.ToLower(grant.GranteeName),
				Email:    "",
			}
		}
	}

	return nil
}

// parallelRoleModify runs endpointFn for each userID concurrently, bounded by MaxConcurrency.
func (c *SnowflakeClient) parallelRoleModify(ctx context.Context, teamID string, userIDs []string,
	endpointFn func(string) string, okStatuses []int, action string,
) error {
	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(c.config.MaxConcurrency)

	for _, userID := range userIDs {
		uid := userID
		g.Go(func() error {
			endpoint := endpointFn(sanitizeUserNameForAPI(uid))
			resp, status, err := c.makeRoleRequest(gctx, teamID, endpoint)
			if err != nil {
				return fmt.Errorf("failed to %s user %s to team %s: %w", action, uid, teamID, err)
			}
			for _, ok := range okStatuses {
				if status == ok {
					return nil
				}
			}
			return fmt.Errorf("failed to %s user %s to team %s, status: %s, body: %s",
				action, uid, teamID, http.StatusText(status), string(resp))
		})
	}
	return g.Wait()
}

// AddUserToTeam adds users to a team (grants role to users) concurrently.
// Concurrency is bounded by SnowflakeConfig.MaxConcurrency (default 10).
func (c *SnowflakeClient) AddUserToTeam(ctx context.Context, teamID string, userIDs []string) error {
	log := logger.Logger(ctx).WithFields(logrus.Fields{
		"service":    "snowflake",
		"teamID":     teamID,
		"user_count": len(userIDs),
	})
	log.Info("adding users to team")

	return c.parallelRoleModify(ctx, teamID, userIDs,
		func(user string) string { return fmt.Sprintf("/api/v2/users/%s/grants", user) },
		[]int{http.StatusOK, http.StatusCreated}, "add")
}

// RemoveUserFromTeam removes users from a team (revokes role from users) concurrently.
// Concurrency is bounded by SnowflakeConfig.MaxConcurrency (default 10).
func (c *SnowflakeClient) RemoveUserFromTeam(ctx context.Context, teamID string, userIDs []string) error {
	log := logger.Logger(ctx).WithFields(logrus.Fields{
		"service":    "snowflake",
		"teamID":     teamID,
		"user_count": len(userIDs),
	})
	log.Info("removing users from team")

	return c.parallelRoleModify(ctx, teamID, userIDs,
		func(user string) string { return fmt.Sprintf("/api/v2/users/%s/grants:revoke", user) },
		[]int{http.StatusOK, http.StatusNoContent}, "remove")
}

// makeRoleRequest sends a role grant/revoke request for a user
func (c *SnowflakeClient) makeRoleRequest(ctx context.Context, teamID, endpoint string) ([]byte, int, error) {
	payload := map[string]interface{}{
		"securable": map[string]string{
			"name": teamID,
		},
		"securable_type": "ROLE",
		"privileges":     []string{},
	}

	resp, _, status, err := c.makeRequestWithPolling(ctx, endpoint, http.MethodPost, payload)
	return resp, status, err
}

func (c *SnowflakeClient) ReconcileGroupParams(
	ctx context.Context, teamID string, groupParams structs.TeamParams) error {
	// TODO: Implement group parameter reconciliation for Snowflake if applicable.
	return nil
}
