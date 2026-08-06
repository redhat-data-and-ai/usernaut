package assets

import (
	"encoding/json"
	"errors"
	"fmt"
	"hash/fnv"
	"reflect"
	"strings"
	"time"

	"github.com/atlanhq/atlan-go/atlan"

	"github.com/atlanhq/atlan-go/atlan/model"
	"github.com/atlanhq/atlan-go/atlan/model/structs"
)

// AtlanObject is an interface that all asset types should implement
type AtlanObject interface {
	MarshalJSON() ([]byte, error) // Used for CRUD operation in Assets
	FromJSON(data []byte) error   // Used for Retrieval of an Asset using GUID
}

// SearchAssets Struct to represent assets for searching
type SearchAssets struct {
	Glossary         *AtlasGlossaryFields
	Table            *AtlasTableFields
	Column           *ColumnFields
	Connection       *ConnectionFields
	MaterialisedView *MaterialisedViewFields
	View             *ViewFields
	AccessControl    *AccessControlFields
	Persona          *PersonaFields
	AuthPolicy       *AuthPolicyFields
	Purpose          *PurposeFields
	// Add other assets here
}

type AttributesFields struct {
	TYPENAME              *KeywordTextField
	GUID                  *KeywordField
	CREATED_BY            *KeywordField
	UPDATED_BY            *KeywordField
	STATUS                *KeywordField
	ATLAN_TAGS            *KeywordTextField
	PROPOGATED_ATLAN_TAGS *KeywordTextField
	ASSIGNED_TERMS        *KeywordTextField
	SUPERTYPE_NAMES       *KeywordTextField
	CREATE_TIME           *NumericField
	UPDATE_TIME           *NumericField
	QUALIFIED_NAME        *KeywordTextField
}

type AssetFields struct {
	AttributesFields
	NAME                       *KeywordTextStemmedField
	DISPLAY_NAME               *KeywordTextField
	DESCRIPTION                *KeywordTextField
	USER_DESCRIPTION           *KeywordTextField
	TENET_ID                   *KeywordField
	CERTIFICATE_STATUS         *KeywordTextField
	CERTIFICATE_STATUS_MESSAGE *KeywordField
	CERTIFICATE_UPDATED_BY     *NumericField
	ANNOUNCEMENT_TITLE         *KeywordField
	ANNOUNCEMENT_MESSAGE       *KeywordTextField
	ANNOUNCEMENT_TYPE          *KeywordField
	ANNOUNCEMENT_UPDATED_AT    *NumericField
	ANNOUNCEMENT_UPDATED_BY    *KeywordField
	OWNER_USERS                *KeywordTextField
	ADMIN_USERS                *KeywordField
	VIEWER_USERS               *KeywordField
	VIEWER_GROUPS              *KeywordField
	CONNECTOR_NAME             *KeywordTextField
	CONNECTION_NAME            *KeywordTextField
	CONNECTION_QUALIFIED_NAME  *KeywordTextField
}

type CatalogFields struct {
	AssetFields
	INPUT_TO_PROCESSES        *RelationField
	OUTPUT_FROM_AIRFLOW_TASKS *RelationField
	INPUT_TO_AIRFLOW_TASKS    *RelationField
	OUTPUT_FROM_PROCESSES     *RelationField
}

type SQLFields struct {
	CatalogFields
	QUERY_COUNT             *NumericField
	QUERY_USER_COUNT        *NumericField
	QUERY_USER_MAP          *KeywordField
	QUERY_COUNT_UPDATED_AT  *NumericField
	DATABASE_NAME           *KeywordTextField
	DATABASE_QUALIFIED_NAME *KeywordField
	SCHEMA_NAME             *KeywordTextField
	SCHEMA_QUALIFIED_NAME   *KeywordField
	TABLE_NAME              *KeywordTextField
	TABLE_QUALIFIED_NAME    *KeywordField
	VIEW_NAME               *KeywordTextField
	VIEW_QUALIFIED_NAME     *KeywordField
	IS_PROFILED             *BooleanField
	LAST_PROFILED_AT        *NumericField
	DBT_SOURCES             *RelationField
	SQL_DBT_MODELS          *RelationField
	SQL_DBT_SOURCES         *RelationField
	DBT_MODELS              *RelationField
	DBT_TESTS               *RelationField
}

// AtlasGlossary represents the AtlasGlossary asset
type AtlasGlossaryFields struct {
	AssetFields
	AtlanObject
}

type AtlasTableFields struct {
	SQLFields
	COLUMN_COUNT             *NumericField
	ROW_COUNT                *NumericField
	SIZE_BYTES               *NumericField
	ALIAS                    *KeywordField
	ISTEMPORARY              *BooleanField
	IS_QUERY_PREVIEW         *BooleanField
	QUERY_PREVIEW_CONFIG     *KeywordField
	EXTERNAL_LOCATION        *KeywordField
	EXTERNAL_LOCATION_REGION *KeywordField
	EXTERNAL_LOCATION_FORMAT *KeywordField
	IS_PARTITIONED           *BooleanField
	PARTITION_STRATEGY       *KeywordField
	PARTITION_COUNT          *NumericField
	PARTITION_LIST           *KeywordField
	PARTITIONS               *RelationField
	COLUMNS                  *RelationField
	QUERIES                  *RelationField
	ATLAN_SCHEMA             *RelationField
	DIMENSIONS               *RelationField
}

type ColumnFields struct {
	SQLFields
	DATA_TYPE                          *KeywordTextField
	SUB_DATA_TYPE                      *KeywordField
	RAW_DATA_TYPE_DEFINITION           *KeywordField
	ORDER                              *NumericField
	NESTED_COLUMN_COUNT                *NumericField
	IS_PARTITION                       *BooleanField
	PARTITION_ORDER                    *NumericField
	IS_CLUSTERED                       *BooleanField
	IS_PRIMARY                         *BooleanField
	IS_FOREIGN                         *BooleanField
	IS_INDEXED                         *BooleanField
	IS_SORT                            *BooleanField
	IS_DIST                            *BooleanField
	IS_PINNED                          *BooleanField
	PINNED_BY                          *KeywordField
	PINNED_AT                          *NumericField
	PRECISION                          *NumericField
	DEFAULT_VALUE                      *KeywordField
	IS_NULLABLE                        *BooleanField
	NUMERIC_SCALE                      *NumericField
	MAX_LENGTH                         *NumericField
	VALIDATIONS                        *KeywordField
	PARENT_COLUMN_QUALIFIED_NAME       *KeywordTextField
	PARENT_COLUMN_NAME                 *KeywordTextField
	COLUMN_DISTINCT_VALUES_COUNT       *NumericField
	COLUMN_DISTINCT_VALUES_COUNT_LONG  *NumericField
	COLUMN_HISTOGRAM                   *KeywordField
	COLUMN_MAX                         *NumericField
	COLUMN_MIN                         *NumericField
	COLUMN_MEAN                        *NumericField
	COLUMN_SUM                         *NumericField
	COLUMN_MEDIAN                      *NumericField
	COLUMN_STANDARD_DEVIATION          *NumericField
	COLUMN_UNIQUE_VALUES_COUNT         *NumericField
	COLUMN_UNIQUE_VALUES_COUNT_LONG    *NumericField
	COLUMN_AVERAGE                     *NumericField
	COLUMN_AVERAGE_LENGTH              *NumericField
	COLUMN_DUPLICATE_VALUES_COUNT      *NumericField
	COLUMN_DUPLICATE_VALUES_COUNT_LONG *NumericField
	COLUMN_MAXIMUM_STRING_LENGTH       *NumericField
	COLUMN_MAXS                        *KeywordField
	COLUMN_MINIMUM_STRING_LENGTH       *NumericField
	COLUMN_MINS                        *KeywordField
	COLUMN_MISSING_VALUES_COUNT        *NumericField
	COLUMN_MISSING_VALUES_COUNT_LONG   *NumericField
	COLUMN_MISSING_VALUES_PERCENTAGE   *NumericField
	COLUMN_UNIQUENESS_PERCENTAGE       *NumericField
	COLUMN_VARIANCE                    *NumericField
	COLUMN_TOP_VALUES                  *KeywordField
	COLUMN_DEPTH_LEVEL                 *NumericField
	SNOWFLAKE_DYNAMIC_TABLE            *RelationField
	VIEW                               *RelationField
	NESTED_COLUMNS                     *RelationField
	DATA_QUALITY_METRIC_DIMENSIONS     *RelationField
	DBT_MODEL_COLUMNS                  *RelationField
	TABLE                              *RelationField
	COLUMN_DBT_MODEL_COLUMNS           *RelationField
	MATERIALIZED_VIEW                  *RelationField
	PARENT_COLUMN                      *RelationField
	QUERIES                            *RelationField
	METRIC_TIMESTAMPS                  *RelationField
	FOREIGN_KEY_TO                     *RelationField
	FOREIGN_KEY_FROM                   *RelationField
	DBT_METRICS                        *RelationField
	TABLE_PARTITION                    *RelationField
}

type ConnectionFields struct {
	AssetFields
	CATEGORY                        *KeywordField
	SUB_CATEGORY                    *KeywordField
	HOST                            *KeywordField
	PORT                            *NumericField
	ALLOW_QUERY                     *BooleanField
	ALLOW_QUERY_PREVIEW             *BooleanField
	QUERY_PREVIEW_CONFIG            *KeywordField
	QUERY_CONFIG                    *KeywordField
	CREDENTIAL_STRATEGY             *KeywordField
	PREVIEW_CREDENTIAL_STRATEGY     *KeywordField
	POLICY_STRATEGY                 *KeywordField
	QUERY_USERNAME_STRATEGY         *KeywordField
	ROW_LIMIT                       *NumericField
	QUERY_TIMEOUT                   *NumericField
	DEFAULT_CREDENTIAL_GUID         *KeywordField
	CONNECTOR_ICON                  *KeywordField
	CONNECTOR_IMAGE                 *KeywordField
	SOURCE_LOGO                     *KeywordField
	IS_SAMPLE_DATA_PREVIEW_ENABLED  *BooleanField
	POPULARITY_INSIGHTS_TIMEFRAME   *NumericField
	HAS_POPULARITY_INSIGHTS         *BooleanField
	CONNECTION_DBT_ENVIRONMENTS     *KeywordField
	CONNECTION_SSO_CREDENTIAL_GUID  *KeywordField
	USE_OBJECT_STORAGE              *BooleanField
	OBJECT_STORAGE_UPLOAD_THRESHOLD *NumericField
	VECTOR_EMBEDDINGS_ENABLED       *BooleanField
	VECTOR_EMBEDDINGS_UPDATED_AT    *NumericField
}

type MaterialisedViewFields struct {
	SQLFields
	REFRESH_MODE         *KeywordField
	REFRESH_METHOD       *KeywordField
	STALENESS            *KeywordField
	STALE_SINCE_DATE     *NumericField
	COLUMN_COUNT         *NumericField
	ROW_COUNT            *NumericField
	SIZE_BYTES           *NumericField
	IS_QUERY_PREVIEW     *BooleanField
	QUERY_PREVIEW_CONFIG *KeywordField
	ALIAS                *KeywordField
	IS_TEMPORARY         *BooleanField
	DEFINITION           *KeywordField
	COLUMNS              *RelationField
	ATLAN_SCHEMA         *RelationField
}

type ViewFields struct {
	SQLFields
	COLUMN_COUNT         *NumericField
	ROW_COUNT            *NumericField
	SIZE_BYTES           *NumericField
	IS_QUERY_PREVIEW     *BooleanField
	QUERY_PREVIEW_CONFIG *KeywordField
	ALIAS                *KeywordField
	IS_TEMPORARY         *BooleanField
	DEFINITION           *KeywordField
	COLUMNS              *RelationField
	QUERIES              *RelationField
	ATLAN_SCHEMA         *RelationField
}

type AccessControlFields struct {
	AssetFields
	IS_ACCESS_CONTROL_ENABLED  *BooleanField
	DENY_CUSTOM_METADATA_GUIDS *KeywordField
	DENY_ASSET_TABS            *KeywordField
	DENY_ASSET_FILTERS         *TextField
	CHANNEL_LINK               *TextField
	DENY_ASSET_TYPES           *TextField
	DENY_NAVIGATION_PAGES      *TextField
	DEFAULT_NAVIGATION         *TextField
	DISPLAY_PREFERENCES        *KeywordField
	POLICIES                   *RelationField
}

type PersonaFields struct {
	AccessControlFields
	PERSONA_GROUPS *KeywordField
	PERSONA_USERS  *KeywordField
	ROLE_ID        *KeywordField
}

