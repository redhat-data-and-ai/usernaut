package structs

// AtlanRole represents a role in Atlan.
type AtlanRole struct {
	// Unique identifier for the role (GUID).
	ID *string `json:"id,omitempty"`
	// Unique name for the role.
	Name *string `json:"name"`
	// Description of the role.
	Description *string `json:"description,omitempty"`
	ClientRole  *bool   `json:"client_role,omitempty"`
	Level       *string `json:"level,omitempty"`
	// Number of users with this role.
	MemberCount *string `json:"member_count,omitempty"`
	UserCount   *string `json:"user_count,omitempty"`
}

// RoleResponse represents the response containing a list of roles in Atlan.
type RoleResponse struct {
	// Total number of roles.
	TotalRecord *int `json:"total_record,omitempty"`
	// Number of roles in the filtered response.
	FilterRecord *int `json:"filter_record,omitempty" `
	// Details of each role included in the response.
	Records *[]AtlanRole `json:"records" `
}
