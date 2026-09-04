package structs

import "github.com/atlanhq/atlan-go/atlan"

type AtlanIcon string

type AtlanConnectorType string

/*
 * Base class for all entities
 */

type Referenceable struct {
	// Type of the asset. For example Table, Column, and so on.
	TypeName *string `json:"typeName"`
	// Globally unique identifier (GUID) of any object in Atlan.
	Guid *string `json:"guid,omitempty"`
	// Atlan user who created this asset.
	CreatedBy *string `json:"createdBy,omitempty"`
	// Atlan user who last updated this asset.
	UpdatedBy *string `json:"updatedBy,omitempty"`
	// Asset status in Atlan (active vs deleted)
	Status *atlan.AtlanStatus `json:"status,omitempty"`
	// All directly-assigned Atlan tags that exist on an asset, searchable by internal hashed-string ID of the Atlan tag.
	AtlanTags *[]AtlanTag `json:"classifications,omitempty"`
	// All propagated Atlan tags that exist on an asset, searchable by internal hashed-string ID of the Atlan tag.
	PropagatedAtlanTags *string `json:"classifications,omitempty"`
	// All terms attached to an asset, searchable by the term's qualifiedName.
	AssignedTerms *[]AtlasGlossaryTerm `json:"meanings,omitempty"`
	// All super types of an asset.
	SuperTypeNames *string `json:"supertypeName,omitempty"`
	// Time (in milliseconds) when the asset was created.
	CreateTime *int `json:"createTime,omitempty"`
	// Time (in milliseconds) when the asset was last updated.
	UpdateTime *string `json:"updateTime,omitempty"`
	// Unique fully-qualified name of the asset in Atlan.
	QualifiedName *string `json:"qualifiedName,omitempty"`
}

/*
 * Base class for all assets.
 */

type CertificateStatus string

type PopularityInsights int

type MCIncident string

type MCMonitor string

type File string

type SchemaRegistrySubject string

type Metric struct {
	// DataQuality
}

type Readme string

type SodaCheck string

type SourceCostUnitType string

type StarredDetails string

type Meaning string

