package v1alpha1

const (
	SuccessfullyReconciled   = "SuccessfullyReconciled"
	ReconcileFailed          = "ReconcileFailed"
	MissingSubGroupsReason   = "MissingSubGroups"
	MembersResolvedCondition = "MembersResolved"

	maxConditionMessageLen = 32768
)
