package assets

import (
	"encoding/json"
	"fmt"

	"github.com/atlanhq/atlan-go/atlan"
	"github.com/atlanhq/atlan-go/atlan/model"
	"github.com/atlanhq/atlan-go/atlan/model/structs"
)

type Persona structs.Persona

// Creator is used to create a new persona asset in memory.
func (p *Persona) Creator(name string) {
	p.TypeName = structs.StringPtr("Persona")
	p.Name = structs.StringPtr(name)
}

/*
	Example Usage :
	persona := &Persona{}
	policy, err := persona.CreateMetadataPolicy(
		"MyPolicy",
		"persona-guid-1234",
		atlan.AuthPolicyTypeMetadata,
		[]atlan.PersonaMetadataAction{
			atlan.PersonaMetadataActionRead,
			atlan.PersonaMetadataActionWrite,
		},
		"connection-guid-5678",
		[]string{"resource1", "resource2"},

	)
*/

// CreateMetadataPolicy creates a new metadata policy for a persona.
func (p *Persona) CreateMetadataPolicy(
	name string,
	personaID string,
	policyType atlan.AuthPolicyType,
	actions []atlan.PersonaMetadataAction,
	connectionQualifiedName string,
	resources []string,
) (*AuthPolicy, error) {
	// Convert actions to their string values
	var policyActions []string
	for _, action := range actions {
		policyActions = append(policyActions, action.String())
	}

	// Create the policy object
	policy := &AuthPolicy{
		Asset: structs.Asset{
			Name:                    structs.StringPtr(name),
			ConnectionQualifiedName: &connectionQualifiedName,
			Referenceable: structs.Referenceable{
				Guid: &personaID,
			},
		},
		PolicyType:             &policyType,
		PolicyCategory:         structs.StringPtr(atlan.AuthPolicyCategoryPersona.String()),
		PolicyResources:        &resources,
		PolicyActions:          &policyActions,
		PolicyResourceCategory: structs.StringPtr(atlan.AuthPolicyResourceCategoryCustom.String()),
		PolicyServiceName:      structs.StringPtr("atlas"),
		PolicySubCategory:      structs.StringPtr("metadata"),
		AccessControl: &structs.AccessControl{
			Asset: structs.Asset{
				Referenceable: structs.Referenceable{
					Guid:     &personaID,
					TypeName: structs.StringPtr("Persona"),
				},
			},
		},
	}
	return policy, nil
}

// CreateDataPolicy creates a new data policy for a persona.
func (p *Persona) CreateDataPolicy(
	name string,
	personaID string,
	policyType atlan.AuthPolicyType,
	connectionQualifiedName string,
	resources []string,
) (*AuthPolicy, error) {
	// Add "entity-type:*" to resources
	resources = append(resources, "entity-type:*")

	// Define default policy actions
	policyActions := []string{atlan.DataActionSelect.String()}

	// Create the policy object
	policy := &AuthPolicy{
		Asset: structs.Asset{
			ConnectionQualifiedName: &connectionQualifiedName,
			Name:                    &name,
			Referenceable: structs.Referenceable{
				Guid: &personaID,
			},
		},
		PolicyType:             &policyType,
		PolicyCategory:         structs.StringPtr(atlan.AuthPolicyCategoryPersona.String()),
		PolicyActions:          &policyActions,
		PolicyResources:        &resources,
		PolicyResourceCategory: structs.StringPtr(atlan.AuthPolicyResourceCategoryEntity.String()),
		PolicyServiceName:      structs.StringPtr("heka"),
		PolicySubCategory:      structs.StringPtr("data"),
		AccessControl: &structs.AccessControl{
			Asset: structs.Asset{
				Referenceable: structs.Referenceable{
					Guid: &personaID,
				},
			},
		},
	}

	return policy, nil
}

// CreateGlossaryPolicy creates a new glossary policy for a persona.
func (p *Persona) CreateGlossaryPolicy(
	name string,
	personaID string,
	policyType atlan.AuthPolicyType,
	actions []atlan.PersonaGlossaryAction,
	resources []string,
) (*AuthPolicy, error) {
	// Convert actions to their string values
	policyActions := make([]string, len(actions))
	for i, action := range actions {
		policyActions[i] = action.String()
	}

	// Create the policy object
	policy := &AuthPolicy{
		Asset: structs.Asset{
			Name: &name,
			Referenceable: structs.Referenceable{
				Guid: &personaID,
			},
		},
		PolicyType:             &policyType,
		PolicyCategory:         structs.StringPtr(atlan.AuthPolicyCategoryPersona.String()),
		PolicyActions:          &policyActions,
		PolicyResources:        &resources,
		PolicyResourceCategory: structs.StringPtr(atlan.AuthPolicyResourceCategoryCustom.String()),
		PolicyServiceName:      structs.StringPtr("atlas"),
		PolicySubCategory:      structs.StringPtr("glossary"),
		AccessControl: &structs.AccessControl{
			Asset: structs.Asset{
				Referenceable: structs.Referenceable{
					Guid: &personaID,
				},
			},
		},
	}

	return policy, nil
}

