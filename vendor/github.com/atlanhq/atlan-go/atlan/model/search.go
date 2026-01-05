// Contains the search model for the Atlas search DSL.

package model

import (
	"encoding/json"
	"fmt"
	"regexp"
	"time"

	"github.com/atlanhq/atlan-go/atlan"
	"github.com/atlanhq/atlan-go/atlan/model/structs"
)

// Query is an interface that represents the base query behavior.
type Query interface {
	ToJSON() map[string]interface{}
}

// TermQuery represents a term query in the Atlas search DSL.
type TermQuery struct {
	Field string
	Value interface{}
}

// BoolQuery represents a boolean query in the Atlas search DSL.
type BoolQuery struct {
	Must               []Query
	Should             []Query
	MustNot            []Query
	Filter             []Query
	TypeName           string
	Boost              *float64
	MinimumShouldMatch *int
}

// MatchAll represents a match_all query in the Atlas search DSL.
type MatchAll struct {
	Boost *float64
}

type MatchNone struct{}

type Exists struct {
	Field string
}

// NestedQuery represents a nested query in the Atlas search DSL.
type NestedQuery struct {
	Path           string
	Query          Query
	ScoreMode      string
	IgnoreUnmapped bool
}

type Terms struct {
	Field  string
	Values []string
	Boost  *float64
}

type PrefixQuery struct {
	Field           string
	Value           interface{}
	Boost           *float64
	CaseInsensitive *bool
	TypeName        string
}

type RangeQuery struct {
	Field    string
	Gt       *float64
	Gte      *float64
	Lt       *float64
	Lte      *float64
	Boost    *float64
	Format   *string
	Relation *string
	TimeZone *string
	TypeName string
}

type WildcardQuery struct {
	Field           string
	Value           string
	Boost           *float64
	CaseInsensitive *bool
}

type RegexpQuery struct {
	Field                 string
	Value                 string
	Boost                 *float64
	CaseInsensitive       *bool
	MaxDeterminizedStates *int
}

type FuzzyQuery struct {
	Field          string
	Value          string
	Fuzziness      *string
	MaxExpansions  *int
	PrefixLength   *int
	Transpositions *bool
	Rewrite        *string
}

type MatchQuery struct {
	Field                           string
	Query                           string
	Analyzer                        *string
	AutoGenerateSynonymsPhraseQuery *bool
	Fuzziness                       *string
	FuzzyTranspositions             *bool
	FuzzyRewrite                    *string
	Lenient                         *bool
	Operator                        *string
	MinimumShouldMatch              *int
	ZeroTermsQuery                  *string
	MaxExpansions                   *int
	PrefixLength                    *int
}

type SortItem struct {
	Field      string
	Order      atlan.SortOrder
	NestedPath *string
}

// ToJSON returns the JSON representation of the TermQuery.
func (t *TermQuery) ToJSON() map[string]interface{} {
	return map[string]interface{}{
		"term": map[string]interface{}{
			t.Field: map[string]interface{}{
				"value": t.Value,
			},
		},
	}
}

// ToJSON returns the JSON representation of the BoolQuery.
func (b *BoolQuery) ToJSON() map[string]interface{} {
	boolQuery := make(map[string]interface{})
	if len(b.Must) > 0 {
		mustClauses := make([]map[string]interface{}, len(b.Must))
		for i, q := range b.Must {
			mustClauses[i] = q.ToJSON()
		}
		boolQuery["must"] = mustClauses
	}
	if len(b.Should) > 0 {
		shouldClauses := make([]map[string]interface{}, len(b.Should))
		for i, q := range b.Should {
			shouldClauses[i] = q.ToJSON()
		}
		boolQuery["should"] = shouldClauses
	}
	if len(b.MustNot) > 0 {
		mustNotClauses := make([]map[string]interface{}, len(b.MustNot))
		for i, q := range b.MustNot {
			mustNotClauses[i] = q.ToJSON()
		}
		boolQuery["must_not"] = mustNotClauses
	}
	if len(b.Filter) > 0 {
		filterClauses := make([]map[string]interface{}, len(b.Filter))
		for i, q := range b.Filter {
			filterClauses[i] = q.ToJSON()
		}
		boolQuery["filter"] = filterClauses
	}
	if b.Boost != nil {
		boolQuery["boost"] = *b.Boost
	}
	if b.MinimumShouldMatch != nil {
		boolQuery["minimum_should_match"] = *b.MinimumShouldMatch
	}
	return map[string]interface{}{"bool": boolQuery}
}

// ToJSON returns the JSON representation of the MatchAll query.
func (m *MatchAll) ToJSON() map[string]interface{} {
	query := make(map[string]interface{})
	if m.Boost != nil {
		query["boost"] = *m.Boost
	}
	return map[string]interface{}{
		"match_all": query,
	}
}

// ToJSON returns the JSON representation of the MatchNone query.
func (m *MatchNone) ToJSON() map[string]interface{} {
	return map[string]interface{}{
		"match_none": map[string]interface{}{},
	}
}

// ToJSON returns the JSON representation of the Exists query.
func (e *Exists) ToJSON() map[string]interface{} {
	return map[string]interface{}{
		"exists": map[string]interface{}{
			"field": e.Field,
		},
	}
}

// ToJSON returns the JSON representation of the NestedQuery.
func (n *NestedQuery) ToJSON() map[string]interface{} {
	query := map[string]interface{}{
		"path":  n.Path,
		"query": n.Query.ToJSON(),
	}
	if n.ScoreMode != "" {
		query["score_mode"] = n.ScoreMode
	}
	if n.IgnoreUnmapped {
		query["ignore_unmapped"] = n.IgnoreUnmapped
	}
	return map[string]interface{}{
		"nested": query,
	}
}

// ToJSON returns the JSON representation of the Terms query.
func (t *Terms) ToJSON() map[string]interface{} {
	query := map[string]interface{}{
		"terms": map[string]interface{}{
			t.Field: t.Values,
		},
	}
	if t.Boost != nil {
		query["terms"].(map[string]interface{})["boost"] = *t.Boost
	}
	return query
}

