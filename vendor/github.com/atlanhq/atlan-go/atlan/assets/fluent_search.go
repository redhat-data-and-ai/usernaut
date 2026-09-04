package assets

import (
	"github.com/atlanhq/atlan-go/atlan"
	"github.com/atlanhq/atlan-go/atlan/model"
)

// FluentSearch is a struct that represents a fluent search query.
type FluentSearch struct {
	Wheres              []model.Query
	WhereNots           []model.Query
	WhereSomes          []model.Query
	MinSomes            *int
	Sorts               []model.SortItem
	PageSize            int
	Aggregations        map[string]interface{}
	IncludesOnResults   []string
	IncludesOnRelations []string
	UtmTags             []string
}

// SetUtmTags sets the UTM tags for tracking the source of requests.
func (fs *FluentSearch) SetUtmTags(tags ...atlan.UTMTags) *FluentSearch {
	// Convert UTMTags to string before assigning to UtmTags
	strTags := make([]string, len(tags))
	for i, tag := range tags {
		strTags[i] = tag.String()
	}
	fs.UtmTags = strTags
	return fs
}

type Aggregation struct{}

// ActiveAssets Returns a query that will only match assets that are active in Atlan.
func (fs *FluentSearch) ActiveAssets() *FluentSearch {
	activeAssetsCondition := &model.TermQuery{
		Field: STATE,
		Value: "ACTIVE",
	}
	fs.Wheres = append(fs.Wheres, activeAssetsCondition)
	return fs
}

// ArchivedAssets Returns a query that will only match assets that are archived (soft-deleted) in Atlan.
func (fs *FluentSearch) ArchivedAssets() *FluentSearch {
	archivedAssetsCondition := &model.TermQuery{
		Field: STATE,
		Value: "Deleted",
	}
	fs.Wheres = append(fs.Wheres, archivedAssetsCondition)
	return fs
}

// AssetType Returns a query that will only match assets of the type provided.
func (fs *FluentSearch) AssetType(of string) *FluentSearch {
	assetTypeCondition := &model.TermQuery{
		Field: TYPE_NAME,
		Value: of,
	}
	fs.Wheres = append(fs.Wheres, assetTypeCondition)
	return fs
}

// AssetTypes Returns a query that will only match assets that are one of the types provided.
func (fs *FluentSearch) AssetTypes(oneOf []string) *FluentSearch {
	assetTypesCondition := &model.Terms{
		Field:  TYPE_NAME,
		Values: oneOf,
	}
	fs.Wheres = append(fs.Wheres, assetTypesCondition)
	return fs
}

func NewFluentSearch() *FluentSearch {
	return &FluentSearch{}
}

// Where adds a TermQuery to the Wheres slice.
func (fs *FluentSearch) Where(queries ...model.Query) *FluentSearch {
	boolQuery := &model.BoolQuery{Filter: queries}
	fs.Wheres = append(fs.Wheres, boolQuery)
	return fs
}

// WhereNot adds a BoolQuery with MustNot clause to the WhereNots slice.
func (fs *FluentSearch) WhereNot(queries ...model.Query) *FluentSearch {
	boolQuery := &model.BoolQuery{MustNot: queries}
	fs.WhereNots = append(fs.WhereNots, boolQuery)
	return fs
}

// WhereSome adds a BoolQuery with Should clause to the WhereSomes slice.
func (fs *FluentSearch) WhereSome(queries ...model.Query) *FluentSearch {
	boolQuery := &model.BoolQuery{Should: queries}
	fs.WhereSomes = append(fs.WhereSomes, boolQuery)
	return fs
}

// MinSome sets the MinSomes field.
func (fs *FluentSearch) MinSome(minSomes int) *FluentSearch {
	fs.MinSomes = &minSomes
	return fs
}

// Sort adds a SortItem to the Sorts slice.
func (fs *FluentSearch) Sort(field string, order atlan.SortOrder) *FluentSearch {
	fs.Sorts = append(fs.Sorts, model.SortItem{Field: field, Order: order})
	return fs
}

