package atlan

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

type Cardinality struct {
	name string
}

func (a Cardinality) String() string {
	return a.name
}

var (
	CardinalitySingle = Cardinality{"SINGLE"}
	CardinatlityList  = Cardinality{"LIST"}
	CardinalitySet    = Cardinality{"SET"}
)

// UnmarshalJSON customizes the unmarshalling of a Cardinality from JSON.
func (c *Cardinality) UnmarshalJSON(data []byte) error {
	var name string
	if err := json.Unmarshal(data, &name); err != nil {
		return err
	}

	switch name {
	case "SINGLE":
		*c = CardinalitySingle
	case "LIST":
		*c = CardinatlityList
	case "SET":
		*c = CardinalitySet
	default:
		*c = Cardinality{name: name}
	}

	return nil
}

// MarshalJSON customizes the marshalling of a Cardinality to JSON.
func (c Cardinality) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.name)
}

type AtlanTagColor struct {
	name string
}

func (a AtlanTagColor) String() string {
	return a.name
}

var (
	AtlanTagColorGreen  = AtlanTagColor{"Green"}
	AtlanTagColorYellow = AtlanTagColor{"Yellow"}
	AtlanTagColorRed    = AtlanTagColor{"Red"}
	AtlanTagColorGray   = AtlanTagColor{"Gray"}
)

// UnmarshalJSON customizes the unmarshalling of an AtlanTagColor from JSON.
func (a *AtlanTagColor) UnmarshalJSON(data []byte) error {
	var name string
	if err := json.Unmarshal(data, &name); err != nil {
		return err
	}

	switch name {
	case "Green":
		*a = AtlanTagColorGreen
	case "Yellow":
		*a = AtlanTagColorYellow
	case "Red":
		*a = AtlanTagColorRed
	case "Gray":
		*a = AtlanTagColorGray
	default:
		*a = AtlanTagColor{name: name}
	}

	return nil
}

// MarshalJSON customizes the marshalling of an AtlanTagColor to JSON.
func (a AtlanTagColor) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.name)
}

type AtlanTypeCategory struct {
	Name string
}

func (a AtlanTypeCategory) String() string {
	return a.Name
}

// Constants representing type categories
var (
	AtlanTypeCategoryEntity           = AtlanTypeCategory{"ENTITY"}
	AtlanTypeCategoryRelationship     = AtlanTypeCategory{"RELATIONSHIP"}
	AtlanTypeCategoryEnum             = AtlanTypeCategory{"ENUM"}
	AtlanTypeCategoryStruct           = AtlanTypeCategory{"STRUCT"}
	AtlanTypeCategoryClassification   = AtlanTypeCategory{"CLASSIFICATION"}
	AtlanTypeCategoryBusinessMetadata = AtlanTypeCategory{"BUSINESS_METADATA"}
)

// UnmarshalJSON customizes the unmarshalling of an AtlanTypeCategory from JSON.
func (a *AtlanTypeCategory) UnmarshalJSON(data []byte) error {
	// Unmarshal the JSON data into a string.
	var categoryName string
	if err := json.Unmarshal(data, &categoryName); err != nil {
		return err
	}

	// Based on the categoryName, set the corresponding AtlanTypeCategory.
	switch categoryName {
	case "ENTITY":
		*a = AtlanTypeCategoryEntity
	case "RELATIONSHIP":
		*a = AtlanTypeCategoryRelationship
	case "ENUM":
		*a = AtlanTypeCategoryEnum
	case "STRUCT":
		*a = AtlanTypeCategoryStruct
	case "CLASSIFICATION":
		*a = AtlanTypeCategoryClassification
	case "BUSINESS_METADATA":
		*a = AtlanTypeCategoryBusinessMetadata
	default:
		// Handle unknown category case, could return an error
		*a = AtlanTypeCategory{Name: categoryName}
	}

	return nil
}

// MarshalJSON customizes the marshalling of an AtlanTypeCategory to JSON.
func (a AtlanTypeCategory) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.Name)
}

// AdminOperationType - Enum for admin operation types.
type AdminOperationType struct {
	Name string
}

func (a AdminOperationType) String() string {
	return a.Name
}

var (
	AdminOperationTypeCreate = AdminOperationType{"CREATE"}
	AdminOperationTypeUpdate = AdminOperationType{"UPDATE"}
	AdminOperationTypeDelete = AdminOperationType{"DELETE"}
	AddAdminOperationType    = AdminOperationType{"ACTION"}
)

func (a *AdminOperationType) UnmarshalJSON(data []byte) error {
	var operationName string
	if err := json.Unmarshal(data, &operationName); err != nil {
		return err
	}

	switch operationName {
	case "CREATE":
		*a = AdminOperationTypeCreate
	case "UPDATE":
		*a = AdminOperationTypeUpdate
	case "DELETE":
		*a = AdminOperationTypeDelete
	case "ACTION":
		*a = AddAdminOperationType
	default:
		*a = AdminOperationType{Name: operationName}
	}

	return nil
}

func (a AdminOperationType) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.Name)
}

// AdminResourceType - Enum for admin resource types.
type AdminResourceType struct {
	Name string
}

func (a AdminResourceType) String() string {
	return a.Name
}

var (
	AdminResourceTypeRealm                       = AdminResourceType{"REALM"}
	AdminResourceTypeRealmRole                   = AdminResourceType{"REALM_ROLE"}
	AdminResourceTypeRealmRoleMapping            = AdminResourceType{"REALM_ROLE_MAPPING"}
	AdminResourceTypeRealmScopeMapping           = AdminResourceType{"REALM_SCOPE_MAPPING"}
	AdminResourceTypeAuthFlow                    = AdminResourceType{"AUTH_FLOW"}
	AdminResourceTypeAuthExecutionFlow           = AdminResourceType{"AUTH_EXECUTION_FLOW"}
	AdminResourceTypeAuthExecution               = AdminResourceType{"AUTH_EXECUTION"}
	AdminResourceTypeAuthenticatorConfig         = AdminResourceType{"AUTHENTICATOR_CONFIG"}
	AdminResourceTypeRequiredAction              = AdminResourceType{"REQUIRED_ACTION"}
	AdminResourceTypeIdentityProvider            = AdminResourceType{"IDENTITY_PROVIDER"}
	AdminResourceTypeIdentityProviderMapper      = AdminResourceType{"IDENTITY_PROVIDER_MAPPER"}
	AdminResourceTypeProtocolMapper              = AdminResourceType{"PROTOCOL_MAPPER"}
	AdminResourceTypeUSER                        = AdminResourceType{"USER"}
	AdminResourceTypeUserLoginFailure            = AdminResourceType{"USER_LOGIN_FAILURE"}
	AdminResourceTypeUserSession                 = AdminResourceType{"USER_SESSION"}
	AdminResourceTypeUserFederationProvider      = AdminResourceType{"USER_FEDERATION_PROVIDER"}
	AdminResourceTypeUserFederationMapper        = AdminResourceType{"USER_FEDERATION_MAPPER"}
	AdminResourceTypeGroup                       = AdminResourceType{"GROUP"}
	AdminResourceTypeGroupMembership             = AdminResourceType{"GROUP_MEMBERSHIP"}
	AdminResourceTypeClient                      = AdminResourceType{"CLIENT"}
	AdminResourceTypeClientInitialAccessModel    = AdminResourceType{"CLIENT_INITIAL_ACCESS_MODEL"}
	AdminResourceTypeClientRole                  = AdminResourceType{"CLIENT_ROLE"}
	AdminResourceTypeClientRoleMapping           = AdminResourceType{"CLIENT_ROLE_MAPPING"}
	AdminResourceTypeClientScope                 = AdminResourceType{"CLIENT_SCOPE"}
	AdminResourceTypeClientScopeMapping          = AdminResourceType{"CLIENT_SCOPE_MAPPING"}
	AdminResourceTypeClientScopeClientMapping    = AdminResourceType{"CLIENT_SCOPE_CLIENT_MAPPING"}
	AdminResourceTypeClusterNode                 = AdminResourceType{"CLUSTER_NODE"}
	AdminResourceTypeComponent                   = AdminResourceType{"COMPONENT"}
	AdminResourceTypeAuthorizationResourceServer = AdminResourceType{"AUTHORIZATION_RESOURCE_SERVER"}
	AdminResourceTypeAuthorizationResource       = AdminResourceType{"AUTHORIZATION_RESOURCE"}
	AdminResourceTypeAuthorizationScope          = AdminResourceType{"AUTHORIZATION_SCOPE"}
	AdminResourceTypeAuthorizationPolicy         = AdminResourceType{"AUTHORIZATION_POLICY"}
	AdminResourceTypeCustom                      = AdminResourceType{"CUSTOM"}
)

func (a *AdminResourceType) UnmarshalJSON(data []byte) error {
	var resourceName string
	if err := json.Unmarshal(data, &resourceName); err != nil {
		return err
	}

	switch resourceName {
	case "REALM":
		*a = AdminResourceTypeRealm
	case "REALM_ROLE":
		*a = AdminResourceTypeRealmRole
	case "REALM_ROLE_MAPPING":
		*a = AdminResourceTypeRealmRoleMapping
	case "REALM_SCOPE_MAPPING":
		*a = AdminResourceTypeRealmScopeMapping
	case "AUTH_FLOW":
		*a = AdminResourceTypeAuthFlow
	case "AUTH_EXECUTION_FLOW":
		*a = AdminResourceTypeAuthExecutionFlow
	case "AUTH_EXECUTION":
		*a = AdminResourceTypeAuthExecution
	case "AUTHENTICATOR_CONFIG":
		*a = AdminResourceTypeAuthenticatorConfig
	case "REQUIRED_ACTION":
		*a = AdminResourceTypeRequiredAction
	case "IDENTITY_PROVIDER":
		*a = AdminResourceTypeIdentityProvider
	case "IDENTITY_PROVIDER_MAPPER":
		*a = AdminResourceTypeIdentityProviderMapper
	case "PROTOCOL_MAPPER":
		*a = AdminResourceTypeProtocolMapper
	case "USER":
		*a = AdminResourceTypeUSER
	case "USER_LOGIN_FAILURE":
		*a = AdminResourceTypeUserLoginFailure
	case "USER_SESSION":
		*a = AdminResourceTypeUserSession
	case "USER_FEDERATION_PROVIDER":
		*a = AdminResourceTypeUserFederationProvider
	case "USER_FEDERATION_MAPPER":
		*a = AdminResourceTypeUserFederationMapper
	case "GROUP":
		*a = AdminResourceTypeGroup
	case "GROUP_MEMBERSHIP":
		*a = AdminResourceTypeGroupMembership
	case "CLIENT":
		*a = AdminResourceTypeClient
	case "CLIENT_INITIAL_ACCESS_MODEL":
		*a = AdminResourceTypeClientInitialAccessModel
	case "CLIENT_ROLE":
		*a = AdminResourceTypeClientRole
	case "CLIENT_ROLE_MAPPING":
		*a = AdminResourceTypeClientRoleMapping
	case "CLIENT_SCOPE":
		*a = AdminResourceTypeClientScope
	case "CLIENT_SCOPE_MAPPING":
		*a = AdminResourceTypeClientScopeMapping
	case "CLIENT_SCOPE_CLIENT_MAPPING":
		*a = AdminResourceTypeClientScopeClientMapping
	case "CLUSTER_NODE":
		*a = AdminResourceTypeClusterNode
	case "COMPONENT":
		*a = AdminResourceTypeComponent
	case "AUTHORIZATION_RESOURCE_SERVER":
		*a = AdminResourceTypeAuthorizationResourceServer
	case "AUTHORIZATION_RESOURCE":
		*a = AdminResourceTypeAuthorizationResource
	case "AUTHORIZATION_SCOPE":
		*a = AdminResourceTypeAuthorizationScope
	case "AUTHORIZATION_POLICY":
		*a = AdminResourceTypeAuthorizationPolicy
	case "CUSTOM":
		*a = AdminResourceTypeCustom
	default:
		*a = AdminResourceType{Name: resourceName}
	}

	return nil
}

func (a AdminResourceType) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.Name)
}

// AnnouncementType represents the type of an announcement.
type AnnouncementType struct {
	Name string
}

func (a AnnouncementType) String() string {
	return a.Name
}

var (
	AnnouncementTypeInformation = AnnouncementType{"information"}
	AnnouncementTypeWARNING     = AnnouncementType{"warning"}
	AnnouncementTypeIssue       = AnnouncementType{"issue"}
)

func (a *AnnouncementType) UnmarshalJSON(data []byte) error {
	var announcementName string
	if err := json.Unmarshal(data, &announcementName); err != nil {
		return err
	}

	switch announcementName {
	case "information":
		*a = AnnouncementTypeInformation
	case "warning":
		*a = AnnouncementTypeWARNING
	case "issue":
		*a = AnnouncementTypeIssue
	default:
		*a = AnnouncementType{Name: announcementName}
	}

	return nil
}

func (a AnnouncementType) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.Name)
}

// AssetSidebarTab represents the tabs available in the asset sidebar.
type AssetSidebarTab struct {
	Name string
}

func (a AssetSidebarTab) String() string {
	return a.Name
}

var (
	AssetSidebarTabOverview         = AssetSidebarTab{"overview"}
	AssetSidebarTabCOLUMNS          = AssetSidebarTab{"Columns"}
	AssetSidebarTabRuns             = AssetSidebarTab{"Runs"}
	AssetSidebarTabTasks            = AssetSidebarTab{"Tasks"}
	AssetSidebarTabComponents       = AssetSidebarTab{"Components"}
	AssetSidebarTabProjects         = AssetSidebarTab{"Projects"}
	AssetSidebarTabCollections      = AssetSidebarTab{"Collections"}
	AssetSidebarTabUsage            = AssetSidebarTab{"Usage"}
	AssetSidebarTabObjects          = AssetSidebarTab{"Objects"}
	AssetSidebarTabLineage          = AssetSidebarTab{"Lineage"}
	AssetSidebarTabIncidents        = AssetSidebarTab{"Incidents"}
	AssetSidebarTabFields           = AssetSidebarTab{"Fields"}
	AssetSidebarTabVisuals          = AssetSidebarTab{"Visuals"}
	AssetSidebarTabVisualizations   = AssetSidebarTab{"Visualizations"}
	AssetSidebarTabSchemaObjects    = AssetSidebarTab{"Schema Objects"}
	AssetSidebarTabRelations        = AssetSidebarTab{"Relations"}
	AssetSidebarTabFactDimRelations = AssetSidebarTab{"Fact-Dim Relations"}
	AssetSidebarTabProfile          = AssetSidebarTab{"Profile"}
	AssetSidebarTabAssets           = AssetSidebarTab{"Assets"}
	AssetSidebarTabActivity         = AssetSidebarTab{"Activity"}
	AssetSidebarTabSchedules        = AssetSidebarTab{"Schedules"}
	AssetSidebarTabResources        = AssetSidebarTab{"Resources"}
	AssetSidebarTabQueries          = AssetSidebarTab{"Queries"}
	AssetSidebarTabRequests         = AssetSidebarTab{"Requests"}
	AssetSidebarTabProperties       = AssetSidebarTab{"Properties"}
	AssetSidebarTabMonteCarlo       = AssetSidebarTab{"Monte Carlo"}
	AssetSidebarTabDbtTest          = AssetSidebarTab{"dbt Test"}
	AssetSidebarTabSoda             = AssetSidebarTab{"Soda"}
)

// AtlanComparisonOperator represents comparison operators in Atlan.
type AtlanComparisonOperator struct {
	Name string
}

func (a AtlanComparisonOperator) String() string {
	return a.Name
}

var (
	AtlanComparisonOperatorLT          = AtlanComparisonOperator{"<"}
	AtlanComparisonOperatorGT          = AtlanComparisonOperator{">"}
	AtlanComparisonOperatorLTE         = AtlanComparisonOperator{"<="}
	AtlanComparisonOperatorGTE         = AtlanComparisonOperator{">="}
	AtlanComparisonOperatorEQ          = AtlanComparisonOperator{"="}
	AtlanComparisonOperatorNEQ         = AtlanComparisonOperator{"!="}
	AtlanComparisonOperatorIn          = AtlanComparisonOperator{"in"}
	AtlanComparisonOperatorLike        = AtlanComparisonOperator{"like"}
	AtlanComparisonOperatorStartsWith  = AtlanComparisonOperator{"startsWith"}
	AtlanComparisonOperatorEndsWith    = AtlanComparisonOperator{"endsWith"}
	AtlanComparisonOperatorContains    = AtlanComparisonOperator{"contains"}
	AtlanComparisonOperatorNotContains = AtlanComparisonOperator{"not_contains"}
	AtlanComparisonOperatorContainsAny = AtlanComparisonOperator{"containsAny"}
	AtlanComparisonOperatorContainsAll = AtlanComparisonOperator{"containsAll"}
	AtlanComparisonOperatorIsNull      = AtlanComparisonOperator{"isNull"}
	AtlanComparisonOperatorNotNull     = AtlanComparisonOperator{"notNull"}
	AtlanComparisonOperatorTimeRange   = AtlanComparisonOperator{"timerange"}
	AtlanComparisonOperatorNotEmpty    = AtlanComparisonOperator{"notEmpty"}
)

// AtlanConnectionCategory represents the category of a connection in Atlan.
type AtlanConnectionCategory struct {
	Name string
}

func (a AtlanConnectionCategory) String() string {
	return a.Name
}

var (
	AtlanConnectionCategoryWAREHOUSE      = AtlanConnectionCategory{"warehouse"}
	AtlanConnectionCategoryBI             = AtlanConnectionCategory{"bi"}
	AtlanConnectionCategoryObjectStore    = AtlanConnectionCategory{"ObjectStore"}
	AtlanConnectionCategorySAAS           = AtlanConnectionCategory{"SaaS"}
	AtlanConnectionCategoryLake           = AtlanConnectionCategory{"lake"}
	AtlanConnectionCategoryQueryEngine    = AtlanConnectionCategory{"queryengine"}
	AtlanConnectionCategoryELT            = AtlanConnectionCategory{"elt"}
	AtlanConnectionCategoryDATABASE       = AtlanConnectionCategory{"database"}
	AtlanConnectionCategoryAPI            = AtlanConnectionCategory{"API"}
	AtlanConnectionCategoryEventBus       = AtlanConnectionCategory{"eventbus"}
	AtlanConnectionCategoryDataQuality    = AtlanConnectionCategory{"data-quality"}
	AtlanConnectionCategorySchemaRegistry = AtlanConnectionCategory{"schema-registry"}
)

// UnmarshalJSON customizes the unmarshalling of an AssetSidebarTab from JSON.
func (a *AssetSidebarTab) UnmarshalJSON(data []byte) error {
	var tabName string
	if err := json.Unmarshal(data, &tabName); err != nil {
		return err
	}

	// Based on the tabName, set the corresponding AssetSidebarTab.
	switch tabName {
	case "overview":
		*a = AssetSidebarTabOverview
	case "Columns":
		*a = AssetSidebarTabCOLUMNS
	case "Runs":
		*a = AssetSidebarTabRuns
	case "Tasks":
		*a = AssetSidebarTabTasks
	case "Components":
		*a = AssetSidebarTabComponents
	case "Projects":
		*a = AssetSidebarTabProjects
	case "Collections":
		*a = AssetSidebarTabCollections
	case "Usage":
		*a = AssetSidebarTabUsage
	case "Objects":
		*a = AssetSidebarTabObjects
	case "Lineage":
		*a = AssetSidebarTabLineage
	case "Incidents":
		*a = AssetSidebarTabIncidents
	case "Fields":
		*a = AssetSidebarTabFields
	case "Visuals":
		*a = AssetSidebarTabVisuals
	case "Visualizations":
		*a = AssetSidebarTabVisualizations
	case "Schema Objects":
		*a = AssetSidebarTabSchemaObjects
	case "Relations":
		*a = AssetSidebarTabRelations
	case "Fact-Dim Relations":
		*a = AssetSidebarTabFactDimRelations
	case "Profile":
		*a = AssetSidebarTabProfile
	case "Assets":
		*a = AssetSidebarTabAssets
	case "Activity":
		*a = AssetSidebarTabActivity
	case "Schedules":
		*a = AssetSidebarTabSchedules
	case "Resources":
		*a = AssetSidebarTabResources
	case "Queries":
		*a = AssetSidebarTabQueries
	case "Requests":
		*a = AssetSidebarTabRequests
	case "Properties":
		*a = AssetSidebarTabProperties
	case "Monte Carlo":
		*a = AssetSidebarTabMonteCarlo
	case "dbt Test":
		*a = AssetSidebarTabDbtTest
	case "Soda":
		*a = AssetSidebarTabSoda
	default:
		*a = AssetSidebarTab{Name: tabName}
	}

	return nil
}

// MarshalJSON customizes the marshalling of an AssetSidebarTab to JSON.
func (a AssetSidebarTab) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.Name)
}

// AtlanConnectorType represents connector types with their categories.
type AtlanConnectorType struct {
	Value    string
	Category AtlanConnectionCategory
}

// ConnectorTypes is a map of all connector types for easy lookup.
var ConnectorTypes = map[string]AtlanConnectorType{
	"snowflake": {Value: "snowflake", Category: AtlanConnectionCategoryWAREHOUSE},
	"tableau":   {Value: "tableau", Category: AtlanConnectionCategoryBI},
	// Add other connectors here
}

// NewAtlanConnectorType creates a new AtlanConnectorType with the given value and category.
func NewAtlanConnectorType(value string, category AtlanConnectionCategory) AtlanConnectorType {
	return AtlanConnectorType{Value: value, Category: category}
}

// ToQualifiedName generates a qualified name for the AtlanConnectorType.
func (a AtlanConnectorType) ToQualifiedName() string {
	return fmt.Sprintf("default/%s/%d", a.Value, time.Now().Unix())
}

// GetConnectorTypeFromQualifiedName attempts to extract an AtlanConnectorType from a qualified name.
func GetConnectorTypeFromQualifiedName(qualifiedName string) (AtlanConnectorType, error) {
	tokens := strings.Split(qualifiedName, "/")
	if len(tokens) > 1 {
		if ct, exists := ConnectorTypes[tokens[1]]; exists {
			return ct, nil
		}
	}
	return AtlanConnectorType{}, fmt.Errorf("could not determine AtlanConnectorType from %s", qualifiedName)
}

