package structs

type Schema struct {
	SQL
}

type Query struct {
	SQL
}

type TablePartition struct {
	SQL
}

type DbtTest struct {
	// Dbt
}

type DbtModel struct {
	// Dbt
}

type DbtSource struct {
	// Dbt
}

type Process struct {
	Asset
}

type AirFlowTask struct {
	Asset
}

type Catalog struct {
	Asset
	InputToProcesses       *[]Process     `json:"inputToProcesses,omitempty"`
	InputToAirFlowTasks    *[]AirFlowTask `json:"inputToAirFlowTasks,omitempty"`
	OutputFromProcesses    *[]Process     `json:"outputFromProcesses,omitempty"`
	OutputFromAirFlowTasks *[]AirFlowTask `json:"outputFromAirFlowTasks,omitempty"`
}

type SQL struct {
	Catalog
	QueryCount            *int32            `json:"queryCount,omitempty"`
	QueryUserCount        *int32            `json:"queryUserCount,omitempty"`
	QueryUserMap          *map[string]int32 `json:"putQueryUserMap,omitempty"`
	QueryCountUpdatedAt   *int32            `json:"queryCountUpdatedAt,omitempty"`
	DatabaseName          *string           `json:"database,omitempty"`
	DatabaseQualifiedName *string           `json:"databaseQualifiedName,omitempty"`
	SchemaName            *string           `json:"schema,omitempty"`
	SchemaQualifiedName   *string           `json:"schemaQualifiedName,omitempty"`
	TableName             *string           `json:"tableName,omitempty"`
	TableQualifiedName    *string           `json:"tableQualifiedName,omitempty"`
	ViewName              *string           `json:"viewName,omitempty"`
	ViewQualifiedName     *string           `json:"viewQualifiedName,omitempty"`
	IsProfiled            *bool             `json:"isProfiled,omitempty"`
	LastProfiledAt        *int64            `json:"lastProfiledAt,omitempty"`
	DbtSources            *[]DbtSource      `json:"dbtSources,omitempty"`
	SqlDBTModels          *[]DbtModel       `json:"sqlDbtModels,omitempty"`
	SqlDBTSources         *[]DbtSource      `json:"sqlDbtSources,omitempty"`
	DbtModels             *[]DbtModel       `json:"dbtModels,omitempty"`
	DbtTests              *[]DbtTest        `json:"dbtTests,omitempty"`
}

type Table struct {
	SQL
	/** Alias for this table. */
	Alias *string `json:"alias,omitempty"`
	/** Number of columns in this table. */
	ColumnCount *int32 `json:"columnCount,omitempty"`
	/** Columns that exist within this table. */
	Columns                *[]Column          `json:"columns,omitempty"`
	Dimensions             *[]Table           `json:"dimensions,omitempty"`
	ExternalLocation       *string            `json:"externalLocation,omitempty"`
	ExternalLocationFormat *string            `json:"externalLocationFormat,omitempty"`
	ExternalLocationRegion *string            `json:"externalLocationRegion,omitempty"`
	Facts                  *[]Table           `json:"facts,omitempty"`
	IsPartitioned          *bool              `json:"isPartitioned,omitempty"`
	IsQueryPreview         *bool              `json:"isQueryPreview,omitempty"`
	IsTemporary            *bool              `json:"isTemporary,omitempty"`
	PartitionCount         *int32             `json:"partitionCount,omitempty"`
	PartitionList          *[]string          `json:"partitionList,omitempty"`
	PartitionStrategy      *string            `json:"partitionStrategy,omitempty"`
	Partitions             *[]TablePartition  `json:"partitions,omitempty"`
	Queries                *[]Query           `json:"queries,omitempty"`
	QueryPreviewConfig     *map[string]string `json:"putQueryPreviewConfig,omitempty"`
	RowCount               *int64             `json:"rowCount,omitempty"`
	Schema                 *Schema            `json:"atlanSchema,omitempty"`
	SizeBytes              *int64             `json:"sizeBytes,omitempty"`
}
