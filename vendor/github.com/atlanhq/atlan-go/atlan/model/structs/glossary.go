package structs

type AtlasGlossaryAttributes AtlasGlossary

type AtlasGlossary struct {
	Relation
	Asset
	ShortDescription     *string                  `json:"shortDescription,omitempty"`
	LongDescription      *string                  `json:"longDescription,omitempty"`
	Language             *string                  `json:"language,omitempty"`
	Usage                *string                  `json:"usage,omitempty"`
	AdditionalAttributes *map[string]string       `json:"additionalAttributes,omitempty"`
	Terms                *[]AtlasGlossaryTerm     `json:"terms,omitempty"`
	Categories           *[]AtlasGlossaryCategory `json:"categories,omitempty"`
}