type Asset struct {
	Referenceable
	// List of groups who administer this asset.
	AdminGroups *[]string `json:"adminGroups,omitempty"`
	// List of roles who administer this asset.
	AdminRoles *[]string `json:"adminRoles,omitempty"`
	// List of users who administer this asset.
	AdminUsers *[]string `json:"adminUsers,omitempty"`
	// Detailed message to include in the announcement on this asset.
	AnnouncementMessage *string `json:"announcementMessage,omitempty"`
	// Brief title for the announcement on this asset.
	AnnouncementTitle *string `json:"announcementTitle,omitempty"`
	// Type of announcement on this asset.
	AnnouncementType *atlan.AnnouncementType `json:"announcementType,omitempty"`
	// Time (epoch) at which the announcement was last updated, in milliseconds.
	AnnouncementUpdatedAt *int64 `json:"announcementUpdatedAt,omitempty"`
	// Name of the user who last updated the announcement.
	AnnouncementUpdatedBy *string `json:"announcementUpdatedBy,omitempty"`
	// Name of the account in which this asset exists in dbt.
	AssetDbtAccountName *string `json:"assetDbtAccountName,omitempty"`
	// Alias of this asset in dbt.
	AssetDbtAlias *string `json:"assetDbtAlias,omitempty"`
	// Version of the environment in which this asset is materialized in dbt.
	AssetDbtEnvironmentDbtVersion *string `json:"assetDbtEnvironmentDbtVersion,omitempty"`
	// Name of the environment in which this asset is materialized in dbt.
	AssetDbtEnvironmentName *string `json:"assetDbtEnvironmentName,omitempty"`
	// Time (epoch) at which the job that materialized this asset in dbt last ran, in milliseconds.
	AssetDbtJobLastRun *int64 `json:"assetDbtJobLastRun,omitempty"`
	// Path in S3 to the artifacts saved from the last run of the job that materialized this asset in dbt.
	AssetDbtJobLastRunArtifactS3Path *string `json:"assetDbtJobLastRunArtifactS3Path,omitempty"`
	// Whether artifacts were saved from the last run of the job that materialized this asset in dbt (true) or not (false).
	AssetDbtJobLastRunArtifactsSaved *bool `json:"assetDbtJobLastRunArtifactsSaved,omitempty"`
	// Time (epoch) at which the job that materialized this asset in dbt was last created, in milliseconds.
	AssetDbtJobLastRunCreatedAt *int64 `json:"assetDbtJobLastRunCreatedAt,omitempty"`
	// Time (epoch) at which the job that materialized this asset in dbt was dequeued, in milliseconds.
	AssetDbtJobLastRunDequedAt *int64 `json:"assetDbtJobLastRunDequedAt,omitempty"`
	// Thread ID of the user who executed the last run of the job that materialized this asset in dbt.
	AssetDbtJobLastRunExecutedByThreadId *string `json:"assetDbtJobLastRunExecutedByThreadId,omitempty"`
	// Branch in git from which the last run of the job that materialized this asset in dbt ran.
	AssetDbtJobLastRunGitBranch *string `json:"assetDbtJobLastRunGitBranch,omitempty"`
	// SHA hash in git for the last run of the job that materialized this asset in dbt.
	AssetDbtJobLastRunGitSha *string `json:"assetDbtJobLastRunGitSha,omitempty"`
	// Whether docs were generated from the last run of the job that materialized this asset in dbt (true) or not (false).
	AssetDbtJobLastRunHasDocsGenerated *bool `json:"assetDbtJobLastRunHasDocsGenerated,omitempty"`
	// Whether sources were generated from the last run of the job that materialized this asset in dbt (true) or not (false).
	AssetDbtJobLastRunHasSourcesGenerated *bool `json:"assetDbtJobLastRunHasSourcesGenerated,omitempty"`
	// Whether notifications were sent from the last run of the job that materialized this asset in dbt (true) or not (false).
	AssetDbtJobLastRunNotificationsSent *bool `json:"assetDbtJobLastRunNotificationsSent,omitempty"`
	// Thread ID of the owner of the last run of the job that materialized this asset in dbt.
	AssetDbtJobLastRunOwnerThreadId *string `json:"assetDbtJobLastRunOwnerThreadId,omitempty"`
	// Total duration the job that materialized this asset in dbt spent being queued.
	AssetDbtJobLastRunQueuedDuration *string `json:"assetDbtJobLastRunQueuedDuration,omitempty"`
	// Human-readable total duration of the last run of the job that materialized this asset in dbt spend being queued.
	AssetDbtJobLastRunQueuedDurationHumanized *string `json:"assetDbtJobLastRunQueuedDurationHumanized,omitempty"`
	// Run duration of the last run of the job that materialized this asset in dbt.
	AssetDbtJobLastRunRunDuration *string `json:"assetDbtJobLastRunRunDuration,omitempty"`
	// Human-readable run duration of the last run of the job that materialized this asset in dbt.
	AssetDbtJobLastRunRunDurationHumanized *string `json:"assetDbtJobLastRunRunDurationHumanized,omitempty"`
	// Time (epoch) at which the job that materialized this asset in dbt was started running, in milliseconds.
	AssetDbtJobLastRunStartedAt *int64 `json:"assetDbtJobLastRunStartedAt,omitempty"`
	// Status message of the last run of the job that materialized this asset in dbt.
	AssetDbtJobLastRunStatusMessage *string `json:"assetDbtJobLastRunStatusMessage,omitempty"`
	// Total duration of the last run of the job that materialized this asset in dbt.
	AssetDbtJobLastRunTotalDuration *string `json:"assetDbtJobLastRunTotalDuration,omitempty"`
	// Human-readable total duration of the last run of the job that materialized this asset in dbt.
	AssetDbtJobLastRunTotalDurationHumanized *string `json:"assetDbtJobLastRunTotalDurationHumanized,omitempty"`
	// Time (epoch) at which the job that materialized this asset in dbt was last updated, in milliseconds.
	AssetDbtJobLastRunUpdatedAt *int64 `json:"assetDbtJobLastRunUpdatedAt,omitempty"`
	// URL of the last run of the job that materialized this asset in dbt.
	AssetDbtJobLastRunUrl *string `json:"assetDbtJobLastRunUrl,omitempty"`
	// Name of the job that materialized this asset in dbt.
	AssetDbtJobName *string `json:"assetDbtJobName,omitempty"`
	// Time (epoch) when the next run of the job that materializes this asset in dbt is scheduled.
	AssetDbtJobNextRun *int64 `json:"assetDbtJobNextRun,omitempty"`
	// Human-readable time when the next run of the job that materializes this asset in dbt is scheduled.
	AssetDbtJobNextRunHumanized *string `json:"assetDbtJobNextRunHumanized,omitempty"`
	// Schedule of the job that materialized this asset in dbt.
	AssetDbtJobSchedule *string `json:"assetDbtJobSchedule,omitempty"`
	// Human-readable cron schedule of the job that materialized this asset in dbt.
	AssetDbtJobScheduleCronHumanized *string `json:"assetDbtJobScheduleCronHumanized,omitempty"`
	// Status of the job that materialized this asset in dbt.
	AssetDbtJobStatus *string `json:"assetDbtJobStatus,omitempty"`
	// Metadata for this asset in dbt, specifically everything under the 'meta' key in the dbt object.
	AssetDbtMeta *string `json:"assetDbtMeta,omitempty"`
	// Name of the package in which this asset exists in dbt.
	AssetDbtPackageName *string `json:"assetDbtPackageName,omitempty"`
	// Name of the project in which this asset exists in dbt.
	AssetDbtProjectName *string `json:"assetDbtProjectName,omitempty"`
	// URL of the semantic layer proxy for this asset in dbt.
	AssetDbtSemanticLayerProxyUrl *string `json:"assetDbtSemanticLayerProxyUrl,omitempty"`
	// Freshness criteria for the source of this asset in dbt.
	AssetDbtSourceFreshnessCriteria *string `json:"assetDbtSourceFreshnessCriteria,omitempty"`
	// List of tags attached to this asset in dbt.
	AssetDbtTags *[]string `json:"assetDbtTags,omitempty"`
	// All associated dbt test statuses.
	AssetDbtTestStatus *string `json:"assetDbtTestStatus,omitempty"`
	// Unique identifier of this asset in dbt.
	AssetDbtUniqueId *string `json:"assetDbtUniqueId,omitempty"`
	// Name of the icon to use for this asset.
	AssetIcon *atlan.AtlanIcon `json:"assetIcon,omitempty"`
	// List of Monte Carlo incident names attached to this asset.
	AssetMcIncidentNames *[]string `json:"assetMcIncidentNames,omitempty"`
	// List of unique Monte Carlo incident names attached to this asset.
	AssetMcIncidentQualifiedNames *[]string `json:"assetMcIncidentQualifiedNames,omitempty"`
	// List of Monte Carlo incident severities associated with this asset.
	AssetMcIncidentSeverities *[]string `json:"assetMcIncidentSeverities,omitempty"`
	// List of Monte Carlo incident states associated with this asset.
	AssetMcIncidentStates *[]string `json:"assetMcIncidentStates,omitempty"`
	// List of Monte Carlo incident sub-types associated with this asset.
	AssetMcIncidentSubTypes *[]string `json:"assetMcIncidentSubTypes,omitempty"`
	// List of Monte Carlo incident types associated with this asset.
	AssetMcIncidentTypes *[]string `json:"assetMcIncidentTypes,omitempty"`
	// Time (epoch) at which this asset was last synced from Monte Carlo.
	AssetMcLastSyncRunAt *int64 `json:"assetMcLastSyncRunAt,omitempty"`
	// List of Monte Carlo monitor names attached to this asset.
	AssetMcMonitorNames *[]string `json:"assetMcMonitorNames,omitempty"`
	// List of unique Monte Carlo monitor names attached to this asset.
	AssetMcMonitorQualifiedNames *[]string `json:"assetMcMonitorQualifiedNames,omitempty"`
	// Schedules of all associated Monte Carlo monitors.
	AssetMcMonitorScheduleTypes *[]string `json:"assetMcMonitorScheduleTypes,omitempty"`
	// Statuses of all associated Monte Carlo monitors.
	AssetMcMonitorStatuses *[]string `json:"assetMcMonitorStatuses,omitempty"`
	// Types of all associated Monte Carlo monitors.
	AssetMcMonitorTypes *[]string `json:"assetMcMonitorTypes,omitempty"`
	// Number of checks done via Soda.
	AssetSodaCheckCount *int64 `json:"assetSodaCheckCount,omitempty"`
	// All associated Soda check statuses.
	AssetSodaCheckStatuses *string `json:"assetSodaCheckStatuses,omitempty"`
	// Status of data quality from Soda.
	AssetSodaDQStatus *string `json:"assetSodaDQStatus,omitempty"`
	// Time (epoch) at which the last scan via Soda occurred, in milliseconds.
	AssetSodaLastScanAt *int64 `json:"assetSodaLastScanAt,omitempty"`
	// Time (epoch) at which this asset was last synced via Soda, in milliseconds.
	AssetSodaLastSyncRunAt *int64 `json:"assetSodaLastSyncRunAt,omitempty"`
	// URL of the source for Soda.
	AssetSodaSourceURL *string `json:"assetSodaSourceURL,omitempty"`
	// List of tags attached to this asset.
	AssetTags *[]string `json:"assetTags,omitempty"`
	// Glossary terms that are linked to this asset.
	AssignedTerms *[]AtlasGlossaryTerm `json:"assignedTerms,omitempty"`
	// Status of this asset's certification.
	CertificateStatus *atlan.CertificateStatus `json:"certificateStatus,omitempty"`
	// Human-readable descriptive message used to provide further detail to certificateStatus.
	CertificateStatusMessage *string `json:"certificateStatusMessage,omitempty"`
	// Time (epoch) at which the certification was last updated, in milliseconds.
	CertificateUpdatedAt *int64 `json:"certificateUpdatedAt,omitempty"`
	// Name of the user who last updated the certification of this asset.
	CertificateUpdatedBy *string `json:"certificateUpdatedBy,omitempty"`
	// Simple name of the connection through which this asset is accessible.
	ConnectionName *string `json:"connectionName,omitempty"`
	// Unique name of the connection through which this asset is accessible.
	ConnectionQualifiedName *string `json:"connectionQualifiedName,omitempty"`
	// Type of the connector through which this asset is accessible.
	ConnectorType *AtlanConnectorType `json:"connectorType,omitempty"`
	// Unique name of this asset in dbt.
	DbtQualifiedName *string `json:"dbtQualifiedName,omitempty"`
	// Description of this asset, for example as crawled from a source.
	Description *string `json:"description,omitempty"`
	// Human-readable name of this asset used for display purposes (in user interface).
	DisplayName *string `json:"displayText,omitempty"`
	// List of files associated with this asset.
	Files *[]File `json:"files,omitempty"`
	// Whether this asset has lineage (true) or not (false).
	HasLineage *bool `json:"hasLineage,omitempty"`
	// Whether this asset is AI-generated (true) or not (false).
	IsAIGenerated *bool `json:"isAIGenerated,omitempty"`
	// Whether this asset is discoverable through the UI (true) or not (false).
	IsDiscoverable *bool `json:"isDiscoverable,omitempty"`
	// Whether this asset can be edited in the UI (true) or not (false).
	IsEditable *bool `json:"isEditable,omitempty"`
	// Time (epoch) of the last operation that inserted, updated, or deleted rows, in milliseconds.
	LastRowChangedAt *int64 `json:"lastRowChangedAt,omitempty"`
	// Name of the last run of the crawler that last synchronized this asset.
	LastSyncRun *string `json:"lastSyncRun,omitempty"`
	// Time (epoch) at which this asset was last crawled, in milliseconds.
	LastSyncRunAt *int64 `json:"lastSyncRunAt,omitempty"`
	// Name of the crawler that last synchronized this asset.
	LastSyncWorkflowName *string `json:"lastSyncWorkflowName,omitempty"`
	// Links that are attached to this asset.
	Links *[]Link `json:"links,omitempty"`
	// Monte Carlo incidents associated with this asset.
	McIncidents *[]MCIncident `json:"mcIncidents,omitempty"`
	// Monte Carlo monitors that observe this asset.
	McMonitors *[]MCMonitor `json:"mcMonitors,omitempty"`
	// Metrics associated with this asset.
	Metrics *[]Metric `json:"metrics,omitempty"`
	// Name of this asset.
	Name *string `json:"name,omitempty"`
	// List of groups who own this asset.
	OwnerGroups *[]string `json:"ownerGroups,omitempty"`
	// List of users who own this asset.
	OwnerUsers *[]string `json:"ownerUsers,omitempty"`
	// Popularity score for this asset.
	PopularityScore *float64 `json:"popularityScore,omitempty"`
	// README that is linked to this asset.
	Readme *Readme `json:"readme,omitempty"`
	// URL for sample data for this asset.
	SampleDataUrl *string `json:"sampleDataUrl,omitempty"`
	// Subjects in the schema registry for this asset.
	SchemaRegistrySubjects *[]SchemaRegistrySubject `json:"schemaRegistrySubjects,omitempty"`
	// Soda checks associated with this asset.
	SodaChecks *[]SodaCheck `json:"sodaChecks,omitempty"`
	// Unit of measure for sourceTotalCost.
	SourceCostUnit *SourceCostUnitType `json:"sourceCostUnit,omitempty"`
	// Time (epoch) at which this asset was created in the source system, in milliseconds.
	SourceCreatedAt *int64 `json:"sourceCreatedAt,omitempty"`
	// User who created this asset in the source system.
	SourceCreatedBy *string `json:"sourceCreatedBy,omitempty"`
	// URL to create an embed for a resource (for example, an image of a dashboard) within Atlan.
	SourceEmbedURL *string `json:"sourceEmbedURL,omitempty"`
	// Timestamp of most recent read operation.
	SourceLastReadAt *int64 `json:"sourceLastReadAt,omitempty"`
	// Owners of this asset in the source system.
	SourceOwners *string `json:"sourceOwners,omitempty"`
	// Records of most expensive warehouse with extra insights.
	SourceQueryComputeCostRecords *[]PopularityInsights `json:"sourceQueryComputeCostRecords,omitempty"`
	// Names of most expensive warehouses.
	SourceQueryComputeCosts *[]string `json:"sourceQueryComputeCosts,omitempty"`
	// Total count of all read operations at source.
	SourceReadCount *int64 `json:"sourceReadCount,omitempty"`
	// Records of most expensive queries that accessed this asset.
	SourceReadExpensiveQueryRecords *[]PopularityInsights `json:"sourceReadExpensiveQueryRecords,omitempty"`
	// Records of most popular queries that accessed this asset.
	SourceReadPopularQueryRecords *[]PopularityInsights `json:"sourceReadPopularQueryRecords,omitempty"`
	// Total cost of read queries at source.
	SourceReadQueryCost *float64 `json:"sourceReadQueryCost,omitempty"`
	// Records of most recent users who read this asset.
	SourceReadRecentUserRecords *[]PopularityInsights `json:"sourceReadRecentUserRecords,omitempty"`
	// Names of most recent users who read this asset.
	SourceReadRecentUsers *[]string `json:"sourceReadRecentUsers,omitempty"`
	// Records of slowest queries that accessed this asset.
	SourceReadSlowQueryRecords *[]PopularityInsights `json:"sourceReadSlowQueryRecords,omitempty"`
	// Records of users who read this asset the most.
	SourceReadTopUserRecords *[]PopularityInsights `json:"sourceReadTopUserRecords,omitempty"`
	// Names of users who read this asset the most.
	SourceReadTopUsers *[]string `json:"sourceReadTopUsers,omitempty"`
	// Total number of unique users that read data from asset.
	SourceReadUserCount *int64 `json:"sourceReadUserCount,omitempty"`
	// Total cost of all operations at source.
	SourceTotalCost *float64 `json:"sourceTotalCost,omitempty"`
	// URL to the resource within the source application.
	SourceURL *string `json:"sourceURL,omitempty"`
	// Time (epoch) at which this asset was last updated in the source system, in milliseconds.
	SourceUpdatedAt *int64 `json:"sourceUpdatedAt,omitempty"`
	// User who last updated this asset in the source system.
	SourceUpdatedBy *string `json:"sourceUpdatedBy,omitempty"`
	// Users who have starred this asset.
	StarredBy *[]string `json:"starredBy,omitempty"`
	// Number of users who have starred this asset.
	StarredCount *int `json:"starredCount,omitempty"`
	// Details of users who have starred this asset.
	StarredDetails *[]StarredDetails `json:"starredDetails,omitempty"`
	// Subtype of this asset.
	SubType *string `json:"subType,omitempty"`
	// Name of the Atlan workspace in which this asset exists.
	TenantId *string `json:"tenantId,omitempty"`
	// Description of this asset as provided by a user.
	UserDescription *string `json:"userDescription,omitempty"`
	// View score for this asset.
	ViewScore *float64 `json:"viewScore,omitempty"`
	// List of groups who can view assets contained in a collection.
	ViewerGroups *[]string `json:"viewerGroups,omitempty"`
	// List of users who can view assets contained in a collection.
	ViewerUsers *[]string `json:"viewerUsers,omitempty"`
	// Internal tracking of fields that should be serialized with null values.
	NullFields *[]string `json:"nullFields,omitempty"`
	// Atlan tags assigned to the asset.
	AtlanTags *[]AtlanTag `json:"classifications,omitempty"`
	// Map of custom metadata attributes and values defined on the asset.
	CustomMetadataSets map[string]map[string]interface{} `json:"customMetadataSets,omitempty"`
	// Time (epoch) at which the asset was created, in milliseconds.
	CreateTime *int64 `json:"createTime,omitempty"`
	// Time (epoch) at which the asset was last updated, in milliseconds.
	UpdateTime *int64 `json:"updateTime,omitempty"`
	// Details on the handler used for deletion of the asset.
	DeleteHandler *string `json:"deleteHandler,omitempty"`
	// Names of the Atlan tags that exist on the asset.
	AtlanTagNames *[]string `json:"classificationNames,omitempty"`
	// Unused.
	IsIncomplete *bool `json:"isIncomplete,omitempty"`
	// Names of terms that have been linked to this asset.
	MeaningNames *[]string `json:"meaningNames,omitempty"`
	// Details of terms that have been linked to this asset.
	Meanings *[]AtlasGlossaryTerm `json:"meanings,omitempty"`
	// Unique identifiers (GUIDs) for any background tasks that are yet to operate on this asset.
	PendingTasks *[]string `json:"pendingTasks,omitempty"`

	DisplayText *string `json:"displayName,omitempty"`
}