// ToJSON returns the JSON representation of the PrefixQuery.
func (p *PrefixQuery) ToJSON() map[string]interface{} {
	prefixQuery := map[string]interface{}{
		"value": p.Value,
	}
	if p.Boost != nil {
		prefixQuery["boost"] = *p.Boost
	}
	if p.CaseInsensitive != nil {
		prefixQuery["case_insensitive"] = *p.CaseInsensitive
	}
	return map[string]interface{}{
		"prefix": map[string]interface{}{
			p.Field: prefixQuery,
		},
	}
}

// ToJSON returns the JSON representation of the RangeQuery.
func (r *RangeQuery) ToJSON() map[string]interface{} {
	rangeQuery := make(map[string]interface{})
	if r.Gt != nil {
		rangeQuery["gt"] = *r.Gt
	}
	if r.Gte != nil {
		rangeQuery["gte"] = *r.Gte
	}
	if r.Lt != nil {
		rangeQuery["lt"] = *r.Lt
	}
	if r.Lte != nil {
		rangeQuery["lte"] = *r.Lte
	}
	if r.Boost != nil {
		rangeQuery["boost"] = *r.Boost
	}
	if r.Format != nil {
		rangeQuery["format"] = *r.Format
	}
	if r.Relation != nil {
		rangeQuery["relation"] = *r.Relation
	}
	if r.TimeZone != nil {
		rangeQuery["time_zone"] = *r.TimeZone
	}
	return map[string]interface{}{
		"range": map[string]interface{}{
			r.Field: rangeQuery,
		},
	}
}

// ToJSON returns the JSON representation of the WildcardQuery.
func (w *WildcardQuery) ToJSON() map[string]interface{} {
	wildcardQuery := map[string]interface{}{
		"value": w.Value,
	}
	if w.Boost != nil {
		wildcardQuery["boost"] = *w.Boost
	}
	if w.CaseInsensitive != nil {
		wildcardQuery["case_insensitive"] = *w.CaseInsensitive
	}
	return map[string]interface{}{
		"wildcard": map[string]interface{}{
			w.Field: wildcardQuery,
		},
	}
}

// ToJSON returns the JSON representation of the RegexpQuery.
func (r *RegexpQuery) ToJSON() map[string]interface{} {
	regexpQuery := map[string]interface{}{
		"value": r.Value,
	}
	if r.Boost != nil {
		regexpQuery["boost"] = *r.Boost
	}
	if r.CaseInsensitive != nil {
		regexpQuery["case_insensitive"] = *r.CaseInsensitive
	}
	if r.MaxDeterminizedStates != nil {
		regexpQuery["max_determinized_states"] = *r.MaxDeterminizedStates
	}
	return map[string]interface{}{
		"regexp": map[string]interface{}{
			r.Field: regexpQuery,
		},
	}
}

// ToJSON returns the JSON representation of the FuzzyQuery.
func (f *FuzzyQuery) ToJSON() map[string]interface{} {
	fuzzyQuery := map[string]interface{}{
		"value": f.Value,
	}
	if f.Fuzziness != nil {
		fuzzyQuery["fuzziness"] = *f.Fuzziness
	}
	if f.MaxExpansions != nil {
		fuzzyQuery["max_expansions"] = *f.MaxExpansions
	}
	if f.PrefixLength != nil {
		fuzzyQuery["prefix_length"] = *f.PrefixLength
	}
	if f.Transpositions != nil {
		fuzzyQuery["transpositions"] = *f.Transpositions
	}
	if f.Rewrite != nil {
		fuzzyQuery["rewrite"] = *f.Rewrite
	}
	return map[string]interface{}{
		"fuzzy": map[string]interface{}{
			f.Field: fuzzyQuery,
		},
	}
}

// ToJSON returns the JSON representation of the MatchQuery.
func (m *MatchQuery) ToJSON() map[string]interface{} {
	matchQuery := map[string]interface{}{
		"query": m.Query,
	}
	if m.Analyzer != nil {
		matchQuery["analyzer"] = *m.Analyzer
	}
	if m.AutoGenerateSynonymsPhraseQuery != nil {
		matchQuery["auto_generate_synonyms_phrase_query"] = *m.AutoGenerateSynonymsPhraseQuery
	}
	if m.Fuzziness != nil {
		matchQuery["fuzziness"] = *m.Fuzziness
	}
	if m.FuzzyTranspositions != nil {
		matchQuery["fuzzy_transpositions"] = *m.FuzzyTranspositions
	}
	if m.FuzzyRewrite != nil {
		matchQuery["fuzzy_rewrite"] = *m.FuzzyRewrite
	}
	if m.Lenient != nil {
		matchQuery["lenient"] = *m.Lenient
	}
	if m.Operator != nil {
		matchQuery["operator"] = *m.Operator
	}
	if m.MinimumShouldMatch != nil {
		matchQuery["minimum_should_match"] = *m.MinimumShouldMatch
	}
	if m.ZeroTermsQuery != nil {
		matchQuery["zero_terms_query"] = *m.ZeroTermsQuery
	}
	if m.MaxExpansions != nil {
		matchQuery["max_expansions"] = *m.MaxExpansions
	}
	if m.PrefixLength != nil {
		matchQuery["prefix_length"] = *m.PrefixLength
	}
	return map[string]interface{}{
		"match": map[string]interface{}{
			m.Field: matchQuery,
		},
	}
}

// ToJSON returns the JSON representation of the SortItem.
func (s *SortItem) ToJSON() map[string]interface{} {
	sortField := map[string]interface{}{"order": s.Order}
	if s.NestedPath != nil {
		sortField["nested"] = map[string]interface{}{"path": *s.NestedPath}
	}
	return map[string]interface{}{s.Field: sortField}
}

/*
I don’t remember why we made a separate `ToJSON()` method for above search queries instead of using
the custom `MarshalJSON()` method, which would automatically be used when preparing a request.
I won’t change these methods for now, but I’ll make a separate JIRA ticket to refactor them and check if anything breaks.
*/