type AuthPolicyFields struct {
	AssetFields
	POLICY_TYPE               *KeywordField
	POLICY_SERVICE_NAME       *KeywordField
	POLICY_CATEGORY           *KeywordField
	POLICY_SUB_CATEGORY       *KeywordField
	POLICY_USERS              *KeywordField
	POLICY_GROUPS             *KeywordField
	POLICY_ROLES              *KeywordField
	POLICY_ACTIONS            *KeywordField
	POLICY_RESOURCES          *KeywordField
	POLICY_RESOURCE_CATEGORY  *KeywordField
	POLICY_PRIORITY           *NumericField
	IS_POLICY_ENABLED         *BooleanField
	POLICY_MASK_TYPE          *KeywordField
	POLICY_VALIDITY_SCHEDULE  *KeywordField
	POLICY_RESOURCE_SIGNATURE *KeywordField
	POLICY_DELEGATE_ADMIN     *BooleanField
	POLICY_CONDITIONS         *KeywordField
	ACCESS_CONTROL            *RelationField
}

type PurposeFields struct {
	AccessControlFields
	PURPOSE_CLASSIFICATIONS *KeywordField
}

type WorkflowFields struct {
	AssetFields
	WORKFLOW_TEMPLATE_GUID  *KeywordField
	WORKFLOW_TYPE           *KeywordField
	WORKFLOW_CONFIG         *TextField
	WORKFLOW_STATUS         *KeywordField
	WORKFLOW_RUN_EXPIRES_IN *TextField
	WORKFLOW_CREATED_BY     *KeywordField
	WORKFLOW_UPDATED_BY     *KeywordField
	WORKFLOW_DELETED_AT     *NumericField
}

// WorkflowRunFields represents the fields for a workflow run.
type WorkflowRunFields struct {
	AssetFields
	WORKFLOW_RUN_WORKFLOW_GUID *KeywordField
	WORKFLOW_RUN_TYPE          *KeywordField
	WORKFLOW_RUN_ON_ASSET_GUID *KeywordField
	WORKFLOW_RUN_COMMENT       *TextField
	WORKFLOW_RUN_CONFIG        *TextField
	WORKFLOW_RUN_STATUS        *KeywordField
	WORKFLOW_RUN_EXPIRES_AT    *NumericField
	WORKFLOW_RUN_CREATED_BY    *KeywordField
	WORKFLOW_RUN_UPDATED_BY    *KeywordField
	WORKFLOW_RUN_DELETED_AT    *NumericField
}

// NewSearchTable returns a new AtlasTable object for Searching
func NewSearchTable() *AtlasTableFields {
	return &AtlasTableFields{
		COLUMN_COUNT:             NewNumericField("columnCount", "columnCount"),
		ROW_COUNT:                NewNumericField("rowCount", "rowCount"),
		SIZE_BYTES:               NewNumericField("sizeBytes", "sizeBytes"),
		ALIAS:                    NewKeywordField("alias", "alias"),
		ISTEMPORARY:              NewBooleanField("isTemporary", "isTemporary"),
		IS_QUERY_PREVIEW:         NewBooleanField("isQueryPreview", "isQueryPreview"),
		QUERY_PREVIEW_CONFIG:     NewKeywordField("queryPreviewConfig", "queryPreviewConfig"),
		EXTERNAL_LOCATION:        NewKeywordField("externalLocation", "externalLocation"),
		EXTERNAL_LOCATION_REGION: NewKeywordField("externalLocationRegion", "externalLocationRegion"),
		EXTERNAL_LOCATION_FORMAT: NewKeywordField("externalLocationFormat", "externalLocationFormat"),
		IS_PARTITIONED:           NewBooleanField("isPartitioned", "isPartitioned"),
		PARTITION_STRATEGY:       NewKeywordField("partitionStrategy", "partitionStrategy"),
		PARTITION_COUNT:          NewNumericField("partitionCount", "partitionCount"),
		PARTITION_LIST:           NewKeywordField("partitionList", "partitionList"),
		PARTITIONS:               NewRelationField("partitions"),
		COLUMNS:                  NewRelationField("columns"),
		QUERIES:                  NewRelationField("queries"),
		ATLAN_SCHEMA:             NewRelationField("atlanSchema"),
		DIMENSIONS:               NewRelationField("dimensions"),
		SQLFields: SQLFields{
			CatalogFields: CatalogFields{
				AssetFields: AssetFields{
					AttributesFields: AttributesFields{
						TYPENAME:              NewKeywordTextField("typeName", "__typeName.keyword", "__typeName"),
						GUID:                  NewKeywordField("guid", "__guid"),
						CREATED_BY:            NewKeywordField("createdBy", "__createdBy"),
						UPDATED_BY:            NewKeywordField("updatedBy", "__modifiedBy"),
						STATUS:                NewKeywordField("status", "__state"),
						ATLAN_TAGS:            NewKeywordTextField("classificationNames", "__traitNames", "__classificationsText"),
						PROPOGATED_ATLAN_TAGS: NewKeywordTextField("classificationNames", "__propagatedTraitNames", "__classificationsText"),
						ASSIGNED_TERMS:        NewKeywordTextField("meanings", "__meanings", "__meaningsText"),
						SUPERTYPE_NAMES:       NewKeywordTextField("typeName", "__superTypeNames.keyword", "__superTypeNames"),
						CREATE_TIME:           NewNumericField("createTime", "__timestamp"),
						UPDATE_TIME:           NewNumericField("updateTime", "__modificationTimestamp"),
						QUALIFIED_NAME:        NewKeywordTextField("qualifiedName", "qualifiedName", "qualifiedName.text"),
					},
					NAME:                       NewKeywordTextStemmedField("name", "name.keyword", "name", "name"),
					DISPLAY_NAME:               NewKeywordTextField("displayName", "displayName.keyword", "displayName"),
					DESCRIPTION:                NewKeywordTextField("description", "description", "description.text"),
					USER_DESCRIPTION:           NewKeywordTextField("userDescription", "userDescription", "userDescription.text"),
					TENET_ID:                   NewKeywordField("tenetId", "tenetId"),
					CERTIFICATE_STATUS:         NewKeywordTextField("certificateStatus", "certificateStatus", "certificateStatus.text"),
					CERTIFICATE_STATUS_MESSAGE: NewKeywordField("certificateStatusMessage", "certificateStatusMessage"),
					CERTIFICATE_UPDATED_BY:     NewNumericField("certificateUpdatedBy", "certificateUpdatedBy"),
					ANNOUNCEMENT_TITLE:         NewKeywordField("announcementTitle", "announcementTitle"),
					ANNOUNCEMENT_MESSAGE:       NewKeywordTextField("announcementMessage", "announcementMessage", "announcementMessage.text"),
					ANNOUNCEMENT_TYPE:          NewKeywordField("announcementType", "announcementType"),
					ANNOUNCEMENT_UPDATED_AT:    NewNumericField("announcementUpdatedAt", "announcementUpdatedAt"),
					ANNOUNCEMENT_UPDATED_BY:    NewKeywordField("announcementUpdatedBy", "announcementUpdatedBy"),
					OWNER_USERS:                NewKeywordTextField("ownerUsers", "ownerUsers", "ownerUsers.text"),
					ADMIN_USERS:                NewKeywordField("adminUsers", "adminUsers"),
					VIEWER_USERS:               NewKeywordField("viewerUsers", "viewerUsers"),
					VIEWER_GROUPS:              NewKeywordField("viewerGroups", "viewerGroups"),
					CONNECTOR_NAME:             NewKeywordTextField("connectorName", "connectorName", "connectorName.text"),
					CONNECTION_QUALIFIED_NAME:  NewKeywordTextField("connectionQualifiedName", "connectionQualifiedName", "connectionQualifiedName.text"),
				},
				INPUT_TO_PROCESSES:        NewRelationField("inputToProcesses"),
				OUTPUT_FROM_AIRFLOW_TASKS: NewRelationField("outputFromAirflowTasks"),
				INPUT_TO_AIRFLOW_TASKS:    NewRelationField("inputToAirflowTasks"),
				OUTPUT_FROM_PROCESSES:     NewRelationField("outputFromProcesses"),
			},
			QUERY_COUNT:             NewNumericField("queryCount", "queryCount"),
			QUERY_USER_COUNT:        NewNumericField("queryUserCount", "queryUserCount"),
			QUERY_USER_MAP:          NewKeywordField("queryUserMap", "queryUserMap"),
			QUERY_COUNT_UPDATED_AT:  NewNumericField("queryCountUpdatedAt", "queryCountUpdatedAt"),
			DATABASE_NAME:           NewKeywordTextField("databaseName", "databaseName.keyword", "databaseName"),
			DATABASE_QUALIFIED_NAME: NewKeywordField("databaseQualifiedName", "databaseQualifiedName"),
			SCHEMA_NAME:             NewKeywordTextField("schemaName", "schemaName.keyword", "schemaName"),
			SCHEMA_QUALIFIED_NAME:   NewKeywordField("schemaQualifiedName", "schemaQualifiedName"),
			TABLE_NAME:              NewKeywordTextField("tableName", "tableName.keyword", "tableName"),
			TABLE_QUALIFIED_NAME:    NewKeywordField("tableQualifiedName", "tableQualifiedName"),
			VIEW_NAME:               NewKeywordTextField("viewName", "viewName.keyword", "viewName"),
			VIEW_QUALIFIED_NAME:     NewKeywordField("viewQualifiedName", "viewQualifiedName"),
			IS_PROFILED:             NewBooleanField("isProfiled", "isProfiled"),
			LAST_PROFILED_AT:        NewNumericField("lastProfiledAt", "lastProfiledAt"),
			DBT_SOURCES:             NewRelationField("dbtSources"),
			SQL_DBT_MODELS:          NewRelationField("sqlDbtModels"),
			SQL_DBT_SOURCES:         NewRelationField("sqlDBTSources"),
			DBT_MODELS:              NewRelationField("dbtModels"),
			DBT_TESTS:               NewRelationField("dbtTests"),
		},
	}
}

