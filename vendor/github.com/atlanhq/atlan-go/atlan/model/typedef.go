package model

import (
	"encoding/json"
	"fmt"

	"github.com/atlanhq/atlan-go/atlan"
)

type IndexType string

// CustomBool is a custom type that implements the json.Unmarshaler interface
type CustomBool bool

type TypeDef interface {
	GetCategory() atlan.AtlanTypeCategory
}

type EnumDef struct {
	TypeDef
}

type StructDef struct {
	TypeDef
}

type EntityDef struct {
	TypeDef
	TypeDefBase
}

type RelationshipDef struct {
	TypeDef
}

func (a *AtlanTagDef) GetCategory() atlan.AtlanTypeCategory {
	return a.Category
}

func (a *CustomMetadataDef) GetCategory() atlan.AtlanTypeCategory {
	return *a.Category
}

type TypeDefResponse struct {
	EnumDefs           []EnumDef           `json:"enumDefs"`
	StructDefs         []StructDef         `json:"structDefs"`
	AtlanTagDefs       []AtlanTagDef       `json:"classificationDefs"`
	EntityDefs         []EntityDef         `json:"entityDefs"`
	RelationshipDefs   []RelationshipDef   `json:"relationshipDefs"`
	CustomMetadataDefs []CustomMetadataDef `json:"businessMetadataDefs"`
}

type TypeDefBase struct {
	Category    atlan.AtlanTypeCategory `json:"category"`
	CreateTime  int64                   `json:"createTime,omitempty"`
	CreatedBy   string                  `json:"createdBy,omitempty"`
	Description string                  `json:"description,omitempty"`
	GUID        string                  `json:"guid,omitempty"`
	Name        string                  `json:"name"`
	TypeVersion string                  `json:"typeVersion,omitempty"`
	UpdateTime  int64                   `json:"updateTime,omitempty"`
	UpdatedBy   string                  `json:"updatedBy,omitempty"`
	Version     int                     `json:"version,omitempty"`
}

// AtlanTagDef represents the AtlanTagDef(Classifications) structure.
type AtlanTagDef struct {
	TypeDefBase
	TypeDef
	Options       map[string]interface{} `json:"options"`
	AttributeDefs []AttributesDefsTags   `json:"attributeDefs"`
	DisplayName   string                 `json:"displayName"`
	EntityTypes   []string               `json:"entityTypes"`
	SubTypes      []string               `json:"subTypes"`
	SuperTypes    []string               `json:"superTypes"`
}

type AttributesDefsTags struct {
	name                  string     `json:"name"`
	typename              string     `json:"typeName"`
	isOptional            CustomBool `json:"isOptional"`
	cardinality           string     `json:"cardinality"`
	valuesmincount        int        `json:"valuesMinCount"`
	valuesmaxcount        int        `json:"valuesMaxCount"`
	isUnique              CustomBool `json:"isUnique"`
	isIndexable           CustomBool `json:"isIndexable"`
	includeInNotification CustomBool `json:"includeInNotification"`
	skipScrubbing         CustomBool `json:"skipScrubbing"`
	searchWeight          int        `json:"searchWeight"`
	displayName           string     `json:"displayName"`
	isDefaultValueNull    CustomBool `json:"isDefaultValueNull"`
}

// AttributeOptions represents options for customizing an attribute.
type AttributeOptions struct {
	CustomMetadataVersion       *string    `json:"customMetadataVersion,omitempty"`
	Description                 *string    `json:"description,omitempty"`
	ApplicableEntityTypes       *string    `json:"applicableEntityTypes,omitempty"`
	CustomApplicableEntityTypes *string    `json:"customApplicableEntityTypes,omitempty"`
	AllowSearch                 *string    `json:"allowSearch,omitempty"`
	MaxStrLength                *string    `json:"maxStrLength,omitempty"`
	AllowFiltering              *string    `json:"allowFiltering,omitempty"`
	MultiValueSelect            *string    `json:"multiValueSelect,omitempty"`
	ShowInOverview              *string    `json:"showInOverview,omitempty"`
	IsDeprecated                *string    `json:"isDeprecated,omitempty"`
	IsEnum                      *string    `json:"isEnum,omitempty"`
	EnumType                    *string    `json:"enumType,omitempty"`
	CustomType                  *string    `json:"customType,omitempty"`
	HasTimePrecision            CustomBool `json:"hasTimePrecision,omitempty"`
	IsArchived                  CustomBool `json:"isArchived,omitempty"`
	ArchivedAt                  *string    `json:"archivedAt,omitempty"` // Using int64 for timestamp
	ArchivedBy                  *string    `json:"archivedBy,omitempty"`
	IsSoftReference             *string    `json:"isSoftReference,omitempty"`
	IsAppendOnPartialUpdate     *string    `json:"isAppendOnPartialUpdate,omitempty"`
	PrimitiveType               *string    `json:"primitiveType,omitempty"`
	ApplicableConnections       *string    `json:"applicableConnections,omitempty"`
	ApplicableGlossaries        *string    `json:"applicableGlossaries,omitempty"`
	ApplicableAssetTypes        *string    `json:"assetTypesList,omitempty"`
	ApplicableGlossaryTypes     *string    `json:"glossaryTypeList,omitempty"`
	ApplicableOtherAssetTypes   *string    `json:"otherAssetTypeList,omitempty"`
}

