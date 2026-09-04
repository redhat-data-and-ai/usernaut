package structs

import (
	"encoding/json"
	"time"

	"github.com/atlanhq/atlan-go/atlan"
)

// PackageParameter represents package-related parameters.
type PackageParameter struct {
	Parameter string          `json:"parameter"`
	Type      string          `json:"type"`
	Body      json.RawMessage `json:"body"`
}

// WorkflowMetadata captures metadata about a workflow.
type WorkflowMetadata struct {
	Annotations       map[string]string `json:"annotations,omitempty"`
	CreationTimestamp *string           `json:"creationTimestamp,omitempty"`
	GenerateName      *string           `json:"generateName,omitempty"`
	Generation        *int              `json:"generation,omitempty"`
	Labels            map[string]string `json:"labels,omitempty"`
	ManagedFields     []interface{}     `json:"managedFields,omitempty"`
	Name              *string           `json:"name,omitempty"`
	Namespace         *string           `json:"namespace,omitempty"`
	ResourceVersion   *string           `json:"resourceVersion,omitempty"`
	UID               *string           `json:"uid,omitempty"`
}

// WorkflowTemplateRef references a specific workflow template.
type WorkflowTemplateRef struct {
	Name         string `json:"name"`
	Template     string `json:"template"`
	ClusterScope bool   `json:"clusterScope"`
}

// NameValuePair represents a name-value pair.
type NameValuePair struct {
	Name  string      `json:"name"`
	Value interface{} `json:"value"`
}

// WorkflowParameters holds parameters for workflow operations.
type WorkflowParameters struct {
	Parameters []NameValuePair `json:"parameters"`
}

// WorkflowTask represents a task in the workflow.
type WorkflowTask struct {
	Name        string              `json:"name"`
	Arguments   WorkflowParameters  `json:"arguments"`
	TemplateRef WorkflowTemplateRef `json:"templateRef"`
}

// WorkflowDAG represents a directed acyclic graph (DAG) of workflow tasks.
type WorkflowDAG struct {
	Tasks []WorkflowTask `json:"tasks"`
}

// WorkflowTemplate captures the structure of a workflow template.
type WorkflowTemplate struct {
	Name     string      `json:"name"`
	Inputs   interface{} `json:"inputs,omitempty"`
	Outputs  interface{} `json:"outputs,omitempty"`
	Metadata interface{} `json:"metadata,omitempty"`
	DAG      WorkflowDAG `json:"dag"`
}

// WorkflowSpec defines the specification of a workflow.
type WorkflowSpec struct {
	Entrypoint          *string                  `json:"entrypoint,omitempty"`
	Arguments           interface{}              `json:"arguments,omitempty"`
	Templates           []WorkflowTemplate       `json:"templates,omitempty"`
	WorkflowTemplateRef map[string]string        `json:"workflowTemplateRef,omitempty"`
	WorkflowMetadata    *WorkflowMetadata        `json:"workflowMetadata,omitempty"`
	Synchronization     *WorkflowSynchronization `json:"synchronization,omitempty"`
}

// Workflow represents the primary workflow object.
type Workflow struct {
	Metadata *WorkflowMetadata  `json:"metadata"`
	Spec     *WorkflowSpec      `json:"spec"`
	Payload  []PackageParameter `json:"payload,omitempty"`
}

// WorkflowAsset defines the Asset structure for a workflow.
type WorkflowAsset struct {
	Asset
	WorkflowAttributes *WorkflowAttributes `json:"Attributes"`
}

// WorkflowAttributes captures attributes of a workflow.
type WorkflowAttributes struct {
	WorkflowTemplateGuid *string    `json:"workflowTemplateGuid,omitempty"`
	WorkflowType         *string    `json:"workflowType,omitempty"`
	WorkflowConfig       *string    `json:"workflowConfig,omitempty"`
	WorkflowStatus       *string    `json:"workflowStatus,omitempty"`
	WorkflowRunExpiresIn *string    `json:"workflowRunExpiresIn,omitempty"`
	WorkflowCreatedBy    *string    `json:"workflowCreatedBy,omitempty"`
	WorkflowUpdatedBy    *string    `json:"workflowUpdatedBy,omitempty"`
	WorkflowDeletedAt    *time.Time `json:"workflowDeletedAt,omitempty"`
}

