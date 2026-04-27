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
	"context"
	"fmt"
	"os"
	"strings"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	k8stypes "k8s.io/apimachinery/pkg/types"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

const (
	roverBackendType           = "rover"
	requiredServiceAccountName = "usernaut-prod"
	defaultNamespace           = "usernaut"
)

func getWatchedNamespace() string {
	if ns := os.Getenv("WATCHED_NAMESPACE"); ns != "" {
		return ns
	}
	return defaultNamespace
}

// validateRoverBackend checks that when any backend has type "rover",
// the required ServiceAccount exists in the watched namespace.
func (g *Group) validateRoverBackend(ctx context.Context) error {
	hasRover := false
	for _, b := range g.Spec.Backends {
		if strings.EqualFold(b.Type, roverBackendType) {
			hasRover = true
			break
		}
	}
	if !hasRover {
		return nil
	}

	c := ClientFromContext(ctx)
	if c == nil {
		return fmt.Errorf("internal error: webhook client not available")
	}

	ns := getWatchedNamespace()
	sa := &corev1.ServiceAccount{}
	err := c.Get(ctx, k8stypes.NamespacedName{
		Name:      requiredServiceAccountName,
		Namespace: ns,
	}, sa)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return fmt.Errorf(
				"backend type %q requires ServiceAccount %q in namespace %q, but it was not found",
				roverBackendType, requiredServiceAccountName, ns,
			)
		}
		return fmt.Errorf("failed to verify ServiceAccount %q in namespace %q: %w",
			requiredServiceAccountName, ns, err,
		)
	}
	return nil
}

func (g *Group) ValidateCreate(ctx context.Context) (admission.Warnings, error) {
	log := logf.FromContext(ctx)
	log.Info("Group.ValidateCreate called", "name", g.Name)

	if err := g.validateRoverBackend(ctx); err != nil {
		return nil, err
	}
	return nil, nil
}

func (g *Group) ValidateUpdate(ctx context.Context, _ runtime.Object) (admission.Warnings, error) {
	log := logf.FromContext(ctx)
	log.Info("Group.ValidateUpdate called", "name", g.Name)

	if err := g.validateRoverBackend(ctx); err != nil {
		return nil, err
	}
	return nil, nil
}

func (g *Group) ValidateDelete(ctx context.Context) (admission.Warnings, error) {
	log := logf.FromContext(ctx)
	log.Info("Group.ValidateDelete called", "name", g.Name)
	return nil, nil
}
