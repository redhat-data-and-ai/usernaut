package assets

import (
	"encoding/json"
	"errors"

	"github.com/atlanhq/atlan-go/atlan"
	"github.com/atlanhq/atlan-go/atlan/model/structs"
)

type AuthPolicy structs.AuthPolicy

func (a *AuthPolicy) Updater(name string, qualifiedName string) error {
	if name == "" || qualifiedName == "" {
		return errors.New("name, qualifiedName are required fields")
	}

	a.Name = &name
	a.QualifiedName = &qualifiedName

	return nil
}

// UnmarshalJSON unmarshal a AuthPolicy from JSON.
func (a *AuthPolicy) UnmarshalJSON(data []byte) error {
	attributes := struct {
		// Base attributes
		QualifiedName *string `json:"qualifiedName,omitempty"`
		Name          *string `json:"name,omitempty"`

		// AuthPolicy specific attributes
		PolicyType              *atlan.AuthPolicyType               `json:"policyType,omitempty"`
		PolicyServiceName       *string                             `json:"policyServiceName,omitempty"`
		PolicyCategory          *string                             `json:"policyCategory,omitempty"`
		PolicySubCategory       *string                             `json:"policySubCategory,omitempty"`
		PolicyUsers             *[]string                           `json:"policyUsers,omitempty"`
		PolicyGroups            *[]string                           `json:"policyGroups,omitempty"`
		PolicyRoles             *[]string                           `json:"policyRoles,omitempty"`
		PolicyActions           *[]string                           `json:"policyActions,omitempty"`
		PolicyResources         *[]string                           `json:"policyResources,omitempty"`
		PolicyResourceCategory  *string                             `json:"policyResourceCategory,omitempty"`
		PolicyPriority          *int                                `json:"policyPriority,omitempty"`
		IsPolicyEnabled         *bool                               `json:"isPolicyEnabled,omitempty"`
		PolicyMaskType          *atlan.DataMaskingType              `json:"policyMaskType,omitempty"`
		PolicyValiditySchedule  *[]atlan.AuthPolicyValiditySchedule `json:"policyValiditySchedule,omitempty"`
		PolicyResourceSignature *string                             `json:"policyResourceSignature,omitempty"`
		PolicyDelegateAdmin     *bool                               `json:"policyDelegateAdmin,omitempty"`
		PolicyConditions        *[]atlan.AuthPolicyCondition        `json:"policyConditions,omitempty"`
		AccessControl           *structs.AccessControl              `json:"accessControl,omitempty"` // Relationship
	}{}

	// Unmarshal Base attributes
	base, err := UnmarshalBaseEntity(data, &attributes)
	if err != nil {
		return err
	}

	// Map base entity fields.
	a.Guid = &base.Entity.Guid
	a.TypeName = &base.Entity.TypeName
	a.IsIncomplete = &base.Entity.IsIncomplete
	a.Status = &base.Entity.Status
	a.CreatedBy = &base.Entity.CreatedBy
	a.UpdatedBy = &base.Entity.UpdatedBy
	a.CreateTime = &base.Entity.CreateTime
	a.UpdateTime = &base.Entity.UpdateTime

	// Map AuthPolicy specific attributes to AuthPolicy fields.
	a.UniqueAttributes.QualifiedName = attributes.QualifiedName
	a.Name = attributes.Name
	a.PolicyType = attributes.PolicyType
	a.PolicyServiceName = attributes.PolicyServiceName
	a.PolicyCategory = attributes.PolicyCategory
	a.PolicySubCategory = attributes.PolicySubCategory
	a.PolicyUsers = attributes.PolicyUsers
	a.PolicyGroups = attributes.PolicyGroups
	a.PolicyRoles = attributes.PolicyRoles
	a.PolicyActions = attributes.PolicyActions
	a.PolicyResources = attributes.PolicyResources
	a.PolicyResourceCategory = attributes.PolicyResourceCategory
	a.PolicyPriority = attributes.PolicyPriority
	a.IsPolicyEnabled = attributes.IsPolicyEnabled
	a.PolicyMaskType = attributes.PolicyMaskType
	a.PolicyValiditySchedule = attributes.PolicyValiditySchedule
	a.PolicyResourceSignature = attributes.PolicyResourceSignature
	a.PolicyDelegateAdmin = attributes.PolicyDelegateAdmin
	a.PolicyConditions = attributes.PolicyConditions
	a.AccessControl = attributes.AccessControl

	return nil
}

