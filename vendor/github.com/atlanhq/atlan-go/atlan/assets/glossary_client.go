package assets

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/atlanhq/atlan-go/atlan"
	"github.com/atlanhq/atlan-go/atlan/model/structs"
)

const (
	MaxRetries    = 5
	RetryInterval = time.Second * 5
)

type AtlasGlossary structs.AtlasGlossary

// Creator is used to create a new glossary asset in memory.
func (ag *AtlasGlossary) Creator(name string, icon atlan.AtlanIcon) {
	ag.TypeName = structs.StringPtr("AtlasGlossary")
	ag.Name = structs.StringPtr(name)
	ag.QualifiedName = structs.StringPtr(name)
	ag.AssetIcon = atlan.AtlanIconPtr(icon)
}

// Updater is used to modify a glossary asset in memory.
func (ag *AtlasGlossary) Updater(name string, qualifiedName string, glossary_guid string) error {
	if name == "" || qualifiedName == "" || glossary_guid == "" {
		return errors.New("name, qualified_name, and glossary_guid are required fields")
	}

	ag.TypeName = structs.StringPtr("AtlasGlossary")
	ag.Name = structs.StringPtr(name)
	ag.Guid = structs.StringPtr(glossary_guid)
	ag.QualifiedName = structs.StringPtr(qualifiedName)

	return nil
}

func (ag *AtlasGlossary) UnmarshalJSON(data []byte) error {
	Attributes := struct {
		Name             *string          `json:"name"`
		QualifiedName    *string          `json:"qualifiedName"`
		AssetIcon        *atlan.AtlanIcon `json:"assetIcon"`
		ShortDescription *string          `json:"shortDescription"`
		LongDescription  *string          `json:"longDescription"`
		Language         *string          `json:"language"`
		Usage            *string          `json:"usage"`
		// Add other attributes as necessary.
	}{}

	base, err := UnmarshalBaseEntity(data, &Attributes)
	if err != nil {
		return err
	}

	// Map Shared Fields
	ag.TypeName = &base.Entity.TypeName
	ag.Guid = &base.Entity.Guid
	ag.IsIncomplete = &base.Entity.IsIncomplete
	ag.Status = &base.Entity.Status
	ag.CreatedBy = &base.Entity.CreatedBy
	ag.UpdatedBy = &base.Entity.UpdatedBy
	ag.CreateTime = &base.Entity.CreateTime
	ag.UpdateTime = &base.Entity.UpdateTime

	// Map Attribute fields
	ag.Name = Attributes.Name
	ag.QualifiedName = Attributes.QualifiedName
	ag.ShortDescription = Attributes.ShortDescription
	ag.LongDescription = Attributes.LongDescription
	ag.Language = Attributes.Language
	ag.Usage = Attributes.Usage
	ag.AssetIcon = Attributes.AssetIcon

	return nil
}

// MarshalJSON filters out entities to only include those with non-empty attributes.
func (ag *AtlasGlossary) MarshalJSON() ([]byte, error) {
	// Construct the custom JSON structure
	customJSON := map[string]interface{}{
		"typeName": "AtlasGlossary",
		"attributes": map[string]interface{}{
			"name": ag.Name,
			// Add other attributes as necessary.
		},
		"relationshipAttributes": make(map[string]interface{}),
	}

	attributes := customJSON["attributes"].(map[string]interface{})

	if ag.QualifiedName != nil && *ag.QualifiedName != "" {
		attributes["qualifiedName"] = *ag.QualifiedName
	}

	if ag.Guid != nil && *ag.Guid != "" {
		customJSON["guid"] = *ag.Guid
	}

	if ag.DisplayName != nil && *ag.DisplayName != "" {
		attributes["displayName"] = *ag.DisplayName
	}
	if ag.Description != nil && *ag.Description != "" {
		attributes["description"] = *ag.Description
	}
	if ag.AnnouncementType != nil {
		attributes["announcementType"] = *ag.AnnouncementType
	}
	if ag.AnnouncementTitle != nil && *ag.AnnouncementTitle != "" {
		attributes["announcementTitle"] = *ag.AnnouncementTitle
	}
	if ag.AnnouncementMessage != nil && *ag.AnnouncementMessage != "" {
		attributes["announcementMessage"] = *ag.AnnouncementMessage
	}
	if ag.CertificateStatus != nil {
		attributes["certificateStatus"] = *ag.CertificateStatus
	}

	// Marshal the custom JSON
	return json.MarshalIndent(customJSON, "", "  ")
}

func (ag *AtlasGlossary) ToJSON() ([]byte, error) {
	return json.MarshalIndent(ag, "", "  ")
}

func (ag *AtlasGlossary) FromJSON(data []byte) error {
	return json.Unmarshal(data, ag)
}