// CreateDomainPolicy creates a new domain policy for a persona.
func (p *Persona) CreateDomainPolicy(
	name string,
	personaID string,
	actions []atlan.PersonaDomainAction,
	resources []string,
) (*AuthPolicy, error) {
	// Convert actions to their string values
	policyActions := make([]string, len(actions))
	for i, action := range actions {
		policyActions[i] = action.String()
	}

	// Create the policy object
	policy := &AuthPolicy{
		Asset: structs.Asset{
			Name: &name,
			Referenceable: structs.Referenceable{
				Guid: &personaID,
			},
		},
		PolicyType:             &atlan.AuthPolicyTypeAllow,
		PolicyCategory:         structs.StringPtr(atlan.AuthPolicyCategoryPersona.String()),
		PolicyActions:          &policyActions,
		PolicyResources:        &resources,
		PolicyResourceCategory: structs.StringPtr(atlan.AuthPolicyResourceCategoryCustom.String()),
		PolicyServiceName:      structs.StringPtr("atlas"),
		PolicySubCategory:      structs.StringPtr("domain"),
		AccessControl: &structs.AccessControl{
			Asset: structs.Asset{
				Referenceable: structs.Referenceable{
					Guid: &personaID,
				},
			},
		},
	}

	return policy, nil
}

func (p *Persona) Updater(qualifiedName, name string, isEnabled bool) error {
	// Validate required fields
	if qualifiedName == "" || name == "" {
		return fmt.Errorf("missing required fields: qualifiedName and name")
	}
	p.QualifiedName = &qualifiedName
	p.Name = &name
	p.IsAccessControlEnabled = &isEnabled
	return nil
}

func FindPersonasByName(name string) (*model.IndexSearchResponse, error) {
	if name == "" {
		return nil, fmt.Errorf("name cannot be empty")
	}
	boolQuery, err := WithActivePersona(name)
	if err != nil {
		return nil, err
	}
	pageSize := 20

	request := model.IndexSearchRequest{
		Dsl: model.Dsl{
			From:           0,
			Size:           pageSize,
			Query:          boolQuery.ToJSON(),
			TrackTotalHits: true,
		},
		SuppressLogs:     true,
		ShowSearchScore:  false,
		ExcludeMeanings:  false,
		ExcludeAtlanTags: false,
	}

	iterator := NewIndexSearchIterator(pageSize, request)

	for iterator.HasMoreResults() {
		response, err := iterator.NextPage()
		if err != nil {
			return nil, fmt.Errorf("error executing search: %v", err)
		}
		fmt.Println("Current Page: ", iterator.CurrentPageNumber())
		for _, entity := range response.Entities {
			if *entity.TypeName == "Persona" {
				return response, err
			}
		}
	}
	return nil, nil
}