// WorkflowSearchRequest captures the request structure for workflow search. (Added here in order to avoid circular imports)
type WorkflowSearchRequest struct {
	From           int        `json:"from"`
	Size           int        `json:"size"`
	TrackTotalHits bool       `json:"track_total_hits"`
	PostFilter     Query      `json:"post_filter,omitempty"`
	Query          Query      `json:"query,omitempty"`
	Sort           []SortItem `json:"sort"`
}

// MarshalJSON marshals the WorkflowSearchRequest to JSON.
func (w WorkflowSearchRequest) MarshalJSON() ([]byte, error) {
	type Alias WorkflowSearchRequest
	alias := Alias(w)

	var queryJSON map[string]interface{}
	if w.Query != nil {
		queryJSON = w.Query.ToJSON()
	}

	var postFilterJSON map[string]interface{}
	if w.PostFilter != nil {
		postFilterJSON = w.PostFilter.ToJSON()
	}

	var sortJSON []map[string]interface{}
	for _, s := range w.Sort {
		sortJSON = append(sortJSON, s.ToJSON())
	}

	return json.Marshal(&struct {
		Query      map[string]interface{}   `json:"query,omitempty"`
		PostFilter map[string]interface{}   `json:"post_filter,omitempty"`
		Sort       []map[string]interface{} `json:"sort,omitempty"`
		Alias
	}{
		Query:      queryJSON,
		PostFilter: postFilterJSON,
		Sort:       sortJSON,
		Alias:      alias,
	})
}

// SearchRequest represents a search request in the Atlas search DSL.
type SearchRequest struct {
	Attributes          []string `json:"attributes,omitempty"`
	Offset              int      `json:"from,omitempty"`
	Size                int      `json:"size,omitempty"`
	RelationsAttributes []string `json:"relationsAttributes,omitempty"`
}

type Metadata struct {
	SaveSearchLog bool     `json:"saveSearchLog,omitempty"`
	UtmTags       []string `json:"utmTags,omitempty"`
}

// IndexSearchRequest represents a search request in the Atlas search DSL.
type IndexSearchRequest struct {
	SearchRequest
	Dsl                   Dsl      `json:"dsl"`
	RelationAttributes    []string `json:"relationAttributes,omitempty"`
	SuppressLogs          bool     `json:"suppressLogs,omitempty"`
	ShowSearchScore       bool     `json:"showSearchScore,omitempty"`
	ExcludeMeanings       bool     `json:"excludeMeanings,omitempty"`
	ExcludeAtlanTags      bool     `json:"excludeClassifications,omitempty"`
	AllowDeletedRelations bool     `json:"allowDeletedRelations,omitempty"`
	Metadata              Metadata `json:"requestMetadata,omitempty"`
}

// Dsl represents the DSL for the Atlas search request.
type Dsl struct {
	From                int                      `json:"from"`
	Size                int                      `json:"size"`
	Aggregation         map[string]interface{}   `json:"aggregation,omitempty"`
	Query               map[string]interface{}   `json:"query"`
	TrackTotalHits      bool                     `json:"track_total_hits"`
	PostFilter          *Query                   `json:"post_filter,omitempty"`
	Sort                []map[string]interface{} `json:"sort,omitempty"`
	IncludesOnResults   []string                 `json:"includesOnResults,omitempty"`
	IncludesOnRelations []string                 `json:"includesOnRelations,omitempty"`
}

// IndexSearchResponse represents a search response in the Atlas search DSL.
type IndexSearchResponse struct {
	QueryType        string           `json:"queryType"`
	SearchParameters SearchParameters `json:"searchParameters"`
	Entities         []SearchAssets   `json:"entities"`
	ApproximateCount int64            `json:"approximateCount"`
}

func (isr *IndexSearchResponse) UnmarshalJSON(data []byte) error {
	// Define an auxiliary struct to decode the JSON
	type AuxIndexSearchResponse struct {
		QueryType        string            `json:"queryType"`
		SearchParameters json.RawMessage   `json:"searchParameters"`
		Entities         []json.RawMessage `json:"entities"`
		ApproximateCount int64             `json:"approximateCount"`
	}

	// Unmarshal into the auxiliary struct
	var aux AuxIndexSearchResponse
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Unmarshal SearchParameters into the desired struct
	var searchParams SearchParameters
	if err := json.Unmarshal(aux.SearchParameters, &searchParams); err != nil {
		return err
	}
	isr.SearchParameters = searchParams

	// Unmarshal each entity into SearchAssets
	var entities []SearchAssets
	for _, entityData := range aux.Entities {
		var sa SearchAssets
		if err := json.Unmarshal(entityData, &sa); err != nil {
			return err
		}
		// Populate custom metadata set for each entity
		sa.CustomMetadataSets = isr.unflattenCustomMetadata(searchParams.Attributes, sa.rawSearchAttributes)
		entities = append(entities, sa)
	}
	isr.Entities = entities
	isr.QueryType = aux.QueryType
	isr.ApproximateCount = aux.ApproximateCount

	return nil
}

// Helper function to unflatten custom metadata structures from index search results
func (isr *IndexSearchResponse) unflattenCustomMetadata(searchParameterAttributes []string, searchAttributes map[string]interface{}) map[string]map[string]interface{} {
	if len(searchParameterAttributes) == 0 || len(searchAttributes) == 0 {
		return nil
	}

	retval := make(map[string]map[string]interface{})
	metadataAttribute := regexp.MustCompile(`(\w+)\.(\w+)`)

	for _, attributeOfInterest := range searchParameterAttributes {
		if metadataAttribute.MatchString(attributeOfInterest) {
			if value, exists := searchAttributes[attributeOfInterest]; exists {
				matches := metadataAttribute.FindStringSubmatch(attributeOfInterest)
				if len(matches) > 2 {
					key := matches[1]
					if _, ok := retval[key]; !ok {
						retval[key] = make(map[string]interface{})
					}
					retval[key][matches[2]] = value
				}
			}
		}
	}
	return retval
}

