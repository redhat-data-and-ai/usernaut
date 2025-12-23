package assets

import (
	"fmt"
	"net/http"
	"net/url"
	"path"
)

const (
	// Types API
	TYPES_API            = "types/"
	TYPEDEFS_API         = TYPES_API + "typedefs/"
	TYPEDEF_BY_NAME      = TYPES_API + "typedef/name/"
	TYPEDEF_BY_GUID      = TYPES_API + "typedef/guid/"
	GET_BY_NAME_TEMPLATE = TYPES_API + "{path_type}/name/{name}"
	GET_BY_GUID_TEMPLATE = TYPES_API + "{path_type}/guid/{guid}"

	// Entities API
	ENTITY_API      = "entity/"
	ENTITY_BULK_API = "entity/bulk/"

	// Files API
	FILES_API = "files/"

	// Users API
	USER_API = "users"

	// Roles API
	ROLES_API = "roles"

	// Groups API
	GROUP_API = "groups"

	// Tokens API
	TOKENS_API = "apikeys"

	// Workflows API
	WORKFLOW_API                        = "workflows"
	WORKFLOW_INDEX_API                  = "workflows/indexsearch"
	WORKFLOW_INDEX_RUN_API              = "runs/indexsearch"
	SCHEDULE_QUERY_WORKFLOWS_SEARCH_API = "runs/cron/scheduleQueriesBetweenDuration"
	SCHEDULE_QUERY_WORKFLOWS_MISSED_API = "runs/cron/missedScheduleQueriesBetweenDuration"
	WORKFLOW_OWNER_RERUN_API            = "workflows/triggerAsOwner"
	WORKFLOW_RERUN_API                  = "workflows/submit"
	WORKFLOW_RUN_API                    = "workflows?submit=true"
	WORKFLOW_SCHEDULE_RUN               = "runs"
)

// API defines the structure of an API call.
type API struct {
	Path     string
	Method   string
	Status   int
	Endpoint Endpoint
	Consumes string
	Produces string
}

type Endpoint struct {
	Atlas string
}

var AtlasEndpoint = Endpoint{
	Atlas: "/api/meta/",
}

var HeraclesEndpoint = Endpoint{
	Atlas: "/api/service/",
}