func NewSearchColumn() *ColumnFields {
	return &ColumnFields{
		SQLFields: SQLFields{
			CatalogFields: CatalogFields{
				AssetFields: AssetFields{
					AttributesFields: AttributesFields{
						TYPENAME:              NewKeywordTextField("typeName", "__typeName.keyword", "__typeName"),
						GUID:                  NewKeywordField("guid", "__guid"),
						CREATED_BY:            NewKeywordField("createdBy", "__createdBy"),
						UPDATED_BY:            NewKeywordField("updatedBy", "__modifiedBy"),
						STATUS:                NewKeywordField("status", "__state"),
						ATLAN_TAGS:            NewKeywordTextField("classificationNames", "__traitNames", "__classificationsText"),
						PROPOGATED_ATLAN_TAGS: NewKeywordTextField("classificationNames", "__propagatedTraitNames", "__classificationsText"),
						ASSIGNED_TERMS:        NewKeywordTextField("meanings", "__meanings", "__meaningsText"),
						SUPERTYPE_NAMES:       NewKeywordTextField("typeName", "__superTypeNames.keyword", "__superTypeNames"),
						CREATE_TIME:           NewNumericField("createTime", "__timestamp"),
						UPDATE_TIME:           NewNumericField("updateTime", "__modificationTimestamp"),
						QUALIFIED_NAME:        NewKeywordTextField("qualifiedName", "qualifiedName", "qualifiedName.text"),
					},
					NAME:                       NewKeywordTextStemmedField("name", "name.keyword", "name", "name"),
					DISPLAY_NAME:               NewKeywordTextField("displayName", "displayName.keyword", "displayName"),
					DESCRIPTION:                NewKeywordTextField("description", "description", "description.text"),
					USER_DESCRIPTION:           NewKeywordTextField("userDescription", "userDescription", "userDescription.text"),
					TENET_ID:                   NewKeywordField("tenetId", "tenetId"),
					CERTIFICATE_STATUS:         NewKeywordTextField("certificateStatus", "certificateStatus", "certificateStatus.text"),
					CERTIFICATE_STATUS_MESSAGE: NewKeywordField("certificateStatusMessage", "certificateStatusMessage"),
					CERTIFICATE_UPDATED_BY:     NewNumericField("certificateUpdatedBy", "certificateUpdatedBy"),
					ANNOUNCEMENT_TITLE:         NewKeywordField("announcementTitle", "announcementTitle"),
					ANNOUNCEMENT_MESSAGE:       NewKeywordTextField("announcementMessage", "announcementMessage", "announcementMessage.text"),
					ANNOUNCEMENT_TYPE:          NewKeywordField("announcementType", "announcementType"),
					ANNOUNCEMENT_UPDATED_AT:    NewNumericField("announcementUpdatedAt", "announcementUpdatedAt"),
					ANNOUNCEMENT_UPDATED_BY:    NewKeywordField("announcementUpdatedBy", "announcementUpdatedBy"),
					OWNER_USERS:                NewKeywordTextField("ownerUsers", "ownerUsers", "ownerUsers.text"),
					ADMIN_USERS:                NewKeywordField("adminUsers", "adminUsers"),
					VIEWER_USERS:               NewKeywordField("viewerUsers", "viewerUsers"),
					VIEWER_GROUPS:              NewKeywordField("viewerGroups", "viewerGroups"),
					CONNECTOR_NAME:             NewKeywordTextField("connectorName", "connectorName", "connectorName.text"),
					CONNECTION_QUALIFIED_NAME:  NewKeywordTextField("connectionQualifiedName", "connectionQualifiedName", "connectionQualifiedName.text"),
				},
				INPUT_TO_PROCESSES:        NewRelationField("inputToProcesses"),
				OUTPUT_FROM_AIRFLOW_TASKS: NewRelationField("outputFromAirflowTasks"),
				INPUT_TO_AIRFLOW_TASKS:    NewRelationField("inputToAirflowTasks"),
				OUTPUT_FROM_PROCESSES:     NewRelationField("outputFromProcesses"),
			},
			QUERY_COUNT:             NewNumericField("queryCount", "queryCount"),
			QUERY_USER_COUNT:        NewNumericField("queryUserCount", "queryUserCount"),
			QUERY_USER_MAP:          NewKeywordField("queryUserMap", "queryUserMap"),
			QUERY_COUNT_UPDATED_AT:  NewNumericField("queryCountUpdatedAt", "queryCountUpdatedAt"),
			DATABASE_NAME:           NewKeywordTextField("databaseName", "databaseName.keyword", "databaseName"),
			DATABASE_QUALIFIED_NAME: NewKeywordField("databaseQualifiedName", "databaseQualifiedName"),
			SCHEMA_NAME:             NewKeywordTextField("schemaName", "schemaName.keyword", "schemaName"),
			SCHEMA_QUALIFIED_NAME:   NewKeywordField("schemaQualifiedName", "schemaQualifiedName"),
			TABLE_NAME:              NewKeywordTextField("tableName", "tableName.keyword", "tableName"),
			TABLE_QUALIFIED_NAME:    NewKeywordField("tableQualifiedName", "tableQualifiedName"),
			VIEW_NAME:               NewKeywordTextField("viewName", "viewName.keyword", "viewName"),
			VIEW_QUALIFIED_NAME:     NewKeywordField("viewQualifiedName", "viewQualifiedName"),
			IS_PROFILED:             NewBooleanField("isProfiled", "isProfiled"),
			LAST_PROFILED_AT:        NewNumericField("lastProfiledAt", "lastProfiledAt"),
			DBT_SOURCES:             NewRelationField("dbtSources"),
			SQL_DBT_MODELS:          NewRelationField("sqlDbtModels"),
			SQL_DBT_SOURCES:         NewRelationField("sqlDBTSources"),
			DBT_MODELS:              NewRelationField("dbtModels"),
			DBT_TESTS:               NewRelationField("dbtTests"),
		},
		DATA_TYPE:                          NewKeywordTextField("dataType", "dataType", "dataType.text"),
		SUB_DATA_TYPE:                      NewKeywordField("subDataType", "subDataType"),
		RAW_DATA_TYPE_DEFINITION:           NewKeywordField("rawDataTypeDefinition", "rawDataTypeDefinition"),
		ORDER:                              NewNumericField("order", "order"),
		NESTED_COLUMN_COUNT:                NewNumericField("nestedColumnCount", "nestedColumnCount"),
		IS_PARTITION:                       NewBooleanField("isPartition", "isPartition"),
		PARTITION_ORDER:                    NewNumericField("partitionOrder", "partitionOrder"),
		IS_CLUSTERED:                       NewBooleanField("isClustered", "isClustered"),
		IS_PRIMARY:                         NewBooleanField("isPrimary", "isPrimary"),
		IS_FOREIGN:                         NewBooleanField("isForeign", "isForeign"),
		IS_INDEXED:                         NewBooleanField("isIndexed", "isIndexed"),
		IS_SORT:                            NewBooleanField("isSort", "isSort"),
		IS_DIST:                            NewBooleanField("isDist", "isDist"),
		IS_PINNED:                          NewBooleanField("isPinned", "isPinned"),
		PINNED_BY:                          NewKeywordField("pinnedBy", "pinnedBy"),
		PINNED_AT:                          NewNumericField("pinnedAt", "pinnedAt"),
		PRECISION:                          NewNumericField("precision", "precision"),
		DEFAULT_VALUE:                      NewKeywordField("defaultValue", "defaultValue"),
		IS_NULLABLE:                        NewBooleanField("isNullable", "isNullable"),
		NUMERIC_SCALE:                      NewNumericField("numericScale", "numericScale"),
		MAX_LENGTH:                         NewNumericField("maxLength", "maxLength"),
		VALIDATIONS:                        NewKeywordField("validations", "validations"),
		PARENT_COLUMN_QUALIFIED_NAME:       NewKeywordTextField("parentColumnQualifiedName", "parentColumnQualifiedName", "parentColumnQualifiedName.text"),
		PARENT_COLUMN_NAME:                 NewKeywordTextField("parentColumnName", "parentColumnName.keyword", "parentColumnName"),
		COLUMN_DISTINCT_VALUES_COUNT:       NewNumericField("columnDistinctValuesCount", "columnDistinctValuesCount"),
		COLUMN_DISTINCT_VALUES_COUNT_LONG:  NewNumericField("columnDistinctValuesCountLong", "columnDistinctValuesCountLong"),
		COLUMN_HISTOGRAM:                   NewKeywordField("columnHistogram", "columnHistogram"),
		COLUMN_MAX:                         NewNumericField("columnMax", "columnMax"),
		COLUMN_MIN:                         NewNumericField("columnMin", "columnMin"),
		COLUMN_MEAN:                        NewNumericField("columnMean", "columnMean"),
		COLUMN_SUM:                         NewNumericField("columnSum", "columnSum"),
		COLUMN_MEDIAN:                      NewNumericField("columnMedian", "columnMedian"),
		COLUMN_STANDARD_DEVIATION:          NewNumericField("columnStandardDeviation", "columnStandardDeviation"),
		COLUMN_UNIQUE_VALUES_COUNT:         NewNumericField("columnUniqueValuesCount", "columnUniqueValuesCount"),
		COLUMN_UNIQUE_VALUES_COUNT_LONG:    NewNumericField("columnUniqueValuesCountLong", "columnUniqueValuesCountLong"),
		COLUMN_AVERAGE:                     NewNumericField("columnAverage", "columnAverage"),
		COLUMN_AVERAGE_LENGTH:              NewNumericField("columnAverageLength", "columnAverageLength"),
		COLUMN_DUPLICATE_VALUES_COUNT:      NewNumericField("columnDuplicateValuesCount", "columnDuplicateValuesCount"),
		COLUMN_DUPLICATE_VALUES_COUNT_LONG: NewNumericField("columnDuplicateValuesCountLong", "columnDuplicateValuesCountLong"),
		COLUMN_MAXIMUM_STRING_LENGTH:       NewNumericField("columnMaximumStringLength", "columnMaximumStringLength"),
		COLUMN_MAXS:                        NewKeywordField("columnMaxs", "columnMaxs"),
		COLUMN_MINIMUM_STRING_LENGTH:       NewNumericField("columnMinimumStringLength", "columnMinimumStringLength"),
		COLUMN_MINS:                        NewKeywordField("columnMins", "columnMins"),
		COLUMN_MISSING_VALUES_COUNT:        NewNumericField("columnMissingValuesCount", "columnMissingValuesCount"),
		COLUMN_MISSING_VALUES_COUNT_LONG:   NewNumericField("columnMissingValuesCountLong", "columnMissingValuesCountLong"),
		COLUMN_MISSING_VALUES_PERCENTAGE:   NewNumericField("columnMissingValuesPercentage", "columnMissingValuesPercentage"),
		COLUMN_UNIQUENESS_PERCENTAGE:       NewNumericField("columnUniquenessPercentage", "columnUniquenessPercentage"),
		COLUMN_VARIANCE:                    NewNumericField("columnVariance", "columnVariance"),
		COLUMN_TOP_VALUES:                  NewKeywordField("columnTopValues", "columnTopValues"),
		COLUMN_DEPTH_LEVEL:                 NewNumericField("columnDepthLevel", "columnDepthLevel"),
		SNOWFLAKE_DYNAMIC_TABLE:            NewRelationField("snowflakeDynamicTable"),
		VIEW:                               NewRelationField("view"),
		NESTED_COLUMNS:                     NewRelationField("nestedColumns"),
		DATA_QUALITY_METRIC_DIMENSIONS:     NewRelationField("dataQualityMetricDimensions"),
		DBT_MODEL_COLUMNS:                  NewRelationField("dbtModelColumns"),
		TABLE:                              NewRelationField("table"),
		COLUMN_DBT_MODEL_COLUMNS:           NewRelationField("columnDbtModelColumns"),
		MATERIALIZED_VIEW:                  NewRelationField("materialisedView"),
		PARENT_COLUMN:                      NewRelationField("parentColumn"),
		QUERIES:                            NewRelationField("queries"),
		METRIC_TIMESTAMPS:                  NewRelationField("metricTimestamps"),
		FOREIGN_KEY_TO:                     NewRelationField("foreignKeyTo"),
		FOREIGN_KEY_FROM:                   NewRelationField("foreignKeyFrom"),
		DBT_METRICS:                        NewRelationField("dbtMetrics"),
		TABLE_PARTITION:                    NewRelationField("tablePartition"),
	}
}