// SearchParameters represents the search parameters in the Atlas search response.
type SearchParameters struct {
	ShowSearchScore       bool     `json:"showSearchScore"`
	SuppressLogs          bool     `json:"suppressLogs"`
	ExcludeMeanings       bool     `json:"excludeMeanings"`
	ExcludeAtlanTags      bool     `json:"excludeClassifications"`
	AllowDeletedRelations bool     `json:"allowDeletedRelations"`
	SaveSearchLog         bool     `json:"saveSearchLog"`
	RequestMetadata       Metadata `json:"requestMetadata"`
	Dsl                   Dsl      `json:"dsl"`
	Query                 string   `json:"query"`
	Attributes            []string `json:"attributes,omitempty"`
}

type SearchAssets struct {
	structs.Asset
	structs.Table
	structs.Column
	structs.AuthPolicy
	structs.Persona
	structs.AccessControl
	QualifiedName       *string           `json:"qualifiedName,omitempty"`
	Name                *string           `json:"name,omitempty"`
	SearchAttributes    *SearchAttributes `json:"Attributes,omitempty"`
	SearchMeanings      []Meanings        `json:"meanings,omitempty"` // If meanings as json is already defined in Asset struct then defining the meanings here would result in an empty response.
	NotNull             *bool             `json:"notNull,omitempty"`
	rawSearchAttributes map[string]interface{}
}

type Meanings struct {
	Guid         string `json:"termGuid,omitempty"`
	RelationGuid string `json:"relationGuid,omitempty"`
	DisplayText  string `json:"displayText,omitempty"`
	Confidence   int    `json:"confidence,omitempty"`
}