// AtlanCustomAttributePrimitiveType simulates an enum for custom attribute primitive types.
type AtlanCustomAttributePrimitiveType struct {
	Name string
}

func (a AtlanCustomAttributePrimitiveType) String() string {
	return a.Name
}

var (
	AtlanCustomAttributeTypeString  = AtlanCustomAttributePrimitiveType{"string"}
	AtlanCustomAttributeTypeInteger = AtlanCustomAttributePrimitiveType{"int"}
	AtlanCustomAttributeTypeDecimal = AtlanCustomAttributePrimitiveType{"float"}
	AtlanCustomAttributeTypeBoolean = AtlanCustomAttributePrimitiveType{"boolean"}
	AtlanCustomAttributeTypeDate    = AtlanCustomAttributePrimitiveType{"date"}
	AtlanCustomAttributeTypeOptions = AtlanCustomAttributePrimitiveType{"enum"}
	AtlanCustomAttributeTypeUsers   = AtlanCustomAttributePrimitiveType{"users"}
	AtlanCustomAttributeTypeGroups  = AtlanCustomAttributePrimitiveType{"groups"}
	AtlanCustomAttributeTypeURL     = AtlanCustomAttributePrimitiveType{"url"}
	AtlanCustomAttributeTypeSQL     = AtlanCustomAttributePrimitiveType{"SQL"}
)

func (a *AtlanCustomAttributePrimitiveType) UnmarshalJSON(data []byte) error {
	var typeName string
	if err := json.Unmarshal(data, &typeName); err != nil {
		return err
	}

	switch typeName {
	case "string":
		*a = AtlanCustomAttributeTypeString
	case "int":
		*a = AtlanCustomAttributeTypeInteger
	case "float":
		*a = AtlanCustomAttributeTypeDecimal
	case "boolean":
		*a = AtlanCustomAttributeTypeBoolean
	case "date":
		*a = AtlanCustomAttributeTypeDate
	case "enum":
		*a = AtlanCustomAttributeTypeOptions
	case "users":
		*a = AtlanCustomAttributeTypeUsers
	case "groups":
		*a = AtlanCustomAttributeTypeGroups
	case "url":
		*a = AtlanCustomAttributeTypeURL
	case "SQL":
		*a = AtlanCustomAttributeTypeSQL
	default:
		*a = AtlanCustomAttributePrimitiveType{Name: typeName}
	}

	return nil
}

func (a AtlanCustomAttributePrimitiveType) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.Name)
}

// AtlanDeleteType simulates an enum for delete types.
type AtlanDeleteType struct {
	Name string
}

func (a AtlanDeleteType) String() string {
	return a.Name
}

var (
	AtlanDeleteTypeHard  = AtlanDeleteType{"HARD"}
	AtlanDeleteTypeSoft  = AtlanDeleteType{"SOFT"}
	AtlanDeleteTypePurge = AtlanDeleteType{"PURGE"}
)

func (a *AtlanDeleteType) UnmarshalJSON(data []byte) error {
	var deleteTypeName string
	if err := json.Unmarshal(data, &deleteTypeName); err != nil {
		return err
	}

	switch deleteTypeName {
	case "HARD":
		*a = AtlanDeleteTypeHard
	case "SOFT":
		*a = AtlanDeleteTypeSoft
	case "PURGE":
		*a = AtlanDeleteTypePurge
	default:
		*a = AtlanDeleteType{Name: deleteTypeName}
	}

	return nil
}

func (a AtlanDeleteType) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.Name)
}

// LiteralState represents the state of something.
type LiteralState struct {
	Name string
}

func (l LiteralState) String() string {
	return l.Name
}

var (
	LiteralStateActive  = LiteralState{"ACTIVE"}
	LiteralStateDeleted = LiteralState{"DELETED"}
	LiteralStatePurged  = LiteralState{"PURGED"}
)

func (l *LiteralState) UnmarshalJSON(data []byte) error {
	var stateName string
	if err := json.Unmarshal(data, &stateName); err != nil {
		return err
	}

	switch stateName {
	case "ACTIVE":
		*l = LiteralStateActive
	case "DELETED":
		*l = LiteralStateDeleted
	case "PURGED":
		*l = LiteralStatePurged
	default:
		*l = LiteralState{Name: stateName}
	}

	return nil
}

func (l LiteralState) MarshalJSON() ([]byte, error) {
	return json.Marshal(l.Name)
}

// SortOrder represents the sorting order.
type SortOrder struct {
	Name string
}

func (s SortOrder) String() string {
	return s.Name
}

var (
	SortOrderAscending  = SortOrder{"asc"}
	SortOrderDescending = SortOrder{"desc"}
)

func (a *SortOrder) UnmarshalJSON(data []byte) error {
	var sortOrderName string
	if err := json.Unmarshal(data, &sortOrderName); err != nil {
		return err
	}

	switch sortOrderName {
	case "asc":
		*a = SortOrderAscending
	case "desc":
		*a = SortOrderDescending
	default:
		*a = SortOrder{Name: sortOrderName}
	}

	return nil
}

// MarshalJSON customizes the marshalling of an AtlanTypeCategory to JSON.
func (a SortOrder) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.Name)
}

type AtlanIcon struct {
	name string
}

