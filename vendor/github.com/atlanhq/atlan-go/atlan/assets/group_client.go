package assets

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/atlanhq/atlan-go/atlan/model/structs"
)

type AtlanGroup structs.AtlanGroup

type GroupClient AtlanClient

// Create creates a new Atlan group with the given alias.
func (g *AtlanGroup) Create(alias string) (*AtlanGroup, error) {
	// Generate group name and attributes.
	name, _ := g.GenerateGroupName(alias)
	attributes := &structs.AtlanGroupAttributes{
		Alias: []string{alias},
	}
	return &AtlanGroup{
		Name:       &name,
		Attributes: attributes,
	}, nil
}

// Updater creates an Atlan group for modification.
func (g *AtlanGroup) Updater(guid, path string) (*AtlanGroup, error) {
	return &AtlanGroup{
		ID:   &guid,
		Path: &path,
		Attributes: &structs.AtlanGroupAttributes{
			Alias:       []string{},
			CreatedAt:   []string{},
			CreatedBy:   []string{},
			UpdatedAt:   []string{},
			UpdatedBy:   []string{},
			Description: []string{},
			IsDefault:   []string{},
			Channels:    []string{},
		},
		Alias:              nil,
		DecentralizedRoles: []interface{}{},
		Name:               nil,
		Personas:           []structs.Persona{},
		Purposes:           []interface{}{},
		UserCount:          nil, // Pointer for optional fields
	}, nil
}

// Update updates the details of an existing group.
// The provided `group` must have its ID populated.
func (gc *GroupClient) Update(group *AtlanGroup) error {
	if group.ID == nil {
		return fmt.Errorf("group ID must be populated")
	}

	api := &UPDATE_GROUP
	api.Path = fmt.Sprintf("groups/%s", *group.ID)

	_, err := DefaultAtlanClient.CallAPI(api, nil, group)
	if err != nil {
		return fmt.Errorf("failed to update group: %w", err)
	}
	return nil
}

// Purge deletes a group by its unique identifier (GUID).
func (gc *GroupClient) Purge(guid string) error {
	if guid == "" {
		return fmt.Errorf("GUID cannot be empty")
	}

	requestPayload := map[string]interface{}{}
	api := &DELETE_GROUP
	api.Path = fmt.Sprintf("groups/%s/delete", guid)

	_, err := DefaultAtlanClient.CallAPI(api, nil, requestPayload)
	if err != nil {
		return fmt.Errorf("failed to delete group: %w", err)
	}

	return nil
}

// GenerateGroupName generates a unique internal name for the group based on its alias.
func (g *AtlanGroup) GenerateGroupName(alias string) (string, error) {
	internal := strings.ToLower(alias)
	return strings.ReplaceAll(internal, " ", "_"), nil
}

// CreateGroupRequest represents the request payload for creating a group.
type CreateGroupRequest struct {
	Group *AtlanGroup `json:"group"`
	Users []string    `json:"users,omitempty"`
}