// AttributeDef represents the definition of an attribute.
type AttributeDef struct {
	IsNew                 CustomBool                    `json:"isNew,omitempty"`
	Cardinality           *atlan.Cardinality            `json:"cardinality,omitempty"`
	Constraints           *[]map[string]interface{}     `json:"constraints,omitempty"`
	EnumValues            *[]string                     `json:"enumValues,omitempty"`
	Description           *string                       `json:"description,omitempty"`
	DefaultValue          *string                       `json:"defaultValue,omitempty"`
	DisplayName           *string                       `json:"displayName,omitempty"`
	Name                  *string                       `json:"name,omitempty"`
	IncludeInNotification CustomBool                    `json:"includeInNotification,omitempty"`
	IndexType             *IndexType                    `json:"indexType,omitempty"`
	IsIndexable           CustomBool                    `json:"isIndexable,omitempty"`
	IsOptional            CustomBool                    `json:"isOptional,omitempty"`
	IsUnique              CustomBool                    `json:"isUnique,omitempty"`
	Options               *AttributeOptions             `json:"options,omitempty"`
	SearchWeight          *float64                      `json:"searchWeight,omitempty"`
	SkipScrubbing         CustomBool                    `json:"skipScrubbing,omitempty"`
	TypeName              *string                       `json:"typeName,omitempty"`
	ValuesMinCount        *float64                      `json:"valuesMinCount,omitempty"`
	ValuesMaxCount        *float64                      `json:"valuesMaxCount,omitempty"`
	IndexTypeESConfig     *map[string]string            `json:"indexTypeESConfig,omitempty"`
	IndexTypeESFields     *map[string]map[string]string `json:"indexTypeESFields,omitempty"`
	IsDefaultValueNull    CustomBool                    `json:"isDefaultValueNull,omitempty"`
}

// CustomMetadataDefOptions represents options for customizing metadata definitions.
type CustomMetadataDefOptions struct {
	Emoji     *string              `json:"emoji,omitempty"`
	ImageID   *string              `json:"imageId,omitempty"`
	IsLocked  string               `json:"isLocked,omitempty"`
	LogoType  *string              `json:"logoType,omitempty"`
	LogoURL   *string              `json:"logoUrl,omitempty"`
	IconColor *atlan.AtlanTagColor `json:"iconColor,omitempty"`
	IconName  *atlan.AtlanIcon     `json:"iconName,omitempty"`
}

// CustomMetadataDef represents the definition of custom metadata.
type CustomMetadataDef struct {
	TypeDefBase
	TypeDef
	AttributeDefs []AttributeDef            `json:"attributeDefs"`
	Category      *atlan.AtlanTypeCategory  `json:"category,omitempty"`
	DisplayName   *string                   `json:"displayName,omitempty"`
	Options       *CustomMetadataDefOptions `json:"options,omitempty"`
}

// Methods

// UnmarshalJSON implements the json.Unmarshaler interface
func (cb *CustomBool) UnmarshalJSON(data []byte) error {
	var b bool
	if err := json.Unmarshal(data, &b); err == nil {
		*cb = CustomBool(b)
		return nil
	}

	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	switch s {
	case "true":
		*cb = CustomBool(true)
	case "false":
		*cb = CustomBool(false)
	default:
		return fmt.Errorf("invalid boolean value: %s", s)
	}

	return nil
}