var (
	AtlanIconAtlanTag                    = AtlanIcon{"atlanTags"}
	AtlanIconAtlanShield                 = AtlanIcon{"atlanShield"}
	AtlanIconAddressBook                 = AtlanIcon{"PhAddressBook"}
	AtlanIconAirTrafficControl           = AtlanIcon{"PhAirTrafficControl"}
	AtlanIconAirplane                    = AtlanIcon{"PhAirplane"}
	AtlanIconAirplaneInFlight            = AtlanIcon{"PhAirplaneInFlight"}
	AtlanIconAirplaneLanding             = AtlanIcon{"PhAirplaneLanding"}
	AtlanIconAirplaneTakeoff             = AtlanIcon{"PhAirplaneTakeoff"}
	AtlanIconAirplaneTilt                = AtlanIcon{"PhAirplaneTilt"}
	AtlanIconAirplay                     = AtlanIcon{"PhAirplay"}
	AtlanIconAlarm                       = AtlanIcon{"PhAlarm"}
	AtlanIconAlien                       = AtlanIcon{"PhAlien"}
	AtlanIconAlignBottom                 = AtlanIcon{"PhAlignBottom"}
	AtlanIconAlignBottomSimple           = AtlanIcon{"PhAlignBottomSimple"}
	AtlanIconAlignCenterHorizontal       = AtlanIcon{"PhAlignCenterHorizontal"}
	AtlanIconAlignCenterHorizontalSimple = AtlanIcon{"PhAlignCenterHorizontalSimple"}
	AtlanIconAlignCenterVertical         = AtlanIcon{"PhAlignCenterVertical"}
	AtlanIconAlignCenterVerticalSimple   = AtlanIcon{"PhAlignCenterVerticalSimple"}
	AtlanIconAlignLeft                   = AtlanIcon{"PhAlignLeft"}
	AtlanIconAlignLeftSimple             = AtlanIcon{"PhAlignLeftSimple"}
	AtlanIconAlignRight                  = AtlanIcon{"PhAlignRight"}
	AtlanIconAlignRightSimple            = AtlanIcon{"PhAlignRightSimple"}
	AtlanIconAlignTop                    = AtlanIcon{"PhAlignTop"}
	AtlanIconAlignTopSimple              = AtlanIcon{"PhAlignTopSimple"}
	AtlanIconAmazonLogo                  = AtlanIcon{"PhAmazonLogo"}
	AtlanIconAnchor                      = AtlanIcon{"PhAnchor"}
	AtlanIconAnchorSimple                = AtlanIcon{"PhAnchorSimple"}
	AtlanIconAndroidLogo                 = AtlanIcon{"PhAndroidLogo"}
	AtlanIconAngularLogo                 = AtlanIcon{"PhAngularLogo"}
	AtlanIconAperture                    = AtlanIcon{"PhAperture"}
	AtlanIconAppStoreLogo                = AtlanIcon{"PhAppStoreLogo"}
	AtlanIconAppWindow                   = AtlanIcon{"PhAppWindow"}
	AtlanIconAppleLogo                   = AtlanIcon{"PhAppleLogo"}
	AtlanIconApplePodcastsLogo           = AtlanIcon{"PhApplePodcastsLogo"}
	AtlanIconArchive                     = AtlanIcon{"PhArchive"}
	AtlanIconArchiveBox                  = AtlanIcon{"PhArchiveBox"}
	AtlanIconArchiveTray                 = AtlanIcon{"PhArchiveTray"}
	AtlanIconArmchair                    = AtlanIcon{"PhArmchair"}
	AtlanIconArrowArcLeft                = AtlanIcon{"PhArrowArcLeft"}
	AtlanIconArrowArcRight               = AtlanIcon{"PhArrowArcRight"}
	AtlanIconArrowBendDoubleUpLeft       = AtlanIcon{"PhArrowBendDoubleUpLeft"}
	AtlanIconArrowBendDoubleUpRight      = AtlanIcon{"PhArrowBendDoubleUpRight"}
	AtlanIconArrowBendDownLeft           = AtlanIcon{"PhArrowBendDownLeft"}
	AtlanIconArrowBendDownRight          = AtlanIcon{"PhArrowBendDownRight"}
	AtlanIconArrowBendLeftDown           = AtlanIcon{"PhArrowBendLeftDown"}
	AtlanIconArrowBendLeftUp             = AtlanIcon{"PhArrowBendLeftUp"}
	AtlanIconArrowBendRightDown          = AtlanIcon{"PhArrowBendRightDown"}
	AtlanIconArrowBendRightUp            = AtlanIcon{"PhArrowBendRightUp"}
	AtlanIconArrowBendUpLeft             = AtlanIcon{"PhArrowBendUpLeft"}
	AtlanIconArrowBendUpRight            = AtlanIcon{"PhArrowBendUpRight"}
	AtlanIconArrowCircleDown             = AtlanIcon{"PhArrowCircleDown"}
	AtlanIconArrowCircleDownLeft         = AtlanIcon{"PhArrowCircleDownLeft"}
	AtlanIconArrowCircleDownRight        = AtlanIcon{"PhArrowCircleDownRight"}
	AtlanIconArrowCircleLeft             = AtlanIcon{"PhArrowCircleLeft"}
	AtlanIconArrowCircleRight            = AtlanIcon{"PhArrowCircleRight"}
	AtlanIconArrowCircleUp               = AtlanIcon{"PhArrowCircleUp"}
	AtlanIconArrowCircleUpLeft           = AtlanIcon{"PhArrowCircleUpLeft"}
	AtlanIconArrowCircleUpRight          = AtlanIcon{"PhArrowCircleUpRight"}
	AtlanIconArrowClockwise              = AtlanIcon{"PhArrowClockwise"}
	AtlanIconArrowCounterClockwise       = AtlanIcon{"PhArrowCounterClockwise"}
	AtlanIconArrowDown                   = AtlanIcon{"PhArrowDown"}
	AtlanIconArrowDownLeft               = AtlanIcon{"PhArrowDownLeft"}
	AtlanIconArrowDownRight              = AtlanIcon{"PhArrowDownRight"}
	AtlanIconArrowElbowDownLeft          = AtlanIcon{"PhArrowElbowDownLeft"}
	AtlanIconArrowElbowDownRight         = AtlanIcon{"PhArrowElbowDownRight"}
	AtlanIconArrowElbowLeft              = AtlanIcon{"PhArrowElbowLeft"}
	AtlanIconArrowElbowLeftDown          = AtlanIcon{"PhArrowElbowLeftDown"}
	AtlanIconArrowElbowLeftUp            = AtlanIcon{"PhArrowElbowLeftUp"}
	AtlanIconArrowElbowRight             = AtlanIcon{"PhArrowElbowRight"}
	AtlanIconArrowElbowRightDown         = AtlanIcon{"PhArrowElbowRightDown"}
	AtlanIconArrowElbowRightUp           = AtlanIcon{"PhArrowElbowRightUp"}
	AtlanIconArrowElbowUpLeft            = AtlanIcon{"PhArrowElbowUpLeft"}
	AtlanIconArrowElbowUpRight           = AtlanIcon{"PhArrowElbowUpRight"}
	AtlanIconArrowFatDown                = AtlanIcon{"PhArrowFatDown"}
	AtlanIconArrowFatLeft                = AtlanIcon{"PhArrowFatLeft"}
	AtlanIconArrowFatLineDown            = AtlanIcon{"PhArrowFatLineDown"}
	AtlanIconArrowFatLineLeft            = AtlanIcon{"PhArrowFatLineLeft"}
	AtlanIconArrowFatLineRight           = AtlanIcon{"PhArrowFatLineRight"}
	AtlanIconArrowFatLineUp              = AtlanIcon{"PhArrowFatLineUp"}
	AtlanIconArrowFatLinesDown           = AtlanIcon{"PhArrowFatLinesDown"}
	AtlanIconArrowFatLinesLeft           = AtlanIcon{"PhArrowFatLinesLeft"}
	AtlanIconArrowFatLinesRight          = AtlanIcon{"PhArrowFatLinesRight"}
	AtlanIconArrowFatLinesUp             = AtlanIcon{"PhArrowFatLinesUp"}
	AtlanIconArrowFatRight               = AtlanIcon{"PhArrowFatRight"}
	AtlanIconArrowFatUp                  = AtlanIcon{"PhArrowFatUp"}
	AtlanIconArrowLeft                   = AtlanIcon{"PhArrowLeft"}
	AtlanIconArrowLineDown               = AtlanIcon{"PhArrowLineDown"}
	AtlanIconArrowLineDownLeft           = AtlanIcon{"PhArrowLineDownLeft"}
	AtlanIconArrowLineDownRight          = AtlanIcon{"PhArrowLineDownRight"}
	AtlanIconArrowLineLeft               = AtlanIcon{"PhArrowLineLeft"}
	AtlanIconArrowLineRight              = AtlanIcon{"PhArrowLineRight"}
	AtlanIconArrowLineUp                 = AtlanIcon{"PhArrowLineUp"}
	AtlanIconArrowLineUpLeft             = AtlanIcon{"PhArrowLineUpLeft"}
	AtlanIconArrowLineUpRight            = AtlanIcon{"PhArrowLineUpRight"}
	AtlanIconArrowRight                  = AtlanIcon{"PhArrowRight"}
	AtlanIconArrowSquareDown             = AtlanIcon{"PhArrowSquareDown"}
	AtlanIconArrowSquareDownLeft         = AtlanIcon{"PhArrowSquareDownLeft"}
	AtlanIconArrowSquareDownRight        = AtlanIcon{"PhArrowSquareDownRight"}
	AtlanIconArrowSquareIn               = AtlanIcon{"PhArrowSquareIn"}
	AtlanIconArrowSquareLeft             = AtlanIcon{"PhArrowSquareLeft"}
	AtlanIconArrowSquareOut              = AtlanIcon{"PhArrowSquareOut"}
	AtlanIconArrowSquareRight            = AtlanIcon{"PhArrowSquareRight"}
	AtlanIconArrowSquareUp               = AtlanIcon{"PhArrowSquareUp"}
	AtlanIconArrowSquareUpLeft           = AtlanIcon{"PhArrowSquareUpLeft"}
	AtlanIconArrowSquareUpRight          = AtlanIcon{"PhArrowSquareUpRight"}
	AtlanIconArrowUDownLeft              = AtlanIcon{"PhArrowUDownLeft"}
	AtlanIconArrowUDownRight             = AtlanIcon{"PhArrowUDownRight"}
	AtlanIconArrowULeftDown              = AtlanIcon{"PhArrowULeftDown"}
	AtlanIconArrowULeftUp                = AtlanIcon{"PhArrowULeftUp"}
	AtlanIconArrowURightDown             = AtlanIcon{"PhArrowURightDown"}
	AtlanIconArrowURightUp               = AtlanIcon{"PhArrowURightUp"}
	AtlanIconArrowUUpLeft                = AtlanIcon{"PhArrowUUpLeft"}
	AtlanIconArrowUUpRight               = AtlanIcon{"PhArrowUUpRight"}
	AtlanIconArrowUp                     = AtlanIcon{"PhArrowUp"}
	AtlanIconArrowUpLeft                 = AtlanIcon{"PhArrowUpLeft"}
	AtlanIconArrowUpRight                = AtlanIcon{"PhArrowUpRight"}
	AtlanIconArrowsClockwise             = AtlanIcon{"PhArrowsClockwise"}
	AtlanIconArrowsCounterClockwise      = AtlanIcon{"PhArrowsCounterClockwise"}
	AtlanIconArrowsDownUp                = AtlanIcon{"PhArrowsDownUp"}
	AtlanIconArrowsHorizontal            = AtlanIcon{"PhArrowsHorizontal"}
	AtlanIconArrowsIn                    = AtlanIcon{"PhArrowsIn"}
	AtlanIconArrowsInCardinal            = AtlanIcon{"PhArrowsInCardinal"}
	AtlanIconArrowsInLineHorizontal      = AtlanIcon{"PhArrowsInLineHorizontal"}
	AtlanIconArrowsInLineVertical        = AtlanIcon{"PhArrowsInLineVertical"}
	AtlanIconArrowsInSimple              = AtlanIcon{"PhArrowsInSimple"}
	AtlanIconArrowsLeftRight             = AtlanIcon{"PhArrowsLeftRight"}
	AtlanIconArrowsMerge                 = AtlanIcon{"PhArrowsMerge"}
	AtlanIconArrowsOut                   = AtlanIcon{"PhArrowsOut"}
	AtlanIconArrowsOutCardinal           = AtlanIcon{"PhArrowsOutCardinal"}
	AtlanIconArrowsOutLineHorizontal     = AtlanIcon{"PhArrowsOutLineHorizontal"}
	AtlanIconArrowsOutLineVertical       = AtlanIcon{"PhArrowsOutLineVertical"}
	AtlanIconArrowsOutSimple             = AtlanIcon{"PhArrowsOutSimple"}
	AtlanIconArrowsSplit                 = AtlanIcon{"PhArrowsSplit"}
	AtlanIconArrowsVertical              = AtlanIcon{"PhArrowsVertical"}
	AtlanIconArticle                     = AtlanIcon{"PhArticle"}
	AtlanIconArticleMedium               = AtlanIcon{"PhArticleMedium"}
	AtlanIconArticleNyTimes              = AtlanIcon{"PhArticleNyTimes"}
	AtlanIconAsterisk                    = AtlanIcon{"PhAsterisk"}
	AtlanIconAsteriskSimple              = AtlanIcon{"PhAsteriskSimple"}
	AtlanIconAt                          = AtlanIcon{"PhAt"}
	AtlanIconAtom                        = AtlanIcon{"PhAtom"}
	AtlanIconBaby                        = AtlanIcon{"PhBaby"}
	AtlanIconBackpack                    = AtlanIcon{"PhBackpack"}
	AtlanIconBackspace                   = AtlanIcon{"PhBackspace"}
	AtlanIconBag                         = AtlanIcon{"PhBag"}
	AtlanIconBagSimple                   = AtlanIcon{"PhBagSimple"}
	AtlanIconBalloon                     = AtlanIcon{"PhBalloon"}
	AtlanIconBandaids                    = AtlanIcon{"PhBandaids"}
	AtlanIconBank                        = AtlanIcon{"PhBank"}
	AtlanIconBarbell                     = AtlanIcon{"PhBarbell"}
	AtlanIconBarcode                     = AtlanIcon{"PhBarcode"}
	AtlanIconBarricade                   = AtlanIcon{"PhBarricade"}
	AtlanIconBaseball                    = AtlanIcon{"PhBaseball"}
	AtlanIconBaseballCap                 = AtlanIcon{"PhBaseballCap"}
	AtlanIconBasket                      = AtlanIcon{"PhBasket"}
	AtlanIconBasketball                  = AtlanIcon{"PhBasketball"}
	AtlanIconBathtub                     = AtlanIcon{"PhBathtub"}
	AtlanIconBatteryCharging             = AtlanIcon{"PhBatteryCharging"}
	AtlanIconBatteryChargingVertical     = AtlanIcon{"PhBatteryChargingVertical"}
	AtlanIconBatteryEmpty                = AtlanIcon{"PhBatteryEmpty"}
	AtlanIconBatteryFull                 = AtlanIcon{"PhBatteryFull"}
	AtlanIconBatteryHigh                 = AtlanIcon{"PhBatteryHigh"}
	AtlanIconBatteryLow                  = AtlanIcon{"PhBatteryLow"}
	AtlanIconBatteryMedium               = AtlanIcon{"PhBatteryMedium"}
	AtlanIconBatteryPlus                 = AtlanIcon{"PhBatteryPlus"}
	AtlanIconBatteryPlusVertical         = AtlanIcon{"PhBatteryPlusVertical"}
	AtlanIconBatteryVerticalEmpty        = AtlanIcon{"PhBatteryVerticalEmpty"}
	AtlanIconBatteryVerticalFull         = AtlanIcon{"PhBatteryVerticalFull"}
	AtlanIconBatteryVerticalHigh         = AtlanIcon{"PhBatteryVerticalHigh"}
	AtlanIconBatteryVerticalLow          = AtlanIcon{"PhBatteryVerticalLow"}
	AtlanIconBatteryVerticalMedium       = AtlanIcon{"PhBatteryVerticalMedium"}
	AtlanIconBatteryWarning              = AtlanIcon{"PhBatteryWarning"}
	AtlanIconBatteryWarningVertical      = AtlanIcon{"PhBatteryWarningVertical"}
	AtlanIconBed                         = AtlanIcon{"PhBed"}
	AtlanIconBeerBottle                  = AtlanIcon{"PhBeerBottle"}
	AtlanIconBeerStein                   = AtlanIcon{"PhBeerStein"}
	AtlanIconBehanceLogo                 = AtlanIcon{"PhBehanceLogo"}
	AtlanIconBell                        = AtlanIcon{"PhBell"}
	AtlanIconBellRinging                 = AtlanIcon{"PhBellRinging"}
	AtlanIconBellSimple                  = AtlanIcon{"PhBellSimple"}
	AtlanIconBellSimpleRinging           = AtlanIcon{"PhBellSimpleRinging"}
	AtlanIconBellSimpleSlash             = AtlanIcon{"PhBellSimpleSlash"}
	AtlanIconBellSimpleZ                 = AtlanIcon{"PhBellSimpleZ"}
	AtlanIconBellSlash                   = AtlanIcon{"PhBellSlash"}
	AtlanIconBellZ                       = AtlanIcon{"PhBellZ"}
	AtlanIconBezierCurve                 = AtlanIcon{"PhBezierCurve"}
	AtlanIconBicycle                     = AtlanIcon{"PhBicycle"}
	AtlanIconBinoculars                  = AtlanIcon{"PhBinoculars"}
	AtlanIconBird                        = AtlanIcon{"PhBird"}
	AtlanIconBluetooth                   = AtlanIcon{"PhBluetooth"}
	AtlanIconBluetoothConnected          = AtlanIcon{"PhBluetoothConnected"}
	AtlanIconBluetoothSlash              = AtlanIcon{"PhBluetoothSlash"}
	AtlanIconBluetoothX                  = AtlanIcon{"PhBluetoothX"}
	AtlanIconBoat                        = AtlanIcon{"PhBoat"}
	AtlanIconBone                        = AtlanIcon{"PhBone"}
	AtlanIconBook                        = AtlanIcon{"PhBook"}
	AtlanIconBookBookmark                = AtlanIcon{"PhBookBookmark"}
	AtlanIconBookOpen                    = AtlanIcon{"PhBookOpen"}
	AtlanIconBookOpenText                = AtlanIcon{"PhBookOpenText"}
	AtlanIconBookmark                    = AtlanIcon{"PhBookmark"}
	AtlanIconBookmarkSimple              = AtlanIcon{"PhBookmarkSimple"}
	AtlanIconBookmarks                   = AtlanIcon{"PhBookmarks"}
	AtlanIconBookmarksSimple             = AtlanIcon{"PhBookmarksSimple"}
	AtlanIconBooks                       = AtlanIcon{"PhBooks"}
	AtlanIconBoot                        = AtlanIcon{"PhBoot"}
	AtlanIconBoundingBox                 = AtlanIcon{"PhBoundingBox"}
	AtlanIconBowlFood                    = AtlanIcon{"PhBowlFood"}
	AtlanIconBracketsAngle               = AtlanIcon{"PhBracketsAngle"}
	AtlanIconBracketsCurly               = AtlanIcon{"PhBracketsCurly"}
	AtlanIconBracketsRound               = AtlanIcon{"PhBracketsRound"}
	AtlanIconBracketsSquare              = AtlanIcon{"PhBracketsSquare"}
	AtlanIconBrain                       = AtlanIcon{"PhBrain"}
	AtlanIconBrandy                      = AtlanIcon{"PhBrandy"}
	AtlanIconBridge                      = AtlanIcon{"PhBridge"}
	AtlanIconBriefcase                   = AtlanIcon{"PhBriefcase"}
	AtlanIconBriefcaseMetal              = AtlanIcon{"PhBriefcaseMetal"}
	AtlanIconBroadcast                   = AtlanIcon{"PhBroadcast"}
	AtlanIconBroom                       = AtlanIcon{"PhBroom"}
	AtlanIconBrowser                     = AtlanIcon{"PhBrowser"}
	AtlanIconBrowsers                    = AtlanIcon{"PhBrowsers"}
	AtlanIconBugBeetle                   = AtlanIcon{"PhBugBeetle"}
	AtlanIconBug                         = AtlanIcon{"PhBug"}
	AtlanIconBugDroid                    = AtlanIcon{"PhBugDroid"}
	AtlanIconBuildings                   = AtlanIcon{"PhBuildings"}
	AtlanIconBus                         = AtlanIcon{"PhBus"}
	AtlanIconButterfly                   = AtlanIcon{"PhButterfly"}
	AtlanIconCactus                      = AtlanIcon{"PhCactus"}
	AtlanIconCake                        = AtlanIcon{"PhCake"}
	AtlanIconCalculator                  = AtlanIcon{"PhCalculator"}
	AtlanIconCalendarBlank               = AtlanIcon{"PhCalendarBlank"}
	AtlanIconCalendar                    = AtlanIcon{"PhCalendar"}
	AtlanIconCalendarCheck               = AtlanIcon{"PhCalendarCheck"}
	AtlanIconCalendarPlus                = AtlanIcon{"PhCalendarPlus"}
	AtlanIconCalendarX                   = AtlanIcon{"PhCalendarX"}
	AtlanIconCallBell                    = AtlanIcon{"PhCallBell"}
	AtlanIconCamera                      = AtlanIcon{"PhCamera"}
	AtlanIconCameraPlus                  = AtlanIcon{"PhCameraPlus"}
	AtlanIconCameraRotate                = AtlanIcon{"PhCameraRotate"}
	AtlanIconCameraSlash                 = AtlanIcon{"PhCameraSlash"}
	AtlanIconCampfire                    = AtlanIcon{"PhCampfire"}
	AtlanIconCar                         = AtlanIcon{"PhCar"}
	AtlanIconCarProfile                  = AtlanIcon{"PhCarProfile"}
	AtlanIconCarSimple                   = AtlanIcon{"PhCarSimple"}
	AtlanIconCardholder                  = AtlanIcon{"PhCardholder"}
	AtlanIconCards                       = AtlanIcon{"PhCards"}
	AtlanIconCaretCircleDoubleDown       = AtlanIcon{"PhCaretCircleDoubleDown"}
	AtlanIconCaretCircleDoubleLeft       = AtlanIcon{"PhCaretCircleDoubleLeft"}
	AtlanIconCaretCircleDoubleRight      = AtlanIcon{"PhCaretCircleDoubleRight"}
	AtlanIconCaretCircleDoubleUp         = AtlanIcon{"PhCaretCircleDoubleUp"}
	AtlanIconCaretCircleDown             = AtlanIcon{"PhCaretCircleDown"}
	AtlanIconCaretCircleLeft             = AtlanIcon{"PhCaretCircleLeft"}
	AtlanIconCaretCircleRight            = AtlanIcon{"PhCaretCircleRight"}
	AtlanIconCaretCircleUp               = AtlanIcon{"PhCaretCircleUp"}
	AtlanIconCaretCircleUpDown           = AtlanIcon{"PhCaretCircleUpDown"}
	AtlanIconCaretDoubleDown             = AtlanIcon{"PhCaretDoubleDown"}
	AtlanIconCaretDoubleLeft             = AtlanIcon{"PhCaretDoubleLeft"}
	AtlanIconCaretDoubleRight            = AtlanIcon{"PhCaretDoubleRight"}
	AtlanIconCaretDoubleUp               = AtlanIcon{"PhCaretDoubleUp"}
	AtlanIconCaretDown                   = AtlanIcon{"PhCaretDown"}
	AtlanIconCaretLeft                   = AtlanIcon{"PhCaretLeft"}
	AtlanIconCaretRight                  = AtlanIcon{"PhCaretRight"}
	AtlanIconCaretUp                     = AtlanIcon{"PhCaretUp"}
	AtlanIconCaretUpDown                 = AtlanIcon{"PhCaretUpDown"}
	AtlanIconCarrot                      = AtlanIcon{"PhCarrot"}
	AtlanIconCassetteTape                = AtlanIcon{"PhCassetteTape"}
	AtlanIconCastleTurret                = AtlanIcon{"PhCastleTurret"}
	AtlanIconCat                         = AtlanIcon{"PhCat"}
	AtlanIconCellSignalFull              = AtlanIcon{"PhCellSignalFull"}
	AtlanIconCellSignalHigh              = AtlanIcon{"PhCellSignalHigh"}
	AtlanIconCellSignalLow               = AtlanIcon{"PhCellSignalLow"}
	AtlanIconCellSignalMedium            = AtlanIcon{"PhCellSignalMedium"}
	AtlanIconCellSignalNone              = AtlanIcon{"PhCellSignalNone"}
	AtlanIconCellSignalSlash             = AtlanIcon{"PhCellSignalSlash"}
	AtlanIconCellSignalX                 = AtlanIcon{"PhCellSignalX"}
	AtlanIconCertificate                 = AtlanIcon{"PhCertificate"}
	AtlanIconChair                       = AtlanIcon{"PhChair"}
	AtlanIconChalkboard                  = AtlanIcon{"PhChalkboard"}
	AtlanIconChalkboardSimple            = AtlanIcon{"PhChalkboardSimple"}
	AtlanIconChalkboardTeacher           = AtlanIcon{"PhChalkboardTeacher"}
	AtlanIconChampagne                   = AtlanIcon{"PhChampagne"}
	AtlanIconChargingStation             = AtlanIcon{"PhChargingStation"}
	AtlanIconChartBar                    = AtlanIcon{"PhChartBar"}
	AtlanIconChartBarHorizontal          = AtlanIcon{"PhChartBarHorizontal"}
	AtlanIconChartDonut                  = AtlanIcon{"PhChartDonut"}
	AtlanIconChartLine                   = AtlanIcon{"PhChartLine"}
	AtlanIconChartLineDown               = AtlanIcon{"PhChartLineDown"}
	AtlanIconChartLineUp                 = AtlanIcon{"PhChartLineUp"}
	AtlanIconChartPie                    = AtlanIcon{"PhChartPie"}
	AtlanIconChartPieSlice               = AtlanIcon{"PhChartPieSlice"}
	AtlanIconChartPolar                  = AtlanIcon{"PhChartPolar"}
	AtlanIconChartScatter                = AtlanIcon{"PhChartScatter"}
	AtlanIconChat                        = AtlanIcon{"PhChat"}
	AtlanIconChatCentered                = AtlanIcon{"PhChatCentered"}
	AtlanIconChatCenteredDots            = AtlanIcon{"PhChatCenteredDots"}
	AtlanIconChatCenteredText            = AtlanIcon{"PhChatCenteredText"}
	AtlanIconChatCircle                  = AtlanIcon{"PhChatCircle"}
	AtlanIconChatCircleDots              = AtlanIcon{"PhChatCircleDots"}
	AtlanIconChatCircleText              = AtlanIcon{"PhChatCircleText"}
	AtlanIconChatDots                    = AtlanIcon{"PhChatDots"}
	AtlanIconChatTeardrop                = AtlanIcon{"PhChatTeardrop"}
	AtlanIconChatTeardropDots            = AtlanIcon{"PhChatTeardropDots"}
	AtlanIconChatTeardropText            = AtlanIcon{"PhChatTeardropText"}
	AtlanIconChatText                    = AtlanIcon{"PhChatText"}
	AtlanIconChats                       = AtlanIcon{"PhChats"}
	AtlanIconChatsCircle                 = AtlanIcon{"PhChatsCircle"}
	AtlanIconChatsTeardrop               = AtlanIcon{"PhChatsTeardrop"}
	AtlanIconCheck                       = AtlanIcon{"PhCheck"}
	AtlanIconCheckCircle                 = AtlanIcon{"PhCheckCircle"}
	AtlanIconCheckFat                    = AtlanIcon{"PhCheckFat"}
	AtlanIconCheckSquare                 = AtlanIcon{"PhCheckSquare"}
	AtlanIconCheckSquareOffset           = AtlanIcon{"PhCheckSquareOffset"}
	AtlanIconChecks                      = AtlanIcon{"PhChecks"}
	AtlanIconChurch                      = AtlanIcon{"PhChurch"}
	AtlanIconCircle                      = AtlanIcon{"PhCircle"}
	AtlanIconCircleDashed                = AtlanIcon{"PhCircleDashed"}
	AtlanIconCircleHalf                  = AtlanIcon{"PhCircleHalf"}
	AtlanIconCircleHalfTilt              = AtlanIcon{"PhCircleHalfTilt"}
	AtlanIconCircleNotch                 = AtlanIcon{"PhCircleNotch"}
	AtlanIconCirclesFour                 = AtlanIcon{"PhCirclesFour"}
	AtlanIconCirclesThree                = AtlanIcon{"PhCirclesThree"}
	AtlanIconCirclesThreePlus            = AtlanIcon{"PhCirclesThreePlus"}
	AtlanIconCircuitry                   = AtlanIcon{"PhCircuitry"}
	AtlanIconClipboard                   = AtlanIcon{"PhClipboard"}
	AtlanIconClipboardText               = AtlanIcon{"PhClipboardText"}
	AtlanIconClockAfternoon              = AtlanIcon{"PhClockAfternoon"}
	AtlanIconClock                       = AtlanIcon{"PhClock"}
	AtlanIconClockClockwise              = AtlanIcon{"PhClockClockwise"}
	AtlanIconClockCountdown              = AtlanIcon{"PhClockCountdown"}
	AtlanIconClockCounterClockwise       = AtlanIcon{"PhClockCounterClockwise"}
	AtlanIconClosedCaptioning            = AtlanIcon{"PhClosedCaptioning"}
	AtlanIconCloudArrowDown              = AtlanIcon{"PhCloudArrowDown"}
	AtlanIconCloudArrowUp                = AtlanIcon{"PhCloudArrowUp"}
	AtlanIconCloud                       = AtlanIcon{"PhCloud"}
	AtlanIconCloudCheck                  = AtlanIcon{"PhCloudCheck"}
	AtlanIconCloudFog                    = AtlanIcon{"PhCloudFog"}
	AtlanIconCloudLightning              = AtlanIcon{"PhCloudLightning"}
	AtlanIconCloudMoon                   = AtlanIcon{"PhCloudMoon"}
	AtlanIconCloudRain                   = AtlanIcon{"PhCloudRain"}
	AtlanIconCloudSlash                  = AtlanIcon{"PhCloudSlash"}
	AtlanIconCloudSnow                   = AtlanIcon{"PhCloudSnow"}
	AtlanIconCloudSun                    = AtlanIcon{"PhCloudSun"}
	AtlanIconCloudWarning                = AtlanIcon{"PhCloudWarning"}
	AtlanIconCloudX                      = AtlanIcon{"PhCloudX"}
	AtlanIconClub                        = AtlanIcon{"PhClub"}
	AtlanIconCoatHanger                  = AtlanIcon{"PhCoatHanger"}
	AtlanIconCodaLogo                    = AtlanIcon{"PhCodaLogo"}
	AtlanIconCodeBlock                   = AtlanIcon{"PhCodeBlock"}
	AtlanIconCode                        = AtlanIcon{"PhCode"}
	AtlanIconCodeSimple                  = AtlanIcon{"PhCodeSimple"}
	AtlanIconCodepenLogo                 = AtlanIcon{"PhCodepenLogo"}
	AtlanIconCodesandboxLogo             = AtlanIcon{"PhCodesandboxLogo"}
	AtlanIconCoffee                      = AtlanIcon{"PhCoffee"}
	AtlanIconCoin                        = AtlanIcon{"PhCoin"}
	AtlanIconCoinVertical                = AtlanIcon{"PhCoinVertical"}
	AtlanIconCoins                       = AtlanIcon{"PhCoins"}
	AtlanIconColumns                     = AtlanIcon{"PhColumns"}
	AtlanIconCommand                     = AtlanIcon{"PhCommand"}
	AtlanIconCompass                     = AtlanIcon{"PhCompass"}
	AtlanIconCompassTool                 = AtlanIcon{"PhCompassTool"}
	AtlanIconComputerTower               = AtlanIcon{"PhComputerTower"}
	AtlanIconConfetti                    = AtlanIcon{"PhConfetti"}
	AtlanIconContactlessPayment          = AtlanIcon{"PhContactlessPayment"}
	AtlanIconControl                     = AtlanIcon{"PhControl"}
	AtlanIconCookie                      = AtlanIcon{"PhCookie"}
	AtlanIconCookingPot                  = AtlanIcon{"PhCookingPot"}
	AtlanIconCopy                        = AtlanIcon{"PhCopy"}
	AtlanIconCopySimple                  = AtlanIcon{"PhCopySimple"}
	AtlanIconCopyleft                    = AtlanIcon{"PhCopyleft"}
	AtlanIconCopyright                   = AtlanIcon{"PhCopyright"}
	AtlanIconCornersIn                   = AtlanIcon{"PhCornersIn"}
	AtlanIconCornersOut                  = AtlanIcon{"PhCornersOut"}
	AtlanIconCouch                       = AtlanIcon{"PhCouch"}
	AtlanIconCpu                         = AtlanIcon{"PhCpu"}
	AtlanIconCreditCard                  = AtlanIcon{"PhCreditCard"}
	AtlanIconCrop                        = AtlanIcon{"PhCrop"}
	AtlanIconCross                       = AtlanIcon{"PhCross"}
	AtlanIconCrosshair                   = AtlanIcon{"PhCrosshair"}
	AtlanIconCrosshairSimple             = AtlanIcon{"PhCrosshairSimple"}
	AtlanIconCrown                       = AtlanIcon{"PhCrown"}
	AtlanIconCrownSimple                 = AtlanIcon{"PhCrownSimple"}
	AtlanIconCube                        = AtlanIcon{"PhCube"}
	AtlanIconCubeFocus                   = AtlanIcon{"PhCubeFocus"}
	AtlanIconCubeTransparent             = AtlanIcon{"PhCubeTransparent"}
	AtlanIconCurrencyBtc                 = AtlanIcon{"PhCurrencyBtc"}
	AtlanIconCurrencyCircleDollar        = AtlanIcon{"PhCurrencyCircleDollar"}
	AtlanIconCurrencyCny                 = AtlanIcon{"PhCurrencyCny"}
	AtlanIconCurrencyDollar              = AtlanIcon{"PhCurrencyDollar"}
	AtlanIconCurrencyDollarSimple        = AtlanIcon{"PhCurrencyDollarSimple"}
	AtlanIconCurrencyEth                 = AtlanIcon{"PhCurrencyEth"}
	AtlanIconCurrencyEur                 = AtlanIcon{"PhCurrencyEur"}
	AtlanIconCurrencyGbp                 = AtlanIcon{"PhCurrencyGbp"}
	AtlanIconCurrencyInr                 = AtlanIcon{"PhCurrencyInr"}
	AtlanIconCurrencyJpy                 = AtlanIcon{"PhCurrencyJpy"}
	AtlanIconCurrencyKrw                 = AtlanIcon{"PhCurrencyKrw"}
	AtlanIconCurrencyKzt                 = AtlanIcon{"PhCurrencyKzt"}
	AtlanIconCurrencyNgn                 = AtlanIcon{"PhCurrencyNgn"}
	AtlanIconCurrencyRub                 = AtlanIcon{"PhCurrencyRub"}
	AtlanIconCursor                      = AtlanIcon{"PhCursor"}
	AtlanIconCursorClick                 = AtlanIcon{"PhCursorClick"}
	AtlanIconCursorText                  = AtlanIcon{"PhCursorText"}
	AtlanIconCylinder                    = AtlanIcon{"PhCylinder"}
	AtlanIconDatabase                    = AtlanIcon{"PhDatabase"}
	AtlanIconDesktop                     = AtlanIcon{"PhDesktop"}
	AtlanIconDesktopTower                = AtlanIcon{"PhDesktopTower"}
	AtlanIconDetective                   = AtlanIcon{"PhDetective"}
	AtlanIconDevToLogo                   = AtlanIcon{"PhDevToLogo"}
	AtlanIconDeviceMobile                = AtlanIcon{"PhDeviceMobile"}
	AtlanIconDeviceMobileCamera          = AtlanIcon{"PhDeviceMobileCamera"}
	AtlanIconDeviceMobileSpeaker         = AtlanIcon{"PhDeviceMobileSpeaker"}
	AtlanIconDeviceTablet                = AtlanIcon{"PhDeviceTablet"}
	AtlanIconDeviceTabletCamera          = AtlanIcon{"PhDeviceTabletCamera"}
	AtlanIconDeviceTabletSpeaker         = AtlanIcon{"PhDeviceTabletSpeaker"}
	AtlanIconDevices                     = AtlanIcon{"PhDevices"}
	AtlanIconDiamond                     = AtlanIcon{"PhDiamond"}
	AtlanIconDiamondsFour                = AtlanIcon{"PhDiamondsFour"}
	AtlanIconDiceFive                    = AtlanIcon{"PhDiceFive"}
	AtlanIconDiceFour                    = AtlanIcon{"PhDiceFour"}
	AtlanIconDiceOne                     = AtlanIcon{"PhDiceOne"}
	AtlanIconDiceSix                     = AtlanIcon{"PhDiceSix"}
	AtlanIconDiceThree                   = AtlanIcon{"PhDiceThree"}
	AtlanIconDiceTwo                     = AtlanIcon{"PhDiceTwo"}
	AtlanIconDisc                        = AtlanIcon{"PhDisc"}
	AtlanIconDiscordLogo                 = AtlanIcon{"PhDiscordLogo"}
	AtlanIconDivide                      = AtlanIcon{"PhDivide"}
	AtlanIconDna                         = AtlanIcon{"PhDna"}
	AtlanIconDog                         = AtlanIcon{"PhDog"}
	AtlanIconDoor                        = AtlanIcon{"PhDoor"}
	AtlanIconDoorOpen                    = AtlanIcon{"PhDoorOpen"}
	AtlanIconDot                         = AtlanIcon{"PhDot"}
	AtlanIconDotOutline                  = AtlanIcon{"PhDotOutline"}
	AtlanIconDotsNine                    = AtlanIcon{"PhDotsNine"}
	AtlanIconDotsSix                     = AtlanIcon{"PhDotsSix"}
	AtlanIconDotsSixVertical             = AtlanIcon{"PhDotsSixVertical"}
	AtlanIconDotsThree                   = AtlanIcon{"PhDotsThree"}
	AtlanIconDotsThreeCircle             = AtlanIcon{"PhDotsThreeCircle"}
	AtlanIconDotsThreeCircleVertical     = AtlanIcon{"PhDotsThreeCircleVertical"}
	AtlanIconDotsThreeOutline            = AtlanIcon{"PhDotsThreeOutline"}
	AtlanIconDotsThreeOutlineVertical    = AtlanIcon{"PhDotsThreeOutlineVertical"}
	AtlanIconDotsThreeVertical           = AtlanIcon{"PhDotsThreeVertical"}
	AtlanIconDownload                    = AtlanIcon{"PhDownload"}
	AtlanIconDownloadSimple              = AtlanIcon{"PhDownloadSimple"}
	AtlanIconDress                       = AtlanIcon{"PhDress"}
	AtlanIconDribbbleLogo                = AtlanIcon{"PhDribbbleLogo"}
	AtlanIconDrop                        = AtlanIcon{"PhDrop"}
	AtlanIconDropHalf                    = AtlanIcon{"PhDropHalf"}
	AtlanIconDropHalfBottom              = AtlanIcon{"PhDropHalfBottom"}
	AtlanIconDropboxLogo                 = AtlanIcon{"PhDropboxLogo"}
	AtlanIconEar                         = AtlanIcon{"PhEar"}
	AtlanIconEarSlash                    = AtlanIcon{"PhEarSlash"}
	AtlanIconEgg                         = AtlanIcon{"PhEgg"}
	AtlanIconEggCrack                    = AtlanIcon{"PhEggCrack"}
	AtlanIconEject                       = AtlanIcon{"PhEject"}
	AtlanIconEjectSimple                 = AtlanIcon{"PhEjectSimple"}
	AtlanIconElevator                    = AtlanIcon{"PhElevator"}
	AtlanIconEngine                      = AtlanIcon{"PhEngine"}
	AtlanIconEnvelope                    = AtlanIcon{"PhEnvelope"}
	AtlanIconEnvelopeOpen                = AtlanIcon{"PhEnvelopeOpen"}
	AtlanIconEnvelopeSimple              = AtlanIcon{"PhEnvelopeSimple"}
	AtlanIconEnvelopeSimpleOpen          = AtlanIcon{"PhEnvelopeSimpleOpen"}
	AtlanIconEqualizer                   = AtlanIcon{"PhEqualizer"}
	AtlanIconEquals                      = AtlanIcon{"PhEquals"}
	AtlanIconEraser                      = AtlanIcon{"PhEraser"}
	AtlanIconEscalatorDown               = AtlanIcon{"PhEscalatorDown"}
	AtlanIconEscalatorUp                 = AtlanIcon{"PhEscalatorUp"}
	AtlanIconExam                        = AtlanIcon{"PhExam"}
	AtlanIconExclude                     = AtlanIcon{"PhExclude"}
	AtlanIconExcludeSquare               = AtlanIcon{"PhExcludeSquare"}
	AtlanIconExport                      = AtlanIcon{"PhExport"}
	AtlanIconEye                         = AtlanIcon{"PhEye"}
	AtlanIconEyeClosed                   = AtlanIcon{"PhEyeClosed"}
	AtlanIconEyeSlash                    = AtlanIcon{"PhEyeSlash"}
	AtlanIconEyedropper                  = AtlanIcon{"PhEyedropper"}
	AtlanIconEyedropperSample            = AtlanIcon{"PhEyedropperSample"}
	AtlanIconEyeglasses                  = AtlanIcon{"PhEyeglasses"}
	AtlanIconFaceMask                    = AtlanIcon{"PhFaceMask"}
	AtlanIconFacebookLogo                = AtlanIcon{"PhFacebookLogo"}
	AtlanIconFactory                     = AtlanIcon{"PhFactory"}
	AtlanIconFaders                      = AtlanIcon{"PhFaders"}
	AtlanIconFadersHorizontal            = AtlanIcon{"PhFadersHorizontal"}
	AtlanIconFan                         = AtlanIcon{"PhFan"}
	AtlanIconFastForward                 = AtlanIcon{"PhFastForward"}
	AtlanIconFastForwardCircle           = AtlanIcon{"PhFastForwardCircle"}
	AtlanIconFeather                     = AtlanIcon{"PhFeather"}
	AtlanIconFigmaLogo                   = AtlanIcon{"PhFigmaLogo"}
	AtlanIconFileArchive                 = AtlanIcon{"PhFileArchive"}
	AtlanIconFileArrowDown               = AtlanIcon{"PhFileArrowDown"}
	AtlanIconFileArrowUp                 = AtlanIcon{"PhFileArrowUp"}
	AtlanIconFileAudio                   = AtlanIcon{"PhFileAudio"}
	AtlanIconFile                        = AtlanIcon{"PhFile"}
	AtlanIconFileCloud                   = AtlanIcon{"PhFileCloud"}
	AtlanIconFileCode                    = AtlanIcon{"PhFileCode"}
	AtlanIconFileCss                     = AtlanIcon{"PhFileCss"}
	AtlanIconFileCsv                     = AtlanIcon{"PhFileCsv"}
	AtlanIconFileDashed                  = AtlanIcon{"PhFileDashed"}
	AtlanIconFileDoc                     = AtlanIcon{"PhFileDoc"}
	AtlanIconFileHtml                    = AtlanIcon{"PhFileHtml"}
	AtlanIconFileImage                   = AtlanIcon{"PhFileImage"}
	AtlanIconFileJpg                     = AtlanIcon{"PhFileJpg"}
	AtlanIconFileJs                      = AtlanIcon{"PhFileJs"}
	AtlanIconFileJsx                     = AtlanIcon{"PhFileJsx"}
	AtlanIconFileLock                    = AtlanIcon{"PhFileLock"}
	AtlanIconFileMagnifyingGlass         = AtlanIcon{"PhFileMagnifyingGlass"}
	AtlanIconFileMinus                   = AtlanIcon{"PhFileMinus"}
	AtlanIconFilePdf                     = AtlanIcon{"PhFilePdf"}
	AtlanIconFilePlus                    = AtlanIcon{"PhFilePlus"}
	AtlanIconFilePng                     = AtlanIcon{"PhFilePng"}
	AtlanIconFilePpt                     = AtlanIcon{"PhFilePpt"}
	AtlanIconFileRs                      = AtlanIcon{"PhFileRs"}
	AtlanIconFileSql                     = AtlanIcon{"PhFileSql"}
	AtlanIconFileSvg                     = AtlanIcon{"PhFileSvg"}
	AtlanIconFileText                    = AtlanIcon{"PhFileText"}
	AtlanIconFileTs                      = AtlanIcon{"PhFileTs"}
	AtlanIconFileTsx                     = AtlanIcon{"PhFileTsx"}
	AtlanIconFileVideo                   = AtlanIcon{"PhFileVideo"}
	AtlanIconFileVue                     = AtlanIcon{"PhFileVue"}
	AtlanIconFileX                       = AtlanIcon{"PhFileX"}
	AtlanIconFileXls                     = AtlanIcon{"PhFileXls"}
	AtlanIconFileZip                     = AtlanIcon{"PhFileZip"}
	AtlanIconFiles                       = AtlanIcon{"PhFiles"}
	AtlanIconFilmReel                    = AtlanIcon{"PhFilmReel"}
	AtlanIconFilmScript                  = AtlanIcon{"PhFilmScript"}
	AtlanIconFilmSlate                   = AtlanIcon{"PhFilmSlate"}
	AtlanIconFilmStrip                   = AtlanIcon{"PhFilmStrip"}
	AtlanIconFingerprint                 = AtlanIcon{"PhFingerprint"}
	AtlanIconFingerprintSimple           = AtlanIcon{"PhFingerprintSimple"}
	AtlanIconFinnTheHuman                = AtlanIcon{"PhFinnTheHuman"}
	AtlanIconFire                        = AtlanIcon{"PhFire"}
	AtlanIconFireExtinguisher            = AtlanIcon{"PhFireExtinguisher"}
	AtlanIconFireSimple                  = AtlanIcon{"PhFireSimple"}
	AtlanIconFirstAid                    = AtlanIcon{"PhFirstAid"}
	AtlanIconFirstAidKit                 = AtlanIcon{"PhFirstAidKit"}
	AtlanIconFish                        = AtlanIcon{"PhFish"}
	AtlanIconFishSimple                  = AtlanIcon{"PhFishSimple"}
	AtlanIconFlagBanner                  = AtlanIcon{"PhFlagBanner"}
	AtlanIconFlag                        = AtlanIcon{"PhFlag"}
	AtlanIconFlagCheckered               = AtlanIcon{"PhFlagCheckered"}
	AtlanIconFlagPennant                 = AtlanIcon{"PhFlagPennant"}
	AtlanIconFlame                       = AtlanIcon{"PhFlame"}
	AtlanIconFlashlight                  = AtlanIcon{"PhFlashlight"}
	AtlanIconFlask                       = AtlanIcon{"PhFlask"}
	AtlanIconFloppyDiskBack              = AtlanIcon{"PhFloppyDiskBack"}
	AtlanIconFloppyDisk                  = AtlanIcon{"PhFloppyDisk"}
	AtlanIconFlowArrow                   = AtlanIcon{"PhFlowArrow"}
	AtlanIconFlower                      = AtlanIcon{"PhFlower"}
	AtlanIconFlowerLotus                 = AtlanIcon{"PhFlowerLotus"}
	AtlanIconFlowerTulip                 = AtlanIcon{"PhFlowerTulip"}
	AtlanIconFlyingSaucer                = AtlanIcon{"PhFlyingSaucer"}
	AtlanIconFolder                      = AtlanIcon{"PhFolder"}
	AtlanIconFolderDashed                = AtlanIcon{"PhFolderDashed"}
	AtlanIconFolderLock                  = AtlanIcon{"PhFolderLock"}
	AtlanIconFolderMinus                 = AtlanIcon{"PhFolderMinus"}
	AtlanIconFolderNotch                 = AtlanIcon{"PhFolderNotch"}
	AtlanIconFolderNotchMinus            = AtlanIcon{"PhFolderNotchMinus"}
	AtlanIconFolderNotchOpen             = AtlanIcon{"PhFolderNotchOpen"}
	AtlanIconFolderNotchPlus             = AtlanIcon{"PhFolderNotchPlus"}
	AtlanIconFolderOpen                  = AtlanIcon{"PhFolderOpen"}
	AtlanIconFolderPlus                  = AtlanIcon{"PhFolderPlus"}
	AtlanIconFolderSimple                = AtlanIcon{"PhFolderSimple"}
	AtlanIconFolderSimpleDashed          = AtlanIcon{"PhFolderSimpleDashed"}
	AtlanIconFolderSimpleLock            = AtlanIcon{"PhFolderSimpleLock"}
	AtlanIconFolderSimpleMinus           = AtlanIcon{"PhFolderSimpleMinus"}
	AtlanIconFolderSimplePlus            = AtlanIcon{"PhFolderSimplePlus"}
	AtlanIconFolderSimpleStar            = AtlanIcon{"PhFolderSimpleStar"}
	AtlanIconFolderSimpleUser            = AtlanIcon{"PhFolderSimpleUser"}
	AtlanIconFolderStar                  = AtlanIcon{"PhFolderStar"}
	AtlanIconFolderUser                  = AtlanIcon{"PhFolderUser"}
	AtlanIconFolders                     = AtlanIcon{"PhFolders"}
	AtlanIconFootball                    = AtlanIcon{"PhFootball"}
	AtlanIconFootprints                  = AtlanIcon{"PhFootprints"}
	AtlanIconForkKnife                   = AtlanIcon{"PhForkKnife"}
	AtlanIconFrameCorners                = AtlanIcon{"PhFrameCorners"}
	AtlanIconFramerLogo                  = AtlanIcon{"PhFramerLogo"}
	AtlanIconFunction                    = AtlanIcon{"PhFunction"}
	AtlanIconFunnel                      = AtlanIcon{"PhFunnel"}
	AtlanIconFunnelSimple                = AtlanIcon{"PhFunnelSimple"}
	AtlanIconGameController              = AtlanIcon{"PhGameController"}
	AtlanIconGarage                      = AtlanIcon{"PhGarage"}
	AtlanIconGasCan                      = AtlanIcon{"PhGasCan"}
	AtlanIconGasPump                     = AtlanIcon{"PhGasPump"}
	AtlanIconGauge                       = AtlanIcon{"PhGauge"}
	AtlanIconGavel                       = AtlanIcon{"PhGavel"}
	AtlanIconGear                        = AtlanIcon{"PhGear"}
	AtlanIconGearFine                    = AtlanIcon{"PhGearFine"}
	AtlanIconGearSix                     = AtlanIcon{"PhGearSix"}
	AtlanIconGenderFemale                = AtlanIcon{"PhGenderFemale"}
	AtlanIconGenderIntersex              = AtlanIcon{"PhGenderIntersex"}
	AtlanIconGenderMale                  = AtlanIcon{"PhGenderMale"}
	AtlanIconGenderNeuter                = AtlanIcon{"PhGenderNeuter"}
	AtlanIconGenderNonbinary             = AtlanIcon{"PhGenderNonbinary"}
	AtlanIconGenderTransgender           = AtlanIcon{"PhGenderTransgender"}
	AtlanIconGhost                       = AtlanIcon{"PhGhost"}
	AtlanIconGif                         = AtlanIcon{"PhGif"}
	AtlanIconGift                        = AtlanIcon{"PhGift"}
	AtlanIconGitBranch                   = AtlanIcon{"PhGitBranch"}
	AtlanIconGitCommit                   = AtlanIcon{"PhGitCommit"}
	AtlanIconGitDiff                     = AtlanIcon{"PhGitDiff"}
	AtlanIconGitFork                     = AtlanIcon{"PhGitFork"}
	AtlanIconGitMerge                    = AtlanIcon{"PhGitMerge"}
	AtlanIconGitPullRequest              = AtlanIcon{"PhGitPullRequest"}
	AtlanIconGithubLogo                  = AtlanIcon{"PhGithubLogo"}
	AtlanIconGitlabLogo                  = AtlanIcon{"PhGitlabLogo"}
	AtlanIconGitlabLogoSimple            = AtlanIcon{"PhGitlabLogoSimple"}
	AtlanIconGlobe                       = AtlanIcon{"PhGlobe"}
	AtlanIconGlobeHemisphereEast         = AtlanIcon{"PhGlobeHemisphereEast"}
	AtlanIconGlobeHemisphereWest         = AtlanIcon{"PhGlobeHemisphereWest"}
	AtlanIconGlobeSimple                 = AtlanIcon{"PhGlobeSimple"}
	AtlanIconGlobeStand                  = AtlanIcon{"PhGlobeStand"}
	AtlanIconGoggles                     = AtlanIcon{"PhGoggles"}
	AtlanIconGoodreadsLogo               = AtlanIcon{"PhGoodreadsLogo"}
	AtlanIconGoogleCardboardLogo         = AtlanIcon{"PhGoogleCardboardLogo"}
	AtlanIconGoogleChromeLogo            = AtlanIcon{"PhGoogleChromeLogo"}
	AtlanIconGoogleDriveLogo             = AtlanIcon{"PhGoogleDriveLogo"}
	AtlanIconGoogleLogo                  = AtlanIcon{"PhGoogleLogo"}
	AtlanIconGooglePhotosLogo            = AtlanIcon{"PhGooglePhotosLogo"}
	AtlanIconGooglePlayLogo              = AtlanIcon{"PhGooglePlayLogo"}
	AtlanIconGooglePodcastsLogo          = AtlanIcon{"PhGooglePodcastsLogo"}
	AtlanIconGradient                    = AtlanIcon{"PhGradient"}
	AtlanIconGraduationCap               = AtlanIcon{"PhGraduationCap"}
	AtlanIconGrains                      = AtlanIcon{"PhGrains"}
	AtlanIconGrainsSlash                 = AtlanIcon{"PhGrainsSlash"}
	AtlanIconGraph                       = AtlanIcon{"PhGraph"}
	AtlanIconGridFour                    = AtlanIcon{"PhGridFour"}
	AtlanIconGridNine                    = AtlanIcon{"PhGridNine"}
	AtlanIconGuitar                      = AtlanIcon{"PhGuitar"}
	AtlanIconHamburger                   = AtlanIcon{"PhHamburger"}
	AtlanIconHammer                      = AtlanIcon{"PhHammer"}
	AtlanIconHand                        = AtlanIcon{"PhHand"}
	AtlanIconHandCoins                   = AtlanIcon{"PhHandCoins"}
	AtlanIconHandEye                     = AtlanIcon{"PhHandEye"}
	AtlanIconHandFist                    = AtlanIcon{"PhHandFist"}
	AtlanIconHandGrabbing                = AtlanIcon{"PhHandGrabbing"}
	AtlanIconHandHeart                   = AtlanIcon{"PhHandHeart"}
	AtlanIconHandPalm                    = AtlanIcon{"PhHandPalm"}
	AtlanIconHandPointing                = AtlanIcon{"PhHandPointing"}
	AtlanIconHandSoap                    = AtlanIcon{"PhHandSoap"}
	AtlanIconHandSwipeLeft               = AtlanIcon{"PhHandSwipeLeft"}
	AtlanIconHandSwipeRight              = AtlanIcon{"PhHandSwipeRight"}
	AtlanIconHandTap                     = AtlanIcon{"PhHandTap"}
	AtlanIconHandWaving                  = AtlanIcon{"PhHandWaving"}
	AtlanIconHandbag                     = AtlanIcon{"PhHandbag"}
	AtlanIconHandbagSimple               = AtlanIcon{"PhHandbagSimple"}
	AtlanIconHandsClapping               = AtlanIcon{"PhHandsClapping"}
	AtlanIconHandsPraying                = AtlanIcon{"PhHandsPraying"}
	AtlanIconHandshake                   = AtlanIcon{"PhHandshake"}
	AtlanIconHardDrive                   = AtlanIcon{"PhHardDrive"}
	AtlanIconHardDrives                  = AtlanIcon{"PhHardDrives"}
	AtlanIconHash                        = AtlanIcon{"PhHash"}
	AtlanIconHashStraight                = AtlanIcon{"PhHashStraight"}
	AtlanIconHeadlights                  = AtlanIcon{"PhHeadlights"}
	AtlanIconHeadphones                  = AtlanIcon{"PhHeadphones"}
	AtlanIconHeadset                     = AtlanIcon{"PhHeadset"}
	AtlanIconHeart                       = AtlanIcon{"PhHeart"}
	AtlanIconHeartBreak                  = AtlanIcon{"PhHeartBreak"}
	AtlanIconHeartHalf                   = AtlanIcon{"PhHeartHalf"}
	AtlanIconHeartStraight               = AtlanIcon{"PhHeartStraight"}
	AtlanIconHeartStraightBreak          = AtlanIcon{"PhHeartStraightBreak"}
	AtlanIconHeartbeat                   = AtlanIcon{"PhHeartbeat"}
	AtlanIconHexagon                     = AtlanIcon{"PhHexagon"}
	AtlanIconHighHeel                    = AtlanIcon{"PhHighHeel"}
	AtlanIconHighlighterCircle           = AtlanIcon{"PhHighlighterCircle"}
	AtlanIconHoodie                      = AtlanIcon{"PhHoodie"}
	AtlanIconHorse                       = AtlanIcon{"PhHorse"}
	AtlanIconHourglass                   = AtlanIcon{"PhHourglass"}
	AtlanIconHourglassHigh               = AtlanIcon{"PhHourglassHigh"}
	AtlanIconHourglassLow                = AtlanIcon{"PhHourglassLow"}
	AtlanIconHourglassMedium             = AtlanIcon{"PhHourglassMedium"}
	AtlanIconHourglassSimple             = AtlanIcon{"PhHourglassSimple"}
	AtlanIconHourglassSimpleHigh         = AtlanIcon{"PhHourglassSimpleHigh"}
	AtlanIconHourglassSimpleLow          = AtlanIcon{"PhHourglassSimpleLow"}
	AtlanIconHourglassSimpleMedium       = AtlanIcon{"PhHourglassSimpleMedium"}
	AtlanIconHouse                       = AtlanIcon{"PhHouse"}
	AtlanIconHouseLine                   = AtlanIcon{"PhHouseLine"}
	AtlanIconHouseSimple                 = AtlanIcon{"PhHouseSimple"}
	AtlanIconIceCream                    = AtlanIcon{"PhIceCream"}
	AtlanIconIdentificationBadge         = AtlanIcon{"PhIdentificationBadge"}
	AtlanIconIdentificationCard          = AtlanIcon{"PhIdentificationCard"}
	AtlanIconImage                       = AtlanIcon{"PhImage"}
	AtlanIconImageSquare                 = AtlanIcon{"PhImageSquare"}
	AtlanIconImages                      = AtlanIcon{"PhImages"}
	AtlanIconImagesSquare                = AtlanIcon{"PhImagesSquare"}
	AtlanIconInfinity                    = AtlanIcon{"PhInfinity"}
	AtlanIconInfo                        = AtlanIcon{"PhInfo"}
	AtlanIconInstagramLogo               = AtlanIcon{"PhInstagramLogo"}
	AtlanIconIntersect                   = AtlanIcon{"PhIntersect"}
	AtlanIconIntersectSquare             = AtlanIcon{"PhIntersectSquare"}
	AtlanIconIntersectThree              = AtlanIcon{"PhIntersectThree"}
	AtlanIconJeep                        = AtlanIcon{"PhJeep"}
	AtlanIconKanban                      = AtlanIcon{"PhKanban"}
	AtlanIconKey                         = AtlanIcon{"PhKey"}
	AtlanIconKeyReturn                   = AtlanIcon{"PhKeyReturn"}
	AtlanIconKeyboard                    = AtlanIcon{"PhKeyboard"}
	AtlanIconKeyhole                     = AtlanIcon{"PhKeyhole"}
	AtlanIconKnife                       = AtlanIcon{"PhKnife"}
	AtlanIconLadder                      = AtlanIcon{"PhLadder"}
	AtlanIconLadderSimple                = AtlanIcon{"PhLadderSimple"}
	AtlanIconLamp                        = AtlanIcon{"PhLamp"}
	AtlanIconLaptop                      = AtlanIcon{"PhLaptop"}
	AtlanIconLayout                      = AtlanIcon{"PhLayout"}
	AtlanIconLeaf                        = AtlanIcon{"PhLeaf"}
	AtlanIconLifebuoy                    = AtlanIcon{"PhLifebuoy"}
	AtlanIconLightbulb                   = AtlanIcon{"PhLightbulb"}
	AtlanIconLightbulbFilament           = AtlanIcon{"PhLightbulbFilament"}
	AtlanIconLighthouse                  = AtlanIcon{"PhLighthouse"}
	AtlanIconLightningA                  = AtlanIcon{"PhLightningA"}
	AtlanIconLightning                   = AtlanIcon{"PhLightning"}
	AtlanIconLightningSlash              = AtlanIcon{"PhLightningSlash"}
	AtlanIconLineSegment                 = AtlanIcon{"PhLineSegment"}
	AtlanIconLineSegments                = AtlanIcon{"PhLineSegments"}
	AtlanIconLink                        = AtlanIcon{"PhLink"}
	AtlanIconLinkBreak                   = AtlanIcon{"PhLinkBreak"}
	AtlanIconLinkSimple                  = AtlanIcon{"PhLinkSimple"}
	AtlanIconLinkSimpleBreak             = AtlanIcon{"PhLinkSimpleBreak"}
	AtlanIconLinkSimpleHorizontal        = AtlanIcon{"PhLinkSimpleHorizontal"}
	AtlanIconLinkSimpleHorizontalBreak   = AtlanIcon{"PhLinkSimpleHorizontalBreak"}
	AtlanIconLinkedinLogo                = AtlanIcon{"PhLinkedinLogo"}
	AtlanIconLinuxLogo                   = AtlanIcon{"PhLinuxLogo"}
	AtlanIconList                        = AtlanIcon{"PhList"}
	AtlanIconListBullets                 = AtlanIcon{"PhListBullets"}
	AtlanIconListChecks                  = AtlanIcon{"PhListChecks"}
	AtlanIconListDashes                  = AtlanIcon{"PhListDashes"}
	AtlanIconListMagnifyingGlass         = AtlanIcon{"PhListMagnifyingGlass"}
	AtlanIconListNumbers                 = AtlanIcon{"PhListNumbers"}
	AtlanIconListPlus                    = AtlanIcon{"PhListPlus"}
	AtlanIconLock                        = AtlanIcon{"PhLock"}
	AtlanIconLockKey                     = AtlanIcon{"PhLockKey"}
	AtlanIconLockKeyOpen                 = AtlanIcon{"PhLockKeyOpen"}
	AtlanIconLockLaminated               = AtlanIcon{"PhLockLaminated"}
	AtlanIconLockLaminatedOpen           = AtlanIcon{"PhLockLaminatedOpen"}
	AtlanIconLockOpen                    = AtlanIcon{"PhLockOpen"}
	AtlanIconLockSimple                  = AtlanIcon{"PhLockSimple"}
	AtlanIconLockSimpleOpen              = AtlanIcon{"PhLockSimpleOpen"}
	AtlanIconLockers                     = AtlanIcon{"PhLockers"}
	AtlanIconMagicWand                   = AtlanIcon{"PhMagicWand"}
	AtlanIconMagnet                      = AtlanIcon{"PhMagnet"}
	AtlanIconMagnetStraight              = AtlanIcon{"PhMagnetStraight"}
	AtlanIconMagnifyingGlass             = AtlanIcon{"PhMagnifyingGlass"}
	AtlanIconMagnifyingGlassMinus        = AtlanIcon{"PhMagnifyingGlassMinus"}
	AtlanIconMagnifyingGlassPlus         = AtlanIcon{"PhMagnifyingGlassPlus"}
	AtlanIconMapPin                      = AtlanIcon{"PhMapPin"}
	AtlanIconMapPinLine                  = AtlanIcon{"PhMapPinLine"}
	AtlanIconMapTrifold                  = AtlanIcon{"PhMapTrifold"}
	AtlanIconMarkerCircle                = AtlanIcon{"PhMarkerCircle"}
	AtlanIconMartini                     = AtlanIcon{"PhMartini"}
	AtlanIconMaskHappy                   = AtlanIcon{"PhMaskHappy"}
	AtlanIconMaskSad                     = AtlanIcon{"PhMaskSad"}
	AtlanIconMathOperations              = AtlanIcon{"PhMathOperations"}
	AtlanIconMedal                       = AtlanIcon{"PhMedal"}
	AtlanIconMedalMilitary               = AtlanIcon{"PhMedalMilitary"}
	AtlanIconMediumLogo                  = AtlanIcon{"PhMediumLogo"}
	AtlanIconMegaphone                   = AtlanIcon{"PhMegaphone"}
	AtlanIconMegaphoneSimple             = AtlanIcon{"PhMegaphoneSimple"}
	AtlanIconMessengerLogo               = AtlanIcon{"PhMessengerLogo"}
	AtlanIconMetaLogo                    = AtlanIcon{"PhMetaLogo"}
	AtlanIconMetronome                   = AtlanIcon{"PhMetronome"}
	AtlanIconMicrophone                  = AtlanIcon{"PhMicrophone"}
	AtlanIconMicrophoneSlash             = AtlanIcon{"PhMicrophoneSlash"}
	AtlanIconMicrophoneStage             = AtlanIcon{"PhMicrophoneStage"}
	AtlanIconMicrosoftExcelLogo          = AtlanIcon{"PhMicrosoftExcelLogo"}
	AtlanIconMicrosoftOutlookLogo        = AtlanIcon{"PhMicrosoftOutlookLogo"}
	AtlanIconMicrosoftPowerpointLogo     = AtlanIcon{"PhMicrosoftPowerpointLogo"}
	AtlanIconMicrosoftTeamsLogo          = AtlanIcon{"PhMicrosoftTeamsLogo"}
	AtlanIconMicrosoftWordLogo           = AtlanIcon{"PhMicrosoftWordLogo"}
	AtlanIconMinus                       = AtlanIcon{"PhMinus"}
	AtlanIconMinusCircle                 = AtlanIcon{"PhMinusCircle"}
	AtlanIconMinusSquare                 = AtlanIcon{"PhMinusSquare"}
	AtlanIconMoney                       = AtlanIcon{"PhMoney"}
	AtlanIconMonitor                     = AtlanIcon{"PhMonitor"}
	AtlanIconMonitorPlay                 = AtlanIcon{"PhMonitorPlay"}
	AtlanIconMoon                        = AtlanIcon{"PhMoon"}
	AtlanIconMoonStars                   = AtlanIcon{"PhMoonStars"}
	AtlanIconMoped                       = AtlanIcon{"PhMoped"}
	AtlanIconMopedFront                  = AtlanIcon{"PhMopedFront"}
	AtlanIconMosque                      = AtlanIcon{"PhMosque"}
	AtlanIconMotorcycle                  = AtlanIcon{"PhMotorcycle"}
	AtlanIconMountains                   = AtlanIcon{"PhMountains"}
	AtlanIconMouse                       = AtlanIcon{"PhMouse"}
	AtlanIconMouseSimple                 = AtlanIcon{"PhMouseSimple"}
	AtlanIconMusicNote                   = AtlanIcon{"PhMusicNote"}
	AtlanIconMusicNoteSimple             = AtlanIcon{"PhMusicNoteSimple"}
	AtlanIconMusicNotes                  = AtlanIcon{"PhMusicNotes"}
	AtlanIconMusicNotesPlus              = AtlanIcon{"PhMusicNotesPlus"}
	AtlanIconMusicNotesSimple            = AtlanIcon{"PhMusicNotesSimple"}
	AtlanIconNavigationArrow             = AtlanIcon{"PhNavigationArrow"}
	AtlanIconNeedle                      = AtlanIcon{"PhNeedle"}
	AtlanIconNewspaper                   = AtlanIcon{"PhNewspaper"}
	AtlanIconNewspaperClipping           = AtlanIcon{"PhNewspaperClipping"}
	AtlanIconNotches                     = AtlanIcon{"PhNotches"}
	AtlanIconNoteBlank                   = AtlanIcon{"PhNoteBlank"}
	AtlanIconNote                        = AtlanIcon{"PhNote"}
	AtlanIconNotePencil                  = AtlanIcon{"PhNotePencil"}
	AtlanIconNotebook                    = AtlanIcon{"PhNotebook"}
	AtlanIconNotepad                     = AtlanIcon{"PhNotepad"}
	AtlanIconNotification                = AtlanIcon{"PhNotification"}
	AtlanIconNotionLogo                  = AtlanIcon{"PhNotionLogo"}
	AtlanIconNumberCircleEight           = AtlanIcon{"PhNumberCircleEight"}
	AtlanIconNumberCircleFive            = AtlanIcon{"PhNumberCircleFive"}
	AtlanIconNumberCircleFour            = AtlanIcon{"PhNumberCircleFour"}
	AtlanIconNumberCircleNine            = AtlanIcon{"PhNumberCircleNine"}
	AtlanIconNumberCircleOne             = AtlanIcon{"PhNumberCircleOne"}
	AtlanIconNumberCircleSeven           = AtlanIcon{"PhNumberCircleSeven"}
	AtlanIconNumberCircleSix             = AtlanIcon{"PhNumberCircleSix"}
	AtlanIconNumberCircleThree           = AtlanIcon{"PhNumberCircleThree"}
	AtlanIconNumberCircleTwo             = AtlanIcon{"PhNumberCircleTwo"}
	AtlanIconNumberCircleZero            = AtlanIcon{"PhNumberCircleZero"}
	AtlanIconNumberEight                 = AtlanIcon{"PhNumberEight"}
	AtlanIconNumberFive                  = AtlanIcon{"PhNumberFive"}
	AtlanIconNumberFour                  = AtlanIcon{"PhNumberFour"}
	AtlanIconNumberNine                  = AtlanIcon{"PhNumberNine"}
	AtlanIconNumberOne                   = AtlanIcon{"PhNumberOne"}
	AtlanIconNumberSeven                 = AtlanIcon{"PhNumberSeven"}
	AtlanIconNumberSix                   = AtlanIcon{"PhNumberSix"}
	AtlanIconNumberSquareEight           = AtlanIcon{"PhNumberSquareEight"}
	AtlanIconNumberSquareFive            = AtlanIcon{"PhNumberSquareFive"}
	AtlanIconNumberSquareFour            = AtlanIcon{"PhNumberSquareFour"}
	AtlanIconNumberSquareNine            = AtlanIcon{"PhNumberSquareNine"}
	AtlanIconNumberSquareOne             = AtlanIcon{"PhNumberSquareOne"}
	AtlanIconNumberSquareSeven           = AtlanIcon{"PhNumberSquareSeven"}
	AtlanIconNumberSquareSix             = AtlanIcon{"PhNumberSquareSix"}
	AtlanIconNumberSquareThree           = AtlanIcon{"PhNumberSquareThree"}
	AtlanIconNumberSquareTwo             = AtlanIcon{"PhNumberSquareTwo"}
	AtlanIconNumberSquareZero            = AtlanIcon{"PhNumberSquareZero"}
	AtlanIconNumberThree                 = AtlanIcon{"PhNumberThree"}
	AtlanIconNumberTwo                   = AtlanIcon{"PhNumberTwo"}
	AtlanIconNumberZero                  = AtlanIcon{"PhNumberZero"}
	AtlanIconNut                         = AtlanIcon{"PhNut"}
	AtlanIconNyTimesLogo                 = AtlanIcon{"PhNyTimesLogo"}
	AtlanIconOctagon                     = AtlanIcon{"PhOctagon"}
	AtlanIconOfficeChair                 = AtlanIcon{"PhOfficeChair"}
	AtlanIconOption                      = AtlanIcon{"PhOption"}
	AtlanIconOrangeSlice                 = AtlanIcon{"PhOrangeSlice"}
	AtlanIconPackage                     = AtlanIcon{"PhPackage"}
	AtlanIconPaintBrush                  = AtlanIcon{"PhPaintBrush"}
	AtlanIconPaintBrushBroad             = AtlanIcon{"PhPaintBrushBroad"}
	AtlanIconPaintBrushHousehold         = AtlanIcon{"PhPaintBrushHousehold"}
	AtlanIconPaintBucket                 = AtlanIcon{"PhPaintBucket"}
	AtlanIconPaintRoller                 = AtlanIcon{"PhPaintRoller"}
	AtlanIconPalette                     = AtlanIcon{"PhPalette"}
	AtlanIconPants                       = AtlanIcon{"PhPants"}
	AtlanIconPaperPlane                  = AtlanIcon{"PhPaperPlane"}
	AtlanIconPaperPlaneRight             = AtlanIcon{"PhPaperPlaneRight"}
	AtlanIconPaperPlaneTilt              = AtlanIcon{"PhPaperPlaneTilt"}
	AtlanIconPaperclip                   = AtlanIcon{"PhPaperclip"}
	AtlanIconPaperclipHorizontal         = AtlanIcon{"PhPaperclipHorizontal"}
	AtlanIconParachute                   = AtlanIcon{"PhParachute"}
	AtlanIconParagraph                   = AtlanIcon{"PhParagraph"}
	AtlanIconParallelogram               = AtlanIcon{"PhParallelogram"}
	AtlanIconPark                        = AtlanIcon{"PhPark"}
	AtlanIconPassword                    = AtlanIcon{"PhPassword"}
	AtlanIconPath                        = AtlanIcon{"PhPath"}
	AtlanIconPatreonLogo                 = AtlanIcon{"PhPatreonLogo"}
	AtlanIconPause                       = AtlanIcon{"PhPause"}
	AtlanIconPauseCircle                 = AtlanIcon{"PhPauseCircle"}
	AtlanIconPawPrint                    = AtlanIcon{"PhPawPrint"}
	AtlanIconPaypalLogo                  = AtlanIcon{"PhPaypalLogo"}
	AtlanIconPeace                       = AtlanIcon{"PhPeace"}
	AtlanIconPen                         = AtlanIcon{"PhPen"}
	AtlanIconPenNib                      = AtlanIcon{"PhPenNib"}
	AtlanIconPenNibStraight              = AtlanIcon{"PhPenNibStraight"}
	AtlanIconPencil                      = AtlanIcon{"PhPencil"}
	AtlanIconPencilCircle                = AtlanIcon{"PhPencilCircle"}
	AtlanIconPencilLine                  = AtlanIcon{"PhPencilLine"}
	AtlanIconPencilSimple                = AtlanIcon{"PhPencilSimple"}
	AtlanIconPencilSimpleLine            = AtlanIcon{"PhPencilSimpleLine"}
	AtlanIconPencilSimpleSlash           = AtlanIcon{"PhPencilSimpleSlash"}
	AtlanIconPencilSlash                 = AtlanIcon{"PhPencilSlash"}
	AtlanIconPentagram                   = AtlanIcon{"PhPentagram"}
	AtlanIconPepper                      = AtlanIcon{"PhPepper"}
	AtlanIconPercent                     = AtlanIcon{"PhPercent"}
	AtlanIconPersonArmsSpread            = AtlanIcon{"PhPersonArmsSpread"}
	AtlanIconPerson                      = AtlanIcon{"PhPerson"}
	AtlanIconPersonSimpleBike            = AtlanIcon{"PhPersonSimpleBike"}
	AtlanIconPersonSimple                = AtlanIcon{"PhPersonSimple"}
	AtlanIconPersonSimpleRun             = AtlanIcon{"PhPersonSimpleRun"}
	AtlanIconPersonSimpleThrow           = AtlanIcon{"PhPersonSimpleThrow"}
	AtlanIconPersonSimpleWalk            = AtlanIcon{"PhPersonSimpleWalk"}
	AtlanIconPerspective                 = AtlanIcon{"PhPerspective"}
	AtlanIconPhone                       = AtlanIcon{"PhPhone"}
	AtlanIconPhoneCall                   = AtlanIcon{"PhPhoneCall"}
	AtlanIconPhoneDisconnect             = AtlanIcon{"PhPhoneDisconnect"}
	AtlanIconPhoneIncoming               = AtlanIcon{"PhPhoneIncoming"}
	AtlanIconPhoneOutgoing               = AtlanIcon{"PhPhoneOutgoing"}
	AtlanIconPhonePlus                   = AtlanIcon{"PhPhonePlus"}
	AtlanIconPhoneSlash                  = AtlanIcon{"PhPhoneSlash"}
	AtlanIconPhoneX                      = AtlanIcon{"PhPhoneX"}
	AtlanIconPhosphorLogo                = AtlanIcon{"PhPhosphorLogo"}
	AtlanIconPi                          = AtlanIcon{"PhPi"}
	AtlanIconPianoKeys                   = AtlanIcon{"PhPianoKeys"}
	AtlanIconPictureInPicture            = AtlanIcon{"PhPictureInPicture"}
	AtlanIconPiggyBank                   = AtlanIcon{"PhPiggyBank"}
	AtlanIconPill                        = AtlanIcon{"PhPill"}
	AtlanIconPinterestLogo               = AtlanIcon{"PhPinterestLogo"}
	AtlanIconPinwheel                    = AtlanIcon{"PhPinwheel"}
	AtlanIconPizza                       = AtlanIcon{"PhPizza"}
	AtlanIconPlaceholder                 = AtlanIcon{"PhPlaceholder"}
	AtlanIconPlanet                      = AtlanIcon{"PhPlanet"}
	AtlanIconPlant                       = AtlanIcon{"PhPlant"}
	AtlanIconPlay                        = AtlanIcon{"PhPlay"}
	AtlanIconPlayCircle                  = AtlanIcon{"PhPlayCircle"}
	AtlanIconPlayPause                   = AtlanIcon{"PhPlayPause"}
	AtlanIconPlaylist                    = AtlanIcon{"PhPlaylist"}
	AtlanIconPlug                        = AtlanIcon{"PhPlug"}
	AtlanIconPlugCharging                = AtlanIcon{"PhPlugCharging"}
	AtlanIconPlugs                       = AtlanIcon{"PhPlugs"}
	AtlanIconPlugsConnected              = AtlanIcon{"PhPlugsConnected"}
	AtlanIconPlus                        = AtlanIcon{"PhPlus"}
	AtlanIconPlusCircle                  = AtlanIcon{"PhPlusCircle"}
	AtlanIconPlusMinus                   = AtlanIcon{"PhPlusMinus"}
	AtlanIconPlusSquare                  = AtlanIcon{"PhPlusSquare"}
	AtlanIconPokerChip                   = AtlanIcon{"PhPokerChip"}
	AtlanIconPoliceCar                   = AtlanIcon{"PhPoliceCar"}
	AtlanIconPolygon                     = AtlanIcon{"PhPolygon"}
	AtlanIconPopcorn                     = AtlanIcon{"PhPopcorn"}
	AtlanIconPottedPlant                 = AtlanIcon{"PhPottedPlant"}
	AtlanIconPower                       = AtlanIcon{"PhPower"}
	AtlanIconPrescription                = AtlanIcon{"PhPrescription"}
	AtlanIconPresentation                = AtlanIcon{"PhPresentation"}
	AtlanIconPresentationChart           = AtlanIcon{"PhPresentationChart"}
	AtlanIconPrinter                     = AtlanIcon{"PhPrinter"}
	AtlanIconProhibit                    = AtlanIcon{"PhProhibit"}
	AtlanIconProhibitInset               = AtlanIcon{"PhProhibitInset"}
	AtlanIconProjectorScreen             = AtlanIcon{"PhProjectorScreen"}
	AtlanIconProjectorScreenChart        = AtlanIcon{"PhProjectorScreenChart"}
	AtlanIconPulse                       = AtlanIcon{"PhPulse"}
	AtlanIconPushPin                     = AtlanIcon{"PhPushPin"}
	AtlanIconPushPinSimple               = AtlanIcon{"PhPushPinSimple"}
	AtlanIconPushPinSimpleSlash          = AtlanIcon{"PhPushPinSimpleSlash"}
	AtlanIconPushPinSlash                = AtlanIcon{"PhPushPinSlash"}
	AtlanIconPuzzlePiece                 = AtlanIcon{"PhPuzzlePiece"}
	AtlanIconQrCode                      = AtlanIcon{"PhQrCode"}
	AtlanIconQuestion                    = AtlanIcon{"PhQuestion"}
	AtlanIconQueue                       = AtlanIcon{"PhQueue"}
	AtlanIconQuotes                      = AtlanIcon{"PhQuotes"}
	AtlanIconRadical                     = AtlanIcon{"PhRadical"}
	AtlanIconRadio                       = AtlanIcon{"PhRadio"}
	AtlanIconRadioButton                 = AtlanIcon{"PhRadioButton"}
	AtlanIconRadioactive                 = AtlanIcon{"PhRadioactive"}
	AtlanIconRainbow                     = AtlanIcon{"PhRainbow"}
	AtlanIconRainbowCloud                = AtlanIcon{"PhRainbowCloud"}
	AtlanIconReadCvLogo                  = AtlanIcon{"PhReadCvLogo"}
	AtlanIconReceipt                     = AtlanIcon{"PhReceipt"}
	AtlanIconReceiptX                    = AtlanIcon{"PhReceiptX"}
	AtlanIconRecord                      = AtlanIcon{"PhRecord"}
	AtlanIconRectangle                   = AtlanIcon{"PhRectangle"}
	AtlanIconRecycle                     = AtlanIcon{"PhRecycle"}
	AtlanIconRedditLogo                  = AtlanIcon{"PhRedditLogo"}
	AtlanIconRepeat                      = AtlanIcon{"PhRepeat"}
	AtlanIconRepeatOnce                  = AtlanIcon{"PhRepeatOnce"}
	AtlanIconRewind                      = AtlanIcon{"PhRewind"}
	AtlanIconRewindCircle                = AtlanIcon{"PhRewindCircle"}
	AtlanIconRoadHorizon                 = AtlanIcon{"PhRoadHorizon"}
	AtlanIconRobot                       = AtlanIcon{"PhRobot"}
	AtlanIconRocket                      = AtlanIcon{"PhRocket"}
	AtlanIconRocketLaunch                = AtlanIcon{"PhRocketLaunch"}
	AtlanIconRows                        = AtlanIcon{"PhRows"}
	AtlanIconRss                         = AtlanIcon{"PhRss"}
	AtlanIconRssSimple                   = AtlanIcon{"PhRssSimple"}
	AtlanIconRug                         = AtlanIcon{"PhRug"}
	AtlanIconRuler                       = AtlanIcon{"PhRuler"}
	AtlanIconScales                      = AtlanIcon{"PhScales"}
	AtlanIconScan                        = AtlanIcon{"PhScan"}
	AtlanIconScissors                    = AtlanIcon{"PhScissors"}
	AtlanIconScooter                     = AtlanIcon{"PhScooter"}
	AtlanIconScreencast                  = AtlanIcon{"PhScreencast"}
	AtlanIconScribbleLoop                = AtlanIcon{"PhScribbleLoop"}
	AtlanIconScroll                      = AtlanIcon{"PhScroll"}
	AtlanIconSeal                        = AtlanIcon{"PhSeal"}
	AtlanIconSealCheck                   = AtlanIcon{"PhSealCheck"}
	AtlanIconSealQuestion                = AtlanIcon{"PhSealQuestion"}
	AtlanIconSealWarning                 = AtlanIcon{"PhSealWarning"}
	AtlanIconSelectionAll                = AtlanIcon{"PhSelectionAll"}
	AtlanIconSelectionBackground         = AtlanIcon{"PhSelectionBackground"}
	AtlanIconSelection                   = AtlanIcon{"PhSelection"}
	AtlanIconSelectionForeground         = AtlanIcon{"PhSelectionForeground"}
	AtlanIconSelectionInverse            = AtlanIcon{"PhSelectionInverse"}
	AtlanIconSelectionPlus               = AtlanIcon{"PhSelectionPlus"}
	AtlanIconSelectionSlash              = AtlanIcon{"PhSelectionSlash"}
	AtlanIconShapes                      = AtlanIcon{"PhShapes"}
	AtlanIconShare                       = AtlanIcon{"PhShare"}
	AtlanIconShareFat                    = AtlanIcon{"PhShareFat"}
	AtlanIconShareNetwork                = AtlanIcon{"PhShareNetwork"}
	AtlanIconShield                      = AtlanIcon{"PhShield"}
	AtlanIconShieldCheck                 = AtlanIcon{"PhShieldCheck"}
	AtlanIconShieldCheckered             = AtlanIcon{"PhShieldCheckered"}
	AtlanIconShieldChevron               = AtlanIcon{"PhShieldChevron"}
	AtlanIconShieldPlus                  = AtlanIcon{"PhShieldPlus"}
	AtlanIconShieldSlash                 = AtlanIcon{"PhShieldSlash"}
	AtlanIconShieldStar                  = AtlanIcon{"PhShieldStar"}
	AtlanIconShieldWarning               = AtlanIcon{"PhShieldWarning"}
	AtlanIconShirtFolded                 = AtlanIcon{"PhShirtFolded"}
	AtlanIconShootingStar                = AtlanIcon{"PhShootingStar"}
	AtlanIconShoppingBag                 = AtlanIcon{"PhShoppingBag"}
	AtlanIconShoppingBagOpen             = AtlanIcon{"PhShoppingBagOpen"}
	AtlanIconShoppingCart                = AtlanIcon{"PhShoppingCart"}
	AtlanIconShoppingCartSimple          = AtlanIcon{"PhShoppingCartSimple"}
	AtlanIconShower                      = AtlanIcon{"PhShower"}
	AtlanIconShrimp                      = AtlanIcon{"PhShrimp"}
	AtlanIconShuffleAngular              = AtlanIcon{"PhShuffleAngular"}
	AtlanIconShuffle                     = AtlanIcon{"PhShuffle"}
	AtlanIconShuffleSimple               = AtlanIcon{"PhShuffleSimple"}
	AtlanIconSidebar                     = AtlanIcon{"PhSidebar"}
	AtlanIconSidebarSimple               = AtlanIcon{"PhSidebarSimple"}
	AtlanIconSigma                       = AtlanIcon{"PhSigma"}
	AtlanIconSignIn                      = AtlanIcon{"PhSignIn"}
	AtlanIconSignOut                     = AtlanIcon{"PhSignOut"}
	AtlanIconSignature                   = AtlanIcon{"PhSignature"}
	AtlanIconSignpost                    = AtlanIcon{"PhSignpost"}
	AtlanIconSimCard                     = AtlanIcon{"PhSimCard"}
	AtlanIconSiren                       = AtlanIcon{"PhSiren"}
	AtlanIconSketchLogo                  = AtlanIcon{"PhSketchLogo"}
	AtlanIconSkipBack                    = AtlanIcon{"PhSkipBack"}
	AtlanIconSkipBackCircle              = AtlanIcon{"PhSkipBackCircle"}
	AtlanIconSkipForward                 = AtlanIcon{"PhSkipForward"}
	AtlanIconSkipForwardCircle           = AtlanIcon{"PhSkipForwardCircle"}
	AtlanIconSkull                       = AtlanIcon{"PhSkull"}
	AtlanIconSlackLogo                   = AtlanIcon{"PhSlackLogo"}
	AtlanIconSliders                     = AtlanIcon{"PhSliders"}
	AtlanIconSlidersHorizontal           = AtlanIcon{"PhSlidersHorizontal"}
	AtlanIconSlideshow                   = AtlanIcon{"PhSlideshow"}
	AtlanIconSmileyAngry                 = AtlanIcon{"PhSmileyAngry"}
	AtlanIconSmileyBlank                 = AtlanIcon{"PhSmileyBlank"}
	AtlanIconSmiley                      = AtlanIcon{"PhSmiley"}
	AtlanIconSmileyMeh                   = AtlanIcon{"PhSmileyMeh"}
	AtlanIconSmileyNervous               = AtlanIcon{"PhSmileyNervous"}
	AtlanIconSmileySad                   = AtlanIcon{"PhSmileySad"}
	AtlanIconSmileySticker               = AtlanIcon{"PhSmileySticker"}
	AtlanIconSmileyWink                  = AtlanIcon{"PhSmileyWink"}
	AtlanIconSmileyXEyes                 = AtlanIcon{"PhSmileyXEyes"}
	AtlanIconSnapchatLogo                = AtlanIcon{"PhSnapchatLogo"}
	AtlanIconSneaker                     = AtlanIcon{"PhSneaker"}
	AtlanIconSneakerMove                 = AtlanIcon{"PhSneakerMove"}
	AtlanIconSnowflake                   = AtlanIcon{"PhSnowflake"}
	AtlanIconSoccerBall                  = AtlanIcon{"PhSoccerBall"}
	AtlanIconSortAscending               = AtlanIcon{"PhSortAscending"}
	AtlanIconSortDescending              = AtlanIcon{"PhSortDescending"}
	AtlanIconSoundcloudLogo              = AtlanIcon{"PhSoundcloudLogo"}
	AtlanIconSpade                       = AtlanIcon{"PhSpade"}
	AtlanIconSparkle                     = AtlanIcon{"PhSparkle"}
	AtlanIconSpeakerHifi                 = AtlanIcon{"PhSpeakerHifi"}
	AtlanIconSpeakerHigh                 = AtlanIcon{"PhSpeakerHigh"}
	AtlanIconSpeakerLow                  = AtlanIcon{"PhSpeakerLow"}
	AtlanIconSpeakerNone                 = AtlanIcon{"PhSpeakerNone"}
	AtlanIconSpeakerSimpleHigh           = AtlanIcon{"PhSpeakerSimpleHigh"}
	AtlanIconSpeakerSimpleLow            = AtlanIcon{"PhSpeakerSimpleLow"}
	AtlanIconSpeakerSimpleNone           = AtlanIcon{"PhSpeakerSimpleNone"}
	AtlanIconSpeakerSimpleSlash          = AtlanIcon{"PhSpeakerSimpleSlash"}
	AtlanIconSpeakerSimpleX              = AtlanIcon{"PhSpeakerSimpleX"}
	AtlanIconSpeakerSlash                = AtlanIcon{"PhSpeakerSlash"}
	AtlanIconSpeakerX                    = AtlanIcon{"PhSpeakerX"}
	AtlanIconSpinner                     = AtlanIcon{"PhSpinner"}
	AtlanIconSpinnerGap                  = AtlanIcon{"PhSpinnerGap"}
	AtlanIconSpiral                      = AtlanIcon{"PhSpiral"}
	AtlanIconSplitHorizontal             = AtlanIcon{"PhSplitHorizontal"}
	AtlanIconSplitVertical               = AtlanIcon{"PhSplitVertical"}
	AtlanIconSpotifyLogo                 = AtlanIcon{"PhSpotifyLogo"}
	AtlanIconSquare                      = AtlanIcon{"PhSquare"}
	AtlanIconSquareHalf                  = AtlanIcon{"PhSquareHalf"}
	AtlanIconSquareHalfBottom            = AtlanIcon{"PhSquareHalfBottom"}
	AtlanIconSquareLogo                  = AtlanIcon{"PhSquareLogo"}
	AtlanIconSquareSplitHorizontal       = AtlanIcon{"PhSquareSplitHorizontal"}
	AtlanIconSquareSplitVertical         = AtlanIcon{"PhSquareSplitVertical"}
	AtlanIconSquaresFour                 = AtlanIcon{"PhSquaresFour"}
	AtlanIconStack                       = AtlanIcon{"PhStack"}
	AtlanIconStackOverflowLogo           = AtlanIcon{"PhStackOverflowLogo"}
	AtlanIconStackSimple                 = AtlanIcon{"PhStackSimple"}
	AtlanIconStairs                      = AtlanIcon{"PhStairs"}
	AtlanIconStamp                       = AtlanIcon{"PhStamp"}
	AtlanIconStarAndCrescent             = AtlanIcon{"PhStarAndCrescent"}
	AtlanIconStar                        = AtlanIcon{"PhStar"}
	AtlanIconStarFour                    = AtlanIcon{"PhStarFour"}
	AtlanIconStarHalf                    = AtlanIcon{"PhStarHalf"}
	AtlanIconStarOfDavid                 = AtlanIcon{"PhStarOfDavid"}
	AtlanIconSteeringWheel               = AtlanIcon{"PhSteeringWheel"}
	AtlanIconSteps                       = AtlanIcon{"PhSteps"}
	AtlanIconStethoscope                 = AtlanIcon{"PhStethoscope"}
	AtlanIconSticker                     = AtlanIcon{"PhSticker"}
	AtlanIconStool                       = AtlanIcon{"PhStool"}
	AtlanIconStop                        = AtlanIcon{"PhStop"}
	AtlanIconStopCircle                  = AtlanIcon{"PhStopCircle"}
	AtlanIconStorefront                  = AtlanIcon{"PhStorefront"}
	AtlanIconStrategy                    = AtlanIcon{"PhStrategy"}
	AtlanIconStripeLogo                  = AtlanIcon{"PhStripeLogo"}
	AtlanIconStudent                     = AtlanIcon{"PhStudent"}
	AtlanIconSubtitles                   = AtlanIcon{"PhSubtitles"}
	AtlanIconSubtract                    = AtlanIcon{"PhSubtract"}
	AtlanIconSubtractSquare              = AtlanIcon{"PhSubtractSquare"}
	AtlanIconSuitcase                    = AtlanIcon{"PhSuitcase"}
	AtlanIconSuitcaseRolling             = AtlanIcon{"PhSuitcaseRolling"}
	AtlanIconSuitcaseSimple              = AtlanIcon{"PhSuitcaseSimple"}
	AtlanIconSun                         = AtlanIcon{"PhSun"}
	AtlanIconSunDim                      = AtlanIcon{"PhSunDim"}
	AtlanIconSunHorizon                  = AtlanIcon{"PhSunHorizon"}
	AtlanIconSunglasses                  = AtlanIcon{"PhSunglasses"}
	AtlanIconSwap                        = AtlanIcon{"PhSwap"}
	AtlanIconSwatches                    = AtlanIcon{"PhSwatches"}
	AtlanIconSwimmingPool                = AtlanIcon{"PhSwimmingPool"}
	AtlanIconSword                       = AtlanIcon{"PhSword"}
	AtlanIconSynagogue                   = AtlanIcon{"PhSynagogue"}
	AtlanIconSyringe                     = AtlanIcon{"PhSyringe"}
	AtlanIconTShirt                      = AtlanIcon{"PhTShirt"}
	AtlanIconTable                       = AtlanIcon{"PhTable"}
	AtlanIconTabs                        = AtlanIcon{"PhTabs"}
	AtlanIconTag                         = AtlanIcon{"PhTag"}
	AtlanIconTagChevron                  = AtlanIcon{"PhTagChevron"}
	AtlanIconTagSimple                   = AtlanIcon{"PhTagSimple"}
	AtlanIconTarget                      = AtlanIcon{"PhTarget"}
	AtlanIconTaxi                        = AtlanIcon{"PhTaxi"}
	AtlanIconTelegramLogo                = AtlanIcon{"PhTelegramLogo"}
	AtlanIconTelevision                  = AtlanIcon{"PhTelevision"}
	AtlanIconTelevisionSimple            = AtlanIcon{"PhTelevisionSimple"}
	AtlanIconTennisBall                  = AtlanIcon{"PhTennisBall"}
	AtlanIconTent                        = AtlanIcon{"PhTent"}
	AtlanIconTerminal                    = AtlanIcon{"PhTerminal"}
	AtlanIconTerminalWindow              = AtlanIcon{"PhTerminalWindow"}
	AtlanIconTestTube                    = AtlanIcon{"PhTestTube"}
	AtlanIconTextAUnderline              = AtlanIcon{"PhTextAUnderline"}
	AtlanIconTextAa                      = AtlanIcon{"PhTextAa"}
	AtlanIconTextAlignCenter             = AtlanIcon{"PhTextAlignCenter"}
	AtlanIconTextAlignJustify            = AtlanIcon{"PhTextAlignJustify"}
	AtlanIconTextAlignLeft               = AtlanIcon{"PhTextAlignLeft"}
	AtlanIconTextAlignRight              = AtlanIcon{"PhTextAlignRight"}
	AtlanIconTextB                       = AtlanIcon{"PhTextB"}
	AtlanIconTextColumns                 = AtlanIcon{"PhTextColumns"}
	AtlanIconTextH                       = AtlanIcon{"PhTextH"}
	AtlanIconTextHFive                   = AtlanIcon{"PhTextHFive"}
	AtlanIconTextHFour                   = AtlanIcon{"PhTextHFour"}
	AtlanIconTextHOne                    = AtlanIcon{"PhTextHOne"}
	AtlanIconTextHSix                    = AtlanIcon{"PhTextHSix"}
	AtlanIconTextHThree                  = AtlanIcon{"PhTextHThree"}
	AtlanIconTextHTwo                    = AtlanIcon{"PhTextHTwo"}
	AtlanIconTextIndent                  = AtlanIcon{"PhTextIndent"}
	AtlanIconTextItalic                  = AtlanIcon{"PhTextItalic"}
	AtlanIconTextOutdent                 = AtlanIcon{"PhTextOutdent"}
	AtlanIconTextStrikethrough           = AtlanIcon{"PhTextStrikethrough"}
	AtlanIconTextT                       = AtlanIcon{"PhTextT"}
	AtlanIconTextUnderline               = AtlanIcon{"PhTextUnderline"}
	AtlanIconTextbox                     = AtlanIcon{"PhTextbox"}
	AtlanIconThermometer                 = AtlanIcon{"PhThermometer"}
	AtlanIconThermometerCold             = AtlanIcon{"PhThermometerCold"}
	AtlanIconThermometerHot              = AtlanIcon{"PhThermometerHot"}
	AtlanIconThermometerSimple           = AtlanIcon{"PhThermometerSimple"}
	AtlanIconThumbsDown                  = AtlanIcon{"PhThumbsDown"}
	AtlanIconThumbsUp                    = AtlanIcon{"PhThumbsUp"}
	AtlanIconTicket                      = AtlanIcon{"PhTicket"}
	AtlanIconTidalLogo                   = AtlanIcon{"PhTidalLogo"}
	AtlanIconTiktokLogo                  = AtlanIcon{"PhTiktokLogo"}
	AtlanIconTimer                       = AtlanIcon{"PhTimer"}
	AtlanIconTipi                        = AtlanIcon{"PhTipi"}
	AtlanIconToggleLeft                  = AtlanIcon{"PhToggleLeft"}
	AtlanIconToggleRight                 = AtlanIcon{"PhToggleRight"}
	AtlanIconToilet                      = AtlanIcon{"PhToilet"}
	AtlanIconToiletPaper                 = AtlanIcon{"PhToiletPaper"}
	AtlanIconToolbox                     = AtlanIcon{"PhToolbox"}
	AtlanIconTooth                       = AtlanIcon{"PhTooth"}
	AtlanIconTote                        = AtlanIcon{"PhTote"}
	AtlanIconToteSimple                  = AtlanIcon{"PhToteSimple"}
	AtlanIconTrademark                   = AtlanIcon{"PhTrademark"}
	AtlanIconTrademarkRegistered         = AtlanIcon{"PhTrademarkRegistered"}
	AtlanIconTrafficCone                 = AtlanIcon{"PhTrafficCone"}
	AtlanIconTrafficSign                 = AtlanIcon{"PhTrafficSign"}
	AtlanIconTrafficSignal               = AtlanIcon{"PhTrafficSignal"}
	AtlanIconTrain                       = AtlanIcon{"PhTrain"}
	AtlanIconTrainRegional               = AtlanIcon{"PhTrainRegional"}
	AtlanIconTrainSimple                 = AtlanIcon{"PhTrainSimple"}
	AtlanIconTram                        = AtlanIcon{"PhTram"}
	AtlanIconTranslate                   = AtlanIcon{"PhTranslate"}
	AtlanIconTrash                       = AtlanIcon{"PhTrash"}
	AtlanIconTrashSimple                 = AtlanIcon{"PhTrashSimple"}
	AtlanIconTray                        = AtlanIcon{"PhTray"}
	AtlanIconTree                        = AtlanIcon{"PhTree"}
	AtlanIconTreeEvergreen               = AtlanIcon{"PhTreeEvergreen"}
	AtlanIconTreePalm                    = AtlanIcon{"PhTreePalm"}
	AtlanIconTreeStructure               = AtlanIcon{"PhTreeStructure"}
	AtlanIconTrendDown                   = AtlanIcon{"PhTrendDown"}
	AtlanIconTrendUp                     = AtlanIcon{"PhTrendUp"}
	AtlanIconTriangle                    = AtlanIcon{"PhTriangle"}
	AtlanIconTrophy                      = AtlanIcon{"PhTrophy"}
	AtlanIconTruck                       = AtlanIcon{"PhTruck"}
	AtlanIconTwitchLogo                  = AtlanIcon{"PhTwitchLogo"}
	AtlanIconTwitterLogo                 = AtlanIcon{"PhTwitterLogo"}
	AtlanIconUmbrella                    = AtlanIcon{"PhUmbrella"}
	AtlanIconUmbrellaSimple              = AtlanIcon{"PhUmbrellaSimple"}
	AtlanIconUnite                       = AtlanIcon{"PhUnite"}
	AtlanIconUniteSquare                 = AtlanIcon{"PhUniteSquare"}
	AtlanIconUpload                      = AtlanIcon{"PhUpload"}
	AtlanIconUploadSimple                = AtlanIcon{"PhUploadSimple"}
	AtlanIconUsb                         = AtlanIcon{"PhUsb"}
	AtlanIconUser                        = AtlanIcon{"PhUser"}
	AtlanIconUserCircle                  = AtlanIcon{"PhUserCircle"}
	AtlanIconUserCircleGear              = AtlanIcon{"PhUserCircleGear"}
	AtlanIconUserCircleMinus             = AtlanIcon{"PhUserCircleMinus"}
	AtlanIconUserCirclePlus              = AtlanIcon{"PhUserCirclePlus"}
	AtlanIconUserFocus                   = AtlanIcon{"PhUserFocus"}
	AtlanIconUserGear                    = AtlanIcon{"PhUserGear"}
	AtlanIconUserList                    = AtlanIcon{"PhUserList"}
	AtlanIconUserMinus                   = AtlanIcon{"PhUserMinus"}
	AtlanIconUserPlus                    = AtlanIcon{"PhUserPlus"}
	AtlanIconUserRectangle               = AtlanIcon{"PhUserRectangle"}
	AtlanIconUserSquare                  = AtlanIcon{"PhUserSquare"}
	AtlanIconUserSwitch                  = AtlanIcon{"PhUserSwitch"}
	AtlanIconUsers                       = AtlanIcon{"PhUsers"}
	AtlanIconUsersFour                   = AtlanIcon{"PhUsersFour"}
	AtlanIconUsersThree                  = AtlanIcon{"PhUsersThree"}
	AtlanIconVan                         = AtlanIcon{"PhVan"}
	AtlanIconVault                       = AtlanIcon{"PhVault"}
	AtlanIconVibrate                     = AtlanIcon{"PhVibrate"}
	AtlanIconVideo                       = AtlanIcon{"PhVideo"}
	AtlanIconVideoCamera                 = AtlanIcon{"PhVideoCamera"}
	AtlanIconVideoCameraSlash            = AtlanIcon{"PhVideoCameraSlash"}
	AtlanIconVignette                    = AtlanIcon{"PhVignette"}
	AtlanIconVinylRecord                 = AtlanIcon{"PhVinylRecord"}
	AtlanIconVirtualReality              = AtlanIcon{"PhVirtualReality"}
	AtlanIconVirus                       = AtlanIcon{"PhVirus"}
	AtlanIconVoicemail                   = AtlanIcon{"PhVoicemail"}
	AtlanIconVolleyball                  = AtlanIcon{"PhVolleyball"}
	AtlanIconWall                        = AtlanIcon{"PhWall"}
	AtlanIconWallet                      = AtlanIcon{"PhWallet"}
	AtlanIconWarehouse                   = AtlanIcon{"PhWarehouse"}
	AtlanIconWarning                     = AtlanIcon{"PhWarning"}
	AtlanIconWarningCircle               = AtlanIcon{"PhWarningCircle"}
	AtlanIconWarningDiamond              = AtlanIcon{"PhWarningDiamond"}
	AtlanIconWarningOctagon              = AtlanIcon{"PhWarningOctagon"}
	AtlanIconWatch                       = AtlanIcon{"PhWatch"}
	AtlanIconWaveSawtooth                = AtlanIcon{"PhWaveSawtooth"}
	AtlanIconWaveSine                    = AtlanIcon{"PhWaveSine"}
	AtlanIconWaveSquare                  = AtlanIcon{"PhWaveSquare"}
	AtlanIconWaveTriangle                = AtlanIcon{"PhWaveTriangle"}
	AtlanIconWaveform                    = AtlanIcon{"PhWaveform"}
	AtlanIconWaves                       = AtlanIcon{"PhWaves"}
	AtlanIconWebcam                      = AtlanIcon{"PhWebcam"}
	AtlanIconWebcamSlash                 = AtlanIcon{"PhWebcamSlash"}
	AtlanIconWebhooksLogo                = AtlanIcon{"PhWebhooksLogo"}
	AtlanIconWechatLogo                  = AtlanIcon{"PhWechatLogo"}
	AtlanIconWhatsappLogo                = AtlanIcon{"PhWhatsappLogo"}
	AtlanIconWheelchair                  = AtlanIcon{"PhWheelchair"}
	AtlanIconWheelchairMotion            = AtlanIcon{"PhWheelchairMotion"}
	AtlanIconWifiHigh                    = AtlanIcon{"PhWifiHigh"}
	AtlanIconWifiLow                     = AtlanIcon{"PhWifiLow"}
	AtlanIconWifiMedium                  = AtlanIcon{"PhWifiMedium"}
	AtlanIconWifiNone                    = AtlanIcon{"PhWifiNone"}
	AtlanIconWifiSlash                   = AtlanIcon{"PhWifiSlash"}
	AtlanIconWifiX                       = AtlanIcon{"PhWifiX"}
	AtlanIconWind                        = AtlanIcon{"PhWind"}
	AtlanIconWindowsLogo                 = AtlanIcon{"PhWindowsLogo"}
	AtlanIconWine                        = AtlanIcon{"PhWine"}
	AtlanIconWrench                      = AtlanIcon{"PhWrench"}
	AtlanIconX                           = AtlanIcon{"PhX"}
	AtlanIconXCircle                     = AtlanIcon{"PhXCircle"}
	AtlanIconXSquare                     = AtlanIcon{"PhXSquare"}
	AtlanIconYinYang                     = AtlanIcon{"PhYinYang"}
	AtlanIconYoutubeLogo                 = AtlanIcon{"PhYoutubeLogo"}
)

