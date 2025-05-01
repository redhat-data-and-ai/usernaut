package fivetran

const (
	AccountReviewerRole  = "Account Reviewer"
	ConnectorAdminRole   = "Connector Administrator"
	ConnectorCreatorRole = "Connector Creator"
)

type Team struct {
	TeamID      string
	TeamName    string
	Role        string
	Description string
}

type UpdateTeam struct {
	ExistingTeamID string
	NewTeamName    string
	NewRole        string
	NewDescription string
}
