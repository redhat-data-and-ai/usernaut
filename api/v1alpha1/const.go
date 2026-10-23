package v1alpha1

const (
	SuccessfullyReconciled   = "SuccessfullyReconciled"
	ReconcileFailed          = "ReconcileFailed"
	MissingSubGroupsReason   = "MissingSubGroups"
	MembersResolvedCondition = "MembersResolved"

	// MaxLDAPQueryDepth is the maximum nesting depth allowed for ldap_query filters.
	MaxLDAPQueryDepth = 4

	maxConditionMessageLen = 32768
)