type Relation struct {
	DisplayText            *string            `json:"displayText,omitempty"`
	EntityStatus           *string            `json:"entityStatus,omitempty"`
	RelationshipType       *string            `json:"relationshipType,omitempty"`
	RelationshipGuid       *string            `json:"relationshipGuid,omitempty"`
	RelationshipStatus     *atlan.AtlanStatus `json:"relationshipStatus,omitempty"`
	relationshipAttributes *[]interface{}     `json:"relationshipAttributes,omitempty"`
	uniqueAttributes       *string            `json:"uniqueAttributes,omitempty"`
}

type Attributes struct {
	PopularityScore                       float64 `json:"popularityScore"`
	AssetMcMonitorNames                   []string
	LastSyncRunAt                         int64 `json:"lastSyncRunAt"`
	__hasLineage                          bool
	SourceQueryComputeCostRecordList      []interface{}
	AssetSodaLastSyncRunAt                int64 `json:"assetSodaLastSyncRunAt"`
	StarredCount                          int
	AdminUsers                            []interface{}
	LastRowChangedAt                      int64 `json:"lastRowChangedAt"`
	SourceReadRecentUserList              []interface{}
	AssetMcIncidentQualifiedNames         []interface{}
	AssetMcIncidentTypes                  []interface{}
	AssetSodaLastScanAt                   int64 `json:"assetSodaLastScanAt"`
	SourceUpdatedAt                       int64 `json:"sourceUpdatedAt"`
	AssetDbtJobLastRunArtifactsSaved      bool
	StarredDetailsList                    []interface{}
	IsEditable                            bool `json:"isEditable"`
	SourceReadCount                       int
	AnnouncementUpdatedAt                 int64 `json:"announcementUpdatedAt"`
	SourceCreatedAt                       int64 `json:"sourceCreatedAt"`
	AssetDbtJobLastRunDequedAt            int64 `json:"assetDbtJobLastRunDequedAt"`
	AssetDbtTags                          []interface{}
	SourceReadSlowQueryRecordList         []interface{}
	QualifiedName                         string `json:"qualifiedName"`
	SourceQueryComputeCostList            []interface{}
	AssetDbtJobLastRunNotificationsSent   bool
	AssetMcMonitorTypes                   []interface{}
	AssetSodaCheckCount                   int
	AssetMcMonitorStatuses                []interface{}
	StarredBy                             []interface{}
	SourceLastReadAt                      int64 `json:"sourceLastReadAt"`
	Name                                  string
	CertificateUpdatedAt                  int64 `json:"certificateUpdatedAt"`
	AssetMcIncidentSeverities             []interface{}
	SourceReadQueryCost                   float64 `json:"sourceReadQueryCost"`
	OwnerUsers                            []interface{}
	AssetDbtJobLastRunHasSourcesGenerated bool
	AssetMcIncidentSubTypes               []interface{}
	IsAIGenerated                         bool `json:"isAIGenerated"`
	AssetDbtJobLastRunHasDocsGenerated    bool `json:"assetDbtJobLastRunHasDocsGenerated"`
	AssetTags                             []interface{}
	AssetMcIncidentStates                 []interface{}
	AssetDbtJobLastRunUpdatedAt           int64 `json:"assetDbtJobLastRunUpdatedAt"`
	OwnerGroups                           []interface{}
	AssetMcMonitorQualifiedNames          []interface{}
	SourceReadExpensiveQueryRecordList    []interface{}
	AssetDbtJobLastRunStartedAt           int64 `json:"assetDbtJobLastRunStartedAt"`
	IsDiscoverable                        bool  `json:"isDiscoverable"`
	IsPartial                             bool  `json:"isPartial"`
	SourceReadTopUserRecordList           []interface{}
	AssetMcMonitorScheduleTypes           []interface{}
	ViewerUsers                           []interface{}
	ViewScore                             float64 `json:"viewScore"`
	SourceReadTopUserList                 []interface{}
	AssetMcIncidentNames                  []interface{}
	AdminRoles                            []interface{}
	AdminGroups                           []interface{}
	AssetDbtJobLastRunCreatedAt           int64 `json:"assetDbtJobLastRunCreatedAt"`
	AssetDbtJobNextRun                    int64 `json:"assetDbtJobNextRun"`
	SourceReadRecentUserRecordList        []interface{}
	AssetIcon                             string `json:"assetIcon"`
	SourceReadPopularQueryRecordList      []interface{}
	SourceTotalCost                       float64 `json:"sourceTotalCost"`
	AssetMcLastSyncRunAt                  int64   `json:"assetMcLastSyncRunAt"`
	SourceReadUserCount                   int
	ViewerGroups                          []interface{}
	AssetDbtJobLastRun                    int64 `json:"assetDbtJobLastRun"`
}

// AtlanTag represents a tag in Atlan.
type AtlanTag struct {
	TypeName                            *string `json:"typeName"`
	EntityGuid                          *string `json:"entityGuid,omitempty"`
	EntityStatus                        *string `json:"entityStatus,omitempty"`
	Propagate                           *bool   `json:"propagate,omitempty"`
	RemovePropagationsOnEntityDelete    *bool   `json:"removePropagationsOnEntityDelete,omitempty"`
	RestrictPropagationThroughLineage   *bool   `json:"restrictPropagationThroughLineage,omitempty"`
	RestrictPropagationThroughHierarchy *bool   `json:"restrictPropagationThroughHierarchy,omitempty"`
}

type Link struct {
	Guid                   string                 `json:"guid"`
	TypeName               string                 `json:"typeName"`
	EntityStatus           string                 `json:"entityStatus"`
	DisplayText            string                 `json:"displayText"`
	RelationshipType       string                 `json:"relationshipType"`
	RelationshipGuid       string                 `json:"relationshipGuid"`
	RelationshipStatus     string                 `json:"relationshipStatus"`
	RelationshipAttributes map[string]interface{} `json:"relationshipAttributes"`
}

func StringPtr(s string) *string {
	return &s
}
