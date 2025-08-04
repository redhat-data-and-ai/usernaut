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

type AtlanGroup struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type AtlanGroupsResponse struct {
	TotalRecord  int          `json:"totalRecord"`
	FilterRecord int          `json:"filterRecord"`
	Records      []AtlanGroup `json:"records"`
}

type SSOGroupMappingConfig struct {
	SyncMode                string `json:"syncMode"`
	Attributes              string `json:"attributes"`
	AreAttributeValuesRegex string `json:"are.attribute.values.regex"`
	AttributeName           string `json:"attribute.name"`
	Group                   string `json:"group"`
	AttributeValue          string `json:"attribute.value"`
}

type SSOGroupMapping struct {
	ID                     string                `json:"id"`
	IdentityProviderAlias  string                `json:"identityProviderAlias"`
	IdentityProviderMapper string                `json:"identityProviderMapper"`
	Name                   string                `json:"name"`
	Config                 SSOGroupMappingConfig `json:"config"`
}

func (ac *AtlanClient) FetchAllTeams(ctx context.Context) (map[string]structs.Team, error) {
	log := logger.Logger(ctx).WithField("service", "atlan")
	log.Info("fetching all teams from Atlan")

	url := fmt.Sprintf("%s/api/service/v2/groups?columns=name", ac.url)
	response, statusCode, err := ac.sendRequest(ctx, url, http.MethodGet, nil, nil, "FetchAllTeams")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch teams from Atlan: %w", err)
	}

	if statusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d when fetching teams from Atlan", statusCode)
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

	response, statusCode, err := ac.sendRequest(ctx, url, http.MethodPost, requestBody, nil, "CreateTeam")
	if err != nil {
		return nil, fmt.Errorf("failed to create team in Atlan: %w", err)
	}

	if statusCode != http.StatusCreated && statusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d when creating team in Atlan", statusCode)
	}

	var createdGroup AtlanGroup
	if err := json.Unmarshal(response, &createdGroup); err != nil {
		return nil, fmt.Errorf("failed to parse created team response from Atlan: %w", err)
	}

	log.WithField("team_id", createdGroup.ID).Info("successfully created team in Atlan")

	return &structs.Team{
		ID:   createdGroup.ID,
		Name: createdGroup.Name,
	}, nil
}

func (ac *AtlanClient) CreateTeamWithSSO(ctx context.Context, team *structs.Team, ssoGroupName string) (*structs.Team, error) {
	createdTeam, err := ac.CreateTeam(ctx, team)
	if err != nil {
		return nil, err
	}

	if ssoGroupName == "" {
		ssoGroupName = createdTeam.Name
	}

	if err := ac.CreateSSOMapping(ctx, createdTeam.ID, createdTeam.Name, ssoGroupName); err != nil {
		log := logger.Logger(ctx)
		log.WithError(err).Error("failed to create SSO group mapping")
	}

	return createdTeam, nil
}

func (ac *AtlanClient) CreateSSOMapping(ctx context.Context, teamID, teamName, ssoGroupName string) error {
	provider := ac.identityProviderAlias
	if provider == "" {
		return fmt.Errorf("identity provider alias not configured - please set identity_provider_alias in backend connection config")
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
	_, statusCode, err := ac.sendRequest(ctx, url, http.MethodPost, groupMapping, nil, "CreateSSOMapping")
	if err != nil {
		return fmt.Errorf("failed to create SSO group mapping: %w", err)
	}

	if statusCode != http.StatusCreated && statusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code %d when creating SSO group mapping", statusCode)
	}

	return nil
}

func (ac *AtlanClient) DeleteSSOMapping(ctx context.Context, teamName string) error {
	mappingID, err := ac.FindSSOMapping(ctx, teamName)
	if err != nil {
		return err
	}

	provider := ac.identityProviderAlias
	if provider == "" {
		return fmt.Errorf("identity provider alias not configured - please set identity_provider_alias in backend connection config")
	}

	url := fmt.Sprintf("%s/api/service/idp/%s/mappers/%s/delete", ac.url, provider, mappingID)
	_, statusCode, err := ac.sendRequest(ctx, url, http.MethodPost, nil, nil, "DeleteSSOMapping")
	if err != nil {
		return fmt.Errorf("failed to delete SSO group mapping: %w", err)
	}

	if statusCode != http.StatusOK && statusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status code %d when deleting SSO group mapping", statusCode)
	}

	return nil
}

