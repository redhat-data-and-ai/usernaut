package structs

import (
	"encoding/json"

	"github.com/atlanhq/atlan-go/atlan"
)

type _Glossary AtlasGlossary

type AtlasGlossaryTerm struct {
	Asset
	ShortDescription     *string              `json:"shortDescription,omitempty"`
	LongDescription      *string              `json:"longDescription,omitempty"`
	Example              *string              `json:"example,omitempty"`
	Abbreviation         *string              `json:"abbreviation,omitempty"`
	Usage                *string              `json:"usage,omitempty"`
	AdditionalAttributes *map[string]string   `json:"additionalAttributes,omitempty"`
	Anchor               *AtlasGlossary       `json:"anchor,omitempty"`
	Antonyms             *[]AtlasGlossaryTerm `json:"antonyms,omitempty"`
	AssignedEntities     *[]Asset             `json:"assignedEntities,omitempty"`
	Categories           *[]AtlasGlossaryTerm `json:"categories,omitempty"`
	ValidValuesFor       *[]AtlasGlossaryTerm `json:"validValuesFor,omitempty"`
	ValidValues          *[]AtlasGlossaryTerm `json:"validValues,omitempty"`
	SeeAlso              *[]AtlasGlossaryTerm `json:"seeAlso,omitempty"`
	IsA                  *[]AtlasGlossaryTerm `json:"isA,omitempty"`
	CLASSIFIES           *[]AtlasGlossaryTerm `json:"classifies,omitempty"`
	PreferredToTerms     *[]AtlasGlossaryTerm `json:"preferredToTerms,omitempty"`
	PreferredTerms       *[]AtlasGlossaryTerm `json:"preferredTerms,omitempty"`
	TranslationTerms     *[]AtlasGlossaryTerm `json:"translationTerms,omitempty"`
	Synonyms             *[]AtlasGlossaryTerm `json:"synonyms,omitempty"`
	ReplacedBy           *[]AtlasGlossaryTerm `json:"replacedBy,omitempty"`
	ReplacementTerms     *[]AtlasGlossaryTerm `json:"replacementTerms,omitempty"`
	TranslatedTerms      *[]AtlasGlossaryTerm `json:"translatedTerms,omitempty"`
}

type AtlasGlossaryTermAttributes struct {
	Asset
	ShortDescription     *string              `json:"shortDescription,omitempty"`
	LongDescription      *string              `json:"longDescription,omitempty"`
	Example              *string              `json:"example,omitempty"`
	Abbreviation         *string              `json:"abbreviation,omitempty"`
	Usage                *string              `json:"usage,omitempty"`
	AdditionalAttributes *map[string]string   `json:"additionalAttributes,omitempty"`
	Anchor               *_Glossary           `json:"anchor,omitempty"`
	Antonyms             *[]AtlasGlossaryTerm `json:"antonyms,omitempty"`
	AssignedEntities     *[]Asset             `json:"assignedEntities,omitempty"`
	Categories           *[]AtlasGlossaryTerm `json:"categories,omitempty"`
	ValidValuesFor       *[]AtlasGlossaryTerm `json:"validValuesFor,omitempty"`
	ValidValues          *[]AtlasGlossaryTerm `json:"validValues,omitempty"`
	SeeAlso              *[]AtlasGlossaryTerm `json:"seeAlso,omitempty"`
	IsA                  *[]AtlasGlossaryTerm `json:"isA,omitempty"`
	CLASSIFIES           *[]AtlasGlossaryTerm `json:"classifies,omitempty"`
	PreferredToTerms     *[]AtlasGlossaryTerm `json:"preferredToTerms,omitempty"`
	PreferredTerms       *[]AtlasGlossaryTerm `json:"preferredTerms,omitempty"`
	TranslationTerms     *[]AtlasGlossaryTerm `json:"translationTerms,omitempty"`
	Synonyms             *[]AtlasGlossaryTerm `json:"synonyms,omitempty"`
	ReplacedBy           *[]AtlasGlossaryTerm `json:"replacedBy,omitempty"`
	ReplacementTerms     *[]AtlasGlossaryTerm `json:"replacementTerms,omitempty"`
	TranslatedTerms      *[]AtlasGlossaryTerm `json:"translatedTerms,omitempty"`
}

// UnmarshalJSON unmarshals the JSON data into a GlossaryTerm object.
func (gt *AtlasGlossaryTerm) UnmarshalJSON(data []byte) error {
	var temp struct {
		Entity struct {
			TypeName               string                      `json:"typeName"`
			Attributes             AtlasGlossaryTermAttributes `json:"attributes"`
			Guid                   string                      `json:"guid"`
			IsIncomplete           bool                        `json:"isIncomplete"`
			Status                 string                      `json:"status"`
			CreatedBy              string                      `json:"createdBy"`
			UpdatedBy              string                      `json:"updatedBy"`
			CreateTime             int64                       `json:"createTime"`
			UpdateTime             int64                       `json:"updateTime"`
			Version                int                         `json:"version"`
			RelationshipAttributes AtlasGlossaryTermAttributes `json:"relationshipAttributes"`
			Tags                   []AtlanTag                  `json:"classifications"`
			Labels                 []interface{}               `json:"labels"`
		} `json:"entity"`
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	// Copy fields
	gt.TypeName = &temp.Entity.TypeName
	gt.ShortDescription = temp.Entity.Attributes.ShortDescription
	gt.LongDescription = temp.Entity.Attributes.LongDescription
	gt.Example = temp.Entity.Attributes.Example
	gt.Abbreviation = temp.Entity.Attributes.Abbreviation
	gt.Usage = temp.Entity.Attributes.Usage
	gt.AdditionalAttributes = temp.Entity.Attributes.AdditionalAttributes
	gt.Guid = &temp.Entity.Guid
	gt.Status = (*atlan.AtlanStatus)(&temp.Entity.Status)
	gt.CreatedBy = &temp.Entity.CreatedBy
	gt.UpdatedBy = &temp.Entity.UpdatedBy
	gt.CreateTime = &temp.Entity.CreateTime
	gt.UpdateTime = &temp.Entity.UpdateTime
	gt.IsIncomplete = &temp.Entity.IsIncomplete

	gt.Anchor = (*AtlasGlossary)(temp.Entity.Attributes.Anchor)

	gt.McMonitors = temp.Entity.RelationshipAttributes.McMonitors
	gt.AtlanTags = &temp.Entity.Tags

	return nil
}

func (ag *AtlasGlossaryTerm) FromJSON(data []byte) error {
	return json.Unmarshal(data, ag)
}

func (gt *AtlasGlossaryTerm) ToJSON() ([]byte, error) {
	return json.Marshal(gt)
}
