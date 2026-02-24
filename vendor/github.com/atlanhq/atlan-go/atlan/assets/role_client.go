package assets

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/atlanhq/atlan-go/atlan/model/structs"
)

type RoleClient struct {
	roleClient *AtlanClient
	structs.AtlanRole
	structs.RoleResponse
}

func NewRoleClient(caller *AtlanClient) *RoleClient {
	return &RoleClient{roleClient: caller}
}

// Get retrieves a RoleResponse containing a list of roles defined in Atlan.
func (r *RoleClient) Get(limit int, postFilter, sort string, count bool, offset int) (*structs.RoleResponse, error) {
	queryParams := map[string]string{
		"count":  strconv.FormatBool(count),
		"offset": strconv.Itoa(offset),
		"limit":  strconv.Itoa(limit),
	}
	if postFilter != "" {
		queryParams["filter"] = postFilter
	}
	if sort != "" {
		queryParams["sort"] = sort
	}

	resp, err := DefaultAtlanClient.CallAPI(&GET_ROLES, queryParams, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch roles: %w", err)
	}

	var roleResponse structs.RoleResponse
	if err := json.Unmarshal(resp, &roleResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal role response: %w", err)
	}
	return &roleResponse, nil
}

// GetAll retrieves all roles defined in Atlan.
func (r *RoleClient) GetAll() (*structs.RoleResponse, error) {
	resp, err := DefaultAtlanClient.CallAPI(&GET_ROLES, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch all roles: %w", err)
	}

	var roleResponse structs.RoleResponse
	if err := json.Unmarshal(resp, &roleResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal role response: %w", err)
	}
	return &roleResponse, nil
}
