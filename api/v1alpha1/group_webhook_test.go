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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func newScheme() *runtime.Scheme {
	s := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(s)
	_ = AddToScheme(s)
	return s
}

func ctxWithFakeClient(objects ...runtime.Object) context.Context {
	s := newScheme()
	builder := fake.NewClientBuilder().WithScheme(s)
	for _, obj := range objects {
		builder = builder.WithRuntimeObjects(obj)
	}
	c := builder.Build()
	return context.WithValue(context.Background(), webhookClientKey, c)
}

func TestValidateCreate_NoBackends_Allowed(t *testing.T) {
	group := &Group{
		ObjectMeta: metav1.ObjectMeta{Name: "test-group", Namespace: "usernaut"},
		Spec:       GroupSpec{Backends: []Backend{}},
	}
	warnings, err := group.ValidateCreate(ctxWithFakeClient())
	require.NoError(t, err)
	assert.Nil(t, warnings)
}

func TestValidateCreate_NonRoverBackend_Allowed(t *testing.T) {
	group := &Group{
		ObjectMeta: metav1.ObjectMeta{Name: "test-group", Namespace: "usernaut"},
		Spec: GroupSpec{
			Backends: []Backend{
				{Name: "snowflake-prod", Type: "snowflake"},
			},
		},
	}
	warnings, err := group.ValidateCreate(ctxWithFakeClient())
	require.NoError(t, err)
	assert.Nil(t, warnings)
}

func TestValidateCreate_RoverBackend_SAExists_Allowed(t *testing.T) {
	t.Setenv("WATCHED_NAMESPACE", "usernaut")
	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "usernaut-prod",
			Namespace: "usernaut",
		},
	}
	group := &Group{
		ObjectMeta: metav1.ObjectMeta{Name: "test-group", Namespace: "usernaut"},
		Spec: GroupSpec{
			Backends: []Backend{
				{Name: "rover-prod", Type: "rover"},
			},
		},
	}
	warnings, err := group.ValidateCreate(ctxWithFakeClient(sa))
	require.NoError(t, err)
	assert.Nil(t, warnings)
}

func TestValidateCreate_RoverBackend_SAMissing_Denied(t *testing.T) {
	t.Setenv("WATCHED_NAMESPACE", "usernaut")
	group := &Group{
		ObjectMeta: metav1.ObjectMeta{Name: "test-group", Namespace: "usernaut"},
		Spec: GroupSpec{
			Backends: []Backend{
				{Name: "rover-prod", Type: "rover"},
			},
		},
	}
	_, err := group.ValidateCreate(ctxWithFakeClient())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "usernaut-prod")
	assert.Contains(t, err.Error(), "not found")
}

func TestValidateCreate_RoverBackend_CaseInsensitive(t *testing.T) {
	t.Setenv("WATCHED_NAMESPACE", "usernaut")
	group := &Group{
		ObjectMeta: metav1.ObjectMeta{Name: "test-group", Namespace: "usernaut"},
		Spec: GroupSpec{
			Backends: []Backend{
				{Name: "rover-prod", Type: "Rover"},
			},
		},
	}
	_, err := group.ValidateCreate(ctxWithFakeClient())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "usernaut-prod")
}

func TestValidateCreate_MixedBackends_RoverPresent_SAMissing_Denied(t *testing.T) {
	t.Setenv("WATCHED_NAMESPACE", "usernaut")
	group := &Group{
		ObjectMeta: metav1.ObjectMeta{Name: "test-group", Namespace: "usernaut"},
		Spec: GroupSpec{
			Backends: []Backend{
				{Name: "snowflake-prod", Type: "snowflake"},
				{Name: "rover-prod", Type: "rover"},
			},
		},
	}
	_, err := group.ValidateCreate(ctxWithFakeClient())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "usernaut-prod")
}

func TestValidateUpdate_RoverBackend_SAMissing_Denied(t *testing.T) {
	t.Setenv("WATCHED_NAMESPACE", "usernaut")
	newGroup := &Group{
		ObjectMeta: metav1.ObjectMeta{Name: "test-group", Namespace: "usernaut"},
		Spec: GroupSpec{
			Backends: []Backend{
				{Name: "rover-prod", Type: "rover"},
			},
		},
	}
	oldGroup := &Group{
		ObjectMeta: metav1.ObjectMeta{Name: "test-group", Namespace: "usernaut"},
		Spec: GroupSpec{
			Backends: []Backend{
				{Name: "snowflake-prod", Type: "snowflake"},
			},
		},
	}
	_, err := newGroup.ValidateUpdate(ctxWithFakeClient(), oldGroup)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "usernaut-prod")
}

func TestValidateUpdate_RoverRemoved_Allowed(t *testing.T) {
	t.Setenv("WATCHED_NAMESPACE", "usernaut")
	newGroup := &Group{
		ObjectMeta: metav1.ObjectMeta{Name: "test-group", Namespace: "usernaut"},
		Spec: GroupSpec{
			Backends: []Backend{
				{Name: "snowflake-prod", Type: "snowflake"},
			},
		},
	}
	oldGroup := &Group{
		ObjectMeta: metav1.ObjectMeta{Name: "test-group", Namespace: "usernaut"},
		Spec: GroupSpec{
			Backends: []Backend{
				{Name: "rover-prod", Type: "rover"},
			},
		},
	}
	warnings, err := newGroup.ValidateUpdate(ctxWithFakeClient(), oldGroup)
	require.NoError(t, err)
	assert.Nil(t, warnings)
}

func TestValidateDelete_Always_Allowed(t *testing.T) {
	group := &Group{
		ObjectMeta: metav1.ObjectMeta{Name: "test-group", Namespace: "usernaut"},
		Spec: GroupSpec{
			Backends: []Backend{
				{Name: "rover-prod", Type: "rover"},
			},
		},
	}
	warnings, err := group.ValidateDelete(ctxWithFakeClient())
	require.NoError(t, err)
	assert.Nil(t, warnings)
}

func TestValidateCreate_NoClient_ReturnsError(t *testing.T) {
	group := &Group{
		ObjectMeta: metav1.ObjectMeta{Name: "test-group", Namespace: "usernaut"},
		Spec: GroupSpec{
			Backends: []Backend{
				{Name: "rover-prod", Type: "rover"},
			},
		},
	}
	// context without client injected
	_, err := group.ValidateCreate(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "webhook client not available")
}

func TestValidateCreate_CustomNamespace(t *testing.T) {
	t.Setenv("WATCHED_NAMESPACE", "custom-ns")
	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "usernaut-prod",
			Namespace: "custom-ns",
		},
	}
	group := &Group{
		ObjectMeta: metav1.ObjectMeta{Name: "test-group", Namespace: "custom-ns"},
		Spec: GroupSpec{
			Backends: []Backend{
				{Name: "rover-prod", Type: "rover"},
			},
		},
	}
	warnings, err := group.ValidateCreate(ctxWithFakeClient(sa))
	require.NoError(t, err)
	assert.Nil(t, warnings)
}