// Aggregate adds an aggregation to the Aggregations map.
func (fs *FluentSearch) Aggregate(name string, aggregation map[string]interface{}) *FluentSearch {
	if fs.Aggregations == nil {
		fs.Aggregations = make(map[string]interface{})
	}
	fs.Aggregations[name] = aggregation
	return fs
}

// PageSize sets the PageSize field.
func (fs *FluentSearch) PageSizes(size int) *FluentSearch {
	fs.PageSize = size
	return fs
}

// IncludeOnResults adds fields to the IncludesOnResults slice.
func (fs *FluentSearch) IncludeOnResults(fields ...string) *FluentSearch {
	fs.IncludesOnResults = append(fs.IncludesOnResults, fields...)
	return fs
}

// IncludeOnRelations adds fields to the IncludesOnRelations slice.
func (fs *FluentSearch) IncludeOnRelations(fields ...string) *FluentSearch {
	fs.IncludesOnRelations = append(fs.IncludesOnRelations, fields...)
	return fs
}

// Execute performs the search and returns the results.
func (fs *FluentSearch) Execute() (*IndexSearchIterator, error) {
	return Search(*fs.ToRequest())
}

// Sort by GUID by default only if not already specified by the developer
func (fs *FluentSearch) SortByGuidDefault() *FluentSearch {
	// Check if "guid" is already in the list of sort criteria
	guidAlreadySorted := false
	for _, item := range fs.Sorts {
		if item.Field == string(GUID) {
			guidAlreadySorted = true
			break
		}
	}

	// If "guid" is not already in the list, add it as the final sort criteria
	if !guidAlreadySorted {
		fs.Sort(GUID, atlan.SortOrderAscending)
	}

	return fs
}

// ToRequest converts FluentSearch to IndexSearchRequest.
func (fs *FluentSearch) ToRequest() *model.IndexSearchRequest {
	// Create a new IndexSearchRequest and set its properties based on FluentSearch
	request := &model.IndexSearchRequest{
		SearchRequest: model.SearchRequest{
			Attributes:          fs.IncludesOnResults,
			RelationsAttributes: fs.IncludesOnRelations,
		},
		Dsl: model.Dsl{
			From:           0,
			Size:           fs.PageSize,
			Aggregation:    fs.Aggregations,
			TrackTotalHits: true,
		},
		Metadata: model.Metadata{
			SaveSearchLog: true,
			UtmTags:       []string{atlan.PROJECT_SDK_GO.String()},
		},
	}

	// Add Wheres to Query
	if len(fs.Wheres) > 0 {
		boolQuery := &model.BoolQuery{Filter: fs.Wheres}
		request.Dsl.Query = boolQuery.ToJSON()
	}

	// Add WhereNots to Query
	if len(fs.WhereNots) > 0 {
		boolQuery := &model.BoolQuery{MustNot: fs.WhereNots}
		if request.Dsl.Query == nil {
			request.Dsl.Query = boolQuery.ToJSON()
		} else {
			request.Dsl.Query = map[string]interface{}{
				"bool": map[string]interface{}{
					"filter":   request.Dsl.Query,
					"must":     request.Dsl.Query,
					"must_not": boolQuery.ToJSON(),
				},
			}
		}
	}

	// Add WhereSomes to Query
	if len(fs.WhereSomes) > 0 {
		boolQuery := &model.BoolQuery{Should: fs.WhereSomes, MinimumShouldMatch: fs.MinSomes}
		if request.Dsl.Query == nil {
			request.Dsl.Query = boolQuery.ToJSON()
		} else {
			request.Dsl.Query = map[string]interface{}{
				"bool": map[string]interface{}{
					"filter": request.Dsl.Query,
					"must":   request.Dsl.Query,
					"should": boolQuery.ToJSON(),
				},
			}
		}
	}

	// Add Sorts to Dsl.Sort
	if len(fs.Sorts) > 0 {
		fs.SortByGuidDefault()
		sortItems := fs.Sorts
		sortItemsJSON := make([]map[string]interface{}, len(sortItems))
		for i, item := range sortItems {
			sortItemsJSON[i] = item.ToJSON()
		}
		request.Dsl.Sort = sortItemsJSON
	}

	// Set UtmTags specified by user
	if fs.UtmTags != nil {
		request.Metadata.UtmTags = fs.UtmTags
	}

	return request
}