func (a *AtlanIcon) UnmarshalJSON(data []byte) error {
	var atlanIconName string
	if err := json.Unmarshal(data, &atlanIconName); err != nil {
		return err
	}

	switch atlanIconName {
	case "PhWind":
		*a = AtlanIconWind
	case "PhAirplaneInFlight":
		*a = AtlanIconAirplaneInFlight
	default:
		*a = AtlanIcon{name: atlanIconName}
	}

	return nil
}

func AtlanIconPtr(value AtlanIcon) *AtlanIcon {
	Icon := AtlanIcon{name: value.name}
	return &Icon
}

func (a AtlanIcon) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.name)
}

// LineageDirection represents the direction of lineage.
type LineageDirection struct {
	Name string
}

func (l LineageDirection) String() string {
	return l.Name
}

var (
	LineageDirectionUpstream   = LineageDirection{"INPUT"}
	LineageDirectionDownstream = LineageDirection{"OUTPUT"}
	LineageDirectionBoth       = LineageDirection{"BOTH"}
)

func (l *LineageDirection) UnmarshalJSON(data []byte) error {
	var directionName string
	if err := json.Unmarshal(data, &directionName); err != nil {
		return err
	}

	switch directionName {
	case "INPUT":
		*l = LineageDirectionUpstream
	case "OUTPUT":
		*l = LineageDirectionDownstream
	case "BOTH":
		*l = LineageDirectionBoth
	default:
		*l = LineageDirection{Name: directionName}
	}

	return nil
}

