package v1alpha1

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGroupUpdateStatus(t *testing.T) {
	t.Parallel()

	t.Run("successful", func(t *testing.T) {
		t.Parallel()
		g := &Group{ObjectMeta: metav1.ObjectMeta{Generation: 4}}
		g.UpdateStatus(SuccessfullyReconciled, "")

		require.Len(t, g.Status.Conditions, 1)
		cond := g.Status.Conditions[0]
		assert.Equal(t, GroupReadyCondition, cond.Type)
		assert.Equal(t, metav1.ConditionTrue, cond.Status)
		assert.Equal(t, SuccessfullyReconciled, cond.Reason)
		assert.Equal(t, "Group reconciled successfully", cond.Message)
		assert.Equal(t, int64(4), g.Status.LastAppliedGeneration)
	})

	t.Run("partially reconciled", func(t *testing.T) {
		t.Parallel()
		g := &Group{ObjectMeta: metav1.ObjectMeta{Generation: 5}}
		g.UpdateStatus(PartiallyReconciled, "Group partially reconciled: 1 user(s) not found or failed: missing-user")

		require.Len(t, g.Status.Conditions, 1)
		cond := g.Status.Conditions[0]
		assert.Equal(t, GroupReadyCondition, cond.Type)
		assert.Equal(t, metav1.ConditionTrue, cond.Status)
		assert.Equal(t, PartiallyReconciled, cond.Reason)
		assert.Equal(t, "Group partially reconciled: 1 user(s) not found or failed: missing-user", cond.Message)
		assert.Equal(t, int64(5), g.Status.LastAppliedGeneration)
	})

	t.Run("failed", func(t *testing.T) {
		t.Parallel()
		g := &Group{ObjectMeta: metav1.ObjectMeta{Generation: 6}}
		g.SetWaiting()
		g.UpdateStatus(ReconcileFailed, "")

		require.Len(t, g.Status.Conditions, 1)
		cond := g.Status.Conditions[0]
		assert.Equal(t, GroupReadyCondition, cond.Type)
		assert.Equal(t, metav1.ConditionFalse, cond.Status)
		assert.Equal(t, ReconcileFailed, cond.Reason)
		assert.Equal(t, "Group reconcile failed", cond.Message)
		assert.Equal(t, int64(0), g.Status.LastAppliedGeneration)
	})
}
