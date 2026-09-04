package assets

import (
	"encoding/json"
	"fmt"

	"github.com/atlanhq/atlan-go/atlan"
	"github.com/atlanhq/atlan-go/atlan/model"
)

type TypeDefClient struct {
	Client *AtlanClient
}

// NewTypeDefClient creates a new instance of TypeDefClient.
func NewTypeDefClient(client *AtlanClient) *TypeDefClient {
	return &TypeDefClient{Client: client}
}

func buildTypeDefRequest(typedef model.TypeDef) (model.TypeDefResponse, error) {
	var payload model.TypeDefResponse

	switch t := typedef.(type) {
	case *model.AtlanTagDef:
		payload = model.TypeDefResponse{
			AtlanTagDefs:       []model.AtlanTagDef{*t},
			EnumDefs:           []model.EnumDef{},
			StructDefs:         []model.StructDef{},
			EntityDefs:         []model.EntityDef{},
			RelationshipDefs:   []model.RelationshipDef{},
			CustomMetadataDefs: []model.CustomMetadataDef{},
		}
	case *model.CustomMetadataDef:
		payload = model.TypeDefResponse{
			AtlanTagDefs:       []model.AtlanTagDef{},
			EnumDefs:           []model.EnumDef{},
			StructDefs:         []model.StructDef{},
			EntityDefs:         []model.EntityDef{},
			RelationshipDefs:   []model.RelationshipDef{},
			CustomMetadataDefs: []model.CustomMetadataDef{*t},
		}
	case *model.EnumDef:
		payload = model.TypeDefResponse{
			AtlanTagDefs:       []model.AtlanTagDef{},
			EnumDefs:           []model.EnumDef{*t},
			StructDefs:         []model.StructDef{},
			EntityDefs:         []model.EntityDef{},
			RelationshipDefs:   []model.RelationshipDef{},
			CustomMetadataDefs: []model.CustomMetadataDef{},
		}
	default:
		return model.TypeDefResponse{}, AtlanError{ErrorCode: errorCodes[UNABLE_TO_UPDATE_TYPEDEF_CATEGORY], Args: []interface{}{t}}
	}

	return payload, nil
}

func NewTypeDefResponse(rawJSON []byte) (*model.TypeDefResponse, error) {
	var response model.TypeDefResponse
	if err := json.Unmarshal(rawJSON, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

func RefreshCaches(typedef model.TypeDef) error {
	switch t := typedef.(type) {
	case *model.AtlanTagDef:
		return GetAtlanTagCache().RefreshCache()
	case *model.CustomMetadataDef:
		return GetCustomMetadataCache().RefreshCache()
	case model.EnumDef:
		// return EnumCache.RefreshCache()
	default:
		return AtlanError{ErrorCode: errorCodes[UNABLE_TO_UPDATE_TYPEDEF_CATEGORY], Args: []interface{}{t}}
	}
	return nil
}

func GetAll() (*model.TypeDefResponse, error) {
	rawJSON, err := DefaultAtlanClient.CallAPI(&GET_ALL_TYPE_DEFS, nil, nil)
	if err != nil {
		return nil, AtlanError{
			ErrorCode: errorCodes[CONNECTION_ERROR],
			Args:      []interface{}{"IOException"},
		}
	}
	return NewTypeDefResponse(rawJSON)
}

// Get retrieves a TypeDefResponse object that contains a list of the specified category type definitions in Atlan.
func Get(typeCategory interface{}) (*model.TypeDefResponse, error) {
	var categories []string
	hasStruct := false

	switch v := typeCategory.(type) {
	case atlan.AtlanTypeCategory:
		if v == atlan.AtlanTypeCategoryStruct {
			hasStruct = true
		}
		categories = append(categories, v.String())
	case []atlan.AtlanTypeCategory:
		for _, tc := range v {
			if tc == atlan.AtlanTypeCategoryStruct {
				hasStruct = true
			}
			categories = append(categories, tc.String())
		}
	default:
		return nil, fmt.Errorf("invalid type category")
	}

	if !hasStruct {
		categories = append(categories, atlan.AtlanTypeCategoryStruct.String())
	}

	queryParams := map[string][]string{
		"type": categories,
	}

	rawJSON, err := DefaultAtlanClient.CallAPI(&GET_ALL_TYPE_DEFS, queryParams, nil)
	if err != nil {
		return nil, AtlanError{
			ErrorCode: errorCodes[CONNECTION_ERROR],
			Args:      []interface{}{"IOException"},
		}
	}

	response, err := NewTypeDefResponse(rawJSON)
	if err != nil {
		return nil, err
	}

	if atlan.Contains(categories, atlan.AtlanTypeCategoryStruct.String()) {
		if response == nil || response.StructDefs == nil || len(response.StructDefs) == 0 {
			return nil, AtlanError{ErrorCode: errorCodes[EXPIRED_API_TOKEN]}
		}
	}

	return response, nil
}

func (c *TypeDefClient) Create(typedef model.TypeDef) (*model.TypeDefResponse, error) {
	payload, err := buildTypeDefRequest(typedef)
	if err != nil {
		return nil, err
	}
	rawJSON, err := c.Client.CallAPI(&CREATE_TYPE_DEFS, nil, payload)
	if err != nil {
		return nil, err
	}
	RefreshCaches(typedef)
	return NewTypeDefResponse(rawJSON)
}

func (c *TypeDefClient) Update(typedef model.TypeDef) (*model.TypeDefResponse, error) {
	payload, err := buildTypeDefRequest(typedef)
	if err != nil {
		return nil, err
	}
	rawJSON, err := c.Client.CallAPI(&UPDATE_TYPE_DEFS, nil, payload)
	if err != nil {
		return nil, err
	}
	RefreshCaches(typedef)
	return NewTypeDefResponse(rawJSON)
}

func (c *TypeDefClient) Purge(name string, typedefType model.TypeDef) error {
	var internalName string
	switch t := typedefType.(type) {
	case *model.CustomMetadataDef:
		internalName, _ = GetCustomMetadataCache().GetIDForName(name)
	case *model.EnumDef:
		// internalName = name
	case *model.AtlanTagDef:
		internalName, _ = GetAtlanTagCache().GetIDForName(name)
	default:
		return fmt.Errorf("unsupported TypeDef type: %T", t)
	}

	if internalName == "" {
		return NotFoundError{AtlanError{ErrorCode: errorCodes[TYPEDEF_NOT_FOUND_BY_NAME], Args: []interface{}{name}}}
	}

	c.Client.CallAPI(&DELETE_TYPE_DEF_BY_NAME, nil, nil)

	switch t := typedefType.(type) {
	case *model.CustomMetadataDef:
		GetCustomMetadataCache().RefreshCache()
	case *model.EnumDef:
		// EnumCache.refreshCache()
	case *model.AtlanTagDef:
		RefreshCache()
	default:
		return fmt.Errorf("unsupported TypeDef type: %T", t)
	}

	return nil
}
