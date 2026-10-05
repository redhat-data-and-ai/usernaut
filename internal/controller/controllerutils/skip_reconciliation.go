/*
Copyright 2026.

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

package controllerutils

import (
	"time"

	"github.com/redhat-data-and-ai/usernaut/pkg/common/constants"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ShouldSkipReconciliation returns true when a CR can safely skip expensive
// reconciliation work (LDAP queries, backend sync, etc.).
//
// Design decisions:
//   - This check lives inside Reconcile, NOT in a predicate. Predicates filter
//     events before Reconcile runs, but RequeueAfter timers are only set by
//     Reconcile's return value. If a predicate blocks the startup Create event
//     on pod restart, Reconcile never runs, the requeue timer is never restored,
//     and the CR is permanently orphaned from reconciliation.
//   - The caller MUST still return RequeueAfter even when skipping, to keep the
//     periodic requeue timer alive. Both skip and non-skip paths must return it.
//   - reconciledWithin should be less than the requeue interval (e.g., requeueAfter/2)
//     so that periodic requeues always exceed the window and trigger a full reconcile.
//     This ensures LDAP-query-based groups get their membership refreshed on schedule.
//   - The force-reconcile label is checked by key existence (not value) to match
//     ForceReconcilePredicate, which triggers on key presence regardless of value.
//
// It skips only when ALL of the following are true:
//   - the force-reconcile label is not present
//   - the CR was successfully reconciled within reconciledWithin
//   - lastAppliedGeneration matches the CR's generation (no spec change)
func ShouldSkipReconciliation(obj client.Object, lastAppliedGeneration int64,
	conditions []metav1.Condition, readyConditionType string, reconciledWithin time.Duration) bool {
	if _, exists := obj.GetLabels()[constants.ForceReconcileLabel]; exists {
		return false
	}
	cond := meta.FindStatusCondition(conditions, readyConditionType)
	if cond == nil || cond.Status != metav1.ConditionTrue || time.Since(cond.LastTransitionTime.Time) > reconciledWithin {
		return false
	}
	return lastAppliedGeneration == obj.GetGeneration()
}