func (p *Persona) UnmarshalJSON(data []byte) error {
	// Unmarshal shared fields and specific attributes.
	attributes := struct {
		// Attributes
		QualifiedName *string `json:"qualifiedName,omitempty"`
		Name          *string `json:"name"`

		// Persona Attributes
		PersonaGroups []string `json:"personaGroups"`
		PersonaUsers  []string `json:"personaUsers"`
		RoleId        *string  `json:"roleId"`

		// Access Control Attributes
		IsAccessControlEnabled  *bool         `json:"isAccessControlEnabled,omitempty"`
		DenyCustomMetadataGuids *[]string     `json:"denyCustomMetadataGuids,omitempty"`
		DenyAssetTabs           *[]string     `json:"denyAssetTabs,omitempty"`
		DenyAssetFilters        *[]string     `json:"denyAssetFilters,omitempty"`
		ChannelLink             *string       `json:"channelLink,omitempty"`
		DenyAssetTypes          *[]string     `json:"denyAssetTypes,omitempty"`
		DenyNavigationPages     *[]string     `json:"denyNavigationPages,omitempty"`
		DefaultNavigation       *string       `json:"defaultNavigation,omitempty"`
		DisplayPreferences      *[]string     `json:"displayPreferences,omitempty"`
		Policies                *[]AuthPolicy `json:"policies,omitempty"`
	}{}
	base, err := UnmarshalBaseEntity(data, &attributes)
	if err != nil {
		return err
	}

	// Map shared fields.
	p.Guid = &base.Entity.Guid
	p.TypeName = &base.Entity.TypeName
	p.IsIncomplete = &base.Entity.IsIncomplete
	p.Status = &base.Entity.Status
	p.CreatedBy = &base.Entity.CreatedBy
	p.UpdatedBy = &base.Entity.UpdatedBy
	p.CreateTime = &base.Entity.CreateTime
	p.UpdateTime = &base.Entity.UpdateTime

	p.QualifiedName = attributes.QualifiedName
	p.Name = attributes.Name

	// Map Persona-specific fields.
	p.PersonaGroups = &attributes.PersonaGroups
	p.PersonaUsers = &attributes.PersonaUsers
	p.RoleId = attributes.RoleId

	// Access Controls
	p.IsAccessControlEnabled = attributes.IsAccessControlEnabled
	p.DenyCustomMetadataGuids = attributes.DenyCustomMetadataGuids
	p.DenyAssetTabs = attributes.DenyAssetTabs
	p.DenyAssetFilters = attributes.DenyAssetFilters
	p.ChannelLink = attributes.ChannelLink
	p.DenyAssetTypes = attributes.DenyAssetTypes
	p.DenyNavigationPages = attributes.DenyNavigationPages
	p.DefaultNavigation = attributes.DefaultNavigation
	p.DisplayPreferences = attributes.DisplayPreferences

	// Unmarshal RelationshipAttributes for Policies
	if base.Entity.RelationshipAttributes != nil {
		relationshipAttributes := struct {
			Policies []structs.AuthPolicy `json:"policies,omitempty"`
		}{}

		if err := json.Unmarshal(base.Entity.RelationshipAttributes, &relationshipAttributes); err != nil {
			return err
		}
		// Map the Policies field
		p.Policies = &relationshipAttributes.Policies
	}

	return nil
}

func (p *Persona) MarshalJSON() ([]byte, error) {
	// Marshal the AccessControl asset into a JSON object.

	// Construct the custom JSON structure
	customJSON := map[string]interface{}{
		"typeName": "Persona",
		"attributes": map[string]interface{}{
			"name":          p.Name,
			"qualifiedName": p.Name,
			//		"displayName":            p.Name,
			"isAccessControlEnabled": true,
			// Add other attributes as necessary.
		},
	}

	attributes := customJSON["attributes"].(map[string]interface{})

	if p.IsAccessControlEnabled != nil {
		attributes["isAccessControlEnabled"] = *p.IsAccessControlEnabled
	}

	if p.QualifiedName != nil && *p.QualifiedName != "" {
		attributes["qualifiedName"] = *p.QualifiedName
	}

	if p.Guid != nil && *p.Guid != "" {
		customJSON["guid"] = *p.Guid
	}

	if p.DisplayName != nil && *p.DisplayName != "" {
		attributes["displayName"] = *p.DisplayName
	}

	if p.Description != nil && *p.Description != "" {
		attributes["description"] = *p.Description
	}

	if p.PersonaUsers != nil {
		attributes["personaUsers"] = p.PersonaUsers
	}

	if p.PersonaGroups != nil {
		attributes["personaGroups"] = p.PersonaGroups
	}
	// Add access control attributes
	if p.IsAccessControlEnabled != nil {
		attributes["isAccessControlEnabled"] = *p.IsAccessControlEnabled
	}
	if p.DenyCustomMetadataGuids != nil {
		attributes["denyCustomMetadataGuids"] = *p.DenyCustomMetadataGuids
	}
	if p.DenyAssetTabs != nil {
		attributes["denyAssetTabs"] = *p.DenyAssetTabs
	}
	if p.DenyAssetFilters != nil {
		attributes["denyAssetFilters"] = *p.DenyAssetFilters
	}
	if p.ChannelLink != nil {
		attributes["channelLink"] = *p.ChannelLink
	}
	if p.DenyAssetTypes != nil {
		attributes["denyAssetTypes"] = *p.DenyAssetTypes
	}
	if p.DenyNavigationPages != nil {
		attributes["denyNavigationPages"] = *p.DenyNavigationPages
	}
	if p.DefaultNavigation != nil {
		attributes["defaultNavigation"] = *p.DefaultNavigation
	}
	if p.DisplayPreferences != nil {
		attributes["displayPreferences"] = *p.DisplayPreferences
	}
	if p.Policies != nil {
		attributes["policies"] = p.Policies // Assuming proper JSON marshalling of structs.AuthPolicy
	}
	if p.RoleId != nil {
		attributes["roleId"] = *p.RoleId
	}

	return json.MarshalIndent(customJSON, "", "  ")
}

func (p *Persona) ToJSON() ([]byte, error) {
	// Marshal the Persona object into a JSON object.
	return json.Marshal(p)
}

func (p *Persona) FromJSON(data []byte) error {
	// Unmarshal the JSON data into the Persona object.
	return json.Unmarshal(data, p)
}
