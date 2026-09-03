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

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type contextKey string

const webhookClientKey contextKey = "webhook-client"

// UnifiedValidator handles validation for all operator.dataverse.redhat.com/v1alpha1 resources.
// It decodes runtime.Unknown objects from the API server into concrete types and delegates
// to per-resource ValidateCreate/ValidateUpdate/ValidateDelete methods.
//
// +kubebuilder:webhook:path=/validate-operator-dataverse-redhat-com-v1alpha1,mutating=false,failurePolicy=fail,sideEffects=None,groups=operator.dataverse.redhat.com,resources=*,verbs=create;update;delete,versions=v1alpha1,name=voperator-dataverse-redhat-com-v1alpha1.kb.io,admissionReviewVersions=v1
// +kubebuilder:object:generate=false
type UnifiedValidator struct {
	Scheme  *runtime.Scheme
	Decoder runtime.Decoder
	Client  client.Client
}

// SetupUnifiedWebhookWithManager registers the unified validating webhook with the manager.
func SetupUnifiedWebhookWithManager(mgr ctrl.Manager) error {
	scheme := mgr.GetScheme()
	codec := serializer.NewCodecFactory(scheme).UniversalDeserializer()

	validator := &UnifiedValidator{
		Scheme:  scheme,
		Decoder: codec,
		Client:  mgr.GetClient(),
	}

	webhook := &admission.Webhook{
		Handler: admission.WithCustomValidator(scheme, &runtime.Unknown{}, validator),
	}

	mgr.GetWebhookServer().Register(
		"/validate-operator-dataverse-redhat-com-v1alpha1",
		webhook,
	)
	return nil
}

// decodeUnknownObject deserializes runtime.Unknown (raw JSON from the API server)
// into a concrete registered type using the scheme's decoder. Wildcard resource
// matching means the API server sends objects wrapped in runtime.Unknown.
func (v *UnifiedValidator) decodeUnknownObject(ctx context.Context, obj runtime.Object, operation string) (runtime.Object, error) {
	log := logf.FromContext(ctx)

	unknown, isUnknown := obj.(*runtime.Unknown)
	if !isUnknown {
		return obj, nil
	}

	gvk := obj.GetObjectKind().GroupVersionKind()
	decoded, _, err := v.Decoder.Decode(unknown.Raw, nil, nil)
	if err != nil {
		log.Error(err, "Failed to decode unknown object", "operation", operation, "gvk", gvk)
		return nil, err
	}
	return decoded, nil
}

func (v *UnifiedValidator) contextWithClient(ctx context.Context) context.Context {
	return context.WithValue(ctx, webhookClientKey, v.Client)
}

// ClientFromContext retrieves the client.Client stored in the context by the unified webhook.
func ClientFromContext(ctx context.Context) client.Client {
	if c, ok := ctx.Value(webhookClientKey).(client.Client); ok {
		return c
	}
	return nil
}

func (v *UnifiedValidator) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	log := logf.FromContext(ctx)
	gvk := obj.GetObjectKind().GroupVersionKind()
	log.Info("UnifiedValidator.ValidateCreate called", "objectType", gvk)

	decoded, err := v.decodeUnknownObject(ctx, obj, "create")
	if err != nil {
		return nil, err
	}
	obj = decoded

	if validator, ok := obj.(interface {
		ValidateCreate(ctx context.Context) (admission.Warnings, error)
	}); ok {
		log.Info("Delegating to resource-specific ValidateCreate", "objectType", gvk)
		return validator.ValidateCreate(v.contextWithClient(ctx))
	}
	log.Info("No resource-specific ValidateCreate found", "objectType", gvk)
	return nil, nil
}

func (v *UnifiedValidator) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	log := logf.FromContext(ctx)
	gvk := newObj.GetObjectKind().GroupVersionKind()
	log.Info("UnifiedValidator.ValidateUpdate called", "objectType", gvk)

	decodedNew, err := v.decodeUnknownObject(ctx, newObj, "update")
	if err != nil {
		return nil, err
	}
	newObj = decodedNew

	decodedOld, err := v.decodeUnknownObject(ctx, oldObj, "update-old")
	if err != nil {
		return nil, err
	}
	oldObj = decodedOld

	if validator, ok := newObj.(interface {
		ValidateUpdate(ctx context.Context, old runtime.Object) (admission.Warnings, error)
	}); ok {
		log.Info("Delegating to resource-specific ValidateUpdate", "objectType", gvk)
		return validator.ValidateUpdate(v.contextWithClient(ctx), oldObj)
	}
	log.Info("No resource-specific ValidateUpdate found", "objectType", gvk)
	return nil, nil
}

func (v *UnifiedValidator) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	log := logf.FromContext(ctx)
	gvk := obj.GetObjectKind().GroupVersionKind()
	log.Info("UnifiedValidator.ValidateDelete called", "objectType", gvk)

	decoded, err := v.decodeUnknownObject(ctx, obj, "delete")
	if err != nil {
		return nil, err
	}
	obj = decoded

	if validator, ok := obj.(interface {
		ValidateDelete(ctx context.Context) (admission.Warnings, error)
	}); ok {
		log.Info("Delegating to resource-specific ValidateDelete", "objectType", gvk)
		return validator.ValidateDelete(v.contextWithClient(ctx))
	}
	log.Info("No resource-specific ValidateDelete found", "objectType", gvk)
	return nil, nil
}
