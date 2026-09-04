package assets

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/atlanhq/atlan-go/atlan"
	"github.com/atlanhq/atlan-go/atlan/model"
	"github.com/atlanhq/atlan-go/atlan/model/structs"
)

const (
	workflowRunSchedule = "orchestration.atlan.com/schedule"
	workflowRunTimezone = "orchestration.atlan.com/timezone"
	MonitorSleepSeconds = 5
)

type WorkflowClient struct {
	*AtlanClient
}

// FindByType searches for workflows by their type prefix.
// Params:
//   - prefix: The workflow package type (atlan.WorkflowPackage) to search for (for example atlan.WorkflowPackageSnowflakeMiner).
//   - maxResults: The maximum number of workflows to return.
//
// Returns:
//   - A slice of structs.WorkflowSearchResult containing the workflows found.
//   - An error if any occurs during the request.
func (w *WorkflowClient) FindByType(prefix atlan.WorkflowPackage, maxResults int) ([]structs.WorkflowSearchResult, error) {
	var query model.Query = &model.BoolQuery{
		Filter: []model.Query{
			&model.NestedQuery{
				Path: "metadata",
				Query: &model.PrefixQuery{
					Field: "metadata.name.keyword",
					Value: prefix.Name,
				},
			},
		},
	}

	nestedPath := "metadata"

	// Add sorting based on creation timestamp in descending order.
	sortItems := []model.SortItem{
		{
			Order:      atlan.SortOrderDescending,
			Field:      "metadata.creationTimestamp",
			NestedPath: &nestedPath,
		},
	}

	request := model.WorkflowSearchRequest{
		Query: query,
		Size:  maxResults,
		Sort:  sortItems,
	}

	rawJSON, err := DefaultAtlanClient.CallAPI(&WORKFLOW_INDEX_SEARCH, nil, &request)
	if err != nil {
		return nil, err
	}
	var response structs.WorkflowSearchResponse
	if err := json.Unmarshal(rawJSON, &response); err != nil {
		return nil, err
	}
	return response.Hits.Hits, nil
}

// FindByID searches for a workflow by its ID.
// Params:
//   - id: The unique ID of the workflow to search for.
//     (e.g: `atlan-snowflake-miner-1714638976-mzdza`)
//
// Returns:
//   - A pointer to a structs.WorkflowSearchResult containing the workflow found, or nil if no workflow is found.
//   - An error if any occurs during the request or unmarshalling.
func (w *WorkflowClient) FindByID(id string) (*structs.WorkflowSearchResult, error) {
	var query model.Query = &model.BoolQuery{
		Filter: []model.Query{
			&model.NestedQuery{
				Path: "metadata",
				Query: &model.TermQuery{
					Field: "metadata.name.keyword",
					Value: id,
				},
			},
		},
	}

	request := model.WorkflowSearchRequest{
		Query: query,
		Size:  1,
	}

	rawJSON, err := DefaultAtlanClient.CallAPI(&WORKFLOW_INDEX_SEARCH, nil, &request)
	if err != nil {
		return nil, err
	}

	var response structs.WorkflowSearchResponse
	if err := json.Unmarshal(rawJSON, &response); err != nil {
		return nil, err
	}

	if len(response.Hits.Hits) > 0 {
		return &response.Hits.Hits[0], nil
	}

	if len(response.Hits.Hits) == 0 {
		return nil, errors.New("no workflow found")
	}

	return nil, nil
}

// FindRunByID searches for a workflow run by its ID.
// Params:
//   - id: The unique ID of the workflow run to search for.
//     (e.g: `atlan-snowflake-miner-1714638976-mzdza`)
//
// Returns:
//   - A pointer to a structs.WorkflowSearchResult containing the workflow run found, or nil if no workflow run is found.
//   - An error if any occurs during the request or unmarshalling.
func (w *WorkflowClient) FindRunByID(id string) (*structs.WorkflowSearchResult, error) {
	var query model.Query = &model.BoolQuery{
		Filter: []model.Query{
			&model.TermQuery{
				Field: "_id",
				Value: id,
			},
		},
	}

	response, err := w.findRuns(query, 0, 1)
	if err != nil {
		return nil, err
	}

	if len(response.Hits.Hits) > 0 {
		return &response.Hits.Hits[0], nil
	}
	return nil, nil
}

