package assets

import (
	"encoding/json"
	"fmt"

	"github.com/atlanhq/atlan-go/atlan"
	"github.com/atlanhq/atlan-go/atlan/model"
	"github.com/atlanhq/atlan-go/atlan/model/structs"
)

type Purpose structs.Purpose

// Creator is used to create a new purpose asset in memory.
func (p *Purpose) Creator(name string, atlanTags []string) error {
	p.TypeName = structs.StringPtr("Purpose")
	p.Name = structs.StringPtr(name)

	var atlanTagValues []structs.AtlanTagName
	for _, tag := range atlanTags {
		newTag, err := NewAtlanTagName(tag)
		if err != nil {
			return fmt.Errorf("failed to create AtlanTagName for %s: %w", tag, err)
		}
		atlanTagValues = append(atlanTagValues, *newTag)
	}

	// Assign the converted tags to PurposeAttributes
	p.Attributes = &structs.PurposeAttributes{
		PurposeAtlanTags: &atlanTagValues,
	}
	return nil
}

/*
	Example Usage:
	purpose := &Purpose{}
	policy, err := purpose.CreateMetadataPolicy(
		"MyPolicy",
		"purpose-guid-1234",
		atlan.AuthPolicyTypeMetadata,
		[]atlan.PurposeMetadataAction{
			atlan.PurposeMetadataActionRead,
			atlan.PurposeMetadataActionWrite,
		},
		[]string{"group1", "group2"},
		[]string{"user1", "user2"},
		true,
	)
*/

// CreateMetadataPolicy creates a new metadata policy for a purpose.
func (p *Purpose) CreateMetadataPolicy(
	name string,
	purposeID string,
	policyType atlan.AuthPolicyType,
	actions []atlan.PurposeMetadataAction,
	policyGroups []string,
	policyUsers []string,
	allUsers bool,
) (*AuthPolicy, error) {
	// Convert actions to their string values
	var policyActions []string
	for _, action := range actions {
		policyActions = append(policyActions, action.String())
	}

	// Create the policy object
	policy := &AuthPolicy{
		Asset: structs.Asset{
			Name: &name,
			Referenceable: structs.Referenceable{
				Guid: &purposeID,
			},
		},
		PolicyType:             &policyType,
		PolicyCategory:         structs.StringPtr(atlan.AuthPolicyCategoryPurpose.String()),
		PolicyActions:          &policyActions,
		PolicyResourceCategory: structs.StringPtr(atlan.AuthPolicyResourceCategoryTag.String()),
		PolicyServiceName:      structs.StringPtr("atlas_tag"),
		PolicySubCategory:      structs.StringPtr("metadata"),
		AccessControl: &structs.AccessControl{
			Asset: structs.Asset{
				Referenceable: structs.Referenceable{
					Guid:     &purposeID,
					TypeName: structs.StringPtr("Purpose"),
				},
			},
		},
	}

	// Assign groups and users to the policy
	if allUsers {
		policy.PolicyGroups = &[]string{"public"}
	} else {
		if len(policyGroups) > 0 {
			policy.PolicyGroups = &policyGroups
		}
		if len(policyUsers) > 0 {
			policy.PolicyUsers = &policyUsers
		}
	}

	if len(*policy.PolicyGroups) == 0 && len(*policy.PolicyUsers) == 0 {
		return nil, fmt.Errorf("no user or group specified for the policy")
	}

	return policy, nil
}

// CreateDataPolicy creates a new data policy for a purpose.
func (p *Purpose) CreateDataPolicy(
	name string,
	purposeID string,
	policyType atlan.AuthPolicyType,
	policyGroups []string,
	policyUsers []string,
	allUsers bool,
) (*AuthPolicy, error) {
	// Default data policy actions
	policyActions := []string{atlan.DataActionSelect.String()}

	policy := &AuthPolicy{
		Asset: structs.Asset{
			Name: &name,
			Referenceable: structs.Referenceable{
				Guid: &purposeID,
			},
		},
		PolicyType:             &policyType,
		PolicyCategory:         structs.StringPtr(atlan.AuthPolicyCategoryPurpose.String()),
		PolicyActions:          &policyActions,
		PolicyResourceCategory: structs.StringPtr(atlan.AuthPolicyResourceCategoryTag.String()),
		PolicyServiceName:      structs.StringPtr("atlas_tag"),
		PolicySubCategory:      structs.StringPtr("data"),
		AccessControl: &structs.AccessControl{
			Asset: structs.Asset{
				Referenceable: structs.Referenceable{
					Guid: &purposeID,
				},
			},
		},
	}

	// Assign groups and users to the policy
	if allUsers {
		policy.PolicyGroups = &[]string{"public"}
	} else {
		if len(policyGroups) > 0 {
			policy.PolicyGroups = &policyGroups
		}
		if len(policyUsers) > 0 {
			policy.PolicyUsers = &policyUsers
		}
	}

	if len(*policy.PolicyGroups) == 0 && len(*policy.PolicyUsers) == 0 {
		return nil, fmt.Errorf("no user or group specified for the policy")
	}

	return policy, nil
}

