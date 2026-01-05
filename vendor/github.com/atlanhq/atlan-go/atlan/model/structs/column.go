package structs

import (
	"time"
)

type Histogram string

type ColumnValueFrequencyMap string

type SnowflakeDynamicTable struct {
	Table
}

type DbtModelColumn struct {
	// Dbt
}

type MaterialisedView struct {
	SQL
}

type DbtMetric struct {
	// Dbt
}

type Column struct {
	SQL
	DataType                       *string                    `json:"dataType,omitempty"`
	SubDataType                    *string                    `json:"subDataType,omitempty"`
	RawDataTypeDefinition          *string                    `json:"rawDataTypeDefinition,omitempty"`
	Order                          *int                       `json:"order,omitempty"`
	NestedColumnCount              *int                       `json:"nestedColumnCount,omitempty"`
	IsPartition                    *bool                      `json:"isPartition,omitempty"`
	PartitionOrder                 *int                       `json:"partitionOrder,omitempty"`
	IsClustered                    *bool                      `json:"isClustered,omitempty"`
	IsPrimary                      *bool                      `json:"isPrimary,omitempty"`
	IsForeign                      *bool                      `json:"isForeign,omitempty"`
	IsIndexed                      *bool                      `json:"isIndexed,omitempty"`
	IsSort                         *bool                      `json:"isSort,omitempty"`
	IsDist                         *bool                      `json:"isDist,omitempty"`
	IsPinned                       *bool                      `json:"isPinned,omitempty"`
	PinnedBy                       *string                    `json:"pinnedBy,omitempty"`
	PinnedAt                       *time.Time                 `json:"pinnedAt,omitempty"`
	Precision                      *int                       `json:"precision,omitempty"`
	DefaultValue                   *string                    `json:"defaultValue,omitempty"`
	IsNullable                     *bool                      `json:"isNullable,omitempty"`
	NumericScale                   *float64                   `json:"numericScale,omitempty"`
	MaxLength                      *int                       `json:"maxLength,omitempty"`
	Validations                    map[string]string          `json:"validations,omitempty"`
	ParentColumnQualifiedName      *string                    `json:"parentColumnQualifiedName,omitempty"`
	ParentColumnName               *string                    `json:"parentColumnName,omitempty"`
	ColumnDistinctValuesCount      *int                       `json:"columnDistinctValuesCount,omitempty"`
	ColumnDistinctValuesCountLong  *int                       `json:"columnDistinctValuesCountLong,omitempty"`
	ColumnHistogram                *Histogram                 `json:"columnHistogram,omitempty"`
	ColumnMax                      *float64                   `json:"columnMax,omitempty"`
	ColumnMin                      *float64                   `json:"columnMin,omitempty"`
	ColumnMean                     *float64                   `json:"columnMean,omitempty"`
	ColumnSum                      *float64                   `json:"columnSum,omitempty"`
	ColumnMedian                   *float64                   `json:"columnMedian,omitempty"`
	ColumnStandardDeviation        *float64                   `json:"columnStandardDeviation,omitempty"`
	ColumnUniqueValuesCount        *int                       `json:"columnUniqueValuesCount,omitempty"`
	ColumnUniqueValuesCountLong    *int                       `json:"columnUniqueValuesCountLong,omitempty"`
	ColumnAverage                  *float64                   `json:"columnAverage,omitempty"`
	ColumnAverageLength            *float64                   `json:"columnAverageLength,omitempty"`
	ColumnDuplicateValuesCount     *int                       `json:"columnDuplicateValuesCount,omitempty"`
	ColumnDuplicateValuesCountLong *int                       `json:"columnDuplicateValuesCountLong,omitempty"`
	ColumnMaximumStringLength      *int                       `json:"columnMaximumStringLength,omitempty"`
	ColumnMaxs                     *map[string]bool           `json:"columnMaxs,omitempty"`
	ColumnMinimumStringLength      *int                       `json:"columnMinimumStringLength,omitempty"`
	ColumnMins                     *map[string]bool           `json:"columnMins,omitempty"`
	ColumnMissingValuesCount       *int                       `json:"columnMissingValuesCount,omitempty"`
	ColumnMissingValuesCountLong   *int                       `json:"columnMissingValuesCountLong,omitempty"`
	ColumnMissingValuesPercentage  *float64                   `json:"columnMissingValuesPercentage,omitempty"`
	ColumnUniquenessPercentage     *float64                   `json:"columnUniquenessPercentage,omitempty"`
	ColumnVariance                 *float64                   `json:"columnVariance,omitempty"`
	ColumnTopValues                []*ColumnValueFrequencyMap `json:"columnTopValues,omitempty"`
	ColumnDepthLevel               *int                       `json:"columnDepthLevel,omitempty"`
	SnowflakeDynamicTable          *SnowflakeDynamicTable     `json:"snowflakeDynamicTable,omitempty"`
	View                           *View                      `json:"view,omitempty"`
	NestedColumns                  []*Column                  `json:"nestedColumns,omitempty"`
	DataQualityMetricDimensions    []*Metric                  `json:"dataQualityMetricDimensions,omitempty"`
	DbtModelColumns                []*DbtModelColumn          `json:"dbtModelColumns,omitempty"`
	Table                          *Table                     `json:"table,omitempty"`
	ColumnDbtModelColumns          []*DbtModelColumn          `json:"columnDbtModelColumns,omitempty"`
	MaterialisedView               *MaterialisedView          `json:"materialisedView,omitempty"`
	ParentColumn                   *Column                    `json:"parentColumn,omitempty"`
	Queries                        []*Query                   `json:"queries,omitempty"`
	MetricTimestamps               []*Metric                  `json:"metricTimestamps,omitempty"`
	ForeignKeyTo                   []*Column                  `json:"foreignKeyTo,omitempty"`
	ForeignKeyFrom                 *Column                    `json:"foreignKeyFrom,omitempty"`
	DbtMetrics                     []*DbtMetric               `json:"dbtMetrics,omitempty"`
	TablePartition                 *TablePartition            `json:"tablePartition,omitempty"`
}
