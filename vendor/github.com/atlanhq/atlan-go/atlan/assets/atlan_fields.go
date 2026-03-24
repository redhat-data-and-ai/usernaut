package assets

import (
	"github.com/atlanhq/atlan-go/atlan"
	"github.com/atlanhq/atlan-go/atlan/model"
)

// Interface for all AtlanField queries
type AtlanField interface {
	GetAtlanFieldName() string
}

type RelationField struct {
	AtlanFieldName string
}

func NewRelationField(atlanFieldName string) *RelationField {
	return &RelationField{AtlanFieldName: atlanFieldName}
}

func (rf *RelationField) GetAtlanFieldName() string {
	return rf.AtlanFieldName
}

// SearchableField represents a field that can be searched on depending upon the type of search
type SearchableField struct {
	AtlanFieldName    string
	ElasticFieldName  string
	InternalFieldName string
}

func NewSearchableField(atlanFieldName, elasticFieldName string) *SearchableField {
	return &SearchableField{
		AtlanFieldName:    atlanFieldName,
		ElasticFieldName:  elasticFieldName,
		InternalFieldName: atlanFieldName,
	}
}

// Getter Methods for SearchableField

func (sf *SearchableField) GetAtlanFieldName() string {
	return sf.AtlanFieldName
}

func (sf *SearchableField) GetElasticFieldName() string {
	return sf.ElasticFieldName
}

func (sf *SearchableField) GetInternalFieldName() string {
	return sf.InternalFieldName
}

// HasAnyValue Returns a query that will only match assets that have some non-null, non-empty value
// (no matter what actual value) for the field.
func (sf *SearchableField) HasAnyValue() model.Query {
	return &model.Exists{Field: sf.ElasticFieldName}
}

// Order Returns a condition to sort results by the field, in the specified order.
func (sf *SearchableField) Order(order atlan.SortOrder) model.SortItem {
	return model.SortItem{
		Field:      sf.ElasticFieldName,
		Order:      order,
		NestedPath: nil,
	}
}

// BooleanField Represents any field in Atlan that can be searched only by truthiness.
type BooleanField struct {
	*SearchableField
	BooleanFieldName string
}

// NewBooleanField Returns a query that will match all assets whose field has a value that exactly equals
// the provided boolean value.
func NewBooleanField(atlanFieldName, booleanFieldName string) *BooleanField {
	searchableField := NewSearchableField(atlanFieldName, booleanFieldName)
	return &BooleanField{
		SearchableField:  searchableField,
		BooleanFieldName: booleanFieldName,
	}
}

func (bf *BooleanField) GetBooleanFieldName() string {
	return bf.BooleanFieldName
}

func (bf *BooleanField) Eq(value bool) model.Query {
	return &model.TermQuery{
		Field: bf.BooleanFieldName,
		Value: value,
	}
}

// KeywordField Represents any field in Atlan that can be searched only by keyword (no text-analyzed fuzziness).
type KeywordField struct {
	*SearchableField
	KeywordFieldName string
}

func NewKeywordField(atlanFieldName, keywordFieldName string) *KeywordField {
	searchableField := NewSearchableField(atlanFieldName, keywordFieldName)
	return &KeywordField{
		SearchableField:  searchableField,
		KeywordFieldName: keywordFieldName,
	}
}

// StartsWith Returns a query that will match all assets whose field has a value that starts with
// the provided value. Note that this can also be a case-insensitive match.
func (kf *KeywordField) StartsWith(value string, caseInsensitive *bool) model.Query {
	return &model.PrefixQuery{
		Field:           kf.KeywordFieldName,
		Value:           value,
		CaseInsensitive: caseInsensitive,
	}
}

func (kf *KeywordField) GetKeywordFieldName() string {
	return kf.KeywordFieldName
}

// StartsWith Returns a query that will match all assets whose field has a value that starts with
// the provided value. Note that this can also be a case-insensitive match.
func (kf *KeywordTextField) StartsWith(value string, caseInsensitive *bool) model.Query {
	return &model.PrefixQuery{
		Field:           kf.KeywordFieldName,
		Value:           value,
		CaseInsensitive: caseInsensitive,
	}
}

// Eq Returns a query that will match all assets whose field has a value that exactly matches
// the provided string value.
func (kf *KeywordTextField) Eq(value string) model.Query {
	return &model.TermQuery{
		Field: kf.KeywordFieldName,
		Value: value,
	}
}