// API calls to various services (Atlas, Heracles etc)
var (
	GET_TYPEDEF_BY_NAME = API{
		Path:     TYPEDEF_BY_NAME,
		Method:   http.MethodGet,
		Status:   http.StatusOK,
		Endpoint: AtlasEndpoint,
	}

	GET_TYPEDEF_BY_GUID = API{
		Path:     TYPEDEF_BY_GUID,
		Method:   http.MethodGet,
		Status:   http.StatusOK,
		Endpoint: AtlasEndpoint,
	}

	GET_ALL_TYPE_DEFS = API{
		Path:     TYPEDEFS_API,
		Method:   http.MethodGet,
		Status:   http.StatusOK,
		Endpoint: AtlasEndpoint,
	}

	GET_ALL_TYPE_DEF_HEADERS = API{
		Path:     TYPEDEFS_API + "headers",
		Method:   http.MethodGet,
		Status:   http.StatusOK,
		Endpoint: AtlasEndpoint,
	}

	UPDATE_TYPE_DEFS = API{
		Path:     TYPEDEFS_API,
		Method:   http.MethodPut,
		Status:   http.StatusOK,
		Endpoint: AtlasEndpoint,
	}

	CREATE_TYPE_DEFS = API{
		Path:     TYPEDEFS_API,
		Method:   http.MethodPut,
		Status:   http.StatusOK,
		Endpoint: AtlasEndpoint,
	}

	DELETE_TYPE_DEFS = API{
		Path:     TYPEDEFS_API,
		Method:   http.MethodDelete,
		Status:   http.StatusNoContent,
		Endpoint: AtlasEndpoint,
	}

	DELETE_TYPE_DEF_BY_NAME = API{
		Path:     TYPEDEF_BY_NAME,
		Method:   http.MethodDelete,
		Status:   http.StatusNoContent,
		Endpoint: AtlasEndpoint,
	}

	GET_ENTITY_BY_GUID = API{
		Path:     ENTITY_API + "guid/",
		Method:   http.MethodGet,
		Status:   http.StatusOK,
		Endpoint: AtlasEndpoint,
	}

	GET_ENTITY_BY_UNIQUE_ATTRIBUTE = API{
		Path:     ENTITY_API + "uniqueAttribute/type/",
		Method:   http.MethodGet,
		Status:   http.StatusOK,
		Endpoint: AtlasEndpoint,
	}

	INDEX_SEARCH = API{
		Path:     "search/indexsearch/",
		Method:   http.MethodPost,
		Status:   http.StatusOK,
		Endpoint: AtlasEndpoint,
	}

	CREATE_ENTITY = API{
		Path:     ENTITY_API,
		Method:   http.MethodPost,
		Status:   http.StatusOK,
		Endpoint: AtlasEndpoint,
	}

	CREATE_ENTITIES = API{
		Path:     ENTITY_BULK_API,
		Method:   http.MethodPost,
		Status:   http.StatusOK,
		Endpoint: AtlasEndpoint,
	}

	DELETE_ENTITIES_BY_GUIDS = API{
		Path:     ENTITY_BULK_API,
		Method:   http.MethodDelete,
		Status:   http.StatusOK,
		Endpoint: AtlasEndpoint,
	}

	PRESIGNED_URL = API{
		Path:     FILES_API + "presignedUrl",
		Method:   http.MethodPost,
		Status:   http.StatusOK,
		Endpoint: HeraclesEndpoint,
	}

	UPDATE_ENTITY_BY_ATTRIBUTE = API{
		Path:     ENTITY_API + "uniqueAttribute/type/",
		Method:   http.MethodPost,
		Status:   http.StatusNoContent,
		Endpoint: AtlasEndpoint,
	}

	PARTIAL_UPDATE_ENTITY_BY_ATTRIBUTE = API{
		Path:     ENTITY_API + "uniqueAttribute/type/",
		Method:   http.MethodPut,
		Status:   http.StatusOK,
		Endpoint: AtlasEndpoint,
	}

	DELETE_ENTITY_BY_ATTRIBUTE = API{
		Path:     ENTITY_API + "uniqueAttribute/type/",
		Method:   http.MethodDelete,
		Status:   http.StatusNoContent,
		Endpoint: AtlasEndpoint,
	}

	// Users API

	CREATE_USERS = API{
		Path:     USER_API,
		Method:   http.MethodPost,
		Status:   http.StatusOK,
		Endpoint: HeraclesEndpoint,
	}

	GET_USERS = API{
		Path:     USER_API,
		Method:   http.MethodGet,
		Status:   http.StatusOK,
		Endpoint: HeraclesEndpoint,
	}

	UPDATE_USERS = API{
		Path:     USER_API,
		Method:   http.MethodPost,
		Status:   http.StatusOK,
		Endpoint: HeraclesEndpoint,
	}

	DELETE_USER = API{
		Path:     USER_API + "/%s/delete",
		Method:   http.MethodPost,
		Status:   http.StatusOK,
		Endpoint: HeraclesEndpoint,
	}

	GET_USER_GROUPS = API{
		Path:     USER_API + "/%s/groups",
		Method:   http.MethodGet,
		Status:   http.StatusOK,
		Endpoint: HeraclesEndpoint,
	}

	ADD_USER_TO_GROUPS = API{
		Path:     USER_API + "/%s/groups",
		Method:   http.MethodPost,
		Status:   http.StatusOK,
		Endpoint: HeraclesEndpoint,
	}

	CHANGE_USER_ROLE = API{
		Path:     USER_API + "/%s/roles/update",
		Method:   http.MethodPost,
		Status:   http.StatusOK,
		Endpoint: HeraclesEndpoint,
	}

	GET_CURRENT_USER = API{
		Path:     USER_API + "/current",
		Method:   http.MethodGet,
		Status:   http.StatusOK,
		Endpoint: HeraclesEndpoint,
	}

	// Roles APIs

	GET_ROLES = API{
		Path:     ROLES_API,
		Method:   http.MethodGet,
		Status:   http.StatusOK,
		Endpoint: HeraclesEndpoint,
	}

	// Group APIs

	GET_GROUPS = API{
		Path:     GROUP_API,
		Method:   http.MethodGet,
		Status:   http.StatusOK,
		Endpoint: HeraclesEndpoint,
	}

	CREATE_GROUP = API{
		Path:     GROUP_API,
		Method:   http.MethodPost,
		Status:   http.StatusOK,
		Endpoint: HeraclesEndpoint,
	}

	UPDATE_GROUP = API{
		Path:     GROUP_API,
		Method:   http.MethodPost,
		Status:   http.StatusOK,
		Endpoint: HeraclesEndpoint,
	}

	DELETE_GROUP = API{
		Path:     GROUP_API + "/%s/delete",
		Method:   http.MethodPost,
		Status:   http.StatusOK,
		Endpoint: HeraclesEndpoint,
	}

	GET_GROUP_MEMBERS = API{
		Path:     GROUP_API + "/%s/members",
		Method:   http.MethodGet,
		Status:   http.StatusOK,
		Endpoint: HeraclesEndpoint,
	}

	REMOVE_USERS_FROM_GROUP = API{
		Path:     GROUP_API + "/%s/members/remove",
		Method:   http.MethodPost,
		Status:   http.StatusOK,
		Endpoint: HeraclesEndpoint,
	}

	// Token APIs

	GET_API_TOKENS = API{
		Path:     TOKENS_API,
		Method:   http.MethodGet,
		Status:   http.StatusOK,
		Endpoint: HeraclesEndpoint,
	}

	UPSERT_API_TOKEN = API{
		Path:     TOKENS_API,
		Method:   http.MethodPost,
		Status:   http.StatusOK,
		Endpoint: HeraclesEndpoint,
	}

	DELETE_API_TOKEN = API{
		Path:     TOKENS_API,
		Method:   http.MethodDelete,
		Status:   http.StatusOK,
		Endpoint: HeraclesEndpoint,
	}

	// Workflows

	SCHEDULE_QUERY_WORKFLOWS_SEARCH = API{
		Path:     SCHEDULE_QUERY_WORKFLOWS_SEARCH_API,
		Method:   http.MethodGet,
		Status:   http.StatusOK,
		Endpoint: HeraclesEndpoint,
	}

	SCHEDULE_QUERY_WORKFLOWS_MISSED = API{
		Path:     SCHEDULE_QUERY_WORKFLOWS_MISSED_API,
		Method:   http.MethodGet,
		Status:   http.StatusOK,
		Endpoint: HeraclesEndpoint,
	}

	WORKFLOW_INDEX_SEARCH = API{
		Path:     WORKFLOW_INDEX_API,
		Method:   http.MethodPost,
		Status:   http.StatusOK,
		Endpoint: HeraclesEndpoint,
	}

	WORKFLOW_INDEX_RUN_SEARCH = API{
		Path:     WORKFLOW_INDEX_RUN_API,
		Method:   http.MethodPost,
		Status:   http.StatusOK,
		Endpoint: HeraclesEndpoint,
	}

	// triggers a workflow using the current user's credentials

	WORKFLOW_RERUN = API{
		Path:     WORKFLOW_RUN_API,
		Method:   http.MethodPost,
		Status:   http.StatusOK,
		Endpoint: HeraclesEndpoint,
	}

	// triggers a workflow using the workflow owner's credentials

	WORKFLOW_OWNER_RERUN = API{
		Path:     WORKFLOW_OWNER_RERUN_API,
		Method:   http.MethodPost,
		Status:   http.StatusOK,
		Endpoint: HeraclesEndpoint,
	}

	WORKFLOW_UPDATE = API{
		Path:     WORKFLOW_API + "/%s",
		Method:   http.MethodPost,
		Status:   http.StatusOK,
		Endpoint: HeraclesEndpoint,
	}

	WORKFLOW_ARCHIVE = API{
		Path:     WORKFLOW_API + "/%s/archive",
		Method:   http.MethodPost,
		Status:   http.StatusOK,
		Endpoint: HeraclesEndpoint,
	}

	GET_ALL_SCHEDULE_RUNS = API{
		Path:     WORKFLOW_SCHEDULE_RUN + "/cron",
		Method:   http.MethodGet,
		Status:   http.StatusOK,
		Endpoint: HeraclesEndpoint,
	}

	GET_SCHEDULE_RUN = API{
		Path:     WORKFLOW_SCHEDULE_RUN + "/cron/%s",
		Method:   http.MethodGet,
		Status:   http.StatusOK,
		Endpoint: HeraclesEndpoint,
	}

	STOP_WORKFLOW_RUN = API{
		Path:     WORKFLOW_SCHEDULE_RUN + "/%s/stop",
		Method:   http.MethodPost,
		Status:   http.StatusOK,
		Endpoint: HeraclesEndpoint,
	}

	WORKFLOW_CHANGE_OWNER = API{
		Path:     WORKFLOW_API + "/%s/changeownership",
		Method:   http.MethodPost,
		Status:   http.StatusOK,
		Endpoint: HeraclesEndpoint,
	}

	WORKFLOW_RUN = API{
		Path:     WORKFLOW_RUN_API,
		Method:   http.MethodPost,
		Status:   http.StatusOK,
		Endpoint: HeraclesEndpoint,
	}
)