func (l LineageDirection) MarshalJSON() ([]byte, error) {
	return json.Marshal(l.Name)
}

// PersonaGlossaryAction represents actions related to persona glossary.
type PersonaGlossaryAction struct {
	Name string
}

func (p PersonaGlossaryAction) String() string {
	return p.Name
}

var (
	PersonaGlossaryActionCreate               = PersonaGlossaryAction{"persona-glossary-create"}
	PersonaGlossaryActionRead                 = PersonaGlossaryAction{"persona-glossary-read"}
	PersonaGlossaryActionUpdate               = PersonaGlossaryAction{"persona-glossary-update"}
	PersonaGlossaryActionDelete               = PersonaGlossaryAction{"persona-glossary-delete"}
	PersonaGlossaryActionUpdateCustomMetadata = PersonaGlossaryAction{"persona-glossary-update-custom-metadata"}
	PersonaGlossaryActionAddAtlanTag          = PersonaGlossaryAction{"persona-glossary-add-classifications"}
	PersonaGlossaryActionUpdateAtlanTag       = PersonaGlossaryAction{"persona-glossary-update-classifications"}
	PersonaGlossaryActionRemoveAtlanTag       = PersonaGlossaryAction{"persona-glossary-delete-classifications"}
)

func (p *PersonaGlossaryAction) UnmarshalJSON(data []byte) error {
	var actionName string
	if err := json.Unmarshal(data, &actionName); err != nil {
		return err
	}

	switch actionName {
	case "persona-glossary-create":
		*p = PersonaGlossaryActionCreate
	case "persona-glossary-read":
		*p = PersonaGlossaryActionRead
	case "persona-glossary-update":
		*p = PersonaGlossaryActionUpdate
	case "persona-glossary-delete":
		*p = PersonaGlossaryActionDelete
	case "persona-glossary-update-custom-metadata":
		*p = PersonaGlossaryActionUpdateCustomMetadata
	case "persona-glossary-add-classifications":
		*p = PersonaGlossaryActionAddAtlanTag
	case "persona-glossary-update-classifications":
		*p = PersonaGlossaryActionUpdateAtlanTag
	case "persona-glossary-delete-classifications":
		*p = PersonaGlossaryActionRemoveAtlanTag
	default:
		*p = PersonaGlossaryAction{Name: actionName}
	}

	return nil
}