type SearchAttributes struct {
	// Column Attributes
	SubDataType                    *string                            `json:"subDataType,omitempty"`
	RawDataTypeDefinition          *string                            `json:"rawDataTypeDefinition,omitempty"`
	Order                          *int                               `json:"order,omitempty"`
	NestedColumnCount              *int                               `json:"nestedColumnCount,omitempty"`
	IsPartition                    *bool                              `json:"isPartition,omitempty"`
	PartitionOrder                 *int                               `json:"partitionOrder,omitempty"`
	IsClustered                    *bool                              `json:"isClustered,omitempty"`
	IsPrimary                      *bool                              `json:"isPrimary,omitempty"`
	IsForeign                      *bool                              `json:"isForeign,omitempty"`
	IsIndexed                      *bool                              `json:"isIndexed,omitempty"`
	IsSort                         *bool                              `json:"isSort,omitempty"`
	IsDist                         *bool                              `json:"isDist,omitempty"`
	IsPinned                       *bool                              `json:"isPinned,omitempty"`
	PinnedBy                       *string                            `json:"pinnedBy,omitempty"`
	PinnedAt                       *time.Time                         `json:"pinnedAt,omitempty"`
	Precision                      *int                               `json:"precision,omitempty"`
	DefaultValue                   *string                            `json:"defaultValue,omitempty"`
	NumericScale                   *float64                           `json:"numericScale,omitempty"`
	Validations                    map[string]string                  `json:"validations,omitempty"`
	ParentColumnQualifiedName      *string                            `json:"parentColumnQualifiedName,omitempty"`
	ParentColumnName               *string                            `json:"parentColumnName,omitempty"`
	ColumnDistinctValuesCount      *int                               `json:"columnDistinctValuesCount,omitempty"`
	ColumnDistinctValuesCountLong  *int                               `json:"columnDistinctValuesCountLong,omitempty"`
	ColumnHistogram                *structs.Histogram                 `json:"columnHistogram,omitempty"`
	ColumnMax                      *float64                           `json:"columnMax,omitempty"`
	ColumnMin                      *float64                           `json:"columnMin,omitempty"`
	ColumnMean                     *float64                           `json:"columnMean,omitempty"`
	ColumnSum                      *float64                           `json:"columnSum,omitempty"`
	ColumnMedian                   *float64                           `json:"columnMedian,omitempty"`
	ColumnStandardDeviation        *float64                           `json:"columnStandardDeviation,omitempty"`
	ColumnUniqueValuesCount        *int                               `json:"columnUniqueValuesCount,omitempty"`
	ColumnUniqueValuesCountLong    *int                               `json:"columnUniqueValuesCountLong,omitempty"`
	ColumnAverage                  *float64                           `json:"columnAverage,omitempty"`
	ColumnAverageLength            *float64                           `json:"columnAverageLength,omitempty"`
	ColumnDuplicateValuesCount     *int                               `json:"columnDuplicateValuesCount,omitempty"`
	ColumnDuplicateValuesCountLong *int                               `json:"columnDuplicateValuesCountLong,omitempty"`
	ColumnMaximumStringLength      *int                               `json:"columnMaximumStringLength,omitempty"`
	ColumnMaxs                     *map[string]bool                   `json:"columnMaxs,omitempty"`
	ColumnMinimumStringLength      *int                               `json:"columnMinimumStringLength,omitempty"`
	ColumnMins                     *map[string]bool                   `json:"columnMins,omitempty"`
	ColumnMissingValuesCount       *int                               `json:"columnMissingValuesCount,omitempty"`
	ColumnMissingValuesCountLong   *int                               `json:"columnMissingValuesCountLong,omitempty"`
	ColumnMissingValuesPercentage  *float64                           `json:"columnMissingValuesPercentage,omitempty"`
	ColumnUniquenessPercentage     *float64                           `json:"columnUniquenessPercentage,omitempty"`
	ColumnVariance                 *float64                           `json:"columnVariance,omitempty"`
	ColumnTopValues                []*structs.ColumnValueFrequencyMap `json:"columnTopValues,omitempty"`
	ColumnDepthLevel               *int                               `json:"columnDepthLevel,omitempty"`
	SnowflakeDynamicTable          *structs.SnowflakeDynamicTable     `json:"snowflakeDynamicTable,omitempty"`
	View                           *structs.View                      `json:"view,omitempty"`
	NestedColumns                  []*structs.Column                  `json:"nestedColumns,omitempty"`
	DataQualityMetricDimensions    []*structs.Metric                  `json:"dataQualityMetricDimensions,omitempty"`
	DbtModelColumns                []*structs.DbtModelColumn          `json:"dbtModelColumns,omitempty"`
	Table                          *structs.Table                     `json:"table,omitempty"`
	ColumnDbtModelColumns          []*structs.DbtModelColumn          `json:"columnDbtModelColumns,omitempty"`
	MaterialisedView               *structs.MaterialisedView          `json:"materialisedView,omitempty"`
	ParentColumn                   *structs.Column                    `json:"parentColumn,omitempty"`
	Queries                        []*Query                           `json:"queries,omitempty"`
	MetricTimestamps               []*structs.Metric                  `json:"metricTimestamps,omitempty"`
	ForeignKeyTo                   []*structs.Column                  `json:"foreignKeyTo,omitempty"`
	ForeignKeyFrom                 *structs.Column                    `json:"foreignKeyFrom,omitempty"`
	DbtMetrics                     []*structs.DbtMetric               `json:"dbtMetrics,omitempty"`
	TablePartition                 *structs.TablePartition            `json:"tablePartition,omitempty"`
	MaxLength                      *int                               `json:"maxLength,omitempty"`

	// Access Control Attributes
	IsAccessControlEnabled  *bool                               `json:"isAccessControlEnabled,omitempty"`
	DenyCustomMetadataGuids *[]string                           `json:"denyCustomMetadataGuids,omitempty"`
	DenyAssetTabs           *[]string                           `json:"denyAssetTabs,omitempty"`
	DenyAssetFilters        *[]string                           `json:"denyAssetFilters,omitempty"`
	ChannelLink             *string                             `json:"channelLink,omitempty"`
	DenyAssetTypes          *[]string                           `json:"denyAssetTypes,omitempty"`
	DenyNavigationPages     *[]string                           `json:"denyNavigationPages,omitempty"`
	DefaultNavigation       *string                             `json:"defaultNavigation,omitempty"`
	DisplayPreferences      *[]string                           `json:"displayPreferences,omitempty"`
	Policies                *[]structs.AuthPolicy               `json:"policies,omitempty"` // Relationship
	PolicyType              *atlan.AuthPolicyType               `json:"policyType,omitempty"`
	PolicyServiceName       *string                             `json:"policyServiceName,omitempty"`
	PolicyCategory          *string                             `json:"policyCategory,omitempty"`
	PolicySubCategory       *string                             `json:"policySubCategory,omitempty"`
	PolicyUsers             *[]string                           `json:"policyUsers,omitempty"`
	PolicyGroups            *[]string                           `json:"policyGroups,omitempty"`
	PolicyRoles             *[]string                           `json:"policyRoles,omitempty"`
	PolicyActions           *[]string                           `json:"policyActions,omitempty"`
	PolicyResources         *[]string                           `json:"policyResources,omitempty"`
	PolicyResourceCategory  *string                             `json:"policyResourceCategory,omitempty"`
	PolicyPriority          *int                                `json:"policyPriority,omitempty"`
	IsPolicyEnabled         *bool                               `json:"isPolicyEnabled,omitempty"`
	PolicyMaskType          *atlan.DataMaskingType              `json:"policyMaskType,omitempty"`
	PolicyValiditySchedule  *[]atlan.AuthPolicyValiditySchedule `json:"policyValiditySchedule,omitempty"`
	PolicyResourceSignature *string                             `json:"policyResourceSignature,omitempty"`
	PolicyDelegateAdmin     *bool                               `json:"policyDelegateAdmin,omitempty"`
	PolicyConditions        *[]atlan.AuthPolicyCondition        `json:"policyConditions,omitempty"`
	AccessControl           *structs.AccessControl              `json:"accessControl,omitempty"` // Relationship
	PersonaGroups           *[]string                           `json:"personaGroups,omitempty"`
	PersonaUsers            *[]string                           `json:"personaUsers,omitempty"`
	RoleId                  *string                             `json:"roleId,omitempty"`

	// Common Attributes
	QualifiedName            *string                  `json:"qualifiedName,omitempty"`
	Name                     *string                  `json:"name,omitempty"`
	UserDescription          *string                  `json:"userDescription,omitempty"`
	Description              *string                  `json:"description,omitempty"`
	DataType                 *string                  `json:"dataType,omitempty"`
	IsNullable               *bool                    `json:"isNullable,omitempty"`
	OwnerGroups              *[]string                `json:"ownerGroups,omitempty"`
	OwnerUsers               *[]string                `json:"ownerUsers,omitempty"`
	AnnouncementType         *atlan.AnnouncementType  `json:"announcementType,omitempty"`
	AnnouncementTitle        *string                  `json:"announcementTitle,omitempty"`
	AnnouncementMessage      *string                  `json:"announcementMessage,omitempty"`
	CertificateStatus        *atlan.CertificateStatus `json:"certificateStatus,omitempty"`
	CertificateStatusMessage *string                  `json:"certificateStatusMessage,omitempty"`
}

// Used in End-to-end bulk update

// Updater is a generic method that updates the required fields of the asset in memory.
func (sa *SearchAssets) Updater() error {
	if sa.TypeName == nil || sa.QualifiedName == nil || sa.Name == nil {
		return fmt.Errorf("missing TypeName or QualifiedName or Name")
	}

	// Ensure required fields are populated
	qualifiedName := *sa.QualifiedName
	TypeName := *sa.TypeName
	name := *sa.Name

	// Modify asset in memory
	sa.TypeName = &TypeName
	sa.QualifiedName = &qualifiedName
	sa.Name = &name

	return nil
}