func NewSearchConnection() *ConnectionFields {
	return &ConnectionFields{
		AssetFields: AssetFields{
			AttributesFields: AttributesFields{
				TYPENAME:              NewKeywordTextField("typeName", "__typeName.keyword", "__typeName"),
				GUID:                  NewKeywordField("guid", "__guid"),
				CREATED_BY:            NewKeywordField("createdBy", "__createdBy"),
				UPDATED_BY:            NewKeywordField("updatedBy", "__modifiedBy"),
				STATUS:                NewKeywordField("status", "__state"),
				ATLAN_TAGS:            NewKeywordTextField("classificationNames", "__traitNames", "__classificationsText"),
				PROPOGATED_ATLAN_TAGS: NewKeywordTextField("classificationNames", "__propagatedTraitNames", "__classificationsText"),
				ASSIGNED_TERMS:        NewKeywordTextField("meanings", "__meanings", "__meaningsText"),
				SUPERTYPE_NAMES:       NewKeywordTextField("typeName", "__superTypeNames.keyword", "__superTypeNames"),
				CREATE_TIME:           NewNumericField("createTime", "__timestamp"),
				UPDATE_TIME:           NewNumericField("updateTime", "__modificationTimestamp"),
				QUALIFIED_NAME:        NewKeywordTextField("qualifiedName", "qualifiedName", "qualifiedName.text"),
			},
			NAME:                       NewKeywordTextStemmedField("name", "name.keyword", "name", "name"),
			DISPLAY_NAME:               NewKeywordTextField("displayName", "displayName.keyword", "displayName"),
			DESCRIPTION:                NewKeywordTextField("description", "description", "description.text"),
			USER_DESCRIPTION:           NewKeywordTextField("userDescription", "userDescription", "userDescription.text"),
			TENET_ID:                   NewKeywordField("tenetId", "tenetId"),
			CERTIFICATE_STATUS:         NewKeywordTextField("certificateStatus", "certificateStatus", "certificateStatus.text"),
			CERTIFICATE_STATUS_MESSAGE: NewKeywordField("certificateStatusMessage", "certificateStatusMessage"),
			CERTIFICATE_UPDATED_BY:     NewNumericField("certificateUpdatedBy", "certificateUpdatedBy"),
			ANNOUNCEMENT_TITLE:         NewKeywordField("announcementTitle", "announcementTitle"),
			ANNOUNCEMENT_MESSAGE:       NewKeywordTextField("announcementMessage", "announcementMessage", "announcementMessage.text"),
			ANNOUNCEMENT_TYPE:          NewKeywordField("announcementType", "announcementType"),
			ANNOUNCEMENT_UPDATED_AT:    NewNumericField("announcementUpdatedAt", "announcementUpdatedAt"),
			ANNOUNCEMENT_UPDATED_BY:    NewKeywordField("announcementUpdatedBy", "announcementUpdatedBy"),
			OWNER_USERS:                NewKeywordTextField("ownerUsers", "ownerUsers", "ownerUsers.text"),
			ADMIN_USERS:                NewKeywordField("adminUsers", "adminUsers"),
			VIEWER_USERS:               NewKeywordField("viewerUsers", "viewerUsers"),
			VIEWER_GROUPS:              NewKeywordField("viewerGroups", "viewerGroups"),
			CONNECTOR_NAME:             NewKeywordTextField("connectorName", "connectorName", "connectorName.text"),
			CONNECTION_QUALIFIED_NAME:  NewKeywordTextField("connectionQualifiedName", "connectionQualifiedName", "connectionQualifiedName.text"),
		},
		CATEGORY:                        NewKeywordField("category", "category"),
		SUB_CATEGORY:                    NewKeywordField("subCategory", "subCategory"),
		HOST:                            NewKeywordField("host", "host"),
		PORT:                            NewNumericField("port", "port"),
		ALLOW_QUERY:                     NewBooleanField("allowQuery", "allowQuery"),
		ALLOW_QUERY_PREVIEW:             NewBooleanField("allowQueryPreview", "allowQueryPreview"),
		QUERY_PREVIEW_CONFIG:            NewKeywordField("queryPreviewConfig", "queryPreviewConfig"),
		QUERY_CONFIG:                    NewKeywordField("queryConfig", "queryConfig"),
		CREDENTIAL_STRATEGY:             NewKeywordField("credentialStrategy", "credentialStrategy"),
		PREVIEW_CREDENTIAL_STRATEGY:     NewKeywordField("previewCredentialStrategy", "previewCredentialStrategy"),
		POLICY_STRATEGY:                 NewKeywordField("policyStrategy", "policyStrategy"),
		QUERY_USERNAME_STRATEGY:         NewKeywordField("queryUsernameStrategy", "queryUsernameStrategy"),
		ROW_LIMIT:                       NewNumericField("rowLimit", "rowLimit"),
		QUERY_TIMEOUT:                   NewNumericField("queryTimeout", "queryTimeout"),
		DEFAULT_CREDENTIAL_GUID:         NewKeywordField("defaultCredentialGuid", "defaultCredentialGuid"),
		CONNECTOR_ICON:                  NewKeywordField("connectorIcon", "connectorIcon"),
		CONNECTOR_IMAGE:                 NewKeywordField("connectorImage", "connectorImage"),
		SOURCE_LOGO:                     NewKeywordField("sourceLogo", "sourceLogo"),
		IS_SAMPLE_DATA_PREVIEW_ENABLED:  NewBooleanField("isSampleDataPreviewEnabled", "isSampleDataPreviewEnabled"),
		POPULARITY_INSIGHTS_TIMEFRAME:   NewNumericField("popularityInsightsTimeframe", "popularityInsightsTimeframe"),
		HAS_POPULARITY_INSIGHTS:         NewBooleanField("hasPopularityInsights", "hasPopularityInsights"),
		CONNECTION_DBT_ENVIRONMENTS:     NewKeywordField("connectionDbtEnvironments", "connectionDbtEnvironments"),
		CONNECTION_SSO_CREDENTIAL_GUID:  NewKeywordField("connectionSSOCredentialGuid", "connectionSSOCredentialGuid"),
		USE_OBJECT_STORAGE:              NewBooleanField("useObjectStorage", "useObjectStorage"),
		OBJECT_STORAGE_UPLOAD_THRESHOLD: NewNumericField("objectStorageUploadThreshold", "objectStorageUploadThreshold"),
		VECTOR_EMBEDDINGS_ENABLED:       NewBooleanField("vectorEmbeddingsEnabled", "vectorEmbeddingsEnabled"),
		VECTOR_EMBEDDINGS_UPDATED_AT:    NewNumericField("vectorEmbeddingsUpdatedAt", "vectorEmbeddingsUpdatedAt"),
	}
}

// NewSearchGlossary returns a new AtlasGlossary object for Searching
func NewSearchGlossary() *AtlasGlossaryFields {
	return &AtlasGlossaryFields{
		AssetFields: AssetFields{
			AttributesFields: AttributesFields{
				TYPENAME:              NewKeywordTextField("typeName", "__typeName.keyword", "__typeName"),
				GUID:                  NewKeywordField("guid", "__guid"),
				CREATED_BY:            NewKeywordField("createdBy", "__createdBy"),
				UPDATED_BY:            NewKeywordField("updatedBy", "__modifiedBy"),
				STATUS:                NewKeywordField("status", "__state"),
				ATLAN_TAGS:            NewKeywordTextField("classificationNames", "__traitNames", "__classificationsText"),
				PROPOGATED_ATLAN_TAGS: NewKeywordTextField("classificationNames", "__propagatedTraitNames", "__classificationsText"),
				ASSIGNED_TERMS:        NewKeywordTextField("meanings", "__meanings", "__meaningsText"),
				SUPERTYPE_NAMES:       NewKeywordTextField("typeName", "__superTypeNames.keyword", "__superTypeNames"),
				CREATE_TIME:           NewNumericField("createTime", "__timestamp"),
				UPDATE_TIME:           NewNumericField("updateTime", "__modificationTimestamp"),
				QUALIFIED_NAME:        NewKeywordTextField("qualifiedName", "qualifiedName", "qualifiedName.text"),
			},
			NAME:                       NewKeywordTextStemmedField("name", "name.keyword", "name", "name"),
			DISPLAY_NAME:               NewKeywordTextField("displayName", "displayName.keyword", "displayName"),
			DESCRIPTION:                NewKeywordTextField("description", "description", "description.text"),
			USER_DESCRIPTION:           NewKeywordTextField("userDescription", "userDescription", "userDescription.text"),
			TENET_ID:                   NewKeywordField("tenetId", "tenetId"),
			CERTIFICATE_STATUS:         NewKeywordTextField("certificateStatus", "certificateStatus", "certificateStatus.text"),
			CERTIFICATE_STATUS_MESSAGE: NewKeywordField("certificateStatusMessage", "certificateStatusMessage"),
			CERTIFICATE_UPDATED_BY:     NewNumericField("certificateUpdatedBy", "certificateUpdatedBy"),
			ANNOUNCEMENT_TITLE:         NewKeywordField("announcementTitle", "announcementTitle"),
			ANNOUNCEMENT_MESSAGE:       NewKeywordTextField("announcementMessage", "announcementMessage", "announcementMessage.text"),
			ANNOUNCEMENT_TYPE:          NewKeywordField("announcementType", "announcementType"),
			ANNOUNCEMENT_UPDATED_AT:    NewNumericField("announcementUpdatedAt", "announcementUpdatedAt"),
			ANNOUNCEMENT_UPDATED_BY:    NewKeywordField("announcementUpdatedBy", "announcementUpdatedBy"),
			OWNER_USERS:                NewKeywordTextField("ownerUsers", "ownerUsers", "ownerUsers.text"),
			ADMIN_USERS:                NewKeywordField("adminUsers", "adminUsers"),
			VIEWER_USERS:               NewKeywordField("viewerUsers", "viewerUsers"),
			VIEWER_GROUPS:              NewKeywordField("viewerGroups", "viewerGroups"),
			CONNECTOR_NAME:             NewKeywordTextField("connectorName", "connectorName", "connectorName.text"),
			CONNECTION_QUALIFIED_NAME:  NewKeywordTextField("connectionQualifiedName", "connectionQualifiedName", "connectionQualifiedName.text"),
		},
	}
}

func NewSearchMaterialisedView() *MaterialisedViewFields {
	return &MaterialisedViewFields{
		SQLFields: SQLFields{
			CatalogFields: CatalogFields{
				AssetFields: AssetFields{
					AttributesFields: AttributesFields{
						TYPENAME:              NewKeywordTextField("typeName", "__typeName.keyword", "__typeName"),
						GUID:                  NewKeywordField("guid", "__guid"),
						CREATED_BY:            NewKeywordField("createdBy", "__createdBy"),
						UPDATED_BY:            NewKeywordField("updatedBy", "__modifiedBy"),
						STATUS:                NewKeywordField("status", "__state"),
						ATLAN_TAGS:            NewKeywordTextField("classificationNames", "__traitNames", "__classificationsText"),
						PROPOGATED_ATLAN_TAGS: NewKeywordTextField("classificationNames", "__propagatedTraitNames", "__classificationsText"),
						ASSIGNED_TERMS:        NewKeywordTextField("meanings", "__meanings", "__meaningsText"),
						SUPERTYPE_NAMES:       NewKeywordTextField("typeName", "__superTypeNames.keyword", "__superTypeNames"),
						CREATE_TIME:           NewNumericField("createTime", "__timestamp"),
						UPDATE_TIME:           NewNumericField("updateTime", "__modificationTimestamp"),
						QUALIFIED_NAME:        NewKeywordTextField("qualifiedName", "qualifiedName", "qualifiedName.text"),
					},
					NAME:                       NewKeywordTextStemmedField("name", "name.keyword", "name", "name"),
					DISPLAY_NAME:               NewKeywordTextField("displayName", "displayName.keyword", "displayName"),
					DESCRIPTION:                NewKeywordTextField("description", "description", "description.text"),
					USER_DESCRIPTION:           NewKeywordTextField("userDescription", "userDescription", "userDescription.text"),
					TENET_ID:                   NewKeywordField("tenetId", "tenetId"),
					CERTIFICATE_STATUS:         NewKeywordTextField("certificateStatus", "certificateStatus", "certificateStatus.text"),
					CERTIFICATE_STATUS_MESSAGE: NewKeywordField("certificateStatusMessage", "certificateStatusMessage"),
					CERTIFICATE_UPDATED_BY:     NewNumericField("certificateUpdatedBy", "certificateUpdatedBy"),
					ANNOUNCEMENT_TITLE:         NewKeywordField("announcementTitle", "announcementTitle"),
					ANNOUNCEMENT_MESSAGE:       NewKeywordTextField("announcementMessage", "announcementMessage", "announcementMessage.text"),
					ANNOUNCEMENT_TYPE:          NewKeywordField("announcementType", "announcementType"),
					ANNOUNCEMENT_UPDATED_AT:    NewNumericField("announcementUpdatedAt", "announcementUpdatedAt"),
					ANNOUNCEMENT_UPDATED_BY:    NewKeywordField("announcementUpdatedBy", "announcementUpdatedBy"),
					OWNER_USERS:                NewKeywordTextField("ownerUsers", "ownerUsers", "ownerUsers.text"),
					ADMIN_USERS:                NewKeywordField("adminUsers", "adminUsers"),
					VIEWER_USERS:               NewKeywordField("viewerUsers", "viewerUsers"),
					VIEWER_GROUPS:              NewKeywordField("viewerGroups", "viewerGroups"),
					CONNECTOR_NAME:             NewKeywordTextField("connectorName", "connectorName", "connectorName.text"),
					CONNECTION_QUALIFIED_NAME:  NewKeywordTextField("connectionQualifiedName", "connectionQualifiedName", "connectionQualifiedName.text"),
				},
				INPUT_TO_PROCESSES:        NewRelationField("inputToProcesses"),
				OUTPUT_FROM_AIRFLOW_TASKS: NewRelationField("outputFromAirflowTasks"),
				INPUT_TO_AIRFLOW_TASKS:    NewRelationField("inputToAirflowTasks"),
				OUTPUT_FROM_PROCESSES:     NewRelationField("outputFromProcesses"),
			},
			QUERY_COUNT:             NewNumericField("queryCount", "queryCount"),
			QUERY_USER_COUNT:        NewNumericField("queryUserCount", "queryUserCount"),
			QUERY_USER_MAP:          NewKeywordField("queryUserMap", "queryUserMap"),
			QUERY_COUNT_UPDATED_AT:  NewNumericField("queryCountUpdatedAt", "queryCountUpdatedAt"),
			DATABASE_NAME:           NewKeywordTextField("databaseName", "databaseName.keyword", "databaseName"),
			DATABASE_QUALIFIED_NAME: NewKeywordField("databaseQualifiedName", "databaseQualifiedName"),
			SCHEMA_NAME:             NewKeywordTextField("schemaName", "schemaName.keyword", "schemaName"),
			SCHEMA_QUALIFIED_NAME:   NewKeywordField("schemaQualifiedName", "schemaQualifiedName"),
			TABLE_NAME:              NewKeywordTextField("tableName", "tableName.keyword", "tableName"),
			TABLE_QUALIFIED_NAME:    NewKeywordField("tableQualifiedName", "tableQualifiedName"),
			VIEW_NAME:               NewKeywordTextField("viewName", "viewName.keyword", "viewName"),
			VIEW_QUALIFIED_NAME:     NewKeywordField("viewQualifiedName", "viewQualifiedName"),
			IS_PROFILED:             NewBooleanField("isProfiled", "isProfiled"),
			LAST_PROFILED_AT:        NewNumericField("lastProfiledAt", "lastProfiledAt"),
			DBT_SOURCES:             NewRelationField("dbtSources"),
			SQL_DBT_MODELS:          NewRelationField("sqlDbtModels"),
			SQL_DBT_SOURCES:         NewRelationField("sqlDBTSources"),
			DBT_MODELS:              NewRelationField("dbtModels"),
			DBT_TESTS:               NewRelationField("dbtTests"),
		},
		REFRESH_MODE:         NewKeywordField("refreshMode", "refreshMode"),
		REFRESH_METHOD:       NewKeywordField("refreshMethod", "refreshMethod"),
		STALENESS:            NewKeywordField("staleness", "staleness"),
		STALE_SINCE_DATE:     NewNumericField("staleSinceDate", "staleSinceDate"),
		COLUMN_COUNT:         NewNumericField("columnCount", "columnCount"),
		ROW_COUNT:            NewNumericField("rowCount", "rowCount"),
		SIZE_BYTES:           NewNumericField("sizeBytes", "sizeBytes"),
		IS_QUERY_PREVIEW:     NewBooleanField("isQueryPreview", "isQueryPreview"),
		QUERY_PREVIEW_CONFIG: NewKeywordField("queryPreviewConfig", "queryPreviewConfig"),
		ALIAS:                NewKeywordField("alias", "alias"),
		IS_TEMPORARY:         NewBooleanField("isTemporary", "isTemporary"),
		DEFINITION:           NewKeywordField("definition", "definition"),
		COLUMNS:              NewRelationField("columns"),
		ATLAN_SCHEMA:         NewRelationField("atlanSchema"),
	}
}

