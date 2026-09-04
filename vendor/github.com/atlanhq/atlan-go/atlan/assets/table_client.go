package assets

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/atlanhq/atlan-go/atlan"
	"github.com/atlanhq/atlan-go/atlan/model/structs"
)

type Table structs.Table

// Creator is used to create a new table asset.
func (t *Table) Creator(name, schemaQualifiedName string) error {
	if name == "" || schemaQualifiedName == "" {
		return errors.New("name and schemaQualifiedName are required fields")
	}

	// Extract the connector type from the schemaQualifiedName
	connectorType, err := atlan.GetConnectorTypeFromQualifiedName(schemaQualifiedName)
	if err != nil {
		return err
	}

	// Split the schemaQualifiedName to extract fields
	fields := strings.Split(schemaQualifiedName, "/")
	if len(fields) < 5 {
		return errors.New("schemaQualifiedName must have at least 5 parts separated by '/'")
	}

	// Construct connectionQualifiedName if not provided
	connectionQn := fields[0] + "/" + connectorType.Value
	connectionQualifiedName := &connectionQn

	// Construct qualified name
	qualifiedName := schemaQualifiedName + "/" + name

	// Set the optional fields if not provided
	databaseName := &fields[3]

	schemaName := &fields[4]

	dbQualifiedName := *connectionQualifiedName + "/" + *databaseName
	databaseQualifiedName := &dbQualifiedName

	// Assign the constructed and provided values to the Table
	t.TypeName = structs.StringPtr("Table")
	t.Name = structs.StringPtr(name)
	t.QualifiedName = structs.StringPtr(qualifiedName)
	t.SchemaQualifiedName = structs.StringPtr(schemaQualifiedName)
	t.SchemaName = schemaName
	t.DatabaseName = databaseName
	t.DatabaseQualifiedName = databaseQualifiedName
	t.ConnectionQualifiedName = connectionQualifiedName

	return nil
}

func (t *Table) CreatorWithParams(name, schemaQualifiedName string, schemaName, databaseName, databaseQualifiedName, connectionQualifiedName *string) error {
	if name == "" || schemaQualifiedName == "" {
		return errors.New("name and schemaQualifiedName are required fields")
	}

	// Extract the connector type from the schemaQualifiedName
	connectorType, err := atlan.GetConnectorTypeFromQualifiedName(schemaQualifiedName)
	if err != nil {
		return err
	}

	// Split the schemaQualifiedName to extract fields
	fields := strings.Split(schemaQualifiedName, "/")
	if len(fields) < 5 {
		return errors.New("schemaQualifiedName must have at least 5 parts separated by '/'")
	}

	// Construct connectionQualifiedName if not provided
	if connectionQualifiedName == nil {
		connectionQn := fields[0] + "/" + connectorType.Value
		connectionQualifiedName = &connectionQn
	}

	// Construct qualified name
	qualifiedName := schemaQualifiedName + "/" + name

	// Set the optional fields if not provided
	if databaseName == nil {
		databaseName = &fields[3]
	}

	if schemaName == nil {
		schemaName = &fields[4]
	}

	if databaseQualifiedName == nil {
		dbQualifiedName := *connectionQualifiedName + "/" + *databaseName
		databaseQualifiedName = &dbQualifiedName
	}

	// Assign the constructed and provided values to the Table
	t.TypeName = structs.StringPtr("Table")
	t.Name = structs.StringPtr(name)
	t.QualifiedName = structs.StringPtr(qualifiedName)
	t.SchemaQualifiedName = structs.StringPtr(schemaQualifiedName)
	t.SchemaName = schemaName
	t.DatabaseName = databaseName
	t.DatabaseQualifiedName = databaseQualifiedName
	t.ConnectionQualifiedName = connectionQualifiedName

	return nil
}

// UnmarshalJSON implements the JSON unmarshal interface for the Table struct.
func (t *Table) UnmarshalJSON(data []byte) error {
	Attributes := struct {
		Name                    *string `json:"name"`
		QualifiedName           *string `json:"qualifiedName"`
		SchemaQualifiedName     *string `json:"schemaQualifiedName"`
		SchemaName              *string `json:"schemaName"`
		DatabaseName            *string `json:"databaseName"`
		DatabaseQualifiedName   *string `json:"databaseQualifiedName"`
		ConnectionQualifiedName *string `json:"connectionQualifiedName"`
		// Add other attributes as necessary.
	}{}

	base, err := UnmarshalBaseEntity(data, &Attributes)
	if err != nil {
		return err
	}

	// Map shared fields
	t.TypeName = &base.Entity.TypeName
	t.Name = Attributes.Name
	t.QualifiedName = Attributes.QualifiedName
	t.Guid = &base.Entity.Guid
	t.SchemaQualifiedName = Attributes.SchemaQualifiedName
	t.SchemaName = Attributes.SchemaName
	t.DatabaseName = Attributes.DatabaseName
	t.DatabaseQualifiedName = Attributes.DatabaseQualifiedName
	t.ConnectionQualifiedName = Attributes.ConnectionQualifiedName
	// Map the rest of the fields
	return nil
}

// Updater is used to modify a Table asset in memory.
func (t *Table) Updater(name string, qualifiedName string) error {
	if name == "" || qualifiedName == "" {
		return errors.New("name, qualified_name are required fields")
	}

	t.TypeName = structs.StringPtr("Table")
	t.Name = structs.StringPtr(name)
	t.QualifiedName = structs.StringPtr(qualifiedName)

	return nil
}

// MarshalJSON filters out entities to only include those with non-empty attributes.
func (t *Table) MarshalJSON() ([]byte, error) {
	// Construct the custom JSON structure
	customJSON := map[string]interface{}{
		"typeName": t.TypeName,
		"attributes": map[string]interface{}{
			"name":                t.Name,
			"schemaQualifiedName": t.SchemaQualifiedName,
			"qualifiedName":       t.QualifiedName,
			// Add other attributes as necessary.
		},
	}

	attributes := customJSON["attributes"].(map[string]interface{})

	if t.Guid != nil && *t.Guid != "" {
		customJSON["guid"] = *t.Guid
	}

	if t.SchemaName != nil && *t.SchemaName != "" {
		attributes["schemaName"] = *t.SchemaName
	}

	if t.DatabaseName != nil && *t.DatabaseName != "" {
		attributes["databaseName"] = *t.DatabaseName
	}

	if t.DatabaseQualifiedName != nil && *t.DatabaseQualifiedName != "" {
		attributes["databaseQualifiedName"] = *t.DatabaseQualifiedName
	}

	if t.ConnectionQualifiedName != nil && *t.ConnectionQualifiedName != "" {
		attributes["connectionQualifiedName"] = *t.ConnectionQualifiedName
	}

	if t.CertificateStatus != nil {
		attributes["certificateStatus"] = *t.CertificateStatus
	}

	// Marshal the custom JSON
	return json.MarshalIndent(customJSON, "", "  ")
}

func (t *Table) ToJSON() ([]byte, error) {
	return json.MarshalIndent(t, "", "  ")
}

func (t *Table) FromJSON(data []byte) error {
	return json.Unmarshal(data, t)
}
