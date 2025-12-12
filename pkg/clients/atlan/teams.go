package atlan

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/redhat-data-and-ai/usernaut/pkg/common/structs"
	"github.com/redhat-data-and-ai/usernaut/pkg/logger"
	"github.com/sirupsen/logrus"
)

func (ac *AtlanClient) FetchAllTeams(ctx context.Context) (map[string]structs.Team, error) {
	log := logger.Logger(ctx).WithField("service", "atlan")
	log.Info("fetching all teams from Atlan")

	url := fmt.Sprintf("%s/api/service/v2/groups?columns=name", ac.url)
	response, err := ac.sendRequest(ctx, url, http.MethodGet, nil, "FetchAllTeams")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch teams from Atlan: %w", err)
	}

	var apiResponse AtlanGroupsResponse
	if err := json.Unmarshal(response, &apiResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response from Atlan: %w", err)
	}

	teams := make(map[string]structs.Team)
	for _, group := range apiResponse.Records {
		teams[group.ID] = structs.Team{
			ID:   group.ID,
			Name: group.Name,
		}
	}

	log.WithField("team_count", len(teams)).Info("successfully fetched teams from Atlan")
	return teams, nil
}

func (ac *AtlanClient) CreateTeam(ctx context.Context, team *structs.Team) (*structs.Team, error) {
	log := logger.Logger(ctx).WithFields(logrus.Fields{
		"service":   "atlan",
		"team_name": team.Name,
	})
	log.Info("creating team in Atlan")

	url := fmt.Sprintf("%s/api/service/groups", ac.url)

	internalName := strings.ToLower(strings.ReplaceAll(team.Name, " ", "_"))
	internalName = strings.ReplaceAll(internalName, "-", "_")

	requestBody := map[string]interface{}{
		"group": map[string]interface{}{
			"attributes": map[string]interface{}{
				"alias":     []string{team.Name},
				"isDefault": []string{"false"},
			},
			"name": internalName,
		},
	}

	response, err := ac.sendRequest(ctx, url, http.MethodPost, requestBody, "CreateTeam")
	if err != nil {
		return nil, fmt.Errorf("failed to create team in Atlan: %w", err)
	}

	// Parse response - Atlan returns {"group": "<id>", "users": null}
	var createResponse struct {
		Group string `json:"group"`
	}
	if err := json.Unmarshal(response, &createResponse); err != nil {
		return nil, fmt.Errorf("failed to parse created team response from Atlan: %w", err)
	}

	createdGroup := AtlanGroup{
		ID:   createResponse.Group,
		Name: internalName,
	}

	log.WithFields(logrus.Fields{"team_id": createdGroup.ID, "team_name": createdGroup.Name}).Info("created team in Atlan")

	// If ldapSync is enabled (depends_on rover is defined), create SSO mapping
	if ac.ldapSync {
		ssoGroupName := ac.ssoGroupName
		if ssoGroupName == "" {
			ssoGroupName = createdGroup.Name
		}

		if err := ac.CreateSSOMapping(ctx, createdGroup.ID, createdGroup.Name, ssoGroupName); err != nil {
			log.WithError(err).Error("failed to create SSO group mapping")
		} else {
			log.Info("created SSO group mapping")
		}
	}

	return &structs.Team{
		ID:   createdGroup.ID,
		Name: createdGroup.Name,
	}, nil
}

func (ac *AtlanClient) CreateSSOMapping(ctx context.Context, teamID, teamName, ssoGroupName string) error {
	log := logger.Logger(ctx).WithFields(logrus.Fields{
		"service":   "atlan",
		"team_id":   teamID,
		"team_name": teamName,
		"sso_group": ssoGroupName,
	})
	log.Info("creating SSO group mapping")

	provider := ac.identityProviderAlias
	if provider == "" {
		return fmt.Errorf("identity provider alias not configured")
	}

	groupMapping := SSOGroupMapping{
		IdentityProviderAlias:  provider,
		IdentityProviderMapper: "saml-group-idp-mapper",
		Name:                   fmt.Sprintf("%s--%d", teamID, time.Now().UnixMilli()),
		Config: SSOGroupMappingConfig{
			SyncMode:                "FORCE",
			Attributes:              "[]",
			AreAttributeValuesRegex: "",
			AttributeName:           "memberOf",
			Group:                   teamName,
			AttributeValue:          ssoGroupName,
		},
	}

	url := fmt.Sprintf("%s/api/service/idp/%s/mappers", ac.url, provider)
	_, err := ac.sendRequest(ctx, url, http.MethodPost, groupMapping, "CreateSSOMapping")
	if err != nil {
		return fmt.Errorf("failed to create SSO group mapping: %w", err)
	}

	log.Info("successfully created SSO group mapping")
	return nil
}

func (ac *AtlanClient) DeleteTeamByID(ctx context.Context, teamID string) error {
	log := logger.Logger(ctx).WithFields(logrus.Fields{
		"service": "atlan",
		"team_id": teamID,
	})
	log.Info("deleting team from Atlan")

	url := fmt.Sprintf("%s/api/service/groups/%s/delete", ac.url, teamID)
	_, err := ac.sendRequest(ctx, url, http.MethodPost, nil, "DeleteTeamByID")
	if err != nil {
		// If error contains 404 or "not found", team is already deleted
		if strings.Contains(err.Error(), "404") || strings.Contains(strings.ToLower(err.Error()), "not found") {
			log.Info("team does not exist in Atlan, nothing to delete")
			return nil
		}
		return fmt.Errorf("failed to delete team from Atlan: %w", err)
	}

	log.Info("successfully deleted team from Atlan")
	return nil
}

// TODO: Implement FetchTeamDetails when needed - currently returns error as it's not used
func (ac *AtlanClient) FetchTeamDetails(ctx context.Context, teamID string) (*structs.Team, error) {
	return nil, fmt.Errorf("FetchTeamDetails is not implemented for Atlan")
}