func NewSearchView() *ViewFields {
	return &ViewFields{
		SQLFields: SQLFields{
			CatalogFields: CatalogFields{
				AssetFields: AssetFields{
					AttributesFields: AttributesFields{
						TYPENAME:              NewKeywordTextField("typeName", "__typeName.keyword", "__typeName"),
						GUID:                  NewKeywordField("guid", "__guid"),
						CREATED_BY:            NewKeywordField("createdBy", "__createdBy"),
						UPDATED_BY:            NewKeywordField("updatedBy", "__modifiedBy"),
						STATUS:                NewKeywordField("status", "__state"),
						ATLAN_TAGS:            NewKeywordTextField("classificationNames", "__traitNames", "__classificationsText"),
						PROPOGATED_ATLAN_TAGS: NewKeywordTextField("classificationNames", "__propagatedTraitNames", "__classificationsText"),
						ASSIGNED_TERMS:        NewKeywordTextField("meanings", "__meanings", "__meaningsText"),
						SUPERTYPE_NAMES:       NewKeywordTextField("typeName", "__superTypeNames.keyword", "__superTypeNames"),
						CREATE_TIME:           NewNumericField("createTime", "__timestamp"),
						UPDATE_TIME:           NewNumericField("updateTime", "__modificationTimestamp"),
						QUALIFIED_NAME:        NewKeywordTextField("qualifiedName", "qualifiedName", "qualifiedName.text"),
					},
					NAME:                       NewKeywordTextStemmedField("name", "name.keyword", "name", "name"),
					DISPLAY_NAME:               NewKeywordTextField("displayName", "displayName.keyword", "displayName"),
					DESCRIPTION:                NewKeywordTextField("description", "description", "description.text"),
					USER_DESCRIPTION:           NewKeywordTextField("userDescription", "userDescription", "userDescription.text"),
					TENET_ID:                   NewKeywordField("tenetId", "tenetId"),
					CERTIFICATE_STATUS:         NewKeywordTextField("certificateStatus", "certificateStatus", "certificateStatus.text"),
					CERTIFICATE_STATUS_MESSAGE: NewKeywordField("certificateStatusMessage", "certificateStatusMessage"),
					CERTIFICATE_UPDATED_BY:     NewNumericField("certificateUpdatedBy", "certificateUpdatedBy"),
					ANNOUNCEMENT_TITLE:         NewKeywordField("announcementTitle", "announcementTitle"),
					ANNOUNCEMENT_MESSAGE:       NewKeywordTextField("announcementMessage", "announcementMessage", "announcementMessage.text"),
					ANNOUNCEMENT_TYPE:          NewKeywordField("announcementType", "announcementType"),
					ANNOUNCEMENT_UPDATED_AT:    NewNumericField("announcementUpdatedAt", "announcementUpdatedAt"),
					ANNOUNCEMENT_UPDATED_BY:    NewKeywordField("announcementUpdatedBy", "announcementUpdatedBy"),
					OWNER_USERS:                NewKeywordTextField("ownerUsers", "ownerUsers", "ownerUsers.text"),
					ADMIN_USERS:                NewKeywordField("adminUsers", "adminUsers"),
					VIEWER_USERS:               NewKeywordField("viewerUsers", "viewerUsers"),
					VIEWER_GROUPS:              NewKeywordField("viewerGroups", "viewerGroups"),
					CONNECTOR_NAME:             NewKeywordTextField("connectorName", "connectorName", "connectorName.text"),
					CONNECTION_QUALIFIED_NAME:  NewKeywordTextField("connectionQualifiedName", "connectionQualifiedName", "connectionQualifiedName.text"),
				},
				INPUT_TO_PROCESSES:        NewRelationField("inputToProcesses"),
				OUTPUT_FROM_AIRFLOW_TASKS: NewRelationField("outputFromAirflowTasks"),
				INPUT_TO_AIRFLOW_TASKS:    NewRelationField("inputToAirflowTasks"),
				OUTPUT_FROM_PROCESSES:     NewRelationField("outputFromProcesses"),
			},
			QUERY_COUNT:             NewNumericField("queryCount", "queryCount"),
			QUERY_USER_COUNT:        NewNumericField("queryUserCount", "queryUserCount"),
			QUERY_USER_MAP:          NewKeywordField("queryUserMap", "queryUserMap"),
			QUERY_COUNT_UPDATED_AT:  NewNumericField("queryCountUpdatedAt", "queryCountUpdatedAt"),
			DATABASE_NAME:           NewKeywordTextField("databaseName", "databaseName.keyword", "databaseName"),
			DATABASE_QUALIFIED_NAME: NewKeywordField("databaseQualifiedName", "databaseQualifiedName"),
			SCHEMA_NAME:             NewKeywordTextField("schemaName", "schemaName.keyword", "schemaName"),
			SCHEMA_QUALIFIED_NAME:   NewKeywordField("schemaQualifiedName", "schemaQualifiedName"),
			TABLE_NAME:              NewKeywordTextField("tableName", "tableName.keyword", "tableName"),
			TABLE_QUALIFIED_NAME:    NewKeywordField("tableQualifiedName", "tableQualifiedName"),
			VIEW_NAME:               NewKeywordTextField("viewName", "viewName.keyword", "viewName"),
			VIEW_QUALIFIED_NAME:     NewKeywordField("viewQualifiedName", "viewQualifiedName"),
			IS_PROFILED:             NewBooleanField("isProfiled", "isProfiled"),
			LAST_PROFILED_AT:        NewNumericField("lastProfiledAt", "lastProfiledAt"),
			DBT_SOURCES:             NewRelationField("dbtSources"),
			SQL_DBT_MODELS:          NewRelationField("sqlDbtModels"),
			SQL_DBT_SOURCES:         NewRelationField("sqlDBTSources"),
			DBT_MODELS:              NewRelationField("dbtModels"),
			DBT_TESTS:               NewRelationField("dbtTests"),
		},
		COLUMN_COUNT:         NewNumericField("columnCount", "columnCount"),
		ROW_COUNT:            NewNumericField("rowCount", "rowCount"),
		SIZE_BYTES:           NewNumericField("sizeBytes", "sizeBytes"),
		IS_QUERY_PREVIEW:     NewBooleanField("isQueryPreview", "isQueryPreview"),
		QUERY_PREVIEW_CONFIG: NewKeywordField("queryPreviewConfig", "queryPreviewConfig"),
		ALIAS:                NewKeywordField("alias", "alias"),
		IS_TEMPORARY:         NewBooleanField("isTemporary", "isTemporary"),
		DEFINITION:           NewKeywordField("definition", "definition"),
		COLUMNS:              NewRelationField("columns"),
		QUERIES:              NewRelationField("queries"),
		ATLAN_SCHEMA:         NewRelationField("atlanSchema"),
	}
}

// NewAccessControlFields initializes a new instance of AccessControlFields.
func NewAccessControlFields() *AccessControlFields {
	return &AccessControlFields{
		IS_ACCESS_CONTROL_ENABLED:  NewBooleanField("isAccessControlEnabled", "isAccessControlEnabled"),
		DENY_CUSTOM_METADATA_GUIDS: NewKeywordField("denyCustomMetadataGuids", "denyCustomMetadataGuids"),
		DENY_ASSET_TABS:            NewKeywordField("denyAssetTabs", "denyAssetTabs"),
		DENY_ASSET_FILTERS:         NewTextField("denyAssetFilters", "denyAssetFilters"),
		CHANNEL_LINK:               NewTextField("channelLink", "channelLink"),
		DENY_ASSET_TYPES:           NewTextField("denyAssetTypes", "denyAssetTypes"),
		DENY_NAVIGATION_PAGES:      NewTextField("denyNavigationPages", "denyNavigationPages"),
		DEFAULT_NAVIGATION:         NewTextField("defaultNavigation", "defaultNavigation"),
		DISPLAY_PREFERENCES:        NewKeywordField("displayPreferences", "displayPreferences"),
		POLICIES:                   NewRelationField("policies"),
		AssetFields: AssetFields{
			AttributesFields: AttributesFields{
				TYPENAME:              NewKeywordTextField("typeName", "__typeName.keyword", "__typeName"),
				GUID:                  NewKeywordField("guid", "__guid"),
				CREATED_BY:            NewKeywordField("createdBy", "__createdBy"),
				UPDATED_BY:            NewKeywordField("updatedBy", "__modifiedBy"),
				STATUS:                NewKeywordField("status", "__state"),
				ATLAN_TAGS:            NewKeywordTextField("classificationNames", "__traitNames", "__classificationsText"),
				PROPOGATED_ATLAN_TAGS: NewKeywordTextField("classificationNames", "__propagatedTraitNames", "__classificationsText"),
				ASSIGNED_TERMS:        NewKeywordTextField("meanings", "__meanings", "__meaningsText"),
				SUPERTYPE_NAMES:       NewKeywordTextField("typeName", "__superTypeNames.keyword", "__superTypeNames"),
				CREATE_TIME:           NewNumericField("createTime", "__timestamp"),
				UPDATE_TIME:           NewNumericField("updateTime", "__modificationTimestamp"),
				QUALIFIED_NAME:        NewKeywordTextField("qualifiedName", "qualifiedName", "qualifiedName.text"),
			},
			NAME:                       NewKeywordTextStemmedField("name", "name.keyword", "name", "name"),
			DISPLAY_NAME:               NewKeywordTextField("displayName", "displayName.keyword", "displayName"),
			DESCRIPTION:                NewKeywordTextField("description", "description", "description.text"),
			USER_DESCRIPTION:           NewKeywordTextField("userDescription", "userDescription", "userDescription.text"),
			TENET_ID:                   NewKeywordField("tenetId", "tenetId"),
			CERTIFICATE_STATUS:         NewKeywordTextField("certificateStatus", "certificateStatus", "certificateStatus.text"),
			CERTIFICATE_STATUS_MESSAGE: NewKeywordField("certificateStatusMessage", "certificateStatusMessage"),
			CERTIFICATE_UPDATED_BY:     NewNumericField("certificateUpdatedBy", "certificateUpdatedBy"),
			ANNOUNCEMENT_TITLE:         NewKeywordField("announcementTitle", "announcementTitle"),
			ANNOUNCEMENT_MESSAGE:       NewKeywordTextField("announcementMessage", "announcementMessage", "announcementMessage.text"),
			ANNOUNCEMENT_TYPE:          NewKeywordField("announcementType", "announcementType"),
			ANNOUNCEMENT_UPDATED_AT:    NewNumericField("announcementUpdatedAt", "announcementUpdatedAt"),
			ANNOUNCEMENT_UPDATED_BY:    NewKeywordField("announcementUpdatedBy", "announcementUpdatedBy"),
			OWNER_USERS:                NewKeywordTextField("ownerUsers", "ownerUsers", "ownerUsers.text"),
			ADMIN_USERS:                NewKeywordField("adminUsers", "adminUsers"),
			VIEWER_USERS:               NewKeywordField("viewerUsers", "viewerUsers"),
			VIEWER_GROUPS:              NewKeywordField("viewerGroups", "viewerGroups"),
			CONNECTOR_NAME:             NewKeywordTextField("connectorName", "connectorName", "connectorName.text"),
			CONNECTION_QUALIFIED_NAME:  NewKeywordTextField("connectionQualifiedName", "connectionQualifiedName", "connectionQualifiedName.text"),
		},
	}
}