func (sa *SearchAssets) MarshalJSON() ([]byte, error) {
	// Construct the custom JSON structure
	customJSON := map[string]interface{}{
		"typeName": sa.TypeName,
		"attributes": map[string]interface{}{
			"name":          sa.Name,
			"qualifiedName": sa.QualifiedName,
			// Add other attributes as necessary.
		},
	}

	attributes := customJSON["attributes"].(map[string]interface{})

	if sa.Guid != nil && *sa.Asset.Guid != "" {
		customJSON["guid"] = *sa.Guid
	}

	if sa.Status != nil {
		customJSON["status"] = *sa.Status
	}

	if sa.CreatedBy != nil {
		customJSON["createdBy"] = *sa.CreatedBy
	}

	if sa.CreateTime != nil {
		customJSON["createTime"] = *sa.CreateTime
	}

	if sa.UpdateTime != nil {
		customJSON["updateTime"] = *sa.UpdateTime
	}

	if sa.UpdatedBy != nil {
		customJSON["updatedBy"] = *sa.UpdatedBy
	}

	if sa.DisplayText != nil && *sa.DisplayText != "" {
		customJSON["DisplayText"] = *sa.DisplayText
	}

	if sa.Asset.DisplayName != nil && *sa.Asset.DisplayName != "" {
		attributes["DisplayText"] = *sa.Asset.DisplayName
	}

	if sa.Asset.Description != nil && *sa.Asset.Description != "" {
		attributes["description"] = *sa.Asset.Description
	}

	if sa.Table.SchemaName != nil && *sa.Table.SchemaName != "" {
		attributes["schemaName"] = *sa.Table.SchemaName
	}

	if sa.Table.DatabaseName != nil && *sa.Table.DatabaseName != "" {
		attributes["databaseName"] = *sa.Table.DatabaseName
	}

	if sa.Table.DatabaseQualifiedName != nil && *sa.Table.DatabaseQualifiedName != "" {
		attributes["databaseQualifiedName"] = *sa.Table.DatabaseQualifiedName
	}

	if sa.Table.ConnectionQualifiedName != nil && *sa.Table.ConnectionQualifiedName != "" {
		attributes["connectionQualifiedName"] = *sa.Table.ConnectionQualifiedName
	}

	if sa.Asset.CertificateStatus != nil {
		attributes["certificateStatus"] = *sa.Asset.CertificateStatus
	}

	// Requires Model Generator for generating other assets

	// Add access control attributes

	if sa.IsAccessControlEnabled != nil {
		attributes["isAccessControlEnabled"] = *sa.IsAccessControlEnabled
	}
	if sa.DenyCustomMetadataGuids != nil {
		attributes["denyCustomMetadataGuids"] = *sa.DenyCustomMetadataGuids
	}
	if sa.DenyAssetTabs != nil {
		attributes["denyAssetTabs"] = *sa.DenyAssetTabs
	}
	if sa.DenyAssetFilters != nil {
		attributes["denyAssetFilters"] = *sa.DenyAssetFilters
	}
	if sa.ChannelLink != nil {
		attributes["channelLink"] = *sa.ChannelLink
	}
	if sa.DenyAssetTypes != nil {
		attributes["denyAssetTypes"] = *sa.DenyAssetTypes
	}
	if sa.DenyNavigationPages != nil {
		attributes["denyNavigationPages"] = *sa.DenyNavigationPages
	}
	if sa.DefaultNavigation != nil {
		attributes["defaultNavigation"] = *sa.DefaultNavigation
	}
	if sa.DisplayPreferences != nil {
		attributes["displayPreferences"] = *sa.DisplayPreferences
	}
	if sa.Policies != nil {
		attributes["policies"] = sa.Policies // Assuming proper JSON marshalling of structs.AuthPolicy
	}
	if sa.PolicyType != nil {
		attributes["policyType"] = *sa.PolicyType
	}
	if sa.PolicyServiceName != nil {
		attributes["policyServiceName"] = *sa.PolicyServiceName
	}
	if sa.PolicyCategory != nil {
		attributes["policyCategory"] = *sa.PolicyCategory
	}
	if sa.PolicySubCategory != nil {
		attributes["policySubCategory"] = *sa.PolicySubCategory
	}
	if sa.PolicyUsers != nil {
		attributes["policyUsers"] = *sa.PolicyUsers
	}
	if sa.PolicyGroups != nil {
		attributes["policyGroups"] = *sa.PolicyGroups
	}
	if sa.PolicyRoles != nil {
		attributes["policyRoles"] = *sa.PolicyRoles
	}
	if sa.PolicyActions != nil {
		attributes["policyActions"] = *sa.PolicyActions
	}
	if sa.PolicyResources != nil {
		attributes["policyResources"] = *sa.PolicyResources
	}
	if sa.PolicyResourceCategory != nil {
		attributes["policyResourceCategory"] = *sa.PolicyResourceCategory
	}
	if sa.PolicyPriority != nil {
		attributes["policyPriority"] = *sa.PolicyPriority
	}
	if sa.IsPolicyEnabled != nil {
		attributes["isPolicyEnabled"] = *sa.IsPolicyEnabled
	}
	if sa.PolicyMaskType != nil {
		attributes["policyMaskType"] = *sa.PolicyMaskType
	}
	if sa.PolicyValiditySchedule != nil {
		attributes["policyValiditySchedule"] = sa.PolicyValiditySchedule
	}
	if sa.PolicyResourceSignature != nil {
		attributes["policyResourceSignature"] = *sa.PolicyResourceSignature
	}
	if sa.PolicyDelegateAdmin != nil {
		attributes["policyDelegateAdmin"] = *sa.PolicyDelegateAdmin
	}
	if sa.PolicyConditions != nil {
		attributes["policyConditions"] = sa.PolicyConditions
	}
	if sa.PersonaGroups != nil {
		attributes["personaGroups"] = *sa.PersonaGroups
	}
	if sa.PersonaUsers != nil {
		attributes["personaUsers"] = *sa.PersonaUsers
	}
	if sa.RoleId != nil {
		attributes["roleId"] = *sa.RoleId
	}

	// Handle nested AccessControl field
	accessControl := map[string]interface{}{}

	if sa.AccessControl.Guid != nil && *sa.AccessControl.Guid != "" {
		accessControl["guid"] = *sa.AccessControl.Guid
	}

	if sa.AccessControl.TypeName != nil && *sa.AccessControl.TypeName != "" {
		accessControl["typeName"] = *sa.AccessControl.TypeName
	}

	if sa.AccessControl.UniqueAttributes.QualifiedName != nil && *sa.AccessControl.UniqueAttributes.QualifiedName != "" {
		accessControl["uniqueAttributes"] = map[string]interface{}{
			"qualifiedName": sa.AccessControl.UniqueAttributes.QualifiedName,
		}
	}

	if len(accessControl) > 0 {
		attributes["accessControl"] = accessControl
	}

	// Marshal the custom JSON
	return json.MarshalIndent(customJSON, "", "  ")
}

