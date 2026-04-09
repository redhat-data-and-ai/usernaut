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

package redhatrover

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	ot "github.com/opentracing/opentracing-go"

	"github.com/redhat-data-and-ai/usernaut/pkg/common/structs"
	"github.com/redhat-data-and-ai/usernaut/pkg/logger"
)

// Fetch all the members and owners of a team by teamID ignoring the serviceaccount members
func (rC *RoverClient) FetchTeamMembersByTeamID(ctx context.Context, teamID string) (map[string]*structs.User, error) {
	span, ctx := ot.StartSpanFromContext(ctx, "backend.redhatrover.FetchTeamMembersByTeamID")
	defer span.Finish()

	log := logger.Logger(ctx)
	log.Info("Fetching team member details from rover group")

	resp, respCode, err := rC.sendRequest(ctx, rC.url+"/v1/groups/"+teamID,
		http.MethodGet, nil,
		headers, "backend.redhatrover.FetchTeamMembersByTeamID")

	if err != nil {
		log.WithError(err).Error("failed to fetch rover group members")
		return nil, err
	}

	if respCode != http.StatusOK {
		log.Error("failed to fetch rover group members")
		return nil, errors.New("failed to fetch rover group members with response code: " + http.StatusText(respCode))
	}

	var roverGroup RoverGroup
	if err := json.Unmarshal(resp, &roverGroup); err != nil {
		log.WithError(err).Error("failed to decode rover group response")
		return nil, errors.New("failed to decode rover group response: " + err.Error())
	}

	members := make(map[string]*structs.User)
	for _, member := range roverGroup.Members {
		if member.Type != MemberTypeUser {
			continue // Only process user type members
		}
		user := &structs.User{
			ID: member.ID,
		}
		members[user.ID] = user
	}

	return members, nil
}

const roverBatchSize = 500

func (rC *RoverClient) modify(
	ctx context.Context,
	spanName string,
	action string,
	teamID string,
	userIDs []string) error {
	span, ctx := ot.StartSpanFromContext(ctx, spanName)
	defer span.Finish()
	log := logger.Logger(ctx)

	if action != "add" && action != "remove" {
		return fmt.Errorf("invalid action:%s", action)
	}

	totalBatches := (len(userIDs) + roverBatchSize - 1) / roverBatchSize
	log.WithField("action", action).WithField("total_users", len(userIDs)).
		WithField("total_batches", totalBatches).
		Infof("%sing users in rover group (batched)", action)

	for batchStart := 0; batchStart < len(userIDs); batchStart += roverBatchSize {
		batchEnd := batchStart + roverBatchSize
		if batchEnd > len(userIDs) {
			batchEnd = len(userIDs)
		}
		batch := userIDs[batchStart:batchEnd]
		batchNum := (batchStart / roverBatchSize) + 1

		log.WithField("batch", fmt.Sprintf("%d/%d", batchNum, totalBatches)).
			WithField("batch_users", len(batch)).
			Infof("processing rover %s batch", action)

		members := make([]Member, 0, len(batch))
		for _, id := range batch {
			members = append(members, Member{ID: id, Type: MemberTypeUser})
		}

		var req MemberModRequest
		switch action {
		case "add":
			req.Additions = members
		case "remove":
			req.Deletions = members
		}

		_, respCode, err := rC.sendRequest(ctx,
			rC.url+"/v1/groups/"+teamID+"/membersMod",
			http.MethodPost,
			req,
			headers,
			spanName)
		if err != nil {
			log.WithError(err).Errorf("failed to %s users in rover group (batch %d/%d)", action, batchNum, totalBatches)
			return err
		}

		if respCode != http.StatusOK {
			log.Errorf("failed to %s users in rover group (batch %d/%d)", action, batchNum, totalBatches)
			return fmt.Errorf("failed to %s users in rover group with response code: %s", action, http.StatusText(respCode))
		}

		log.WithField("batch", fmt.Sprintf("%d/%d", batchNum, totalBatches)).
			Infof("rover %s batch completed", action)
	}

	return nil
}

// AddUserToTeam adds a user to a team in Rover by teamID and userID
func (rC *RoverClient) AddUserToTeam(ctx context.Context, teamID string, userIDs []string) error {
	return rC.modify(ctx, "backend.redhatrover.AddUserToTeam", "add", teamID, userIDs)
}

// RemoveUserFromTeam removes a user from a team in Rover by teamID and userID
func (rC *RoverClient) RemoveUserFromTeam(ctx context.Context, teamID string, userIDs []string) error {
	return rC.modify(ctx, "backend.redhatrover.RemoveUserFromTeam", "remove", teamID, userIDs)
}

func (rC *RoverClient) ReconcileGroupParams(ctx context.Context, teamID string, groupParams structs.TeamParams) error {
	// TODO: Implement group parameter reconciliation for Red Hat Rover if applicable.
	return nil
}