func (ac *AtlanClient) DeleteTeamByID(ctx context.Context, teamID string) error {
	log := logger.Logger(ctx).WithFields(logrus.Fields{
		"service": "atlan",
		"team_id": teamID,
	})
	log.Info("deleting team from Atlan")

	url := fmt.Sprintf("%s/api/service/groups/%s", ac.url, teamID)
	_, statusCode, err := ac.sendRequest(ctx, url, http.MethodDelete, nil, nil, "DeleteTeamByID")
	if err != nil {
		return fmt.Errorf("failed to delete team from Atlan: %w", err)
	}

	if statusCode != http.StatusOK && statusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status code %d when deleting team from Atlan", statusCode)
	}

	log.Info("successfully deleted team from Atlan")
	return nil
}

// This function pulls all SSO mappings and returns the ID of the mapping for the given team name. This needs to be added to the Cache later on as an enhancement
func (ac *AtlanClient) FindSSOMapping(ctx context.Context, teamName string) (string, error) {
	provider := ac.identityProviderAlias
	if provider == "" {
		return "", fmt.Errorf("identity provider alias not configured - please set identity_provider_alias in backend connection config")
	}

	url := fmt.Sprintf("%s/api/service/idp/%s/mappers", ac.url, provider)
	response, statusCode, err := ac.sendRequest(ctx, url, http.MethodGet, nil, nil, "FindSSOMapping")
	if err != nil {
		return "", fmt.Errorf("failed to get SSO mappings: %w", err)
	}

	if statusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code %d when getting SSO mappings", statusCode)
	}

	var mappings []SSOGroupMapping
	if err := json.Unmarshal(response, &mappings); err != nil {
		return "", fmt.Errorf("failed to parse SSO mappings response: %w", err)
	}

	for _, mapping := range mappings {
		if mapping.Config.Group == teamName {
			return mapping.ID, nil
		}
	}

	return "", fmt.Errorf("SSO mapping not found for team: %s", teamName)
}

func (ac *AtlanClient) UpdateSSOMapping(ctx context.Context, teamName, newSSOGroupName string) error {
	mappingID, err := ac.FindSSOMapping(ctx, teamName)
	if err != nil {
		return err
	}

	provider := ac.identityProviderAlias
	if provider == "" {
		return fmt.Errorf("identity provider alias not configured - please set identity_provider_alias in backend connection config")
	}

	groupMapping := SSOGroupMapping{
		ID:                     mappingID,
		IdentityProviderAlias:  provider,
		IdentityProviderMapper: "saml-group-idp-mapper",
		Name:                   fmt.Sprintf("%s--%d", mappingID, time.Now().UnixMilli()),
		Config: SSOGroupMappingConfig{
			SyncMode:                "FORCE",
			Attributes:              "[]",
			AreAttributeValuesRegex: "",
			AttributeName:           "memberOf",
			Group:                   teamName,
			AttributeValue:          newSSOGroupName,
		},
	}

	url := fmt.Sprintf("%s/api/service/idp/%s/mappers/%s", ac.url, provider, mappingID)
	_, statusCode, err := ac.sendRequest(ctx, url, http.MethodPost, groupMapping, nil, "UpdateSSOMapping")
	if err != nil {
		return fmt.Errorf("failed to update SSO group mapping: %w", err)
	}

	if statusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code %d when updating SSO group mapping", statusCode)
	}

	return nil
}

func (ac *AtlanClient) FetchTeamDetails(ctx context.Context, teamID string) (*structs.Team, error) {
	return nil, fmt.Errorf("FetchTeamDetails is not implemented for Atlan")
}