// WorkflowSearchResultStatus captures the status of a workflow search result.
type WorkflowSearchResultStatus struct {
	ArtifactGCStatus           map[string]interface{}    `json:"artifactGCStatus,omitempty"`
	ArtifactRepositoryRef      interface{}               `json:"artifactRepositoryRef,omitempty"`
	CompressedNodes            *string                   `json:"compressedNodes,omitempty"`
	EstimatedDuration          *int                      `json:"estimatedDuration,omitempty"`
	Conditions                 []interface{}             `json:"conditions,omitempty"`
	Message                    *string                   `json:"message,omitempty"`
	FinishedAt                 *string                   `json:"finishedAt,omitempty"`
	Nodes                      interface{}               `json:"nodes,omitempty"`
	Outputs                    *WorkflowParameters       `json:"outputs,omitempty"`
	Phase                      *atlan.AtlanWorkflowPhase `json:"phase,omitempty"`
	Progress                   *string                   `json:"progress,omitempty"`
	ResourcesDuration          map[string]int            `json:"resourcesDuration,omitempty"`
	StartedAt                  *string                   `json:"startedAt,omitempty"`
	StoredTemplates            interface{}               `json:"storedTemplates,omitempty"`
	StoredWorkflowTemplateSpec interface{}               `json:"storedWorkflowTemplateSpec,omitempty"`
	Synchronization            map[string]interface{}    `json:"synchronization,omitempty"`
}

// WorkflowSearchResultDetail contains detailed information about a workflow search result.
type WorkflowSearchResultDetail struct {
	APIVersion string                      `json:"apiVersion"`
	Kind       string                      `json:"kind"`
	Metadata   WorkflowMetadata            `json:"metadata"`
	Spec       WorkflowSpec                `json:"spec"`
	Status     *WorkflowSearchResultStatus `json:"status,omitempty"`
}

// WorkflowSearchResult captures the search result for a workflow.
type WorkflowSearchResult struct {
	Index       string                     `json:"_index"`
	Type        string                     `json:"_type"`
	ID          string                     `json:"_id"`
	SeqNo       interface{}                `json:"_seq_no"`
	PrimaryTerm interface{}                `json:"_primary_term"`
	Sort        []interface{}              `json:"sort"`
	Source      WorkflowSearchResultDetail `json:"_source"`
}

// Status returns the workflow phase if available.
func (w *WorkflowSearchResult) Status() *atlan.AtlanWorkflowPhase {
	if w.Source.Status != nil {
		return w.Source.Status.Phase
	}
	return nil
}

// ToWorkflow converts a WorkflowSearchResult into a Workflow object.
func (w *WorkflowSearchResult) ToWorkflow() *Workflow {
	return &Workflow{
		Spec:     &w.Source.Spec,
		Metadata: &w.Source.Metadata,
	}
}

// WorkflowSearchHits contains hits from a workflow search response.
type WorkflowSearchHits struct {
	Total map[string]interface{} `json:"total"`
	Hits  []WorkflowSearchResult `json:"hits,omitempty"`
}

// WorkflowSearchResponse represents the response from a workflow search.
type WorkflowSearchResponse struct {
	Took   *int                   `json:"took,omitempty"`
	Hits   WorkflowSearchHits     `json:"hits"`
	Shards map[string]interface{} `json:"_shards"`
}

// ReRunRequest captures a request to rerun a workflow.
type ReRunRequest struct {
	Namespace    string `json:"namespace"`
	ResourceKind string `json:"resourceKind"`
	ResourceName string `json:"resourceName"`
}

// WorkflowResponse represents a generic workflow response.
type WorkflowResponse struct {
	Metadata *WorkflowMetadata `json:"metadata"`
	Spec     *WorkflowSpec     `json:"spec"`
	Payload  []interface{}     `json:"payload,omitempty"`
}