func (p PersonaGlossaryAction) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.Name)
}

// PersonaMetadataAction represents actions related to persona metadata.
type PersonaMetadataAction struct {
	Name string
}

func (p PersonaMetadataAction) String() string {
	return p.Name
}

var (
	PersonaMetadataActionCreate               = PersonaMetadataAction{"persona-api-create"}
	PersonaMetadataActionRead                 = PersonaMetadataAction{"persona-asset-read"}
	PersonaMetadataActionUpdate               = PersonaMetadataAction{"persona-asset-update"}
	PersonaMetadataActionDelete               = PersonaMetadataAction{"persona-api-delete"}
	PersonaMetadataActionUpdateCustomMetadata = PersonaMetadataAction{"persona-business-update-metadata"}
	PersonaMetadataActionAddAtlanTag          = PersonaMetadataAction{"persona-entity-add-classification"}
	PersonaMetadataActionUpdateAtlanTag       = PersonaMetadataAction{"persona-entity-update-classification"}
	PersonaMetadataActionRemoveAtlanTag       = PersonaMetadataAction{"persona-entity-remove-classification"}
	PersonaMetadataActionAttachTerms          = PersonaMetadataAction{"persona-add-terms"}
	PersonaMetadataActionDetachTerms          = PersonaMetadataAction{"persona-remove-terms"}
)