// findRuns retrieves existing workflow runs based on a given query.
// Params:
//   - query: The query to filter the workflow runs.
//   - from: The starting index to retrieve the workflow runs (default: `0`).
//   - size: The number of workflow runs to retrieve (default: `100`).
//
// Returns:
//   - A pointer to structs.WorkflowSearchResponse containing the retrieved workflow runs.
//   - An error if any occurs during the request or unmarshalling.
func (w *WorkflowClient) findRuns(query model.Query, from, size int) (*structs.WorkflowSearchResponse, error) {
	request := model.WorkflowSearchRequest{
		Query: query,
		From:  from,
		Size:  size,
	}

	rawJSON, err := DefaultAtlanClient.CallAPI(&WORKFLOW_INDEX_RUN_SEARCH, nil, &request)
	if err != nil {
		return nil, err
	}

	var response structs.WorkflowSearchResponse
	if err := json.Unmarshal(rawJSON, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

// FindLatestRun retrieves the latest run of a given workflow by its name.
// Params:
//   - workflowName: The name of the workflow to search for.
//
// Returns:
//   - A pointer to a structs.WorkflowSearchResult containing the latest workflow run found, or nil if no run is found.
//   - An error if any occurs during the request or unmarshalling.
func (w *WorkflowClient) FindLatestRun(workflowName string) (*structs.WorkflowSearchResult, error) {
	var query model.Query = &model.BoolQuery{
		Filter: []model.Query{
			&model.NestedQuery{
				Path: "spec",
				Query: &model.TermQuery{
					Field: "spec.workflowTemplateRef.name.keyword",
					Value: workflowName,
				},
			},
		},
	}

	response, err := w.findRuns(query, 0, 1)
	if err != nil {
		return nil, err
	}

	if len(response.Hits.Hits) > 0 {
		return &response.Hits.Hits[0], nil
	}
	return nil, nil
}

// FindCurrentRun retrieves the most current, still-running workflow for a given workflow name.
// Params:
//   - workflowName: The name of the workflow to search for.
//
// Returns:
//   - A pointer to a structs.WorkflowSearchResult containing the currently running workflow, or nil if no run is found or it's not running.
//   - An error if any occurs during the request or unmarshalling.
func (w *WorkflowClient) FindCurrentRun(workflowName string) (*structs.WorkflowSearchResult, error) {
	var query model.Query = &model.BoolQuery{
		Filter: []model.Query{
			&model.NestedQuery{
				Path: "spec",
				Query: &model.TermQuery{
					Field: "spec.workflowTemplateRef.name.keyword",
					Value: workflowName,
				},
			},
		},
	}

	response, err := w.findRuns(query, 0, 50)
	if err != nil {
		return nil, err
	}

	for _, result := range response.Hits.Hits {
		if *result.Source.Status.Phase == atlan.AtlanWorkflowPhasePending || *result.Source.Status.Phase == atlan.AtlanWorkflowPhaseRunning {
			return &result, nil
		}
	}
	return nil, nil
}

// GetRuns retrieves all workflow runs filtered by workflow name and phase.
// Params:
//   - workflowName: The name of the workflow to filter by.
//   - workflowPhase: The phase of the workflow to filter by.
//   - from: The starting index to retrieve the workflow runs.
//   - size: The number of workflow runs to retrieve.
//
// Returns:
//   - A slice of structs.WorkflowSearchResult containing the workflow runs found.
//   - An error if any occurs during the request or unmarshalling.
func (w *WorkflowClient) GetRuns(workflowName string, workflowPhase atlan.AtlanWorkflowPhase, from, size int) ([]structs.WorkflowSearchResult, error) {
	var query model.Query = &model.BoolQuery{
		Must: []model.Query{
			&model.NestedQuery{
				Path: "spec",
				Query: &model.TermQuery{
					Field: "spec.workflowTemplateRef.name.keyword",
					Value: workflowName,
				},
			},
		},
		Filter: []model.Query{
			&model.TermQuery{
				Field: "status.phase.keyword",
				Value: workflowPhase.Name,
			},
		},
	}

	response, err := w.findRuns(query, from, size)
	if err != nil {
		return nil, err
	}

	return response.Hits.Hits, nil
}

// Stop stops a running workflow by its run ID.
// Params:
//   - workflowRunID: The unique ID of the workflow run to stop.
//
// Returns:
//   - A pointer to structs.WorkflowRunResponse containing the result of the stop action.
//   - An error if any occurs during the request or unmarshalling.
func (w *WorkflowClient) Stop(workflowRunID string) (*structs.WorkflowRunResponse, error) {
	api := &STOP_WORKFLOW_RUN
	api.Path = fmt.Sprintf("runs/%s/stop", workflowRunID)

	rawJSON, err := DefaultAtlanClient.CallAPI(api, nil, "")
	if err != nil {
		return nil, err
	}

	var response structs.WorkflowRunResponse
	if err := json.Unmarshal(rawJSON, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

// Delete archives (deletes) the provided workflow by its name.
// Params:
//   - workflowName: The name of the workflow to archive.
//
// Returns:
//   - An error if any occurs during the request or API call.
func (w *WorkflowClient) Delete(workflowName string) error {
	api := &WORKFLOW_ARCHIVE
	api.Path = fmt.Sprintf("workflows/%s/archive", workflowName)
	_, err := DefaultAtlanClient.CallAPI(api, nil, "")
	return err
}

// handleWorkflowTypes determines the workflow details based on its type.
// Params:
//   - workflow: The workflow to handle, which can be a WorkflowPackage, WorkflowSearchResult, WorkflowSearchResultDetail, or Workflow.
//
// Returns:
//   - A pointer to structs.WorkflowSearchResultDetail containing the details of the workflow.
//   - An error if the workflow type is invalid or any issue occurs.
func (w *WorkflowClient) handleWorkflowTypes(workflow interface{}) (*structs.WorkflowSearchResultDetail, error) {
	switch wf := workflow.(type) {
	case atlan.WorkflowPackage:
		results, err := w.FindByType(wf, 1) // Fetching at most 1 result
		if err != nil {
			return nil, err
		}
		if len(results) == 0 {
			return nil, fmt.Errorf("no prior run available for workflow: %v", wf)
		}
		return &results[0].Source, nil

	case structs.WorkflowSearchResult:
		return &wf.Source, nil

	case *structs.WorkflowSearchResultDetail:
		return wf, nil
	case *structs.Workflow:
		return convertWorkflowToSearchResult(wf)

	default:
		return nil, fmt.Errorf("invalid workflow type provided")
	}
}

// convertWorkflowToSearchResult converts a Workflow into a WorkflowSearchResultDetail.
// Params:
//   - wf: The Workflow to convert.
//
// Returns:
//   - A pointer to structs.WorkflowSearchResultDetail containing the converted details of the workflow.
//   - An error if any issues occur while converting.
func convertWorkflowToSearchResult(wf *structs.Workflow) (*structs.WorkflowSearchResultDetail, error) {
	if wf == nil {
		return nil, fmt.Errorf("workflow is nil")
	}
	if wf.Metadata == nil || wf.Spec == nil {
		return nil, fmt.Errorf("workflow metadata or spec is nil")
	}

	return &structs.WorkflowSearchResultDetail{
		Metadata: *wf.Metadata,
		Spec:     *wf.Spec,
	}, nil
}

// Rerun executes the workflow immediately if it has been run before.
// Params:
//   - workflow: The workflow to rerun, which can be a WorkflowPackage, WorkflowSearchResult, WorkflowSearchResultDetail, or Workflow.
//   - idempotent: A boolean indicating whether the workflow should only be rerun if not already running.
//
// Returns:
//   - A pointer to structs.WorkflowRunResponse containing the result of the rerun.
//   - An error if any occurs during the rerun process.
func (w *WorkflowClient) Rerun(workflow interface{}, idempotent bool) (*structs.WorkflowRunResponse, error) {
	detail, err := w.handleWorkflowTypes(workflow)
	if err != nil {
		return nil, err
	}

	if idempotent && *detail.Metadata.Name != "" {
		// Wait before checking the current workflow run status
		time.Sleep(10 * time.Second)

		currentRun, err := w.FindCurrentRun(*detail.Metadata.Name)
		if err == nil && currentRun != nil && currentRun.Source.Status != nil {
			return &structs.WorkflowRunResponse{
				WorkflowResponse: structs.WorkflowResponse{
					Metadata: &currentRun.Source.Metadata,
					Spec:     &currentRun.Source.Spec,
				},
				Status: *currentRun.Source.Status,
			}, nil
		}
	}

	request := structs.ReRunRequest{
		Namespace:    *detail.Metadata.Namespace,
		ResourceName: *detail.Metadata.Name,
	}

	rawJSON, err := DefaultAtlanClient.CallAPI(&WORKFLOW_RERUN, nil, &request)
	if err != nil {
		return nil, err
	}

	var response structs.WorkflowRunResponse
	if err := json.Unmarshal(rawJSON, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

// Update modifies the configuration of an existing workflow.
// Params:
//   - workflow: The workflow to update, which is a pointer to structs.Workflow.
//
// Returns:
//   - A pointer to structs.WorkflowResponse containing the updated workflow result.
//   - An error if any occurs during the update process.
func (w *WorkflowClient) Update(workflow *structs.Workflow) (*structs.WorkflowResponse, error) {
	api := &WORKFLOW_UPDATE
	api.Path = fmt.Sprintf("workflows/%s", *workflow.Metadata.Name)

	rawJSON, err := DefaultAtlanClient.CallAPI(api, nil, workflow)
	if err != nil {
		return nil, err
	}

	var response structs.WorkflowResponse
	if err := json.Unmarshal(rawJSON, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

// UpdateOwner assigns a new owner to the workflow.
// Params:
//   - workflowName: The name of the workflow to update.
//   - username: The username of the new owner to assign.
//
// Returns:
//   - A pointer to structs.WorkflowResponse containing the updated workflow result.
//   - An error if any occurs during the ownership change.
func (w *WorkflowClient) UpdateOwner(workflowName, username string) (*structs.WorkflowResponse, error) {
	api := &WORKFLOW_CHANGE_OWNER
	api.Path = fmt.Sprintf("workflows/%s/changeownership", workflowName)

	queryParams := map[string]string{"username": username}

	rawJSON, err := DefaultAtlanClient.CallAPI(api, queryParams, nil)
	if err != nil {
		return nil, err
	}

	var response structs.WorkflowResponse
	if err := json.Unmarshal(rawJSON, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

// Methods related to workflow schedules

// Monitor the status of the workflow's run.
// Params:
//   - workflowResponse: The response containing the workflow details to monitor.
//   - logger: An optional logger for printing the workflow status during monitoring.
//
// Returns:
//   - The current workflow phase (atlan.AtlanWorkflowPhase) indicating the status of the workflow run.
//   - An error if any occurs during the monitoring process.
func (w *WorkflowClient) Monitor(workflowResponse *structs.WorkflowResponse, logger *log.Logger) (*atlan.AtlanWorkflowPhase, error) {
	if workflowResponse.Metadata == nil || *workflowResponse.Metadata.Name == "" {
		if logger != nil {
			logger.Println("Skipping workflow monitoring — nothing to monitor.")
		}
		return nil, nil
	}

	name := workflowResponse.Metadata.Name
	var status *atlan.AtlanWorkflowPhase

	for status == nil || (*status != atlan.AtlanWorkflowPhaseSuccess && *status != atlan.AtlanWorkflowPhaseError && *status != atlan.AtlanWorkflowPhaseFailed) {
		time.Sleep(MonitorSleepSeconds * time.Second)
		runDetails, _ := w.FindLatestRun(*name)
		if runDetails != nil {
			status = runDetails.Status()
		}
		if logger != nil {
			logger.Printf("Workflow status: %s\n", status)
		}
	}
	if logger != nil {
		logger.Printf("Workflow completion status: %s\n", status)
	}
	return status, nil
}

func (w *WorkflowClient) addSchedule(workflow *structs.WorkflowSearchResultDetail, schedule *structs.WorkflowSchedule) {
	if workflow.Metadata.Annotations == nil {
		workflow.Metadata.Annotations = make(map[string]string)
	}
	workflow.Metadata.Annotations[workflowRunSchedule] = schedule.CronSchedule
	workflow.Metadata.Annotations[workflowRunTimezone] = schedule.Timezone
}

// AddSchedule adds a schedule for an existing workflow run.
//
// This method attaches a cron schedule and timezone to the given workflow.
//
// Param:
//   - workflow: The workflow object to schedule, can be of type WorkflowSearchResultDetail or similar.
//   - schedule: A WorkflowSchedule object containing:
//   - Cron schedule expression (e.g., `5 4 * * *`).
//   - Timezone for the cron schedule (e.g., `Europe/Paris`).
//
// Returns:
//   - A WorkflowResponse object containing the details of the scheduled workflow.
//   - Error if any occurred during the process.
func (w *WorkflowClient) AddSchedule(workflow interface{}, schedule *structs.WorkflowSchedule) (*structs.WorkflowResponse, error) {
	workflowToUpdate, err := w.handleWorkflowTypes(workflow)
	if err != nil {
		return nil, err
	}

	w.addSchedule(workflowToUpdate, schedule)

	api := &WORKFLOW_UPDATE
	api.Path = fmt.Sprintf("workflows/%s", *workflowToUpdate.Metadata.Name)

	rawJSON, err := DefaultAtlanClient.CallAPI(api, nil, workflowToUpdate)
	if err != nil {
		return nil, err
	}

	var response structs.WorkflowResponse
	if err := json.Unmarshal(rawJSON, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

// RemoveSchedule removes the schedule from an existing workflow run.
//
// This method removes the cron schedule and timezone annotation from the given workflow.
//
// Param:
//   - workflow: The workflow object to remove the schedule from, can be of type WorkflowSearchResultDetail or similar.
//
// Returns:
//   - A WorkflowResponse object with the updated workflow details.
//   - Error if any occurred during the process.
func (w *WorkflowClient) RemoveSchedule(workflow interface{}) (*structs.WorkflowResponse, error) {
	workflowToUpdate, err := w.handleWorkflowTypes(workflow)
	if err != nil {
		return nil, err
	}

	if workflowToUpdate.Metadata.Annotations != nil {
		delete(workflowToUpdate.Metadata.Annotations, workflowRunSchedule)
	}

	api := &WORKFLOW_UPDATE
	api.Path = fmt.Sprintf("workflows/%s", *workflowToUpdate.Metadata.Name)

	rawJSON, err := DefaultAtlanClient.CallAPI(api, nil, workflowToUpdate)
	if err != nil {
		return nil, err
	}

	var response structs.WorkflowResponse
	if err := json.Unmarshal(rawJSON, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

// GetAllScheduledRuns retrieves all scheduled runs for workflows.
//
// This method fetches the list of all scheduled workflow runs.
//
// Returns:
//   - A WorkflowScheduleResponse containing the list of scheduled workflows.
//   - Error if any occurred during the API call.
func (w *WorkflowClient) GetAllScheduledRuns() (*structs.WorkflowScheduleResponse, error) {
	rawJSON, err := DefaultAtlanClient.CallAPI(&GET_ALL_SCHEDULE_RUNS, nil, nil)
	if err != nil {
		return nil, err
	}

	var response *structs.WorkflowScheduleResponse

	if err := json.Unmarshal(rawJSON, &response); err != nil {
		return nil, err
	}

	return response, nil
}

// GetScheduledRun retrieves an existing scheduled run for a workflow.
//
// This method fetches the scheduled workflow run for the given workflow name.
//
// Param:
//   - workflowName: The name of the workflow (e.g., `atlan-snowflake-miner-1714638976`).
//
// Returns:
//   - A WorkflowScheduleResponse containing the scheduled run details.
//   - Error if any occurred during the API call.
func (w *WorkflowClient) GetScheduledRun(workflowName string) (*structs.WorkflowScheduleResponse, error) {
	api := &GET_SCHEDULE_RUN
	api.Path = fmt.Sprintf("runs/cron/%s-cron", workflowName)

	rawJSON, err := DefaultAtlanClient.CallAPI(api, nil, nil)
	if err != nil {
		return nil, err
	}

	var response structs.WorkflowScheduleResponse
	if err := json.Unmarshal(rawJSON, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

// FindScheduleQuery searches for scheduled query workflows by their saved query identifier.
//
// This method retrieves scheduled workflows related to a specific saved query ID.
//
// Param:
//   - savedQueryID: The identifier of the saved query.
//   - maxResults: The maximum number of results to retrieve. Defaults to `10`.
//
// Returns:
//   - A slice of WorkflowSearchResult containing the matching scheduled workflows.
//   - Error if any occurred during the search process.
func (w *WorkflowClient) FindScheduleQuery(savedQueryID string, maxResults int) ([]structs.WorkflowSearchResult, error) {
	if maxResults <= 0 {
		maxResults = 10
	}

	var query model.Query = &model.BoolQuery{
		Filter: []model.Query{
			&model.NestedQuery{
				Path: "metadata",
				Query: &model.PrefixQuery{
					Field: "metadata.name.keyword",
					Value: fmt.Sprintf("asq-%s", savedQueryID),
				},
			},
			&model.NestedQuery{
				Path: "metadata",
				Query: &model.TermQuery{
					Field: "metadata.annotations.package.argoproj.io/name.keyword",
					Value: "@atlan/schedule-query",
				},
			},
		},
	}

	request := model.WorkflowSearchRequest{
		Query: query,
		Size:  maxResults,
	}

	rawJSON, err := DefaultAtlanClient.CallAPI(&WORKFLOW_INDEX_SEARCH, nil, request)
	if err != nil {
		return nil, err
	}

	var response structs.WorkflowSearchResponse
	if err := json.Unmarshal(rawJSON, &response); err != nil {
		return nil, err
	}

	if response.Hits.Hits != nil {
		return response.Hits.Hits, nil
	}
	return nil, nil
}

// ReRunScheduleQuery re-triggers a scheduled query workflow by its schedule query ID.
//
// This method re-runs a scheduled workflow using the given schedule query identifier.
//
// Param:
//   - scheduleQueryID: The identifier of the schedule query workflow.
//
// Returns:
//   - A WorkflowRunResponse containing details of the re-triggered workflow.
//   - Error if any occurred during the re-run process.
func (w *WorkflowClient) ReRunScheduleQuery(scheduleQueryID string) (*structs.WorkflowRunResponse, error) {
	request := structs.ReRunRequest{
		Namespace:    "default",
		ResourceName: scheduleQueryID,
	}

	rawJSON, err := DefaultAtlanClient.CallAPI(&WORKFLOW_OWNER_RERUN, nil, &request)
	if err != nil {
		return nil, err
	}

	var response structs.WorkflowRunResponse
	if err := json.Unmarshal(rawJSON, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

// FindScheduleQueryBetween searches for scheduled query workflows within a specific date range.
//
// This method retrieves scheduled workflows that fall between the specified start and end dates.
//
// Param:
//   - request: A ScheduleQueriesSearchRequest object containing the start and end dates in ISO 8601 format.
//   - missed: If true, searches for missed scheduled workflows.
//
// Returns:
//   - A slice of WorkflowRunResponse containing the matching scheduled workflows within the date range.
//   - Error if any occurred during the search process.
func (w *WorkflowClient) FindScheduleQueryBetween(request structs.ScheduleQueriesSearchRequest, missed bool) ([]structs.WorkflowRunResponse, error) {
	queryParams := map[string]string{
		"startDate": request.StartDate,
		"endDate":   request.EndDate,
	}

	searchAPI := SCHEDULE_QUERY_WORKFLOWS_SEARCH
	if missed {
		searchAPI = SCHEDULE_QUERY_WORKFLOWS_MISSED
	}

	rawJSON, err := DefaultAtlanClient.CallAPI(&searchAPI, queryParams, nil)
	if err != nil {
		return nil, err
	}

	var response []structs.WorkflowRunResponse
	if err := json.Unmarshal(rawJSON, &response); err != nil {
		return nil, err
	}
	return response, nil
}

// Run executes an Atlan workflow with an optional schedule.
//
// This method triggers the workflow and attaches a schedule if provided.
//
// Params:
//   - workflow: The workflow object to execute, can be a JSON string or an object (WorkflowResponse, WorkflowSearchResult, etc.).
//   - schedule: A WorkflowSchedule object containing cron expression and timezone.
//
// Returns:
//   - A WorkflowResponse object containing the details of the executed workflow.
//   - Error if any occurred during the execution process.
func (w *WorkflowClient) Run(workflow interface{}, schedule *structs.WorkflowSchedule) (*structs.WorkflowResponse, error) {
	if workflow == nil {
		return nil, errors.New("workflow cannot be nil")
	}

	var workflowPayload interface{}

	switch v := workflow.(type) {
	case string: // If workflow is a JSON string, unmarshal it
		if err := json.Unmarshal([]byte(v), &workflowPayload); err != nil {
			return nil, fmt.Errorf("invalid JSON workflow: %w", err)
		}
	default:
		workflowPayload = workflow
	}

	if schedule != nil {
		workflowToUpdate, _ := w.handleWorkflowTypes(workflowPayload)
		w.addSchedule(workflowToUpdate, schedule)
	}

	responseData, err := DefaultAtlanClient.CallAPI(&WORKFLOW_RUN, nil, workflowPayload)
	if err != nil {
		return nil, fmt.Errorf("error executing workflow: %w", err)
	}

	var workflowResponse structs.WorkflowResponse
	if err := json.Unmarshal(responseData, &workflowResponse); err != nil {
		return nil, fmt.Errorf("failed to parse workflow response: %w", err)
	}

	return &workflowResponse, nil
}
