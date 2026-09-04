package structs

import "time"

type MaterializedView struct {
	SQL
	RefreshMode        *string            `json:"refreshMode,omitempty"`
	RefreshMethod      *string            `json:"refreshMethod,omitempty"`
	Staleness          *string            `json:"staleness,omitempty"`
	StaleSinceDate     *time.Time         `json:"staleSinceDate,omitempty"`
	ColumnCount        *int               `json:"columnCount,omitempty"`
	RowCount           *int               `json:"rowCount,omitempty"`
	SizeBytes          *int               `json:"sizeBytes,omitempty"`
	IsQueryPreview     *bool              `json:"isQueryPreview,omitempty"`
	QueryPreviewConfig *map[string]string `json:"queryPreviewConfig,omitempty"`
	Alias              *string            `json:"alias,omitempty"`
	IsTemporary        *bool              `json:"isTemporary,omitempty"`
	Definition         *string            `json:"definition,omitempty"`
	Columns            []*Column          `json:"columns,omitempty"`
	AtlanSchema        *Schema            `json:"atlanSchema,omitempty"`
}
