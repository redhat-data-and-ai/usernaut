package v1alpha1

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGroupUpdateValidationFailed(t *testing.T) {
	group := &Group{}
	group.UpdateStatusWithErrMessage(`spec.group_name must start with "aif-"`)

	if len(group.Status.Conditions) != 1 {
		t.Fatalf("expected 1 condition, got %d", len(group.Status.Conditions))
	}

	condition := group.Status.Conditions[0]
	if condition.Type != GroupReadyCondition {
		t.Fatalf("expected type %q, got %q", GroupReadyCondition, condition.Type)
	}
	if condition.Status != metav1.ConditionFalse {
		t.Fatalf("expected status False, got %q", condition.Status)
	}
	if condition.Reason != ReconcileFailed {
		t.Fatalf("expected reason %q, got %q", ReconcileFailed, condition.Reason)
	}
	if condition.Message != `spec.group_name must start with "aif-"` {
		t.Fatalf("unexpected message: %q", condition.Message)
	}
}
