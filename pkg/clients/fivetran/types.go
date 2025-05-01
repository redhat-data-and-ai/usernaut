package fivetran

const (
	AccountReviewerRole  = "Account Reviewer"
	ConnectorAdminRole   = "Connector Administrator"
	ConnectorCreatorRole = "Connector Creator"
)

type UpdateTeam struct {
	ExistingTeamID string
	NewTeamName    string
	NewRole        string
	NewDescription string
}
