package atlan

// AuthPolicyValiditySchedule: Validity schedule struct for policy

// AuthPolicyValiditySchedule represents the struct for AuthPolicyValiditySchedule.
type AuthPolicyValiditySchedule struct {
	Policyvalidityschedulestarttime *string `json:"policyvalidityschedulestarttime,omitempty"`
	Policyvalidityscheduleendtime   *string `json:"policyvalidityscheduleendtime,omitempty"`
	Policyvalidityscheduletimezone  *string `json:"policyvalidityscheduletimezone,omitempty"`
}

// AuthPolicyCondition: Policy condition schedule struct

// AuthPolicyCondition represents the struct for AuthPolicyCondition.
type AuthPolicyCondition struct {
	Policyconditiontype   *string   `json:"policyconditiontype,omitempty"`
	Policyconditionvalues []*string `json:"policyconditionvalues,omitempty"`
}