// UnmarshalJSON customizes the unmarshalling of a PersonaMetadataAction from JSON.
func (p *PersonaMetadataAction) UnmarshalJSON(data []byte) error {
	var actionName string
	if err := json.Unmarshal(data, &actionName); err != nil {
		return err
	}

	switch actionName {
	case "persona-api-create":
		*p = PersonaMetadataActionCreate
	case "persona-asset-read":
		*p = PersonaMetadataActionRead
	case "persona-asset-update":
		*p = PersonaMetadataActionUpdate
	case "persona-api-delete":
		*p = PersonaMetadataActionDelete
	case "persona-business-update-metadata":
		*p = PersonaMetadataActionUpdateCustomMetadata
	case "persona-entity-add-classification":
		*p = PersonaMetadataActionAddAtlanTag
	case "persona-entity-update-classification":
		*p = PersonaMetadataActionUpdateAtlanTag
	case "persona-entity-remove-classification":
		*p = PersonaMetadataActionRemoveAtlanTag
	case "persona-add-terms":
		*p = PersonaMetadataActionAttachTerms
	case "persona-remove-terms":
		*p = PersonaMetadataActionDetachTerms
	default:
		*p = PersonaMetadataAction{Name: actionName}
	}

	return nil
}

// MarshalJSON customizes the marshalling of a PersonaMetadataAction to JSON.
func (p PersonaMetadataAction) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.Name)
}

// PurposeMetadataAction represents actions related to purpose metadata.
type PurposeMetadataAction struct {
	Name string
}

func (p PurposeMetadataAction) String() string {
	return p.Name
}

var (
	PurposeMetadataActionCreate               = PurposeMetadataAction{"entity-create"}
	PurposeMetadataActionRead                 = PurposeMetadataAction{"entity-read"}
	PurposeMetadataActionUpdate               = PurposeMetadataAction{"entity-update"}
	PurposeMetadataActionDelete               = PurposeMetadataAction{"entity-delete"}
	PurposeMetadataActionUpdateCustomMetadata = PurposeMetadataAction{"entity-update-business-metadata"}
	PurposeMetadataActionAddAtlanTag          = PurposeMetadataAction{"entity-add-classification"}
	PurposeMetadataActionReadAtlanTag         = PurposeMetadataAction{"entity-read-classification"}
	PurposeMetadataActionUpdateAtlanTag       = PurposeMetadataAction{"entity-update-classification"}
	PurposeMetadataActionRemoveAtlanTag       = PurposeMetadataAction{"entity-remove-classification"}
	PurposeMetadataActionAttachTerms          = PurposeMetadataAction{"purpose-add-terms"}
	PurposeMetadataActionDetachTerms          = PurposeMetadataAction{"purpose-remove-terms"}
)

// UnmarshalJSON customizes the unmarshalling of a PurposeMetadataAction from JSON.
func (p *PurposeMetadataAction) UnmarshalJSON(data []byte) error {
	var actionName string
	if err := json.Unmarshal(data, &actionName); err != nil {
		return err
	}

	switch actionName {
	case "entity-create":
		*p = PurposeMetadataActionCreate
	case "entity-read":
		*p = PurposeMetadataActionRead
	case "entity-update":
		*p = PurposeMetadataActionUpdate
	case "entity-delete":
		*p = PurposeMetadataActionDelete
	case "entity-update-business-metadata":
		*p = PurposeMetadataActionUpdateCustomMetadata
	case "entity-add-classification":
		*p = PurposeMetadataActionAddAtlanTag
	case "entity-read-classification":
		*p = PurposeMetadataActionReadAtlanTag
	case "entity-update-classification":
		*p = PurposeMetadataActionUpdateAtlanTag
	case "entity-remove-classification":
		*p = PurposeMetadataActionRemoveAtlanTag
	case "purpose-add-terms":
		*p = PurposeMetadataActionAttachTerms
	case "purpose-remove-terms":
		*p = PurposeMetadataActionDetachTerms
	default:
		*p = PurposeMetadataAction{Name: actionName}
	}

	return nil
}

// MarshalJSON customizes the marshalling of a PurposeMetadataAction to JSON.
func (p PurposeMetadataAction) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.Name)
}

// QueryParserSourceType represents types of query parsers.
type QueryParserSourceType struct {
	Name string
}

func (q QueryParserSourceType) String() string {
	return q.Name
}

var (
	QueryParserSourceTypeAnsi       = QueryParserSourceType{"ansi"}
	QueryParserSourceTypeBigquery   = QueryParserSourceType{"bigquery"}
	QueryParserSourceTypeHana       = QueryParserSourceType{"hana"}
	QueryParserSourceTypeHive       = QueryParserSourceType{"hive"}
	QueryParserSourceTypeMssql      = QueryParserSourceType{"mssql"}
	QueryParserSourceTypeMysql      = QueryParserSourceType{"mysql"}
	QueryParserSourceTypeOracle     = QueryParserSourceType{"oracle"}
	QueryParserSourceTypePostgresql = QueryParserSourceType{"postgresql"}
	QueryParserSourceTypeRedshift   = QueryParserSourceType{"redshift"}
	QueryParserSourceTypeSnowflake  = QueryParserSourceType{"snowflake"}
	QueryParserSourceTypeSparksql   = QueryParserSourceType{"sparksql"}
	QueryParserSourceTypeAthena     = QueryParserSourceType{"athena"}
)

func (q *QueryParserSourceType) UnmarshalJSON(data []byte) error {
	var sourceType string
	if err := json.Unmarshal(data, &sourceType); err != nil {
		return err
	}

	switch sourceType {
	case "ansi":
		*q = QueryParserSourceTypeAnsi
	case "bigquery":
		*q = QueryParserSourceTypeBigquery
	case "hana":
		*q = QueryParserSourceTypeHana
	case "hive":
		*q = QueryParserSourceTypeHive
	case "mssql":
		*q = QueryParserSourceTypeMssql
	case "mysql":
		*q = QueryParserSourceTypeMysql
	case "oracle":
		*q = QueryParserSourceTypeOracle
	case "postgresql":
		*q = QueryParserSourceTypePostgresql
	case "redshift":
		*q = QueryParserSourceTypeRedshift
	case "snowflake":
		*q = QueryParserSourceTypeSnowflake
	case "sparksql":
		*q = QueryParserSourceTypeSparksql
	case "athena":
		*q = QueryParserSourceTypeAthena
	default:
		*q = QueryParserSourceType{Name: sourceType}
	}

	return nil
}

func (q QueryParserSourceType) MarshalJSON() ([]byte, error) {
	return json.Marshal(q.Name)
}

// TagIconType represents types of tag icons.
type TagIconType struct {
	Name string
}

func (t TagIconType) String() string {
	return t.Name
}

var (
	TagIconTypeImage = TagIconType{"image"}
	TagIconTypeIcon  = TagIconType{"icon"}
	TagIconTypeEmoji = TagIconType{"emoji"}
	TagIconTypeNone  = TagIconType{""}
)

func (t *TagIconType) UnmarshalJSON(data []byte) error {
	var iconType string
	if err := json.Unmarshal(data, &iconType); err != nil {
		return err
	}

	switch iconType {
	case "image":
		*t = TagIconTypeImage
	case "icon":
		*t = TagIconTypeIcon
	case "emoji":
		*t = TagIconTypeEmoji
	default:
		*t = TagIconType{Name: iconType}
	}

	return nil
}

func (t TagIconType) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.Name)
}

// TypeName represents types of names.
type TypeName struct {
	Name string
}

func (t TypeName) String() string {
	return t.Name
}

var (
	TypeNameString      = TypeName{"string"}
	TypeNameArrayString = TypeName{"array<string>"}
)

func (t *TypeName) UnmarshalJSON(data []byte) error {
	var typeName string
	if err := json.Unmarshal(data, &typeName); err != nil {
		return err
	}

	switch typeName {
	case "string":
		*t = TypeNameString
	case "array<string>":
		*t = TypeNameArrayString
	default:
		*t = TypeName{Name: typeName}
	}

	return nil
}

func (t TypeName) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.Name)
}

// AtlanWorkflowPhase represents phases of Atlan workflows.
type AtlanWorkflowPhase struct {
	Name string
}

func (a AtlanWorkflowPhase) String() string {
	return a.Name
}

var (
	AtlanWorkflowPhaseSuccess = AtlanWorkflowPhase{"Succeeded"}
	AtlanWorkflowPhaseRunning = AtlanWorkflowPhase{"Running"}
	AtlanWorkflowPhaseFailed  = AtlanWorkflowPhase{"Failed"}
	AtlanWorkflowPhaseError   = AtlanWorkflowPhase{"Error"}
	AtlanWorkflowPhasePending = AtlanWorkflowPhase{"Pending"}
)

func (a *AtlanWorkflowPhase) UnmarshalJSON(data []byte) error {
	var phaseName string
	if err := json.Unmarshal(data, &phaseName); err != nil {
		return err
	}

	switch phaseName {
	case "Succeeded":
		*a = AtlanWorkflowPhaseSuccess
	case "Running":
		*a = AtlanWorkflowPhaseRunning
	case "Failed":
		*a = AtlanWorkflowPhaseFailed
	case "Error":
		*a = AtlanWorkflowPhaseError
	case "Pending":
		*a = AtlanWorkflowPhasePending
	default:
		*a = AtlanWorkflowPhase{Name: phaseName}
	}

	return nil
}

func (a AtlanWorkflowPhase) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.Name)
}

// ChildScoreMode represents modes of child scores.
type ChildScoreMode struct {
	Name string
}

func (c ChildScoreMode) String() string {
	return c.Name
}

var (
	ChildScoreModeNone = ChildScoreMode{"none"}
	ChildScoreModeAvg  = ChildScoreMode{"avg"}
	ChildScoreModeSum  = ChildScoreMode{"sum"}
	ChildScoreModeMax  = ChildScoreMode{"max"}
	ChildScoreModeMin  = ChildScoreMode{"min"}
)

// UnmarshalJSON customizes the unmarshalling of a ChildScoreMode from JSON.
func (c *ChildScoreMode) UnmarshalJSON(data []byte) error {
	var modeName string
	if err := json.Unmarshal(data, &modeName); err != nil {
		return err
	}

	// Set the corresponding ChildScoreMode based on the modeName.
	switch modeName {
	case "none":
		*c = ChildScoreModeNone
	case "avg":
		*c = ChildScoreModeAvg
	case "sum":
		*c = ChildScoreModeSum
	case "max":
		*c = ChildScoreModeMax
	case "min":
		*c = ChildScoreModeMin
	default:
		*c = ChildScoreMode{Name: modeName}
	}

	return nil
}

// MarshalJSON customizes the marshalling of a ChildScoreMode to JSON.
func (c ChildScoreMode) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.Name)
}

// WorkflowPackage represents types of workflow packages.
type WorkflowPackage struct {
	Name string
}

func (w WorkflowPackage) String() string {
	return w.Name
}

var (
	WorkflowPackageAirflow                    = WorkflowPackage{"atlan-airflow"}
	WorkflowPackageAthena                     = WorkflowPackage{"atlan-athena"}
	WorkflowPackageAwsLambdaTrigger           = WorkflowPackage{"atlan-aws-lambda-trigger"}
	WorkflowPackageAzureEventHub              = WorkflowPackage{"atlan-azure-event-hub"}
	WorkflowPackageBigquery                   = WorkflowPackage{"atlan-bigquery"}
	WorkflowPackageBigqueryMiner              = WorkflowPackage{"atlan-bigquery-miner"}
	WorkflowPackageConnectionDelete           = WorkflowPackage{"atlan-connection-delete"}
	WorkflowPackageDatabricks                 = WorkflowPackage{"atlan-databricks"}
	WorkflowPackageDatabricksLineage          = WorkflowPackage{"atlan-databricks-lineage"}
	WorkflowPackageDbt                        = WorkflowPackage{"atlan-dbt"}
	WorkflowPackageFivetran                   = WorkflowPackage{"atlan-fivetran"}
	WorkflowPackageGlue                       = WorkflowPackage{"atlan-glue"}
	WorkflowPackageHive                       = WorkflowPackage{"atlan-hive"}
	WorkflowPackageHiveMiner                  = WorkflowPackage{"atlan-hive-miner"}
	WorkflowPackageKafka                      = WorkflowPackage{"atlan-kafka"}
	WorkflowPackageKafkaAiven                 = WorkflowPackage{"atlan-kafka-aiven"}
	WorkflowPackageKafkaConfluentCloud        = WorkflowPackage{"atlan-kafka-confluent-cloud"}
	WorkflowPackageKafkaRedpanda              = WorkflowPackage{"atlan-kafka-redpanda"}
	WorkflowPackageLooker                     = WorkflowPackage{"atlan-looker"}
	WorkflowPackageMatillion                  = WorkflowPackage{"atlan-matillion"}
	WorkflowPackageMetabase                   = WorkflowPackage{"atlan-metabase"}
	WorkflowPackageMicrostrategy              = WorkflowPackage{"atlan-microstrategy"}
	WorkflowPackageMode                       = WorkflowPackage{"atlan-mode"}
	WorkflowPackageMonteCarlo                 = WorkflowPackage{"atlan-monte-carlo"}
	WorkflowPackageMssql                      = WorkflowPackage{"atlan-mssql"}
	WorkflowPackageMssqlMiner                 = WorkflowPackage{"atlan-mssql-miner"}
	WorkflowPackageMysql                      = WorkflowPackage{"atlan-mysql"}
	WorkflowPackageOracle                     = WorkflowPackage{"atlan-oracle"}
	WorkflowPackagePostgres                   = WorkflowPackage{"atlan-postgres"}
	WorkflowPackagePowerbi                    = WorkflowPackage{"atlan-powerbi"}
	WorkflowPackagePowerbiMiner               = WorkflowPackage{"atlan-powerbi-miner"}
	WorkflowPackagePresto                     = WorkflowPackage{"atlan-presto"}
	WorkflowPackageQlikSense                  = WorkflowPackage{"atlan-qlik-sense"}
	WorkflowPackageQlikSenseEnterpriseWindows = WorkflowPackage{"atlan-qlik-sense-enterprise-windows"}
	WorkflowPackageQuicksight                 = WorkflowPackage{"atlan-quicksight"}
	WorkflowPackageRedash                     = WorkflowPackage{"atlan-redash"}
	WorkflowPackageRedshift                   = WorkflowPackage{"atlan-redshift"}
	WorkflowPackageRedshiftMiner              = WorkflowPackage{"atlan-redshift-miner"}
	WorkflowPackageSalesforce                 = WorkflowPackage{"atlan-salesforce"}
	WorkflowPackageSapHana                    = WorkflowPackage{"atlan-sap-hana"}
	WorkflowPackageSchemaRegistryConfluent    = WorkflowPackage{"atlan-schema-registry-confluent"}
	WorkflowPackageSigma                      = WorkflowPackage{"atlan-sigma"}
	WorkflowPackageSnowflake                  = WorkflowPackage{"atlan-snowflake"}
	WorkflowPackageSnowflakeMiner             = WorkflowPackage{"atlan-snowflake-miner"}
	WorkflowPackageSoda                       = WorkflowPackage{"atlan-soda"}
	WorkflowPackageSynapse                    = WorkflowPackage{"atlan-synapse"}
	WorkflowPackageTableau                    = WorkflowPackage{"atlan-tableau"}
	WorkflowPackageTeradata                   = WorkflowPackage{"atlan-teradata"}
	WorkflowPackageTeradataMiner              = WorkflowPackage{"atlan-teradata-miner"}
	WorkflowPackageThoughtspot                = WorkflowPackage{"atlan-thoughtspot"}
	WorkflowPackageTrino                      = WorkflowPackage{"atlan-trino"}
)

