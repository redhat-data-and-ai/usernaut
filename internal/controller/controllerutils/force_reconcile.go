package controllerutils

import (
	"context"

	"github.com/redhat-data-and-ai/usernaut/pkg/common/constants"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

func ForceReconcilePredicate() predicate.Predicate {
	return CustomLabelKeyChangedPredicate{LabelKey: constants.ForceReconcileLabel}
}

// Custom Predicate to filter by a specific label key
type CustomLabelKeyChangedPredicate struct {
	LabelKey string
	predicate.Funcs
}

// Custom Predicate label to force reconciliation on label addition
func (p CustomLabelKeyChangedPredicate) Update(e event.UpdateEvent) bool {
	if e.ObjectOld == nil || e.ObjectNew == nil {
		return false
	}

	oldLabels := e.ObjectOld.GetLabels()
	newLabels := e.ObjectNew.GetLabels()

	_, oldExists := oldLabels[p.LabelKey]
	_, newExists := newLabels[p.LabelKey]

	// Trigger reconciliation only if the label is added
	if !oldExists && newExists {
		return true
	}

	return false
}

func RemoveForceReconcileLabel(ctx context.Context, c client.Client, obj client.Object) error {
	labels := obj.GetLabels()
	// if there are no labels, there is nothing to do, return nil
	if labels == nil {
		return nil
	}

	// if the force reconcile label is not present, return nil, nothing to do here
	_, ok := labels[constants.ForceReconcileLabel]
	if !ok {
		return nil
	}

	// if the force reconcile label is present, remove it and update the object
	delete(labels, constants.ForceReconcileLabel)
	obj.SetLabels(labels)
	return c.Update(ctx, obj)
}
