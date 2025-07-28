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

// GroupSpec defines the desired state of Group
type GroupSpec struct {
	GroupName string    `json:"group_name"`
	Members   []string  `json:"members"`
	Backends  []Backend `json:"backends"`
}

// GroupStatus defines the observed state of Group
type GroupStatus struct {
	Conditions            []metav1.Condition `json:"conditions,omitempty"`
	LastAppliedGeneration int64              `json:"lastAppliedGeneration,omitempty"`
	BackendsStatus        []BackendStatus    `json:"backends,omitempty"`
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
	condition := metav1.Condition{
		Type:               GroupReadyCondition,
		LastTransitionTime: metav1.Now(),
		Status:             metav1.ConditionUnknown,
		Message:            "Group is getting reconciled",
		Reason:             "Waiting",
	}
	for i, currentCondition := range c.Status.Conditions {
		if currentCondition.Type == condition.Type {
			c.Status.Conditions[i] = condition
			return
		}
	}
	c.Status.Conditions = append(c.Status.Conditions, condition)
}

func (c *Group) UpdateStatus(isError bool) {
	condition := metav1.Condition{
		Type:               GroupReadyCondition,
		LastTransitionTime: metav1.Now(),
	}
	if !isError {
		condition.Status = metav1.ConditionTrue
		condition.Message = "Group reconciled successfully"
		condition.Reason = SuccessfullyReconciled

		c.Status.LastAppliedGeneration = c.Generation
	} else {
		condition.Status = metav1.ConditionFalse
		condition.Message = "Group reconcile failed"
		condition.Reason = ReconcileFailed
	}
	for i, currentCondition := range c.Status.Conditions {
		if currentCondition.Type == condition.Type {
			c.Status.Conditions[i] = condition
			return
		}
	}
	c.Status.Conditions = append(c.Status.Conditions, condition)
}