// NewPersonaFields initializes a new instance of PersonaFields.
func NewPersonaFields() *PersonaFields {
	return &PersonaFields{
		AccessControlFields: AccessControlFields{
			IS_ACCESS_CONTROL_ENABLED:  NewBooleanField("isAccessControlEnabled", "isAccessControlEnabled"),
			DENY_CUSTOM_METADATA_GUIDS: NewKeywordField("denyCustomMetadataGuids", "denyCustomMetadataGuids"),
			DENY_ASSET_TABS:            NewKeywordField("denyAssetTabs", "denyAssetTabs"),
			DENY_ASSET_FILTERS:         NewTextField("denyAssetFilters", "denyAssetFilters"),
			CHANNEL_LINK:               NewTextField("channelLink", "channelLink"),
			DENY_ASSET_TYPES:           NewTextField("denyAssetTypes", "denyAssetTypes"),
			DENY_NAVIGATION_PAGES:      NewTextField("denyNavigationPages", "denyNavigationPages"),
			DEFAULT_NAVIGATION:         NewTextField("defaultNavigation", "defaultNavigation"),
			DISPLAY_PREFERENCES:        NewKeywordField("displayPreferences", "displayPreferences"),
			POLICIES:                   NewRelationField("policies"),
			AssetFields: AssetFields{
				AttributesFields: AttributesFields{
					TYPENAME:              NewKeywordTextField("typeName", "__typeName.keyword", "__typeName"),
					GUID:                  NewKeywordField("guid", "__guid"),
					CREATED_BY:            NewKeywordField("createdBy", "__createdBy"),
					UPDATED_BY:            NewKeywordField("updatedBy", "__modifiedBy"),
					STATUS:                NewKeywordField("status", "__state"),
					ATLAN_TAGS:            NewKeywordTextField("classificationNames", "__traitNames", "__classificationsText"),
					PROPOGATED_ATLAN_TAGS: NewKeywordTextField("classificationNames", "__propagatedTraitNames", "__classificationsText"),
					ASSIGNED_TERMS:        NewKeywordTextField("meanings", "__meanings", "__meaningsText"),
					SUPERTYPE_NAMES:       NewKeywordTextField("typeName", "__superTypeNames.keyword", "__superTypeNames"),
					CREATE_TIME:           NewNumericField("createTime", "__timestamp"),
					UPDATE_TIME:           NewNumericField("updateTime", "__modificationTimestamp"),
					QUALIFIED_NAME:        NewKeywordTextField("qualifiedName", "qualifiedName", "qualifiedName.text"),
				},
				NAME:                       NewKeywordTextStemmedField("name", "name.keyword", "name", "name"),
				DISPLAY_NAME:               NewKeywordTextField("displayName", "displayName.keyword", "displayName"),
				DESCRIPTION:                NewKeywordTextField("description", "description", "description.text"),
				USER_DESCRIPTION:           NewKeywordTextField("userDescription", "userDescription", "userDescription.text"),
				TENET_ID:                   NewKeywordField("tenetId", "tenetId"),
				CERTIFICATE_STATUS:         NewKeywordTextField("certificateStatus", "certificateStatus", "certificateStatus.text"),
				CERTIFICATE_STATUS_MESSAGE: NewKeywordField("certificateStatusMessage", "certificateStatusMessage"),
				CERTIFICATE_UPDATED_BY:     NewNumericField("certificateUpdatedBy", "certificateUpdatedBy"),
				ANNOUNCEMENT_TITLE:         NewKeywordField("announcementTitle", "announcementTitle"),
				ANNOUNCEMENT_MESSAGE:       NewKeywordTextField("announcementMessage", "announcementMessage", "announcementMessage.text"),
				ANNOUNCEMENT_TYPE:          NewKeywordField("announcementType", "announcementType"),
				ANNOUNCEMENT_UPDATED_AT:    NewNumericField("announcementUpdatedAt", "announcementUpdatedAt"),
				ANNOUNCEMENT_UPDATED_BY:    NewKeywordField("announcementUpdatedBy", "announcementUpdatedBy"),
				OWNER_USERS:                NewKeywordTextField("ownerUsers", "ownerUsers", "ownerUsers.text"),
				ADMIN_USERS:                NewKeywordField("adminUsers", "adminUsers"),
				VIEWER_USERS:               NewKeywordField("viewerUsers", "viewerUsers"),
				VIEWER_GROUPS:              NewKeywordField("viewerGroups", "viewerGroups"),
				CONNECTOR_NAME:             NewKeywordTextField("connectorName", "connectorName", "connectorName.text"),
				CONNECTION_QUALIFIED_NAME:  NewKeywordTextField("connectionQualifiedName", "connectionQualifiedName", "connectionQualifiedName.text"),
			},
		},
		PERSONA_GROUPS: NewKeywordField("personaGroups", "personaGroups"),
		PERSONA_USERS:  NewKeywordField("personaUsers", "personaUsers"),
		ROLE_ID:        NewKeywordField("roleId", "roleId"),
	}
}

// NewAuthPolicyFields initializes a new instance of AuthPolicyFields.
func NewAuthPolicyFields() *AuthPolicyFields {
	return &AuthPolicyFields{
		AssetFields: AssetFields{
			AttributesFields: AttributesFields{
				TYPENAME:              NewKeywordTextField("typeName", "__typeName.keyword", "__typeName"),
				GUID:                  NewKeywordField("guid", "__guid"),
				CREATED_BY:            NewKeywordField("createdBy", "__createdBy"),
				UPDATED_BY:            NewKeywordField("updatedBy", "__modifiedBy"),
				STATUS:                NewKeywordField("status", "__state"),
				ATLAN_TAGS:            NewKeywordTextField("classificationNames", "__traitNames", "__classificationsText"),
				PROPOGATED_ATLAN_TAGS: NewKeywordTextField("classificationNames", "__propagatedTraitNames", "__classificationsText"),
				ASSIGNED_TERMS:        NewKeywordTextField("meanings", "__meanings", "__meaningsText"),
				SUPERTYPE_NAMES:       NewKeywordTextField("typeName", "__superTypeNames.keyword", "__superTypeNames"),
				CREATE_TIME:           NewNumericField("createTime", "__timestamp"),
				UPDATE_TIME:           NewNumericField("updateTime", "__modificationTimestamp"),
				QUALIFIED_NAME:        NewKeywordTextField("qualifiedName", "qualifiedName", "qualifiedName.text"),
			},
			NAME:                       NewKeywordTextStemmedField("name", "name.keyword", "name", "name"),
			DISPLAY_NAME:               NewKeywordTextField("displayName", "displayName.keyword", "displayName"),
			DESCRIPTION:                NewKeywordTextField("description", "description", "description.text"),
			USER_DESCRIPTION:           NewKeywordTextField("userDescription", "userDescription", "userDescription.text"),
			TENET_ID:                   NewKeywordField("tenetId", "tenetId"),
			CERTIFICATE_STATUS:         NewKeywordTextField("certificateStatus", "certificateStatus", "certificateStatus.text"),
			CERTIFICATE_STATUS_MESSAGE: NewKeywordField("certificateStatusMessage", "certificateStatusMessage"),
			CERTIFICATE_UPDATED_BY:     NewNumericField("certificateUpdatedBy", "certificateUpdatedBy"),
			ANNOUNCEMENT_TITLE:         NewKeywordField("announcementTitle", "announcementTitle"),
			ANNOUNCEMENT_MESSAGE:       NewKeywordTextField("announcementMessage", "announcementMessage", "announcementMessage.text"),
			ANNOUNCEMENT_TYPE:          NewKeywordField("announcementType", "announcementType"),
			ANNOUNCEMENT_UPDATED_AT:    NewNumericField("announcementUpdatedAt", "announcementUpdatedAt"),
			ANNOUNCEMENT_UPDATED_BY:    NewKeywordField("announcementUpdatedBy", "announcementUpdatedBy"),
			OWNER_USERS:                NewKeywordTextField("ownerUsers", "ownerUsers", "ownerUsers.text"),
			ADMIN_USERS:                NewKeywordField("adminUsers", "adminUsers"),
			VIEWER_USERS:               NewKeywordField("viewerUsers", "viewerUsers"),
			VIEWER_GROUPS:              NewKeywordField("viewerGroups", "viewerGroups"),
			CONNECTOR_NAME:             NewKeywordTextField("connectorName", "connectorName", "connectorName.text"),
			CONNECTION_QUALIFIED_NAME:  NewKeywordTextField("connectionQualifiedName", "connectionQualifiedName", "connectionQualifiedName.text"),
		},
		POLICY_TYPE:               NewKeywordField("policyType", "policyType"),
		POLICY_SERVICE_NAME:       NewKeywordField("policyServiceName", "policyServiceName"),
		POLICY_CATEGORY:           NewKeywordField("policyCategory", "policyCategory"),
		POLICY_SUB_CATEGORY:       NewKeywordField("policySubCategory", "policySubCategory"),
		POLICY_USERS:              NewKeywordField("policyUsers", "policyUsers"),
		POLICY_GROUPS:             NewKeywordField("policyGroups", "policyGroups"),
		POLICY_ROLES:              NewKeywordField("policyRoles", "policyRoles"),
		POLICY_ACTIONS:            NewKeywordField("policyActions", "policyActions"),
		POLICY_RESOURCES:          NewKeywordField("policyResources", "policyResources"),
		POLICY_RESOURCE_CATEGORY:  NewKeywordField("policyResourceCategory", "policyResourceCategory"),
		POLICY_PRIORITY:           NewNumericField("policyPriority", "policyPriority"),
		IS_POLICY_ENABLED:         NewBooleanField("isPolicyEnabled", "isPolicyEnabled"),
		POLICY_MASK_TYPE:          NewKeywordField("policyMaskType", "policyMaskType"),
		POLICY_VALIDITY_SCHEDULE:  NewKeywordField("policyValiditySchedule", "policyValiditySchedule"),
		POLICY_RESOURCE_SIGNATURE: NewKeywordField("policyResourceSignature", "policyResourceSignature"),
		POLICY_DELEGATE_ADMIN:     NewBooleanField("policyDelegateAdmin", "policyDelegateAdmin"),
		POLICY_CONDITIONS:         NewKeywordField("policyConditions", "policyConditions"),
		ACCESS_CONTROL:            NewRelationField("accessControl"),
	}
}