// MarshalJSON Marshals the AuthPolicy asset into a JSON object.
func (a *AuthPolicy) MarshalJSON() ([]byte, error) {
	// Construct the custom JSON structure
	customJSON := map[string]interface{}{
		"typeName": "AuthPolicy",
		"attributes": map[string]interface{}{
			"name":          a.Name,
			"qualifiedName": a.Name,
			// Add other attributes as necessary.
		},
	}

	attributes := customJSON["attributes"].(map[string]interface{})

	if a.QualifiedName != nil && *a.QualifiedName != "" {
		attributes["qualifiedName"] = *a.QualifiedName
	}

	//	if a.Guid != nil && *a.Guid != "" {
	//		customJSON["guid"] = *a.Guid
	//	}

	if a.DisplayName != nil && *a.DisplayName != "" {
		attributes["displayName"] = *a.DisplayName
	}

	if a.PolicyType != nil {
		attributes["policyType"] = *a.PolicyType
	}

	if a.PolicyCategory != nil && *a.PolicyCategory != "" {
		attributes["policyCategory"] = *a.PolicyCategory
	}

	if a.PolicyResources != nil {
		attributes["policyResources"] = *a.PolicyResources
	}

	if a.PolicyActions != nil {
		attributes["policyActions"] = *a.PolicyActions
	}

	if a.PolicyResourceCategory != nil && *a.PolicyResourceCategory != "" {
		attributes["policyResourceCategory"] = *a.PolicyResourceCategory
	}

	if a.PolicyServiceName != nil && *a.PolicyServiceName != "" {
		attributes["policyServiceName"] = *a.PolicyServiceName
	}

	if a.PolicySubCategory != nil && *a.PolicySubCategory != "" {
		attributes["policySubCategory"] = *a.PolicySubCategory
	}

	if a.ConnectionQualifiedName != nil && *a.ConnectionQualifiedName != "" {
		attributes["connectionQualifiedName"] = *a.ConnectionQualifiedName
	}

	if a.PolicyGroups != nil && len(*a.PolicyGroups) > 0 {
		attributes["policyGroups"] = *a.PolicyGroups
	}

	if a.PolicyUsers != nil && len(*a.PolicyUsers) > 0 {
		attributes["policyUsers"] = *a.PolicyUsers
	}

	if a.PolicyMaskType != nil {
		attributes["policyMaskType"] = *a.PolicyMaskType
	}

	// Handle nested AccessControl field
	if a.AccessControl != nil {
		accessControl := map[string]interface{}{}

		if a.AccessControl.Guid != nil && *a.AccessControl.Guid != "" {
			accessControl["guid"] = *a.AccessControl.Guid
		}

		if a.AccessControl.TypeName != nil && *a.AccessControl.TypeName != "" {
			accessControl["typeName"] = *a.AccessControl.TypeName
		}

		if a.AccessControl.UniqueAttributes.QualifiedName != nil && *a.AccessControl.UniqueAttributes.QualifiedName != "" {
			accessControl["uniqueAttributes"] = map[string]interface{}{
				"qualifiedName": *a.AccessControl.UniqueAttributes.QualifiedName,
			}
		}

		attributes["accessControl"] = accessControl
	}

	return json.MarshalIndent(customJSON, "", "  ")
}

func (a *AuthPolicy) ToJSON() ([]byte, error) {
	// Marshal the Persona object into a JSON object.
	return json.Marshal(a)
}

func (a *AuthPolicy) FromJSON(data []byte) error {
	// Unmarshal the JSON data into the Persona object.
	return json.Unmarshal(data, a)
}
