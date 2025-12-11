package structs

import "strings"

// AtlanGroup represents a group in Atlan with detailed attributes.
type AtlanGroup struct {
	Alias              *string               `json:"alias,omitempty"`              // Name of the group as it appears in the UI.
	Attributes         *AtlanGroupAttributes `json:"attributes,omitempty"`         // Detailed attributes of the group.
	DecentralizedRoles []interface{}         `json:"decentralizedRoles,omitempty"` // Decentralized roles associated with the group (TBC).
	ID                 *string               `json:"id,omitempty"`                 // Unique identifier for the group (GUID).
	Name               *string               `json:"name,omitempty"`               // Unique (internal) name for the group.
	Path               *string               `json:"path,omitempty"`               // TBC
	Personas           []Persona             `json:"personas,omitempty"`           // Personas the group is associated with.
	Purposes           []interface{}         `json:"purposes,omitempty"`           // Purposes the group is associated with.
	UserCount          *int                  `json:"userCount,omitempty"`          // Number of users in the group.
}

// AtlanGroupAttributes represents detailed attributes of an Atlan group.
type AtlanGroupAttributes struct {
	Alias       []string `json:"alias,omitempty"`       // Name of the group as it appears in the UI.
	CreatedAt   []string `json:"createdAt,omitempty"`   // Time (epoch) at which the group was created, in milliseconds.
	CreatedBy   []string `json:"createdBy,omitempty"`   // User who created the group.
	UpdatedAt   []string `json:"updatedAt,omitempty"`   // Time (epoch) at which the group was last updated, in milliseconds.
	UpdatedBy   []string `json:"updatedBy,omitempty"`   // User who last updated the group.
	Description []string `json:"description,omitempty"` // Description of the group.
	IsDefault   []string `json:"isDefault,omitempty"`   // Whether this group should be auto-assigned to all new users or not.
	Channels    []string `json:"channels,omitempty"`    // Slack channels for this group.
}

// GroupRequest represents the request for querying groups.
type GroupRequest struct {
	PostFilter *string `json:"postFilter,omitempty"` // Criteria for filtering groups.
	Sort       string  `json:"sort,omitempty"`       // Property to sort groups by (default: "name").
	Count      bool    `json:"count,omitempty"`      // Include the overall count of groups.
	Offset     int     `json:"offset,omitempty"`     // Starting offset for pagination.
	Limit      int     `json:"limit,omitempty"`      // Maximum number of groups to return per page.
}

// QueryParams converts the request into a map of query parameters.
func (gr *GroupRequest) QueryParams() map[string]interface{} {
	qp := map[string]interface{}{
		"count":  gr.Count,
		"offset": gr.Offset,
		"limit":  gr.Limit,
	}
	if gr.PostFilter != nil && *gr.PostFilter != "" {
		qp["filter"] = *gr.PostFilter
	}
	if gr.Sort != "" {
		qp["sort"] = gr.Sort
	}
	return qp
}

// UserStatus represents the status of a user association.
type UserStatus struct {
	Status        *int    // Response code for the association (200 is success).
	StatusMessage *string // Status message for the association ("success" means the association was successful).
}

// WasSuccessful checks if the user association was successful.
func (us *UserStatus) WasSuccessful() bool {
	return (us.Status != nil && *us.Status == 200) ||
		(us.StatusMessage != nil && strings.ToLower(*us.StatusMessage) == "success")
}

// CreateGroupResponse represents the response for creating a group.
type CreateGroupResponse struct {
	Group string                 // Unique identifier (GUID) of the group that was created.
	Users map[string]*UserStatus // Map of user association statuses, keyed by user GUID.
}