// NewPurposeFields initializes a new instance of PersonaFields.
func NewPurposeFields() *PurposeFields {
	return &PurposeFields{
		AccessControlFields: AccessControlFields{
			IS_ACCESS_CONTROL_ENABLED:  NewBooleanField("isAccessControlEnabled", "isAccessControlEnabled"),
			DENY_CUSTOM_METADATA_GUIDS: NewKeywordField("denyCustomMetadataGuids", "denyCustomMetadataGuids"),
			DENY_ASSET_TABS:            NewKeywordField("denyAssetTabs", "denyAssetTabs"),
			DENY_ASSET_FILTERS:         NewTextField("denyAssetFilters", "denyAssetFilters"),
			CHANNEL_LINK:               NewTextField("channelLink", "channelLink"),
			DENY_ASSET_TYPES:           NewTextField("denyAssetTypes", "denyAssetTypes"),
			DENY_NAVIGATION_PAGES:      NewTextField("denyNavigationPages", "denyNavigationPages"),
			DEFAULT_NAVIGATION:         NewTextField("defaultNavigation", "defaultNavigation"),
			DISPLAY_PREFERENCES:        NewKeywordField("displayPreferences", "displayPreferences"),
			POLICIES:                   NewRelationField("policies"),
			AssetFields: AssetFields{
				AttributesFields: AttributesFields{
					TYPENAME:              NewKeywordTextField("typeName", "__typeName.keyword", "__typeName"),
					GUID:                  NewKeywordField("guid", "__guid"),
					CREATED_BY:            NewKeywordField("createdBy", "__createdBy"),
					UPDATED_BY:            NewKeywordField("updatedBy", "__modifiedBy"),
					STATUS:                NewKeywordField("status", "__state"),
					ATLAN_TAGS:            NewKeywordTextField("classificationNames", "__traitNames", "__classificationsText"),
					PROPOGATED_ATLAN_TAGS: NewKeywordTextField("classificationNames", "__propagatedTraitNames", "__classificationsText"),
					ASSIGNED_TERMS:        NewKeywordTextField("meanings", "__meanings", "__meaningsText"),
					SUPERTYPE_NAMES:       NewKeywordTextField("typeName", "__superTypeNames.keyword", "__superTypeNames"),
					CREATE_TIME:           NewNumericField("createTime", "__timestamp"),
					UPDATE_TIME:           NewNumericField("updateTime", "__modificationTimestamp"),
					QUALIFIED_NAME:        NewKeywordTextField("qualifiedName", "qualifiedName", "qualifiedName.text"),
				},
				NAME:                       NewKeywordTextStemmedField("name", "name.keyword", "name", "name"),
				DISPLAY_NAME:               NewKeywordTextField("displayName", "displayName.keyword", "displayName"),
				DESCRIPTION:                NewKeywordTextField("description", "description", "description.text"),
				USER_DESCRIPTION:           NewKeywordTextField("userDescription", "userDescription", "userDescription.text"),
				TENET_ID:                   NewKeywordField("tenetId", "tenetId"),
				CERTIFICATE_STATUS:         NewKeywordTextField("certificateStatus", "certificateStatus", "certificateStatus.text"),
				CERTIFICATE_STATUS_MESSAGE: NewKeywordField("certificateStatusMessage", "certificateStatusMessage"),
				CERTIFICATE_UPDATED_BY:     NewNumericField("certificateUpdatedBy", "certificateUpdatedBy"),
				ANNOUNCEMENT_TITLE:         NewKeywordField("announcementTitle", "announcementTitle"),
				ANNOUNCEMENT_MESSAGE:       NewKeywordTextField("announcementMessage", "announcementMessage", "announcementMessage.text"),
				ANNOUNCEMENT_TYPE:          NewKeywordField("announcementType", "announcementType"),
				ANNOUNCEMENT_UPDATED_AT:    NewNumericField("announcementUpdatedAt", "announcementUpdatedAt"),
				ANNOUNCEMENT_UPDATED_BY:    NewKeywordField("announcementUpdatedBy", "announcementUpdatedBy"),
				OWNER_USERS:                NewKeywordTextField("ownerUsers", "ownerUsers", "ownerUsers.text"),
				ADMIN_USERS:                NewKeywordField("adminUsers", "adminUsers"),
				VIEWER_USERS:               NewKeywordField("viewerUsers", "viewerUsers"),
				VIEWER_GROUPS:              NewKeywordField("viewerGroups", "viewerGroups"),
				CONNECTOR_NAME:             NewKeywordTextField("connectorName", "connectorName", "connectorName.text"),
				CONNECTION_QUALIFIED_NAME:  NewKeywordTextField("connectionQualifiedName", "connectionQualifiedName", "connectionQualifiedName.text"),
			},
		},
		PURPOSE_CLASSIFICATIONS: NewKeywordField("purposeClassifications", "purposeClassifications"),
	}
}

// NewWorkflowFields initializes and returns a WorkflowFields struct.
func NewWorkflowFields() *WorkflowFields {
	return &WorkflowFields{
		AssetFields: AssetFields{
			AttributesFields: AttributesFields{
				TYPENAME:              NewKeywordTextField("typeName", "__typeName.keyword", "__typeName"),
				GUID:                  NewKeywordField("guid", "__guid"),
				CREATED_BY:            NewKeywordField("createdBy", "__createdBy"),
				UPDATED_BY:            NewKeywordField("updatedBy", "__modifiedBy"),
				STATUS:                NewKeywordField("status", "__state"),
				ATLAN_TAGS:            NewKeywordTextField("classificationNames", "__traitNames", "__classificationsText"),
				PROPOGATED_ATLAN_TAGS: NewKeywordTextField("classificationNames", "__propagatedTraitNames", "__classificationsText"),
				ASSIGNED_TERMS:        NewKeywordTextField("meanings", "__meanings", "__meaningsText"),
				SUPERTYPE_NAMES:       NewKeywordTextField("typeName", "__superTypeNames.keyword", "__superTypeNames"),
				CREATE_TIME:           NewNumericField("createTime", "__timestamp"),
				UPDATE_TIME:           NewNumericField("updateTime", "__modificationTimestamp"),
				QUALIFIED_NAME:        NewKeywordTextField("qualifiedName", "qualifiedName", "qualifiedName.text"),
			},
			NAME:                       NewKeywordTextStemmedField("name", "name.keyword", "name", "name"),
			DISPLAY_NAME:               NewKeywordTextField("displayName", "displayName.keyword", "displayName"),
			DESCRIPTION:                NewKeywordTextField("description", "description", "description.text"),
			USER_DESCRIPTION:           NewKeywordTextField("userDescription", "userDescription", "userDescription.text"),
			TENET_ID:                   NewKeywordField("tenetId", "tenetId"),
			CERTIFICATE_STATUS:         NewKeywordTextField("certificateStatus", "certificateStatus", "certificateStatus.text"),
			CERTIFICATE_STATUS_MESSAGE: NewKeywordField("certificateStatusMessage", "certificateStatusMessage"),
			CERTIFICATE_UPDATED_BY:     NewNumericField("certificateUpdatedBy", "certificateUpdatedBy"),
			ANNOUNCEMENT_TITLE:         NewKeywordField("announcementTitle", "announcementTitle"),
			ANNOUNCEMENT_MESSAGE:       NewKeywordTextField("announcementMessage", "announcementMessage", "announcementMessage.text"),
			ANNOUNCEMENT_TYPE:          NewKeywordField("announcementType", "announcementType"),
			ANNOUNCEMENT_UPDATED_AT:    NewNumericField("announcementUpdatedAt", "announcementUpdatedAt"),
			ANNOUNCEMENT_UPDATED_BY:    NewKeywordField("announcementUpdatedBy", "announcementUpdatedBy"),
			OWNER_USERS:                NewKeywordTextField("ownerUsers", "ownerUsers", "ownerUsers.text"),
			ADMIN_USERS:                NewKeywordField("adminUsers", "adminUsers"),
			VIEWER_USERS:               NewKeywordField("viewerUsers", "viewerUsers"),
			VIEWER_GROUPS:              NewKeywordField("viewerGroups", "viewerGroups"),
			CONNECTOR_NAME:             NewKeywordTextField("connectorName", "connectorName", "connectorName.text"),
			CONNECTION_QUALIFIED_NAME:  NewKeywordTextField("connectionQualifiedName", "connectionQualifiedName", "connectionQualifiedName.text"),
		},
		WORKFLOW_TEMPLATE_GUID:  NewKeywordField("workflowTemplateGuid", "workflowTemplateGuid"),
		WORKFLOW_TYPE:           NewKeywordField("workflowType", "workflowType"),
		WORKFLOW_CONFIG:         NewTextField("workflowConfig", "workflowConfig"),
		WORKFLOW_STATUS:         NewKeywordField("workflowStatus", "workflowStatus"),
		WORKFLOW_RUN_EXPIRES_IN: NewTextField("workflowRunExpiresIn", "workflowRunExpiresIn"),
		WORKFLOW_CREATED_BY:     NewKeywordField("workflowCreatedBy", "workflowCreatedBy"),
		WORKFLOW_UPDATED_BY:     NewKeywordField("workflowUpdatedBy", "workflowUpdatedBy"),
		WORKFLOW_DELETED_AT:     NewNumericField("workflowDeletedAt", "workflowDeletedAt"),
	}
}

// NewWorkflowRunFields initializes and returns a WorkflowRunFields struct.
func NewWorkflowRunFields() *WorkflowRunFields {
	return &WorkflowRunFields{
		AssetFields: AssetFields{
			AttributesFields: AttributesFields{
				TYPENAME:              NewKeywordTextField("typeName", "__typeName.keyword", "__typeName"),
				GUID:                  NewKeywordField("guid", "__guid"),
				CREATED_BY:            NewKeywordField("createdBy", "__createdBy"),
				UPDATED_BY:            NewKeywordField("updatedBy", "__modifiedBy"),
				STATUS:                NewKeywordField("status", "__state"),
				ATLAN_TAGS:            NewKeywordTextField("classificationNames", "__traitNames", "__classificationsText"),
				PROPOGATED_ATLAN_TAGS: NewKeywordTextField("classificationNames", "__propagatedTraitNames", "__classificationsText"),
				ASSIGNED_TERMS:        NewKeywordTextField("meanings", "__meanings", "__meaningsText"),
				SUPERTYPE_NAMES:       NewKeywordTextField("typeName", "__superTypeNames.keyword", "__superTypeNames"),
				CREATE_TIME:           NewNumericField("createTime", "__timestamp"),
				UPDATE_TIME:           NewNumericField("updateTime", "__modificationTimestamp"),
				QUALIFIED_NAME:        NewKeywordTextField("qualifiedName", "qualifiedName", "qualifiedName.text"),
			},
			NAME:                       NewKeywordTextStemmedField("name", "name.keyword", "name", "name"),
			DISPLAY_NAME:               NewKeywordTextField("displayName", "displayName.keyword", "displayName"),
			DESCRIPTION:                NewKeywordTextField("description", "description", "description.text"),
			USER_DESCRIPTION:           NewKeywordTextField("userDescription", "userDescription", "userDescription.text"),
			TENET_ID:                   NewKeywordField("tenetId", "tenetId"),
			CERTIFICATE_STATUS:         NewKeywordTextField("certificateStatus", "certificateStatus", "certificateStatus.text"),
			CERTIFICATE_STATUS_MESSAGE: NewKeywordField("certificateStatusMessage", "certificateStatusMessage"),
			CERTIFICATE_UPDATED_BY:     NewNumericField("certificateUpdatedBy", "certificateUpdatedBy"),
			ANNOUNCEMENT_TITLE:         NewKeywordField("announcementTitle", "announcementTitle"),
			ANNOUNCEMENT_MESSAGE:       NewKeywordTextField("announcementMessage", "announcementMessage", "announcementMessage.text"),
			ANNOUNCEMENT_TYPE:          NewKeywordField("announcementType", "announcementType"),
			ANNOUNCEMENT_UPDATED_AT:    NewNumericField("announcementUpdatedAt", "announcementUpdatedAt"),
			ANNOUNCEMENT_UPDATED_BY:    NewKeywordField("announcementUpdatedBy", "announcementUpdatedBy"),
			OWNER_USERS:                NewKeywordTextField("ownerUsers", "ownerUsers", "ownerUsers.text"),
			ADMIN_USERS:                NewKeywordField("adminUsers", "adminUsers"),
			VIEWER_USERS:               NewKeywordField("viewerUsers", "viewerUsers"),
			VIEWER_GROUPS:              NewKeywordField("viewerGroups", "viewerGroups"),
			CONNECTOR_NAME:             NewKeywordTextField("connectorName", "connectorName", "connectorName.text"),
			CONNECTION_QUALIFIED_NAME:  NewKeywordTextField("connectionQualifiedName", "connectionQualifiedName", "connectionQualifiedName.text"),
		},
		WORKFLOW_RUN_WORKFLOW_GUID: NewKeywordField("workflowRunWorkflowGuid", "workflowRunWorkflowGuid"),
		WORKFLOW_RUN_TYPE:          NewKeywordField("workflowRunType", "workflowRunType"),
		WORKFLOW_RUN_ON_ASSET_GUID: NewKeywordField("workflowRunOnAssetGuid", "workflowRunOnAssetGuid"),
		WORKFLOW_RUN_COMMENT:       NewTextField("workflowRunComment", "workflowRunComment"),
		WORKFLOW_RUN_CONFIG:        NewTextField("workflowRunConfig", "workflowRunConfig"),
		WORKFLOW_RUN_STATUS:        NewKeywordField("workflowRunStatus", "workflowRunStatus"),
		WORKFLOW_RUN_EXPIRES_AT:    NewNumericField("workflowRunExpiresAt", "workflowRunExpiresAt"),
		WORKFLOW_RUN_CREATED_BY:    NewKeywordField("workflowRunCreatedBy", "workflowRunCreatedBy"),
		WORKFLOW_RUN_UPDATED_BY:    NewKeywordField("workflowRunUpdatedBy", "workflowRunUpdatedBy"),
		WORKFLOW_RUN_DELETED_AT:    NewNumericField("workflowRunDeletedAt", "workflowRunDeletedAt"),
	}
}

// Methods on assets