func (p *Purpose) Updater(qualifiedName, name string, isEnabled bool) error {
	if qualifiedName == "" || name == "" {
		return fmt.Errorf("missing required fields: qualifiedName and name")
	}
	p.QualifiedName = &qualifiedName
	p.Name = &name
	p.IsAccessControlEnabled = &isEnabled
	return nil
}

func FindPurposesByName(name string) (*model.IndexSearchResponse, error) {
	if name == "" {
		return nil, fmt.Errorf("name cannot be empty")
	}

	// Construct the boolean query for active purposes with the given name
	boolQuery, err := WithActivePurpose(name)
	if err != nil {
		return nil, err
	}

	pageSize := 20

	// Create the search request
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

	// Iterate through the pages
	for iterator.HasMoreResults() {
		response, err := iterator.NextPage()
		if err != nil {
			return nil, fmt.Errorf("error executing search: %v", err)
		}
		fmt.Println("Current Page: ", iterator.CurrentPageNumber())

		// Check each entity in the current page
		for _, entity := range response.Entities {
			if *entity.TypeName == "Purpose" {
				return response, nil
			}
		}
	}

	// If no purpose is found, return nil without an error
	return nil, nil
}

// NewAtlanTagName creates a new AtlanTagName instance, validating against the cache.
func NewAtlanTagName(displayText string) (*structs.AtlanTagName, error) {
	id, _ := GetAtlanTagIDForName(displayText)
	if id == "" {
		return nil, fmt.Errorf("%s is not a valid Classification", displayText)
	}
	return &structs.AtlanTagName{
		DisplayText: displayText,
		ID:          id,
	}, nil
}

// UnmarshalJSON unmarshal a Purpose from JSON.
func (p *Purpose) UnmarshalJSON(data []byte) error {
	attributes := struct {
		// Base attributes
		QualifiedName *string `json:"qualifiedName,omitempty"`
		Name          *string `json:"name"`

		// Purpose-specific attributes
		PurposeAtlanTags *[]structs.AtlanTagName `json:"purposeAtlanTags,omitempty"`

		// Access Control-specific Attributes
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

	// Map base fields
	p.Guid = &base.Entity.Guid
	p.TypeName = &base.Entity.TypeName
	p.IsIncomplete = &base.Entity.IsIncomplete
	p.Status = &base.Entity.Status
	p.CreatedBy = &base.Entity.CreatedBy
	p.UpdatedBy = &base.Entity.UpdatedBy
	p.CreateTime = &base.Entity.CreateTime
	p.UpdateTime = &base.Entity.UpdateTime

	// Map Purpose-specific fields
	p.QualifiedName = attributes.QualifiedName
	p.Name = attributes.Name
	p.Attributes = &structs.PurposeAttributes{
		PurposeAtlanTags: attributes.PurposeAtlanTags,
	}

	// Map Access Control Attributes
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

// MarshalJSON marshals a Purpose into JSON.
func (p *Purpose) MarshalJSON() ([]byte, error) {
	customJSON := map[string]interface{}{
		"typeName": "Purpose",
		"attributes": map[string]interface{}{
			"displayName":            p.Name,
			"name":                   p.Name,
			"qualifiedName":          p.Name,
			"isAccessControlEnabled": true,
		},
	}

	attributes := customJSON["attributes"].(map[string]interface{})

	if p.QualifiedName != nil && *p.QualifiedName != "" {
		attributes["qualifiedName"] = *p.QualifiedName
	}

	if p.IsAccessControlEnabled != nil {
		attributes["isAccessControlEnabled"] = *p.IsAccessControlEnabled
	}

	if p.Description != nil && *p.Description != "" {
		attributes["description"] = *p.Description
	}

	if p.DisplayName != nil && *p.DisplayName != "" {
		attributes["displayName"] = *p.DisplayName
	}

	if p.Guid != nil && *p.Guid != "" {
		customJSON["guid"] = *p.Guid
	}

	if p.Attributes != nil && p.Attributes.PurposeAtlanTags != nil && len(*p.Attributes.PurposeAtlanTags) > 0 {
		attributes["purposeClassifications"] = p.Attributes.PurposeAtlanTags
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

	// Add other attributes here

	return json.MarshalIndent(customJSON, "", "  ")
}

// ToJSON converts a Purpose object to JSON.
func (p *Purpose) ToJSON() ([]byte, error) {
	return json.Marshal(p)
}

// FromJSON populates a Purpose object from JSON.
func (p *Purpose) FromJSON(data []byte) error {
	return json.Unmarshal(data, p)
}
