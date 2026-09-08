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

package preset

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/redhat-data-and-ai/usernaut/pkg/common/structs"
	"github.com/redhat-data-and-ai/usernaut/pkg/logger"
	"github.com/sirupsen/logrus"
)

// FetchTeamMembersByTeamID retrieves all members of a SCIM group
func (pc *PresetClient) FetchTeamMembersByTeamID(ctx context.Context, teamID string) (map[string]*structs.User, error) {
	log := logger.Logger(ctx).WithFields(logrus.Fields{
		"service": "preset",
		"teamID":  teamID,
	})
	log.Info("fetching SCIM group members from Preset")

	memberList, err := pc.fetchSCIMGroupMembers(ctx, teamID)
	if err != nil {
		log.WithError(err).Error("failed to fetch SCIM group members from Preset")
		return nil, fmt.Errorf("failed to fetch SCIM group members from Preset: %w", err)
	}

	members := scimMembersToUserMap(memberList)
	log.WithFields(logrus.Fields{
		"member_count": len(members),
	}).Info("fetched SCIM group members from Preset")
	return members, nil
}

func scimMembersToUserMap(memberList []scimMember) map[string]*structs.User {
	members := make(map[string]*structs.User, len(memberList))
	for _, m := range memberList {
		members[m.Value] = &structs.User{
			ID:          m.Value,
			DisplayName: m.Display,
		}
	}
	return members
}

// AddUserToTeam adds users to a SCIM group via batched PATCH operations.
func (pc *PresetClient) AddUserToTeam(ctx context.Context, teamID string, userIDs []string) error {
	if len(userIDs) == 0 {
		return nil
	}

	log := logger.Logger(ctx).WithFields(logrus.Fields{
		"service":    "preset",
		"teamID":     teamID,
		"user_count": len(userIDs),
	})
	log.Info("adding users to SCIM group in Preset")

	if err := pc.runMembershipBatches(ctx, log, teamID, userIDs, pc.patchAddMembers,
		"failed to add user batch to SCIM group in Preset",
		"failed to add users to SCIM group in Preset",
	); err != nil {
		return err
	}

	log.Info("successfully added users to SCIM group in Preset")
	return nil
}

func (pc *PresetClient) patchAddMembers(ctx context.Context, teamID string, userIDs []string) error {
	members := make([]scimMember, len(userIDs))
	for i, uid := range userIDs {
		members[i] = scimMember{Value: uid}
	}

	patchReq := scimPatchRequest{
		Schemas: []string{scimPatchSchema},
		Operations: []scimPatchOperation{
			{Op: "add", Path: "members", Value: members},
		},
	}

	reqURL := pc.groupURL(teamID)
	_, err := pc.sendRequest(ctx, reqURL, http.MethodPatch, patchReq)
	return err
}

// RemoveUserFromTeam removes users from a SCIM group via batched PATCH operations.
func (pc *PresetClient) RemoveUserFromTeam(ctx context.Context, teamID string, userIDs []string) error {
	if len(userIDs) == 0 {
		return nil
	}

	log := logger.Logger(ctx).WithFields(logrus.Fields{
		"service":    "preset",
		"teamID":     teamID,
		"user_count": len(userIDs),
	})
	log.Info("removing users from SCIM group in Preset")

	if err := pc.runMembershipBatches(ctx, log, teamID, userIDs, pc.patchRemoveMembers,
		"failed to remove user batch from SCIM group in Preset",
		"failed to remove users from SCIM group in Preset",
	); err != nil {
		return err
	}

	log.Info("successfully removed users from SCIM group in Preset")
	return nil
}

func (pc *PresetClient) patchRemoveMembers(ctx context.Context, teamID string, userIDs []string) error {
	operations := make([]scimPatchOperation, len(userIDs))
	for i, uid := range userIDs {
		operations[i] = scimPatchOperation{
			Op:   "remove",
			Path: fmt.Sprintf(`members[value eq "%s"]`, escapeSCIMLiteral(uid)),
		}
	}

	patchReq := scimPatchRequest{
		Schemas:    []string{scimPatchSchema},
		Operations: operations,
	}

	reqURL := pc.groupURL(teamID)
	_, err := pc.sendRequest(ctx, reqURL, http.MethodPatch, patchReq)
	return err
}

type membershipBatchFn func(ctx context.Context, teamID string, userIDs []string) error

func (pc *PresetClient) runMembershipBatches(
	ctx context.Context,
	log *logrus.Entry,
	teamID string,
	userIDs []string,
	batchFn membershipBatchFn,
	batchFailMsg, overallFailMsg string,
) error {
	if len(userIDs) == 0 {
		return nil
	}

	totalBatches := (len(userIDs) + scimMembershipBatchSize - 1) / scimMembershipBatchSize
	var errs []error

	for start := 0; start < len(userIDs); start += scimMembershipBatchSize {
		batchNum := start/scimMembershipBatchSize + 1
		end := start + scimMembershipBatchSize
		if end > len(userIDs) {
			end = len(userIDs)
		}
		batch := userIDs[start:end]

		log.WithFields(logrus.Fields{
			"batch":         fmt.Sprintf("%d/%d", batchNum, totalBatches),
			"batch_size":    len(batch),
			"total_batches": totalBatches,
		}).Info("processing SCIM group membership batch")

		if err := batchFn(ctx, teamID, batch); err != nil {
			log.WithError(err).WithFields(logrus.Fields{
				"batch":         batchNum,
				"total_batches": totalBatches,
				"batch_size":    len(batch),
			}).Error(batchFailMsg)
			errs = append(errs, fmt.Errorf("batch %d/%d: %w", batchNum, totalBatches, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("%s: %w", overallFailMsg, errors.Join(errs...))
	}
	return nil
}

// ReconcileGroupParams is a no-op for the Preset backend.
func (pc *PresetClient) ReconcileGroupParams(
	ctx context.Context, teamID string, groupParams structs.TeamParams,
) error {
	return nil
}
