/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	GroupReadyCondition = "GroupReadyCondition"
)

type BackendStatus struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Status  bool   `json:"status"`
	Message string `json:"message"`
}

type Backend struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type LDAPFilter struct {
	// +kubebuilder:validation:Enum=givenName;displayName;rhatJobTitle;title;employeeType;manager;rhatCostCenter;rhatCostCenterDesc;rhatGeo;co;st;rhatLocation;rhatOfficeLocation;rhatOfficeFloor;roomNumber
	Key string `json:"key"`
	// +kubebuilder:validation:Enum=equals;contains;not
	Criteria string `json:"criteria"`
	Value    string `json:"value"`
}

type LDAPQuery struct {
	// +kubebuilder:validation:Enum=and;or
	Operator string       `json:"operator"`
	Filters  []LDAPFilter `json:"filters"`
	Options  *LDAPOptions `json:"options,omitempty"`
}

type LDAPOptions struct {
	IncludeIndirectReports bool `json:"include_indirect_reports,omitempty"`
	IncludeManager         bool `json:"include_manager,omitempty"`
}

// GroupSpec defines the desired state of Group
type GroupSpec struct {
	GroupName   string       `json:"group_name"`
	Members     Members      `json:"members"`
	GroupParams []GroupParam `json:"group_params,omitempty"`
	Backends    []Backend    `json:"backends"`
}

// Members defines how group membership is resolved. At least one of ldap_query,
// a non-empty users list, or a non-empty groups list must be provided.
//
// +kubebuilder:validation:XValidation:rule="has(self.ldap_query) || (has(self.users) && size(self.users) > 0) || (has(self.groups) && size(self.groups) > 0)",message="users or groups must be a non-empty list when ldap_query is omitted"
type Members struct {
	Groups    []string   `json:"groups,omitempty"`
	Users     []string   `json:"users,omitempty"`
	LDAPQuery *LDAPQuery `json:"ldap_query,omitempty"`
}

type GroupParam struct {
	Backend  string `json:"backend"`
	Name     string `json:"name"`
	Property string `json:"property"`
	// +kubebuilder:validation:MinItems=1
	Value []string `json:"value"`
}

// GroupStatus defines the observed state of Group
type GroupStatus struct {
	ReconciledUsers       []string           `json:"reconciledUsers,omitempty"`
	Conditions            []metav1.Condition `json:"conditions,omitempty"`
	LastAppliedGeneration int64              `json:"lastAppliedGeneration,omitempty"`
	BackendsStatus        []BackendStatus    `json:"backends,omitempty"`
	MissingSubGroups      []string           `json:"missingSubGroups,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.conditions[?(@.type=="GroupReadyCondition")].status`
// +kubebuilder:printcolumn:name="Message",type=string,JSONPath=`.status.conditions[?(@.type=="GroupReadyCondition")].message`

// Group is the Schema for the groups API
type Group struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GroupSpec   `json:"spec,omitempty"`
	Status GroupStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// GroupList contains a list of Group
type GroupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Group `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Group{}, &GroupList{})
}

func (c *Group) SetWaiting() {
	c.setReadyCondition(metav1.ConditionUnknown, "Waiting", "Group is getting reconciled")
}

func (c *Group) UpdateStatus(isError bool) {
	if isError {
		c.setReadyCondition(metav1.ConditionFalse, ReconcileFailed, "Group reconcile failed")
		return
	}

	c.Status.LastAppliedGeneration = c.Generation
	if len(c.Status.MissingSubGroups) > 0 {
		c.setReadyCondition(metav1.ConditionFalse, MissingSubGroupsReason,
			fmt.Sprintf("Reconciled with %d missing sub-groups: %s",
				len(c.Status.MissingSubGroups), strings.Join(c.Status.MissingSubGroups, ", ")))
		return
	}
	c.setReadyCondition(metav1.ConditionTrue, SuccessfullyReconciled, "Group reconciled successfully")
}

func (c *Group) UpdateStatusWithErrMessage(errMessage string) {
	c.setReadyCondition(metav1.ConditionFalse, ReconcileFailed, errMessage)
}

func (c *Group) SetMissingSubGroups(missing []string) {
	c.Status.MissingSubGroups = missing
	if len(missing) == 0 {
		c.setCondition(MembersResolvedCondition, metav1.ConditionTrue, "Resolved", "All sub-groups resolved")
	} else {
		c.setCondition(MembersResolvedCondition, metav1.ConditionFalse, MissingSubGroupsReason,
			fmt.Sprintf("Missing sub-groups: %s", strings.Join(missing, ", ")))
	}
}

func (c *Group) setCondition(condType string, status metav1.ConditionStatus, reason, message string) {
	if len(message) > maxConditionMessageLen {
		message = message[:maxConditionMessageLen-3] + "..."
	}
	condition := metav1.Condition{
		Type: condType, Status: status, Reason: reason,
		Message: message, LastTransitionTime: metav1.Now(),
	}
	for i, cur := range c.Status.Conditions {
		if cur.Type == condType {
			c.Status.Conditions[i] = condition
			return
		}
	}
	c.Status.Conditions = append(c.Status.Conditions, condition)
}

func (c *Group) setReadyCondition(status metav1.ConditionStatus, reason, message string) {
	c.setCondition(GroupReadyCondition, status, reason, message)
}
