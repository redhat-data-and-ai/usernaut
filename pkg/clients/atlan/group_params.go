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

package atlan

import (
	"context"

	"github.com/redhat-data-and-ai/usernaut/pkg/common/structs"
	"github.com/redhat-data-and-ai/usernaut/pkg/logger"
	"github.com/sirupsen/logrus"
)

// ReconcileGroupParams reconciles backend-specific parameters for a group/team in Atlan.
// For Atlan, this handles persona assignment based on the group_params configuration.
func (ac *AtlanClient) ReconcileGroupParams(
	ctx context.Context, teamID string, groupParams structs.TeamParams,
) error {
	log := logger.Logger(ctx).WithFields(logrus.Fields{
		"service":  "atlan",
		"teamID":   teamID,
		"teamName": groupParams.TeamName,
		"property": groupParams.Property,
	})
	log.Info("reconciling group params for Atlan")

	if groupParams.TeamName == "" {
		log.Warn("team name is empty, cannot reconcile group params")
		return nil
	}

	// Convert to Atlan's internal name format (lowercase, underscores instead of dashes/spaces)
	// This matches the transformation done in CreateTeam
	atlanGroupName := ToAtlanInternalName(groupParams.TeamName)
	log.WithField("atlanGroupName", atlanGroupName).Debug("converted team name to Atlan internal format")

	switch groupParams.Property {
	case "persona":
		// Handle persona assignment using the Atlan internal group name
		additionalPersonas := groupParams.Value
		if err := ac.AddGroupToPersonas(ctx, atlanGroupName, additionalPersonas); err != nil {
			log.WithError(err).Error("error assigning group to personas")
			return err
		}
		log.Info("successfully assigned group to personas")

	case "":
		// No explicit group params specified, but still assign default persona if configured
		if ac.defaultPersona != "" {
			log.Info("no explicit persona property, assigning default persona")
			if err := ac.AddGroupToPersonas(ctx, atlanGroupName, nil); err != nil {
				log.WithError(err).Error("error assigning group to default persona")
				return err
			}
			log.Info("successfully assigned group to default persona")
		}

	default:
		log.WithField("property", groupParams.Property).Warn("unsupported group property for atlan backend")
	}

	return nil
}
