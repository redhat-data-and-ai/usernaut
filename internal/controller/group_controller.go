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

package controller

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	usernautdevv1alpha1 "github.com/redhat-data-and-ai/usernaut/api/v1alpha1"
	"github.com/redhat-data-and-ai/usernaut/pkg/logger"
	"github.com/sirupsen/logrus"
)

// GroupReconciler reconciles a Group object
type GroupReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=usernaut.dev,resources=groups,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=usernaut.dev,resources=groups/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=usernaut.dev,resources=groups/finalizers,verbs=update

func (r *GroupReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	ctx = logger.WithRequestId(ctx, controller.ReconcileIDFromContext(ctx))
	log := logger.Logger(ctx).WithFields(logrus.Fields{
		"request": req.NamespacedName.String(),
	})

	log.Info("reconciling the group")

	groupCR := &usernautdevv1alpha1.Group{}
	if err := r.Client.Get(ctx, req.NamespacedName, groupCR); err != nil {
		log.Error(err, "error fetching the group CR")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *GroupReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&usernautdevv1alpha1.Group{}).
		WithEventFilter(predicate.GenerationChangedPredicate{}).
		Complete(r)
}