// GetbyGuid retrieves an asset by guid
func GetByGuid[T AtlanObject](guid string) (T, error) {
	var asset T

	if DefaultAtlanClient == nil {
		return asset, fmt.Errorf("default AtlanClient not initialized")
	}

	api := &GET_ENTITY_BY_GUID
	api.Path = fmt.Sprintf("entity/guid/%s", guid) // Adjust to the actual API path structure

	response, err := DefaultAtlanClient.CallAPI(api, nil, nil)
	if err != nil {
		return asset, err
	}

	// Create a new instance of T using reflection
	assetType := reflect.TypeOf(asset).Elem()
	newAsset := reflect.New(assetType).Interface().(T)

	err = newAsset.FromJSON(response)
	if err != nil {
		return asset, err
	}

	return newAsset, nil
}

func ModifyTags(api API,
	assetType reflect.Type,
	qualifiedName string,
	atlanTagNames []string,
	propagate bool,
	removePropagationOnDelete bool,
	restrictLineagePropagation bool,
	restrictPropagationThroughHierarchy bool,
) error {
	var atlanTags []structs.AtlanTag

	for _, name := range atlanTagNames {
		TagName, _ := GetAtlanTagIDForName(name)
		atlanTags = append(atlanTags, structs.AtlanTag{
			TypeName:                            &TagName,
			Propagate:                           &propagate,
			RemovePropagationsOnEntityDelete:    &removePropagationOnDelete,
			RestrictPropagationThroughLineage:   &restrictLineagePropagation,
			RestrictPropagationThroughHierarchy: &restrictPropagationThroughHierarchy,
		})
	}

	queryParams := map[string]string{
		"attr:qualifiedName": qualifiedName,
	}

	API, _ := api.FormatPathWithParams(assetType.Name(), "classifications")

	_, err := DefaultAtlanClient.CallAPI(
		API,
		queryParams,
		atlanTags,
	)
	if err != nil {
		return fmt.Errorf("failed to modify tags: %w", err)
	}

	return nil
}

func AddAtlanTags[T AtlanObject](
	qualifiedName string,
	atlanTagNames []string,
	propagate bool,
	removePropagationOnDelete bool,
	restrictLineagePropagation bool,
	restrictPropagationThroughHierarchy bool,
) error {
	var asset T
	assetType := reflect.TypeOf(asset).Elem()

	err := ModifyTags(
		UPDATE_ENTITY_BY_ATTRIBUTE,
		assetType,
		qualifiedName,
		atlanTagNames,
		propagate,
		removePropagationOnDelete,
		restrictLineagePropagation,
		restrictPropagationThroughHierarchy,
	)
	if err != nil {
		return fmt.Errorf("failed to add Atlan tags: %w", err)
	}
	return nil
}

func UpdateAtlanTags[T AtlanObject](
	qualifiedName string,
	atlanTagNames []string,
	propagate bool,
	removePropagationOnDelete bool,
	restrictLineagePropagation bool,
	restrictPropagationThroughHierarchy bool,
) error {
	var asset T
	assetType := reflect.TypeOf(asset).Elem()

	err := ModifyTags(
		PARTIAL_UPDATE_ENTITY_BY_ATTRIBUTE,
		assetType,
		qualifiedName,
		atlanTagNames,
		propagate,
		removePropagationOnDelete,
		restrictLineagePropagation,
		restrictPropagationThroughHierarchy,
	)
	if err != nil {
		return fmt.Errorf("failed to modify Atlan tags: %w", err)
	}
	return nil
}

func RemoveAtlanTag[T AtlanObject](
	qualifiedName string,
	atlanTagName string,
) error {
	var api API = DELETE_ENTITY_BY_ATTRIBUTE
	var asset T
	assetType := reflect.TypeOf(asset).Elem()

	// Get the internal ID for the tag name
	classificationID, err := GetAtlanTagIDForName(atlanTagName)
	if err != nil {
		return fmt.Errorf("failed to get Atlan tag ID for name %s: %w", atlanTagName, err)
	}

	// If classification ID is not found, return an error
	if classificationID == "" {
		return fmt.Errorf("Atlan tag not found: %s", atlanTagName)
	}

	// Set query params with the qualified name
	queryParams := map[string]string{
		"attr:qualifiedName": qualifiedName,
	}

	// Construct the API path for deleting the tag
	API, _ := api.FormatPathWithParams(assetType.Name(), "classification", classificationID)

	// Call the Atlan API to remove the tag
	_, err = DefaultAtlanClient.CallAPI(API, queryParams, nil)
	if err != nil {
		return fmt.Errorf("failed to remove Atlan tag: %w", err)
	}

	return nil
}

// GetByQualifiedName retrieves an asset by guid
func GetByQualifiedName[T AtlanObject](qualifiedName string) (T, error) {
	var asset T

	if DefaultAtlanClient == nil {
		return asset, fmt.Errorf("default AtlanClient not initialized")
	}

	api := &GET_ENTITY_BY_UNIQUE_ATTRIBUTE
	api.Path += reflect.TypeOf(asset).Elem().Name()

	queryParams := map[string]string{
		"attr:qualifiedName": qualifiedName,
	}

	response, err := DefaultAtlanClient.CallAPI(api, queryParams, nil)
	if err != nil {
		return asset, err
	}

	// Create a new instance of T using reflection
	assetType := reflect.TypeOf(asset).Elem()
	newAsset := reflect.New(assetType).Interface().(T)

	err = newAsset.FromJSON(response)
	if err != nil {
		return asset, err
	}

	return newAsset, nil
}

// RetrieveMinimal retrieves an asset by its GUID, without any of its relationships.
func RetrieveMinimal(guid string) (*structs.Asset, error) {
	if DefaultAtlanClient == nil {
		return nil, fmt.Errorf("default AtlanClient not initialized")
	}

	api := &GET_ENTITY_BY_GUID
	originalPath := api.Path

	api.Path += guid

	// Add query parameters to ignore relationships
	queryParams := make(map[string]string)
	queryParams["min_ext_info"] = "true"
	queryParams["ignore_relationships"] = "true"

	response, err := DefaultAtlanClient.CallAPI(api, queryParams, nil)
	if err != nil {
		return nil, err
	}

	// Unmarshal the response into asset json structure
	var assetresponse structs.Asset
	err = json.Unmarshal(response, &assetresponse)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling asset response: %v", err)
	}

	api.Path = originalPath // Reset the api.Path to its original value
	return &assetresponse, nil
}

// PurgeByGuid HARD deletes assets by their GUIDs.
func PurgeByGuid(guids []string) (*model.AssetMutationResponse, error) {
	if len(guids) == 0 {
		return nil, fmt.Errorf("no GUIDs provided for deletion")
	}

	api := &DELETE_ENTITIES_BY_GUIDS

	// Construct the query parameters
	queryParams := make(map[string]string)
	queryParams["deleteType"] = "HARD"

	// Convert the GUIDs slice to a comma-separated string
	guidString := strings.Join(guids, ",")

	// Add the comma-separated string of GUIDs to the query parameters
	queryParams["guid"] = guidString

	// Call the API
	resp, err := DefaultAtlanClient.CallAPI(api, queryParams, nil)
	if err != nil {
		return nil, err
	}

	// Unmarshal the response into the AssetMutationResponse struct
	var response model.AssetMutationResponse
	err = json.Unmarshal(resp, &response)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %v", err)
	}

	return &response, nil
}

// DeleteByGuid SOFT deletes assets by their GUIDs.
func DeleteByGuid(guids []string) (*model.AssetMutationResponse, error) {
	if len(guids) == 0 {
		return nil, fmt.Errorf("no GUIDs provided for deletion")
	}

	for _, guid := range guids {
		asset, err := RetrieveMinimal(guid)
		if err != nil {
			return nil, fmt.Errorf("error retrieving asset: %v", err)
		}

		// Assuming the asset has a CanBeArchived field that indicates if it can be archived
		if *asset.TypeName == "AtlasGlossaryCategory" {
			return nil, fmt.Errorf("asset %s of type %s cannot be archived", guid, *asset.TypeName)
		}
	}

	api := &DELETE_ENTITIES_BY_GUIDS

	// Construct the query parameters
	queryParams := make(map[string]string)
	queryParams["deleteType"] = "SOFT"

	// Convert the GUIDs slice to a comma-separated string
	guidString := strings.Join(guids, ",")

	// Add the comma-separated string of GUIDs to the query parameters
	queryParams["guid"] = guidString

	fmt.Println("Query Params:", queryParams)

	// Call the API
	resp, err := DefaultAtlanClient.CallAPI(api, queryParams, nil)
	if err != nil {
		DefaultAtlanClient.logger.Errorf("Error soft deleting assets: %v", err)
		return nil, err
	}

	// Unmarshal the response into the AssetMutationResponse struct
	var response model.AssetMutationResponse
	err = json.Unmarshal(resp, &response)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %v", err)
	}

	// Wait until each asset is deleted
	for _, guid := range guids {
		err = WaitTillDeleted(guid)
		if err != nil {
			return nil, err
		}
	}

	return &response, nil
}

// WaitTillDeleted waits for an asset to be deleted.
func WaitTillDeleted(guid string) error {
	for i := 0; i < MaxRetries; i++ {
		asset, err := RetrieveMinimal(guid)
		if err != nil {
			return fmt.Errorf("error retrieving asset: %v", err)
		}

		if *asset.Status == "DELETED" {
			return nil
		}

		// If the asset is not deleted, wait for a while before retrying
		time.Sleep(RetryInterval)
	}

	// If the asset is still not deleted after all retries, return an error
	return errors.New("retry limit overrun waiting for asset to be deleted")
}

type SaveRequest struct {
	Entities []AtlanObject `json:"entities"`
}

// Save saves the assets in memory to the Atlas server.
func Save(assets ...AtlanObject) (*model.AssetMutationResponse, error) {
	request := SaveRequest{
		Entities: assets,
	}

	api := &CREATE_ENTITIES
	resp, err := DefaultAtlanClient.CallAPI(api, nil, request)
	if err != nil {
		return nil, err
	}

	var response model.AssetMutationResponse
	err = json.Unmarshal(resp, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

func generateCacheKey(baseURL, apiKey string) string {
	h := fnv.New32a()
	_, _ = h.Write([]byte(fmt.Sprintf("%s/%s", baseURL, apiKey)))
	return fmt.Sprintf("%d", h.Sum32())
}

// Used in End-to-end bulk update

// TrimToRequired trims a SearchAsset to its required attributes and returns an SearchAsset Object.
func TrimToRequired(asset model.SearchAssets) (*model.SearchAssets, error) {
	// Validate required fields
	if asset.TypeName == nil || asset.QualifiedName == nil {
		return nil, fmt.Errorf("asset must have TypeName and QualifiedName")
	}

	instance := &model.SearchAssets{}
	instance.TypeName = asset.TypeName
	instance.QualifiedName = asset.QualifiedName
	instance.Name = asset.Name
	instance.Guid = asset.Guid
	// Call the generic Updater method directly on the asset
	err := instance.Updater()
	if err != nil {
		return nil, fmt.Errorf("failed to trim to required fields for asset type: %s, error: %w", *asset.TypeName, err)
	}

	return instance, nil
}

// BaseEntity represents the base entity structure returned by the Atlan while fetching an entity.
type BaseEntity struct {
	ReferredEntities map[string]interface{} `json:"referredEntities"`
	Entity           Entity                 `json:"entity"`
}

type Entity struct {
	TypeName               string            `json:"typeName"`
	AttributesJSON         json.RawMessage   `json:"attributes"`
	Guid                   string            `json:"guid"`
	IsIncomplete           bool              `json:"isIncomplete"`
	Status                 atlan.AtlanStatus `json:"status"`
	CreatedBy              string            `json:"createdBy"`
	UpdatedBy              string            `json:"updatedBy"`
	CreateTime             int64             `json:"createTime"`
	UpdateTime             int64             `json:"updateTime"`
	Version                int               `json:"version"`
	RelationshipAttributes json.RawMessage   `json:"relationshipAttributes"`
	Labels                 []interface{}     `json:"labels"`
}

// UnmarshalBaseEntity unmarshals the base entity and the attributes into the specific entity struct.
func UnmarshalBaseEntity(data []byte, entity interface{}) (*BaseEntity, error) {
	// Unmarshal the base entity.
	var base BaseEntity
	if err := json.Unmarshal(data, &base); err != nil {
		return nil, err
	}

	// Unmarshal the attributes into the specific entity struct.
	if err := json.Unmarshal(base.Entity.AttributesJSON, entity); err != nil {
		return nil, err
	}

	return &base, nil
}