// WorkflowRunResponse extends WorkflowResponse with status information.
type WorkflowRunResponse struct {
	WorkflowResponse
	Status WorkflowSearchResultStatus `json:"status"`
}

// ScheduleQueriesSearchRequest defines a request for searching schedule queries.
type ScheduleQueriesSearchRequest struct {
	StartDate string `json:"startDate"`
	EndDate   string `json:"endDate"`
}

// WorkflowSchedule defines a workflow schedule.
type WorkflowSchedule struct {
	Timezone     string `json:"timezone"`
	CronSchedule string `json:"cronSchedule"`
}

// WorkflowScheduleSpec specifies details for a workflow schedule.
type WorkflowScheduleSpec struct {
	Schedule                   *string       `json:"schedule,omitempty"`
	Timezone                   *string       `json:"timezone,omitempty"`
	WorkflowSpec               *WorkflowSpec `json:"workflowSpec,omitempty"`
	ConcurrencyPolicy          *string       `json:"concurrencyPolicy,omitempty"`
	StartingDeadlineSeconds    *int          `json:"startingDeadlineSeconds,omitempty"`
	SuccessfulJobsHistoryLimit *int          `json:"successfulJobsHistoryLimit,omitempty"`
	FailedJobsHistoryLimit     *int          `json:"failedJobsHistoryLimit,omitempty"`
}

// WorkflowScheduleStatus captures the status of a workflow schedule.
type WorkflowScheduleStatus struct {
	Active            interface{} `json:"active,omitempty"`
	Conditions        interface{} `json:"conditions,omitempty"`
	LastScheduledTime *string     `json:"lastScheduledTime,omitempty"`
}

// WorkflowScheduleResponse captures the response for a workflow schedule.
type WorkflowScheduleResponse struct {
	Metadata         *WorkflowMetadata       `json:"metadata,omitempty"`
	Spec             *WorkflowScheduleSpec   `json:"spec,omitempty"`
	Status           *WorkflowScheduleStatus `json:"status,omitempty"`
	WorkflowMetadata *WorkflowMetadata       `json:"workflowMetadata,omitempty"`
}

// WorkflowRun defines a workflow run as an Asset.
type WorkflowRun struct {
	Asset
	WorkflowRunAttributes *WorkflowRunAttributes `json:"Attributes"`
}

type WorkflowRunAttributes struct {
	WorkflowRunGuid        *string `json:"workflowRunGuid,omitempty"`
	WorkflowRunType        *string `json:"workflowRunType,omitempty"`
	WorkflowRunOnAssetGuid *string `json:"workflowRunOnAssetGuid,omitempty"`
	WorkflowRunComment     *string `json:"workflowRunComment,omitempty"`
	WorkflowRunConfig      *string `json:"workflowRunConfig,omitempty"`
	WorkflowRunStatus      *string `json:"workflowRunStatus,omitempty"`
	WorkflowRunExpiresAt   *string `json:"workflowRunExpiresAt,omitempty"`
	WorkflowRunCreatedBy   *string `json:"workflowRunCreatedBy,omitempty"`
	WorkflowRunUpdatedBy   *string `json:"workflowRunUpdatedBy,omitempty"`
	WorkflowRunDeletedAt   *string `json:"workflowRunDeletedAt,omitempty"`
}

// WorkflowSynchronization represents the synchronization settings for a workflow.
type WorkflowSynchronization struct {
	Semaphore *WorkflowSemaphore `json:"semaphore,omitempty"`
}

// WorkflowSemaphore represents a semaphore for workflow concurrency control.
type WorkflowSemaphore struct {
	ConfigMapKeyRef *WorkflowConfigMapKeyRef `json:"configMapKeyRef,omitempty"`
}

// WorkflowConfigMapKeyRef represents a reference to a ConfigMap key.
type WorkflowConfigMapKeyRef struct {
	Name string `json:"name"`
	Key  string `json:"key"`
}