// Within Returns a query that will match all assets whose field has a value that exactly matches
// at least one of the provided string values.
func (kf *KeywordField) Within(values []string) model.Query {
	return &model.Terms{
		Field:  kf.KeywordFieldName,
		Values: values,
	}
}

func (kf *KeywordField) Eq(value string) model.Query {
	return &model.TermQuery{
		Field: kf.KeywordFieldName,
		Value: value,
	}
}

// KeywordTextField Represents any field in Atlan that can be searched by keyword or text-based search operations.
type KeywordTextField struct {
	*SearchableField
	KeywordFieldName string
	TextFieldName    string
}

func NewKeywordTextField(atlanFieldName, keywordFieldName string, textfieldname string) *KeywordTextField {
	searchableField := NewSearchableField(atlanFieldName, keywordFieldName)
	return &KeywordTextField{
		SearchableField:  searchableField,
		KeywordFieldName: keywordFieldName,
		TextFieldName:    textfieldname,
	}
}

// TextField Represents any field in Atlan that can only be searched using text-related search operations.
type TextField struct {
	*SearchableField
	TextFieldName string
}

func NewTextField(atlanFieldName, textFieldName string) *TextField {
	searchableField := NewSearchableField(atlanFieldName, textFieldName)
	return &TextField{
		SearchableField: searchableField,
		TextFieldName:   textFieldName,
	}
}

// GetTextFieldName Returns the name of the text field index for this attribute in Elastic.
func (tf *TextField) GetTextFieldName() string {
	return tf.TextFieldName
}

// Match Returns a query that will textually match the provided value against the field. This
// analyzes the provided value according to the same analysis carried out on the field
// (for example, tokenization, stemming, and so on).
func (tf *TextField) Match(value string) model.Query {
	return &model.MatchQuery{
		Field: tf.TextFieldName,
		Query: value,
	}
}

// NumericField Represents any field in Atlan that can be searched using only numeric search operations.
type NumericField struct {
	*SearchableField
	NumericFieldName string
}

func NewNumericField(atlanFieldName, numericFieldName string) *NumericField {
	searchableField := NewSearchableField(atlanFieldName, numericFieldName)
	return &NumericField{
		SearchableField:  searchableField,
		NumericFieldName: numericFieldName,
	}
}

// GetNumericFieldName Returns the name of the numeric field index for this attribute in Elastic.
func (nf *NumericField) GetNumericFieldName() string {
	return nf.NumericFieldName
}

// Eq Returns a query that will match all assets whose field has a value that exactly
// matches the provided numeric value.
func (nf *NumericField) Eq(value interface{}) model.Query {
	return &model.TermQuery{
		Field: nf.NumericFieldName,
		Value: value,
	}
}

// Gt Returns a query that will match all assets whose field has a value that is strictly
// greater than the provided numeric value.
func (nf *NumericField) Gt(value *float64) model.Query {
	return &model.RangeQuery{
		Field: nf.NumericFieldName,
		Gt:    value,
	}
}

// Gte Returns a query that will match all assets whose field has a value that is greater
// than or equal to the provided numeric value.
func (nf *NumericField) Gte(value *float64) model.Query {
	return &model.RangeQuery{
		Field: nf.NumericFieldName,
		Gte:   value,
	}
}

// Lt Returns a query that will match all assets whose field has a value that is strictly
// less than the provided numeric value.
func (nf *NumericField) Lt(value *float64) model.Query {
	return &model.RangeQuery{
		Field: nf.NumericFieldName,
		Lt:    value,
	}
}

// Lte Returns a query that will match all assets whose field has a value that is less
// than or equal to the provided numeric value.
func (nf *NumericField) Lte(value *float64) model.Query {
	return &model.RangeQuery{
		Field: nf.NumericFieldName,
		Lte:   value,
	}
}

// Between Returns a query that will match all assets whose field has a value between the minimum and
// maximum specified values, inclusive.
func (nf *NumericField) Between(minimum, maximum *float64) model.Query {
	return &model.RangeQuery{
		Field: nf.NumericFieldName,
		Gte:   minimum,
		Lte:   maximum,
	}
}

// InternalKeywordTextField Represents any field in Atlan that can be searched by keyword or text-based search operations, and can also
// be searched against a special internal field directly within Atlan
type InternalKeywordTextField struct {
	*KeywordTextField
	InternalFieldName string
}

func NewInternalKeywordTextField(atlanFieldName, keywordFieldName, textFieldName, internalFieldName string) *InternalKeywordTextField {
	keywordTextField := NewKeywordTextField(atlanFieldName, keywordFieldName, textFieldName)
	return &InternalKeywordTextField{
		KeywordTextField:  keywordTextField,
		InternalFieldName: internalFieldName,
	}
}