// UnmarshalJSON customizes the unmarshalling of a WorkflowPackage from JSON.
func (w *WorkflowPackage) UnmarshalJSON(data []byte) error {
	var packageName string
	if err := json.Unmarshal(data, &packageName); err != nil {
		return err
	}

	// Based on the packageName, set the corresponding WorkflowPackage.
	switch packageName {
	case "atlan-airflow":
		*w = WorkflowPackageAirflow
	case "atlan-athena":
		*w = WorkflowPackageAthena
	case "atlan-aws-lambda-trigger":
		*w = WorkflowPackageAwsLambdaTrigger
	case "atlan-azure-event-hub":
		*w = WorkflowPackageAzureEventHub
	case "atlan-bigquery":
		*w = WorkflowPackageBigquery
	case "atlan-bigquery-miner":
		*w = WorkflowPackageBigqueryMiner
	case "atlan-connection-delete":
		*w = WorkflowPackageConnectionDelete
	case "atlan-databricks":
		*w = WorkflowPackageDatabricks
	case "atlan-databricks-lineage":
		*w = WorkflowPackageDatabricksLineage
	case "atlan-dbt":
		*w = WorkflowPackageDbt
	case "atlan-fivetran":
		*w = WorkflowPackageFivetran
	case "atlan-glue":
		*w = WorkflowPackageGlue
	case "atlan-hive":
		*w = WorkflowPackageHive
	case "atlan-hive-miner":
		*w = WorkflowPackageHiveMiner
	case "atlan-kafka":
		*w = WorkflowPackageKafka
	case "atlan-kafka-aiven":
		*w = WorkflowPackageKafkaAiven
	case "atlan-kafka-confluent-cloud":
		*w = WorkflowPackageKafkaConfluentCloud
	case "atlan-kafka-redpanda":
		*w = WorkflowPackageKafkaRedpanda
	case "atlan-looker":
		*w = WorkflowPackageLooker
	case "atlan-matillion":
		*w = WorkflowPackageMatillion
	case "atlan-metabase":
		*w = WorkflowPackageMetabase
	case "atlan-microstrategy":
		*w = WorkflowPackageMicrostrategy
	case "atlan-mode":
		*w = WorkflowPackageMode
	case "atlan-monte-carlo":
		*w = WorkflowPackageMonteCarlo
	case "atlan-mssql":
		*w = WorkflowPackageMssql
	case "atlan-mssql-miner":
		*w = WorkflowPackageMssqlMiner
	case "atlan-mysql":
		*w = WorkflowPackageMysql
	case "atlan-oracle":
		*w = WorkflowPackageOracle
	case "atlan-postgres":
		*w = WorkflowPackagePostgres
	case "atlan-powerbi":
		*w = WorkflowPackagePowerbi
	case "atlan-powerbi-miner":
		*w = WorkflowPackagePowerbiMiner
	case "atlan-presto":
		*w = WorkflowPackagePresto
	case "atlan-qlik-sense":
		*w = WorkflowPackageQlikSense
	case "atlan-qlik-sense-enterprise-windows":
		*w = WorkflowPackageQlikSenseEnterpriseWindows
	case "atlan-quicksight":
		*w = WorkflowPackageQuicksight
	case "atlan-redash":
		*w = WorkflowPackageRedash
	case "atlan-redshift":
		*w = WorkflowPackageRedshift
	case "atlan-redshift-miner":
		*w = WorkflowPackageRedshiftMiner
	case "atlan-salesforce":
		*w = WorkflowPackageSalesforce
	case "atlan-sap-hana":
		*w = WorkflowPackageSapHana
	case "atlan-schema-registry-confluent":
		*w = WorkflowPackageSchemaRegistryConfluent
	case "atlan-sigma":
		*w = WorkflowPackageSigma
	case "atlan-snowflake":
		*w = WorkflowPackageSnowflake
	case "atlan-snowflake-miner":
		*w = WorkflowPackageSnowflakeMiner
	case "atlan-soda":
		*w = WorkflowPackageSoda
	case "atlan-synapse":
		*w = WorkflowPackageSynapse
	case "atlan-tableau":
		*w = WorkflowPackageTableau
	case "atlan-teradata":
		*w = WorkflowPackageTeradata
	case "atlan-teradata-miner":
		*w = WorkflowPackageTeradataMiner
	case "atlan-thoughtspot":
		*w = WorkflowPackageThoughtspot
	case "atlan-trino":
		*w = WorkflowPackageTrino
	default:
		return errors.New("unknown workflow package: " + packageName)
	}

	return nil
}

// MarshalJSON customizes the marshalling of a WorkflowPackage to JSON.
func (w WorkflowPackage) MarshalJSON() ([]byte, error) {
	return json.Marshal(w.Name)
}

type AtlanStatus string

func AtlanStatusPtr(value string) *AtlanStatus {
	status := AtlanStatus(value)
	return &status
}

type QueryUsernameStrategy struct {
	Name string
}

func (q QueryUsernameStrategy) String() string {
	return q.Name
}

var (
	CONNECTION_USERNAME = QueryUsernameStrategy{"connectionUsername"}
	ATLAN_USERNAME      = QueryUsernameStrategy{Name: "atlanUsername"}
)

func (q QueryUsernameStrategy) MarshalJSON() ([]byte, error) {
	return json.Marshal(q.Name)
}

func (q *QueryUsernameStrategy) UnmarshalJSON(data []byte) error {
	var UserStrategy string
	if err := json.Unmarshal(data, &UserStrategy); err != nil {
		return err
	}

	// Set the corresponding ChildScoreMode based on the modeName.
	switch UserStrategy {
	case "connectionUsername":
		*q = CONNECTION_USERNAME
	case "atlanUsername":
		*q = ATLAN_USERNAME

	default:
		*q = QueryUsernameStrategy{Name: UserStrategy}
	}

	return nil
}

type UTMTags struct {
	Name string
}

func (u UTMTags) String() string {
	return u.Name
}

var (
	// PAGE_ entries indicate where the action was taken.

	// Search was made from the home page.
	PAGE_HOME = UTMTags{"page_home"}
	// Search was made from the assets (discovery) page.
	PAGE_ASSETS = UTMTags{"page_assets"}
	// Asset was viewed from within a glossary.
	PAGE_GLOSSARY = UTMTags{"page_glossary"}
	// Asset was viewed from within insights.
	PAGE_INSIGHTS = UTMTags{"page_insights"}

	// PROJECT_ entries indicate how (via what application) the action was taken.

	// Search was made via the webapp (UI)
	PROJECT_WEBAPP = UTMTags{"project_webapp"}
	// Search was made via the Java SDK.
	PROJECT_SDK_JAVA = UTMTags{"project_sdk_java"}
	// Search was made via the Python SDK.
	PROJECT_SDK_PYTHON = UTMTags{"project_sdk_python"}
	// Search was made via the Go SDK.
	PROJECT_SDK_GO = UTMTags{"project_sdk_go"}
	// Search was made via the atlan cli.
	PROJECT_SDK_CLI = UTMTags{"project_sdk_cli"}

	// ACTION_ entries dictate the specific action that was taken.

	// Assets were searched.
	ACTION_SEARCHED = UTMTags{"action_searched"}
	// Search was run through the Cmd-K popup.
	ACTION_CMD_K = UTMTags{"action_cmd_k"}
	// Search was through changing a filter in the UI (discovery).
	ACTION_FILTER_CHANGED = UTMTags{"action_filter_changed"}
	// Search was through changing a type filter (pill) in the UI (discovery)
	ACTION_ASSET_TYPE_CHANGED = UTMTags{"action_asset_type_changed"}
	// Asset was viewed, rather than an explicit search.
	ACTION_ASSET_VIEWED = UTMTags{"action_asset_viewed"}

	// Others indicate any special mechanisms used for the action.

	// Search was run using the UI popup searchbar.
	UI_POPUP_SEARCHBAR = UTMTags{"ui_popup_searchbar"}
	// Search was through a UI filter (discovery).
	UI_FILTERS = UTMTags{"ui_filters"}
	// View was done via the UI's sidebar.
	UI_SIDEBAR = UTMTags{"ui_sidebar"}
	// View was done of the full asset profile, not only sidebar.
	UI_PROFILE = UTMTags{"ui_profile"}
	// Listing of assets, usually by a particular type, in the discovery page.
	UI_MAIN_LIST = UTMTags{"ui_main_list"}
)

func (u UTMTags) MarshalJSON() ([]byte, error) {
	return json.Marshal(u.Name)
}

func (u *UTMTags) UnmarshalJSON(data []byte) error {
	var UtmTags string
	if err := json.Unmarshal(data, &UtmTags); err != nil {
		return err
	}

	switch UtmTags { //nolint:gocritic
	case "page_home":
		*u = PAGE_HOME
		switch UtmTags {
		case "page_home":
			*u = PAGE_HOME
		case "page_assets":
			*u = PAGE_ASSETS
		case "page_glossary":
			*u = PAGE_GLOSSARY
		case "page_insights":
			*u = PAGE_INSIGHTS
		case "project_webapp":
			*u = PROJECT_WEBAPP
		case "project_sdk_java":
			*u = PROJECT_SDK_JAVA
		case "project_sdk_python":
			*u = PROJECT_SDK_PYTHON
		case "project_sdk_go":
			*u = PROJECT_SDK_GO
		case "project_sdk_cli":
			*u = PROJECT_SDK_CLI
		case "action_searched":
			*u = ACTION_SEARCHED
		case "action_cmd_k":
			*u = ACTION_CMD_K
		case "action_filter_changed":
			*u = ACTION_FILTER_CHANGED
		case "action_asset_type_changed":
			*u = ACTION_ASSET_TYPE_CHANGED
		case "action_asset_viewed":
			*u = ACTION_ASSET_VIEWED
		case "ui_popup_searchbar":
			*u = UI_POPUP_SEARCHBAR
		case "ui_filters":
			*u = UI_FILTERS
		case "ui_sidebar":
			*u = UI_SIDEBAR
		case "ui_profile":
			*u = UI_PROFILE
		case "ui_main_list":
			*u = UI_MAIN_LIST
		default:
			*u = UTMTags{Name: UtmTags}
		}
	}
	return nil
}

type CustomMetadataHandling struct {
	Name string
}

func (c CustomMetadataHandling) String() string {
	return c.Name
}

var (
	IGNORE    = CustomMetadataHandling{"ignore"}
	OVERWRITE = CustomMetadataHandling{Name: "overwrite"}
	MERGE     = CustomMetadataHandling{Name: "merge"}
)

func (c CustomMetadataHandling) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.Name)
}

func (c *CustomMetadataHandling) UnmarshalJSON(data []byte) error {
	var CustomMetadataType string
	if err := json.Unmarshal(data, &CustomMetadataType); err != nil {
		return err
	}

	switch CustomMetadataType {
	case "ignore":
		*c = IGNORE
	case "overwrite":
		*c = OVERWRITE
	case "merge":
		*c = MERGE

	default:
		*c = CustomMetadataHandling{Name: CustomMetadataType}
	}

	return nil
}

type CertificateStatus struct {
	Name string
}

func (a CertificateStatus) String() string {
	return a.Name
}

var (
	CertificateStatusDeprecated = CertificateStatus{"DEPRECATED"}
	CertificateStatusDraft      = CertificateStatus{"DRAFT"}
	CertificateStatusVerified   = CertificateStatus{"VERIFIED"}
)

// UnmarshalJSON customizes the unmarshalling of a certificate_status from JSON.
func (c *CertificateStatus) UnmarshalJSON(data []byte) error {
	var name string
	if err := json.Unmarshal(data, &name); err != nil {
		return err
	}

	switch name {
	case "DEPRECATED":
		*c = CertificateStatusDeprecated

	case "DRAFT":
		*c = CertificateStatusDraft

	case "VERIFIED":
		*c = CertificateStatusVerified
	default:
		*c = CertificateStatus{Name: name}
	}

	return nil
}

// MarshalJSON customizes the marshalling of a certificate_status to JSON.
func (c CertificateStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.Name)
}

type AuthPolicyType struct {
	Name string
}

func (a AuthPolicyType) String() string {
	return a.Name
}

var (
	AuthPolicyTypeAllow           = AuthPolicyType{"allow"}
	AuthPolicyTypeDeny            = AuthPolicyType{"deny"}
	AuthPolicyTypeAllowexceptions = AuthPolicyType{"allowExceptions"}
	AuthPolicyTypeDenyexceptions  = AuthPolicyType{"denyExceptions"}
	AuthPolicyTypeDatamask        = AuthPolicyType{"dataMask"}
	AuthPolicyTypeRowfilter       = AuthPolicyType{"rowFilter"}
)

// UnmarshalJSON customizes the unmarshalling of a AuthPolicyType from JSON.
func (c *AuthPolicyType) UnmarshalJSON(data []byte) error {
	var name string
	if err := json.Unmarshal(data, &name); err != nil {
		return err
	}

	switch name {
	case "allow":
		*c = AuthPolicyTypeAllow

	case "deny":
		*c = AuthPolicyTypeDeny

	case "allowExceptions":
		*c = AuthPolicyTypeAllowexceptions

	case "denyExceptions":
		*c = AuthPolicyTypeDenyexceptions

	case "dataMask":
		*c = AuthPolicyTypeDatamask

	case "rowFilter":
		*c = AuthPolicyTypeRowfilter
	default:
		*c = AuthPolicyType{Name: name}
	}

	return nil
}

// MarshalJSON customizes the marshalling of a AuthPolicyType to JSON.
func (c AuthPolicyType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.Name)
}

type AuthPolicyCategory struct {
	Name string
}

func (a AuthPolicyCategory) String() string {
	return a.Name
}

var (
	AuthPolicyCategoryBootstrap = AuthPolicyCategory{"bootstrap"}
	AuthPolicyCategoryPersona   = AuthPolicyCategory{"persona"}
	AuthPolicyCategoryPurpose   = AuthPolicyCategory{"purpose"}
)

// UnmarshalJSON customizes the unmarshalling of a AuthPolicyCategory from JSON.
func (c *AuthPolicyCategory) UnmarshalJSON(data []byte) error {
	var name string
	if err := json.Unmarshal(data, &name); err != nil {
		return err
	}

	switch name {
	case "bootstrap":
		*c = AuthPolicyCategoryBootstrap

	case "persona":
		*c = AuthPolicyCategoryPersona

	case "purpose":
		*c = AuthPolicyCategoryPurpose
	default:
		*c = AuthPolicyCategory{Name: name}
	}

	return nil
}

// MarshalJSON customizes the marshalling of a AuthPolicyCategory to JSON.
func (c AuthPolicyCategory) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.Name)
}

type AuthPolicyResourceCategory struct {
	Name string
}

func (a AuthPolicyResourceCategory) String() string {
	return a.Name
}

var (
	AuthPolicyResourceCategoryEntity       = AuthPolicyResourceCategory{"ENTITY"}
	AuthPolicyResourceCategoryRelationship = AuthPolicyResourceCategory{"RELATIONSHIP"}
	AuthPolicyResourceCategoryTag          = AuthPolicyResourceCategory{"TAG"}
	AuthPolicyResourceCategoryCustom       = AuthPolicyResourceCategory{"CUSTOM"}
	AuthPolicyResourceCategoryTypedefs     = AuthPolicyResourceCategory{"TYPEDEFS"}
	AuthPolicyResourceCategoryAdmin        = AuthPolicyResourceCategory{"ADMIN"}
)

// UnmarshalJSON customizes the unmarshalling of a AuthPolicyResourceCategory from JSON.
func (c *AuthPolicyResourceCategory) UnmarshalJSON(data []byte) error {
	var name string
	if err := json.Unmarshal(data, &name); err != nil {
		return err
	}

	switch name {
	case "ENTITY":
		*c = AuthPolicyResourceCategoryEntity

	case "RELATIONSHIP":
		*c = AuthPolicyResourceCategoryRelationship

	case "TAG":
		*c = AuthPolicyResourceCategoryTag

	case "CUSTOM":
		*c = AuthPolicyResourceCategoryCustom

	case "TYPEDEFS":
		*c = AuthPolicyResourceCategoryTypedefs

	case "ADMIN":
		*c = AuthPolicyResourceCategoryAdmin
	default:
		*c = AuthPolicyResourceCategory{Name: name}
	}

	return nil
}

// MarshalJSON customizes the marshalling of a AuthPolicyResourceCategory to JSON.
func (c AuthPolicyResourceCategory) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.Name)
}

type DataAction struct {
	Name string
}

func (a DataAction) String() string {
	return a.Name
}

var DataActionSelect = DataAction{"select"}

// UnmarshalJSON customizes the unmarshalling of a DataAction from JSON.
func (c *DataAction) UnmarshalJSON(data []byte) error {
	var name string
	if err := json.Unmarshal(data, &name); err != nil {
		return err
	}

	switch name {
	case "select":
		*c = DataActionSelect
	default:
		*c = DataAction{Name: name}
	}
	return nil
}

// MarshalJSON customizes the marshalling of a DataAction to JSON.
func (c DataAction) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.Name)
}

type PersonaDomainAction struct {
	Name string
}

func (a PersonaDomainAction) String() string {
	return a.Name
}

var (
	PersonaDomainActionCreate                        = PersonaDomainAction{"persona-domain-create"}
	PersonaDomainActionRead                          = PersonaDomainAction{"persona-domain-read"}
	PersonaDomainActionUpdate                        = PersonaDomainAction{"persona-domain-update"}
	PersonaDomainActionDelete                        = PersonaDomainAction{"persona-domain-delete"}
	PersonaDomainActionCreateSubdomain               = PersonaDomainAction{"persona-domain-sub-domain-create"}
	PersonaDomainActionReadSubdomain                 = PersonaDomainAction{"persona-domain-sub-domain-read"}
	PersonaDomainActionUpdateSubdomain               = PersonaDomainAction{"persona-domain-sub-domain-update"}
	PersonaDomainActionDeleteSubdomain               = PersonaDomainAction{"persona-domain-sub-domain-delete"}
	PersonaDomainActionCreateProducts                = PersonaDomainAction{"persona-domain-product-create"}
	PersonaDomainActionReadProducts                  = PersonaDomainAction{"persona-domain-product-read"}
	PersonaDomainActionUpdateProducts                = PersonaDomainAction{"persona-domain-product-update"}
	PersonaDomainActionDeleteProducts                = PersonaDomainAction{"persona-domain-product-delete"}
	PersonaDomainActionUpdateDomainCustomMetadata    = PersonaDomainAction{"persona-domain-business-update-metadata"}
	PersonaDomainActionUpdateSubdomainCustomMetadata = PersonaDomainAction{"persona-domain-sub-domain-business-update-metadata"}
	PersonaDomainActionUpdateProductCustomMetadata   = PersonaDomainAction{"persona-domain-product-business-update-metadata"}
)

// UnmarshalJSON customizes the unmarshalling of a PersonaDomainAction from JSON.
func (c *PersonaDomainAction) UnmarshalJSON(data []byte) error {
	var name string
	if err := json.Unmarshal(data, &name); err != nil {
		return err
	}

	switch name {
	case "persona-domain-create":
		*c = PersonaDomainActionCreate
	case "persona-domain-read":
		*c = PersonaDomainActionRead
	case "persona-domain-update":
		*c = PersonaDomainActionUpdate
	case "persona-domain-delete":
		*c = PersonaDomainActionDelete
	case "persona-domain-sub-domain-create":
		*c = PersonaDomainActionCreateSubdomain
	case "persona-domain-sub-domain-read":
		*c = PersonaDomainActionReadSubdomain
	case "persona-domain-sub-domain-update":
		*c = PersonaDomainActionUpdateSubdomain
	case "persona-domain-sub-domain-delete":
		*c = PersonaDomainActionDeleteSubdomain
	case "persona-domain-product-create":
		*c = PersonaDomainActionCreateProducts
	case "persona-domain-product-read":
		*c = PersonaDomainActionReadProducts
	case "persona-domain-product-update":
		*c = PersonaDomainActionUpdateProducts
	case "persona-domain-product-delete":
		*c = PersonaDomainActionDeleteProducts
	case "persona-domain-business-update-metadata":
		*c = PersonaDomainActionUpdateDomainCustomMetadata
	case "persona-domain-sub-domain-business-update-metadata":
		*c = PersonaDomainActionUpdateSubdomainCustomMetadata
	case "persona-domain-product-business-update-metadata":
		*c = PersonaDomainActionUpdateProductCustomMetadata
	default:
		*c = PersonaDomainAction{Name: name}
	}
	return nil
}

// MarshalJSON customizes the marshalling of a PersonaDomainAction to JSON.
func (c PersonaDomainAction) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.Name)
}

type AssetFilterGroup struct {
	Name string
}

func (a AssetFilterGroup) String() string {
	return a.Name
}

var (
	AssetFilterGroupTerms       = AssetFilterGroup{"terms"}
	AssetFilterGroupTags        = AssetFilterGroup{"__traitNames"}
	AssetFilterGroupOwners      = AssetFilterGroup{"owners"}
	AssetFilterGroupUsage       = AssetFilterGroup{"usage"}
	AssetFilterGroupProperties  = AssetFilterGroup{"properties"}
	AssetFilterGroupCertificate = AssetFilterGroup{"certificateStatus"}
)

// UnmarshalJSON customizes the unmarshalling of a AssetFilterGroup from JSON.
func (c *AssetFilterGroup) UnmarshalJSON(data []byte) error {
	var name string
	if err := json.Unmarshal(data, &name); err != nil {
		return err
	}
	switch name {
	case "terms":
		*c = AssetFilterGroupTerms
	case "__traitNames":
		*c = AssetFilterGroupTags
	case "owners":
		*c = AssetFilterGroupOwners
	case "usage":
		*c = AssetFilterGroupUsage
	case "properties":
		*c = AssetFilterGroupProperties
	case "certificateStatus":
		*c = AssetFilterGroupCertificate
	default:
		*c = AssetFilterGroup{Name: name}
	}
	return nil
}

type DataMaskingType struct {
	Name string
}

func (d DataMaskingType) String() string {
	return d.Name
}

var (
	DataMaskingTypeSHOWFIRST4 = DataMaskingType{"MASK_SHOW_FIRST_4"}
	DataMaskingTypeSHOWLAST4  = DataMaskingType{"MASK_SHOW_LAST_4"}
	DataMaskingTypeHASH       = DataMaskingType{"MASK_HASH"}
	DataMaskingTypeNULLIFY    = DataMaskingType{"MASK_NULL"}
	DataMaskingTypeREDACT     = DataMaskingType{"MASK_REDACT"}
)

// UnmarshalJSON customizes the unmarshalling of a DataMaskingType from JSON.
func (d *DataMaskingType) UnmarshalJSON(data []byte) error {
	var name string
	if err := json.Unmarshal(data, &name); err != nil {
		return err
	}

	switch name {
	case "MASK_SHOW_FIRST_4":
		*d = DataMaskingTypeSHOWFIRST4

	case "MASK_SHOW_LAST_4":
		*d = DataMaskingTypeSHOWLAST4

	case "MASK_HASH":
		*d = DataMaskingTypeHASH
	case "MASK_NULL":
		*d = DataMaskingTypeNULLIFY
	case "MASK_REDACT":
		*d = DataMaskingTypeREDACT
	default:
		*d = DataMaskingType{Name: name}
	}

	return nil
}

// MarshalJSON customizes the marshalling of a DataMaskingType to JSON.
func (d DataMaskingType) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.Name)
}