// Create creates a new group in Atlan.
func (gc *GroupClient) Create(group *AtlanGroup, userIDs []string) (*structs.CreateGroupResponse, error) {
	if group == nil {
		return nil, fmt.Errorf("group cannot be nil")
	}

	payload := CreateGroupRequest{
		Group: group,
	}

	// Only add users to the payload if userIDs is not nil or empty
	if len(userIDs) > 0 {
		payload.Users = userIDs
	}

	responseData, err := DefaultAtlanClient.CallAPI(&CREATE_GROUP, nil, payload)
	if err != nil {
		return nil, err
	}

	var response structs.CreateGroupResponse
	if err := json.Unmarshal(responseData, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	return &response, nil
}

// Get retrieves a list of groups with optional filters.
func (gc *GroupClient) Get(limit int, postFilter, sort string, count bool, offset int) (*GroupResponse, error) {
	request := &structs.GroupRequest{
		PostFilter: &postFilter,
		Sort:       sort,
		Count:      count,
		Offset:     offset,
		Limit:      limit,
	}

	queryParams := request.QueryParams()

	responseData, err := DefaultAtlanClient.CallAPI(&GET_GROUPS, queryParams, nil)
	if err != nil {
		return nil, err
	}

	var response GroupResponse
	if err := json.Unmarshal(responseData, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}
	return &response, nil
}

// GetAll retrieves all groups in Atlan.
func (gc *GroupClient) GetAll(limit, offset int, sort string) ([]*AtlanGroup, error) {
	groupResponse, err := gc.Get(limit, "", sort, true, offset)
	if err != nil {
		return nil, err
	}
	return groupResponse.Records, nil
}

// GetByName retrieves groups with names containing the provided alias.
func (gc *GroupClient) GetByName(alias string, limit, offset int) ([]*AtlanGroup, error) {
	postFilter := fmt.Sprintf(`{"$and":[{"alias":{"$ilike":"%%%s%%"}}]}`, alias)
	groupResponse, err := gc.Get(limit, postFilter, "", true, offset)
	if err != nil {
		return nil, err
	}
	return groupResponse.Records, nil
}

// GetMembers retrieves members of a group by GUID.
func (gc *GroupClient) GetMembers(guid string, request *structs.UserRequest) ([]AtlanUser, error) {
	if guid == "" {
		return nil, fmt.Errorf("guid cannot be empty")
	}

	if request == nil {
		request = &structs.UserRequest{}
	}

	api := &GET_GROUP_MEMBERS
	api.Path = fmt.Sprintf("groups/%s/members", guid)

	responseData, err := DefaultAtlanClient.CallAPI(api, request.QueryParams(), nil)
	if err != nil {
		return nil, err
	}

	var response UserResponse
	if err := json.Unmarshal(responseData, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}
	return response.Records, nil
}

// RemoveFromGroupRequest represents the request payload to remove users from a group.
type RemoveFromGroupRequest struct {
	Users []string `json:"users,omitempty"` // List of user GUIDs to remove from the group (optional).
}

// RemoveUsers removes users from a group by GUID.
func (gc *GroupClient) RemoveUsers(guid string, userIDs []string) error {
	if guid == "" {
		return fmt.Errorf("guid cannot be empty")
	}

	request := RemoveFromGroupRequest{
		Users: userIDs,
	}

	api := &REMOVE_USERS_FROM_GROUP
	api.Path = fmt.Sprintf("groups/%s/members/remove", guid)
	_, err := DefaultAtlanClient.CallAPI(api, nil, request)
	return err
}

// GroupResponse

// GroupResponse represents the response containing group details.
type GroupResponse struct {
	Size         int                  // Page size.
	Start        int                  // Starting offset.
	Endpoint     *API                 // API endpoint.
	Client       *AtlanClient         // API client.
	Criteria     structs.GroupRequest // Group request criteria.
	TotalRecords *int                 `json:"totalRecord,omitempty"`  // Total number of groups.
	FilterRecord *int                 `json:"filterRecord,omitempty"` // Number of filtered groups.
	Records      []*AtlanGroup        `json:"records,omitempty"`      // Details of each group.
}

// CurrentPage returns the current page of group records.
func (gr *GroupResponse) CurrentPage() []*AtlanGroup {
	return gr.Records
}

// NextPage fetches the next page of group records.
func (gr *GroupResponse) NextPage(start, size *int) bool {
	if start != nil {
		gr.Start = *start
	} else {
		gr.Start += gr.Size
	}
	if size != nil {
		gr.Size = *size
	}
	return gr.getNextPage()
}

// getNextPage fetches the next page of records from the API.
func (gr *GroupResponse) getNextPage() bool {
	gr.Criteria.Offset = gr.Start
	gr.Criteria.Limit = gr.Size

	queryParams := gr.Criteria.QueryParams()
	rawJSON, err := gr.Client.CallAPI(gr.Endpoint, queryParams, nil)
	if err != nil {
		gr.Records = nil
		return false
	}

	if err := json.Unmarshal(rawJSON, &gr); err != nil {
		fmt.Printf("Error parsing JSON: %v\n", err)
		gr.Records = nil
		return false
	}

	return len(gr.Records) > 0
}

// Iterator for iterating through all group records.
func (gr *GroupResponse) Iterator() <-chan *AtlanGroup {
	ch := make(chan *AtlanGroup)
	go func() {
		defer close(ch)
		for {
			for _, record := range gr.CurrentPage() {
				ch <- record
			}
			if !gr.NextPage(nil, nil) {
				break
			}
		}
	}()
	return ch
}
