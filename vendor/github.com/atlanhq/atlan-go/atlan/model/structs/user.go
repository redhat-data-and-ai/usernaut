package structs

// AtlanUser represents an Atlan user.
type AtlanUser struct {
	Asset
	ID             string          `json:"id"`                      // Unique identifier (GUID) of the user within Atlan.
	Username       *string         `json:"username,omitempty"`      // Username of the user within Atlan.
	WorkspaceRole  string          `json:"workspaceRole"`           // Name of the role of the user within Atlan.
	Email          string          `json:"email"`                   // Email address of the user.
	EmailVerified  *bool           `json:"emailVerified,omitempty"` // When true, the email address of the user has been verified.
	Enabled        *bool           `json:"enabled,omitempty"`       // When true, the user is enabled. When false, the user has been deactivated.
	FirstName      *string         `json:"firstName,omitempty"`     // First name of the user.
	LastName       *string         `json:"lastName,omitempty"`      // Last name (surname) of the user.
	Attributes     *UserAttributes `json:"attributes,omitempty"`    // Detailed attributes of the user.
	CreatedAt      *string         `json:"CreatedAt,omitempty"`     // Time (epoch) at which the user was created, in milliseconds.
	LastLoginTime  *int            `json:"lastLoginTime,omitempty"` // Time (epoch) at which the user last logged into Atlan
	GroupCount     *int            `json:"groupCount,omitempty"`    // Number of groups to which the user belongs
	DefaultRoles   *[]string       `json:"defaultRoles,omitempty"`
	Roles          *[]string       `json:"roles,omitempty"`
	DecentralRoles interface{}     `json:"decentralizedRoles,omitempty"` // TBD
	Personas       *[]Persona      `json:"personas,omitempty"`           // Personas the user is associated with.
	Purposes       []interface{}   `json:"purposes,omitempty"`           // TBD
	AdminEvents    *[]AdminEvent   `json:"adminEvents,omitempty"`        // List of administration-related events for this user.
	LoginEvents    *[]LoginEvent   `json:"loginEvents,omitempty"`        // List of login-related events for this user.
}

// UserAttributes represents detailed attributes of an Atlan user.
type UserAttributes struct {
	// Designation for the user, such as an honorific or title.
	Designation *[]string `json:"designation,omitempty"`
	// Skills the user possesses.
	Skills *[]string `json:"skills,omitempty"`
	// Unique Slack member identifier.
	Slack *[]string `json:"slack,omitempty"`
	// Unique JIRA user identifier.
	Jira *[]string `json:"jira,omitempty"`
	// Time at which the user was invited (as a formatted string).
	InvitedAt *[]string `json:"invitedAt,omitempty"`
	// User who invited this user.
	InvitedBy *[]string `json:"invitedBy,omitempty"`
	ByName    *[]string `json:"invitedByName,omitempty"`
}

// LoginEvent represents a login event for an Atlan user.
type LoginEvent struct {
	ClientID  *string `json:"clientID,omitempty"` // Where the login occurred (usually `atlan-frontend`).
	Details   any     `json:"details,omitempty"`
	IPAddress *string `json:"ipAddress,omitempty"` // IP address from which the user logged in.
	RealmID   *string `json:"realmID,omitempty"`
	SessionID *string `json:"sessionID,omitempty"` // Unique identifier (GUID) of the session for the login.
	Time      *int64  `json:"time,omitempty"`      // Time (epoch) when the login occurred, in milliseconds.
	Type      *string `json:"type,omitempty"`      // Type of login event that occurred (usually `LOGIN`).
	UserID    *string `json:"userID,omitempty"`    // Unique identifier (GUID) of the user that logged in.
}

// AuthDetails represents authentication details for admin operations.
type AuthDetails struct {
	ClientID  *string `json:"clientID,omitempty"`
	IPAddress *string `json:"ipAddress,omitempty"`
	RealmID   *string `json:"realmID,omitempty"`
	UserID    *string `json:"userID,omitempty"`
}

// AdminEvent represents an admin operation event for an Atlan user.
type AdminEvent struct {
	OperationType  *string      `json:"operationType,omitempty"` // Type of admin operation that occurred.
	RealmID        *string      `json:"realmID,omitempty"`
	Representation *string      `json:"representation,omitempty"`
	ResourcePath   *string      `json:"resourcePath,omitempty"`
	ResourceType   *string      `json:"resourceType,omitempty"` // Type of resource for the admin operation that occurred.
	Time           *int64       `json:"time,omitempty"`         // Time (epoch) when the admin operation occurred, in milliseconds.
	AuthDetails    *AuthDetails `json:"authDetails,omitempty"`
}

type UserRequest struct {
	MaxLoginEvents int      `json:"maxLoginEvents,omitempty"` // Maximum login events to include
	PostFilter     *string  `json:"post_filter,omitempty"`    // Filter criteria for the user list
	Sort           *string  `json:"sort,omitempty"`           // Property to sort the list of users
	Count          bool     `json:"count"`                    // Whether to include a count of users
	Offset         int      `json:"offset,omitempty"`         // Starting point for paging
	Limit          int      `json:"limit,omitempty"`          // Maximum number of users per page
	Columns        []string `json:"columns,omitempty"`        // List of columns to be returned in the response
}

// QueryParams converts the UserRequest to a map of query parameters.
func (r *UserRequest) QueryParams() map[string]interface{} {
	qp := make(map[string]interface{})

	if r.PostFilter != nil && *r.PostFilter != "" {
		qp["filter"] = *r.PostFilter
	}
	if r.Sort != nil && *r.Sort != "" {
		qp["sort"] = *r.Sort
	}
	if len(r.Columns) > 0 {
		qp["columns"] = r.Columns
	}
	qp["count"] = r.Count
	qp["offset"] = r.Offset
	qp["limit"] = r.Limit
	qp["maxLoginEvents"] = r.MaxLoginEvents

	return qp
}
