package structs

type View struct {
	SQL
	ColumnCount        *int               `json:"columnCount,omitempty"`
	RowCount           *int               `json:"rowCount,omitempty"`
	SizeBytes          *int               `json:"sizeBytes,omitempty"`
	IsQueryPreview     *bool              `json:"isQueryPreview,omitempty"`
	QueryPreviewConfig *map[string]string `json:"queryPreviewConfig,omitempty"`
	Alias              *string            `json:"alias,omitempty"`
	IsTemporary        *bool              `json:"isTemporary,omitempty"`
	Definition         *string            `json:"definition,omitempty"`
	Columns            []*Column          `json:"columns,omitempty"`
	Queries            []*Query           `json:"queries,omitempty"`
	AtlanSchema        *Schema            `json:"atlanSchema,omitempty"`
}