// Constants for the Atlas search DSL
const (
	// TermAttributes Constants
	CONNECTOR_NAME           = "connectorName"
	CATEGORIES               = "__categories"
	CREATE_TIME_AS_TIMESTAMP = "__timestamp"
	CREATED_BY               = "__createdBy"
	GLOSSARY                 = "__glossary"
	GUID                     = "__guid"
	HAS_LINEAGE              = "__hasLineage"
	MEANINGS                 = "__meanings"
	MODIFIED_BY              = "__modifiedBy"
	NAME                     = "name.keyword"
	OWNER_GROUPS             = "ownerGroups"
	OWNER_USERS              = "ownerUsers"
	PARENT_CATEGORY          = "__parentCategory"
	POPULARITY_SCORE         = "popularityScore"
	QUALIFIED_NAME           = "qualifiedName"
	STATE                    = "__state"
	SUPER_TYPE_NAMES         = "__superTypeNames.keyword"
	TYPE_NAME                = "__typeName.keyword"
	UPDATE_TIME_AS_TIMESTAMP = "__modificationTimestamp"
	CERTIFICATE_STATUS       = "certificateStatus"

	// TextAttributes Constants
	CLASSIFICATION_NAMES            = "__classificationNames"
	CLASSIFICATIONS_TEXT            = "__classificationsText"
	CREATE_TIME_AS_DATE             = "__timestamp.date"
	DESCRIPTION                     = "description"
	MEANINGS_TEXT                   = "__meaningsText"
	NAME_TEXT                       = "name"
	QUALIFIED_NAME_TEXT             = "qualifiedName.text"
	PROPAGATED_CLASSIFICATION_NAMES = "__propagatedClassificationNames"
	PROPAGATED_TRAIT_NAMES          = "__propagatedTraitNames"
	SUPER_TYPE_NAMES_TEXT           = "__superTypeNames"
	TRAIT_NAMES                     = "__traitNames"
	UPDATE_TIME_AS_DATE             = "__modificationTimestamp.date"
	USER_DESCRIPTION                = "userDescription"
)

// FormatPathWithParams returns a new API object with the path formatted by joining the provided parameters.
func (api *API) FormatPathWithParams(params ...string) (*API, error) {
	// Join the base path with the additional params
	requestPath, err := MultipartURLJoin(api.Path, params...)
	if err != nil {
		return nil, fmt.Errorf("failed to join URL parts: %w", err)
	}

	// Return a new API object with the formatted path
	return &API{
		Path:     requestPath,
		Method:   api.Method,
		Status:   api.Status,
		Endpoint: api.Endpoint,
		Consumes: api.Consumes,
		Produces: api.Produces,
	}, nil
}

// MultipartURLJoin joins the base path with the provided segments.
func MultipartURLJoin(basePath string, params ...string) (string, error) {
	// Parse the base path as a URL
	u, err := url.Parse(basePath)
	if err != nil {
		return "", fmt.Errorf("invalid base path: %w", err)
	}

	// Join additional path segments
	u.Path = path.Join(u.Path, path.Join(params...))

	// Return the final formatted URL
	return u.String(), nil
}
