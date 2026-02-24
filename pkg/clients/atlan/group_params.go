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
	ctx context.Context, teamID string, teamName string, groupParams structs.TeamParams,
) error {
	log := logger.Logger(ctx).WithFields(logrus.Fields{
		"service":  "atlan",
		"teamID":   teamID,
		"teamName": teamName,
		"property": groupParams.Property,
	})
	log.Info("reconciling group params for Atlan")

	switch groupParams.Property {
	case "persona":
		// Handle persona assignment using teamName (the transformed group name)
		if teamName == "" {
			log.Warn("team name is empty, cannot assign personas")
			return nil
		}

		additionalPersonas := groupParams.Value
		if err := ac.AddGroupToPersonas(ctx, teamName, additionalPersonas); err != nil {
			log.WithError(err).Error("error assigning group to personas")
			return err
		}
		log.Info("successfully assigned group to personas")

	default:
		log.WithField("property", groupParams.Property).Warn("unsupported group property for atlan backend")
	}

	return nil
}
