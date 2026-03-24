package assets

import (
	"encoding/json"

	"github.com/atlanhq/atlan-go/atlan/model/structs"
)

type DataContract structs.DataContract

type DataContractClient struct {
	client *AtlanClient
}

func NewDataContractClient(ac *AtlanClient) *DataContractClient {
	return &DataContractClient{client: ac}
}

func (dc *DataContract) Creator(name string) {
	dc.TypeName = structs.StringPtr("DataContract")
	dc.Name = structs.StringPtr(name)
}

func (dc *DataContract) MarshalJSON() ([]byte, error) {
	// Construct the custom JSON structure
	customJSON := map[string]interface{}{
		"typeName": "DataContract",
		"attributes": map[string]interface{}{
			"name": dc.Name,
			// Add other attributes as necessary.
		},
		"relationshipAttributes": make(map[string]interface{}),
	}

	if dc.QualifiedName != nil && *dc.QualifiedName != "" {
		customJSON["attributes"].(map[string]interface{})["qualifiedName"] = *dc.QualifiedName
	}

	if dc.Guid != nil && *dc.Guid != "" {
		customJSON["guid"] = *dc.Guid
	}

	if dc.DisplayName != nil && *dc.DisplayName != "" {
		customJSON["attributes"].(map[string]interface{})["displayName"] = *dc.DisplayName
	}

	// Marshal the custom JSON
	return json.MarshalIndent(customJSON, "", "  ")
}