// KeywordTextStemmedField Represents any field in Atlan that can be searched by keyword or text-based search operations,
// including a stemmed variation of the text analyzers.
type KeywordTextStemmedField struct {
	*KeywordTextField
	StemmedFieldName string
}

func NewKeywordTextStemmedField(atlanFieldName, keywordFieldName, textFieldName, stemmedFieldName string) *KeywordTextStemmedField {
	keywordTextField := NewKeywordTextField(atlanFieldName, keywordFieldName, textFieldName)
	return &KeywordTextStemmedField{
		KeywordTextField: keywordTextField,
		StemmedFieldName: stemmedFieldName,
	}
}

// MatchStemmed Returns a query that will textually match the provided value against the field. This
// analyzes the provided value according to the same analysis carried out on the field
func (ktsf *KeywordTextStemmedField) MatchStemmed(value string) model.Query {
	return &model.MatchQuery{
		Field: ktsf.StemmedFieldName,
		Query: value,
	}
}

/*
CustomMetadataField Utility class to simplify searching for values on custom metadata attributes.
*/
type CustomMetadataField struct {
	*SearchableField
	SetName       string
	AttributeName string
	AttributeDef  model.AttributeDef
}

func NewCustomMetadataField(setName, attributeName string) (*CustomMetadataField, error) {
	elasticFieldName, err := GetCustomMetadataCache().GetAttrIDForName(setName, attributeName)
	if err != nil {
		return nil, err
	}
	searchableField := NewSearchableField(attributeName, elasticFieldName)
	return &CustomMetadataField{
		SearchableField: searchableField,
		SetName:         setName,
		AttributeName:   attributeName,
		AttributeDef:    GetAttributeDef(elasticFieldName),
	}, nil
}

func (cmf *CustomMetadataField) Eq(value string) model.Query {
	return &model.TermQuery{
		Field: cmf.ElasticFieldName,
		Value: value,
	}
}

func (cmf *CustomMetadataField) StartsWith(value string, caseInsensitive *bool) model.Query {
	return &model.PrefixQuery{
		Field:           cmf.ElasticFieldName,
		Value:           value,
		CaseInsensitive: caseInsensitive,
	}
}

func (cmf *CustomMetadataField) Within(values []string) model.Query {
	return &model.Terms{
		Field:  cmf.ElasticFieldName,
		Values: values,
	}
}

func (cmf *CustomMetadataField) Match(value string) model.Query {
	return &model.MatchQuery{
		Field: cmf.ElasticFieldName,
		Query: value,
	}
}

func (cmf *CustomMetadataField) Gt(value *float64) model.Query {
	return &model.RangeQuery{
		Field: cmf.ElasticFieldName,
		Gt:    value,
	}
}

func (cmf *CustomMetadataField) Gte(value *float64) model.Query {
	return &model.RangeQuery{
		Field: cmf.ElasticFieldName,
		Gte:   value,
	}
}

func (cmf *CustomMetadataField) Lt(value *float64) model.Query {
	return &model.RangeQuery{
		Field: cmf.ElasticFieldName,
		Lt:    value,
	}
}

func (cmf *CustomMetadataField) Lte(value *float64) model.Query {
	return &model.RangeQuery{
		Field: cmf.ElasticFieldName,
		Lt:    value,
	}
}

func (cmf *CustomMetadataField) Between(minimum, maximum *float64) model.Query {
	return &model.RangeQuery{
		Field: cmf.ElasticFieldName,
		Gte:   minimum,
		Lte:   maximum,
	}
}

/*
// Needs to Implement Custom Metadata and Aggregations

func (nf *NumericField) Avg() Aggregation {
	return Aggregation{
		Root: map[string]interface{}{
			"avg": map[string]interface{}{
				"field": nf.NumericFieldName,
			},
		},
	}
}

func (nf *NumericField) Sum() Aggregation {
	return Aggregation{
		Root: map[string]interface{}{
			"sum": map[string]interface{}{
				"field": nf.NumericFieldName,
			},
		},
	}
}

func (nf *NumericField) Min() Aggregation {
	return Aggregation{
		Root: map[string]interface{}{
			"min": map[string]interface{}{
				"field": nf.NumericFieldName,
			},
		},
	}
}

func (nf *NumericField) Max() Aggregation {
	return Aggregation{
		Root: map[string]interface{}{
			"max": map[string]interface{}{
				"field": nf.NumericFieldName,
			},
		},
	}
}
*/
