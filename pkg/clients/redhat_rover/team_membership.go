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

// AddUserToTeam adds a user to a team in Rover by teamID and userID
func (rC *RoverClient) AddUserToTeam(ctx context.Context, teamID, userID string) error {
	span, ctx := ot.StartSpanFromContext(ctx, "backend.redhatrover.AddUserToTeam")
	defer span.Finish()

	log := logger.Logger(ctx)
	log.Info("Adding team member to the rover group")

	req := MemberModRequest{
		Additions: []Member{
			{
				ID:   userID,
				Type: MemberTypeUser,
			},
		},
	}

	_, respCode, err := rC.sendRequest(ctx, rC.url+"/v1/groups/"+teamID+"/membersMod",
		http.MethodPost, req,
		headers, "backend.redhatrover.AddUserToTeam")

	if err != nil {
		log.WithError(err).Error("failed to add user to rover group")
		return err
	}

	if respCode != http.StatusOK {
		log.Error("failed to add user to rover group")
		return errors.New("failed to add user to rover group with response code: " + http.StatusText(respCode))
	}

	return nil
}

// RemoveUserFromTeam removes a user from a team in Rover by teamID and userID
func (rC *RoverClient) RemoveUserFromTeam(ctx context.Context, teamID, userID string) error {
	span, ctx := ot.StartSpanFromContext(ctx, "backend.redhatrover.RemoveUserFromTeam")
	defer span.Finish()

	log := logger.Logger(ctx)
	log.Info("Removing team member from the rover group")

	req := MemberModRequest{
		Deletions: []Member{
			{
				ID:   userID,
				Type: MemberTypeUser,
			},
		},
	}

	_, respCode, err := rC.sendRequest(ctx, rC.url+"/v1/groups/"+teamID+"/membersMod",
		http.MethodPost, req,
		headers, "backend.redhatrover.RemoveUserFromTeam")
	if err != nil {
		log.WithError(err).Error("failed to remove user from rover group")
		return err
	}

	if respCode != http.StatusOK {
		log.Error("failed to remove user from rover group")
		return errors.New("failed to remove user from rover group with response code: " + http.StatusText(respCode))
	}

	return nil
}