func (sa *SearchAssets) UnmarshalJSON(data []byte) error {
	// Define an auxiliary struct to decode the JSON
	type AuxSearchAssets struct {
		structs.Asset
		structs.Table
		structs.Column
		structs.AuthPolicy
		structs.Persona
		structs.Purpose
		structs.AccessControl
		QualifiedName    *string           `json:"qualifiedName,omitempty"`
		Name             *string           `json:"name,omitempty"`
		SearchAttributes *SearchAttributes `json:"attributes,omitempty"`
		SearchMeanings   []Meanings        `json:"meanings,omitempty"`
		NotNull          *bool             `json:"notNull,omitempty"`

		rawSearchAttributes map[string]interface{}
	}

	// Decode into the auxiliary struct
	var aux AuxSearchAssets
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Directly handling the SearchMeanings if present (This is done because Meanings in IndexSearchResponse conflicits with meanings of type AtlasGlossaryTerm defined in /structs/asset.go .
	type SearchMeaningsData struct {
		SearchMeanings []Meanings `json:"meanings,omitempty"`
	}

	var meaningsData SearchMeaningsData
	if err := json.Unmarshal(data, &meaningsData); err != nil {
		return err
	}

	// Requires Model Generator for All Assets
	// Copy fields from auxiliary struct to the main struct
	sa.Asset = aux.Asset
	sa.Table = aux.Table
	sa.Column = aux.Column
	sa.NotNull = aux.NotNull
	sa.SearchMeanings = meaningsData.SearchMeanings
	sa.AuthPolicy = aux.AuthPolicy
	sa.AccessControl = aux.AccessControl

	// Check if any search attributes are present
	if aux.SearchAttributes != nil {
		sa.Name = aux.SearchAttributes.Name
		sa.QualifiedName = aux.SearchAttributes.QualifiedName
		sa.DataType = aux.SearchAttributes.DataType
		sa.Description = aux.SearchAttributes.Description
		sa.UserDescription = aux.SearchAttributes.UserDescription
		sa.IsPrimary = aux.SearchAttributes.IsPrimary
		sa.IsNullable = aux.SearchAttributes.IsNullable
		sa.OwnerGroups = aux.SearchAttributes.OwnerGroups
		sa.OwnerUsers = aux.SearchAttributes.OwnerUsers
		sa.AnnouncementType = aux.SearchAttributes.AnnouncementType
		sa.AnnouncementTitle = aux.SearchAttributes.AnnouncementTitle
		sa.AnnouncementMessage = aux.SearchAttributes.AnnouncementMessage
		sa.CertificateStatus = aux.SearchAttributes.CertificateStatus
		sa.CertificateStatusMessage = aux.SearchAttributes.CertificateStatusMessage
		// Column Attributes
		sa.MaxLength = aux.SearchAttributes.MaxLength
		sa.Precision = aux.SearchAttributes.Precision
		sa.NumericScale = aux.SearchAttributes.NumericScale
		sa.IsPartition = aux.SearchAttributes.IsPartition
		sa.SubDataType = aux.SearchAttributes.SubDataType
		sa.RawDataTypeDefinition = aux.SearchAttributes.RawDataTypeDefinition
		sa.Order = aux.SearchAttributes.Order
		sa.NestedColumnCount = aux.SearchAttributes.NestedColumnCount
		sa.IsPartition = aux.SearchAttributes.IsPartition
		sa.PartitionOrder = aux.SearchAttributes.PartitionOrder
		sa.IsClustered = aux.SearchAttributes.IsClustered
		sa.IsPrimary = aux.SearchAttributes.IsPrimary
		sa.IsForeign = aux.SearchAttributes.IsForeign
		sa.IsIndexed = aux.SearchAttributes.IsIndexed
		sa.IsSort = aux.SearchAttributes.IsSort
		sa.IsDist = aux.SearchAttributes.IsDist
		sa.IsPinned = aux.SearchAttributes.IsPinned
		sa.PinnedBy = aux.SearchAttributes.PinnedBy
		sa.PinnedAt = aux.SearchAttributes.PinnedAt
		sa.DefaultValue = aux.SearchAttributes.DefaultValue
		sa.Validations = aux.SearchAttributes.Validations
		sa.ParentColumnQualifiedName = aux.SearchAttributes.ParentColumnQualifiedName
		sa.ParentColumnName = aux.SearchAttributes.ParentColumnName
		sa.ColumnDistinctValuesCount = aux.SearchAttributes.ColumnDistinctValuesCount
		sa.ColumnDistinctValuesCountLong = aux.SearchAttributes.ColumnDistinctValuesCountLong
		sa.ColumnHistogram = aux.SearchAttributes.ColumnHistogram
		sa.ColumnMax = aux.SearchAttributes.ColumnMax
		sa.ColumnMin = aux.SearchAttributes.ColumnMin
		sa.ColumnMean = aux.SearchAttributes.ColumnMean
		sa.ColumnSum = aux.SearchAttributes.ColumnSum
		sa.ColumnMedian = aux.SearchAttributes.ColumnMedian
		sa.ColumnStandardDeviation = aux.SearchAttributes.ColumnStandardDeviation
		sa.ColumnUniqueValuesCount = aux.SearchAttributes.ColumnUniqueValuesCount
		sa.ColumnUniqueValuesCountLong = aux.SearchAttributes.ColumnUniqueValuesCountLong
		sa.ColumnAverage = aux.SearchAttributes.ColumnAverage
		sa.ColumnAverageLength = aux.SearchAttributes.ColumnAverageLength
		sa.ColumnDuplicateValuesCount = aux.SearchAttributes.ColumnDuplicateValuesCount
		sa.ColumnDuplicateValuesCountLong = aux.SearchAttributes.ColumnDuplicateValuesCountLong
		sa.ColumnMaximumStringLength = aux.SearchAttributes.ColumnMaximumStringLength
		sa.ColumnMaxs = aux.SearchAttributes.ColumnMaxs
		sa.ColumnMinimumStringLength = aux.SearchAttributes.ColumnMinimumStringLength
		sa.ColumnMins = aux.SearchAttributes.ColumnMins
		sa.ColumnMissingValuesCount = aux.SearchAttributes.ColumnMissingValuesCount
		sa.ColumnMissingValuesCountLong = aux.SearchAttributes.ColumnMissingValuesCountLong
		sa.ColumnMissingValuesPercentage = aux.SearchAttributes.ColumnMissingValuesPercentage
		sa.ColumnUniquenessPercentage = aux.SearchAttributes.ColumnUniquenessPercentage
		sa.ColumnVariance = aux.SearchAttributes.ColumnVariance
		sa.ColumnTopValues = aux.SearchAttributes.ColumnTopValues
		sa.ColumnDepthLevel = aux.SearchAttributes.ColumnDepthLevel
		sa.SnowflakeDynamicTable = aux.SearchAttributes.SnowflakeDynamicTable
		sa.View = aux.SearchAttributes.View
		sa.NestedColumns = aux.SearchAttributes.NestedColumns
		sa.DataQualityMetricDimensions = aux.SearchAttributes.DataQualityMetricDimensions
		sa.DbtModelColumns = aux.SearchAttributes.DbtModelColumns
		sa.ColumnDbtModelColumns = aux.SearchAttributes.ColumnDbtModelColumns
		sa.MaterialisedView = aux.SearchAttributes.MaterialisedView
		sa.ParentColumn = aux.SearchAttributes.ParentColumn
		sa.MetricTimestamps = aux.SearchAttributes.MetricTimestamps
		sa.ForeignKeyTo = aux.SearchAttributes.ForeignKeyTo
		sa.ForeignKeyFrom = aux.SearchAttributes.ForeignKeyFrom
		sa.DbtMetrics = aux.SearchAttributes.DbtMetrics
		sa.TablePartition = aux.SearchAttributes.TablePartition

		// Access Control Attributes
		sa.IsAccessControlEnabled = aux.SearchAttributes.IsAccessControlEnabled
		sa.DenyCustomMetadataGuids = aux.SearchAttributes.DenyCustomMetadataGuids
		sa.DenyAssetTabs = aux.SearchAttributes.DenyAssetTabs
		sa.DenyAssetFilters = aux.SearchAttributes.DenyAssetFilters
		sa.ChannelLink = aux.SearchAttributes.ChannelLink
		sa.DenyAssetTypes = aux.SearchAttributes.DenyAssetTypes
		sa.DenyNavigationPages = aux.SearchAttributes.DenyNavigationPages
		sa.DefaultNavigation = aux.SearchAttributes.DefaultNavigation
		sa.DisplayPreferences = aux.SearchAttributes.DisplayPreferences
		sa.Policies = aux.SearchAttributes.Policies
		sa.PolicyType = aux.SearchAttributes.PolicyType
		sa.PolicyServiceName = aux.SearchAttributes.PolicyServiceName
		sa.PolicyCategory = aux.SearchAttributes.PolicyCategory
		sa.PolicySubCategory = aux.SearchAttributes.PolicySubCategory
		sa.PolicyUsers = aux.SearchAttributes.PolicyUsers
		sa.PolicyGroups = aux.SearchAttributes.PolicyGroups
		sa.PolicyRoles = aux.SearchAttributes.PolicyRoles
		sa.PolicyActions = aux.SearchAttributes.PolicyActions
		sa.PolicyResources = aux.SearchAttributes.PolicyResources
		sa.PolicyResourceCategory = aux.SearchAttributes.PolicyResourceCategory
		sa.PolicyPriority = aux.SearchAttributes.PolicyPriority
		sa.IsPolicyEnabled = aux.SearchAttributes.IsPolicyEnabled
		sa.PolicyMaskType = aux.SearchAttributes.PolicyMaskType
		sa.PolicyValiditySchedule = aux.SearchAttributes.PolicyValiditySchedule
		sa.PolicyResourceSignature = aux.SearchAttributes.PolicyResourceSignature
		sa.PolicyDelegateAdmin = aux.SearchAttributes.PolicyDelegateAdmin
		sa.PolicyConditions = aux.SearchAttributes.PolicyConditions
		sa.PersonaGroups = aux.SearchAttributes.PersonaGroups
		sa.PersonaUsers = aux.SearchAttributes.PersonaUsers
		sa.RoleId = aux.SearchAttributes.RoleId

		if aux.SearchAttributes.AccessControl != nil {
			// Attributes under AccessControl struct
			sa.AccessControl.TypeName = aux.SearchAttributes.AccessControl.TypeName
			sa.AccessControl.Guid = aux.SearchAttributes.AccessControl.Guid
			sa.AccessControl.UniqueAttributes.QualifiedName = aux.SearchAttributes.AccessControl.UniqueAttributes.QualifiedName
		}

		// Populate `rawSearchAttributes` (necessary for setting `SearchAssets.CustomMetadataSets`)
		// First, unmarshal the data into a `rawSearchAsset` map
		var rawSearchAsset map[string]interface{}
		if err := json.Unmarshal(data, &rawSearchAsset); err != nil {
			return err
		}
		// Extract `SearchAttributes` if present
		if attributesData, ok := rawSearchAsset["attributes"]; ok {
			attributesDataBytes, err := json.Marshal(attributesData)
			if err != nil {
				return err
			}
			if err := json.Unmarshal(attributesDataBytes, &aux.rawSearchAttributes); err != nil {
				return err
			}
		}
		sa.rawSearchAttributes = aux.rawSearchAttributes
	}
	return nil
}

func (sa *SearchAssets) FromJSON(data []byte) error {
	return json.Unmarshal(data, sa)
}
