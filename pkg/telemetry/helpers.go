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

package telemetry

import (
	"go.opentelemetry.io/otel/attribute"
)

// naming conventions for metric names
const (
	MetricNameSuffixTotal    = "_total"
	MetricNameSuffixDuration = "_duration_seconds"
	MetricNameSuffixBytes    = "_bytes"
	MetricNameSuffixCount    = "_count"
)

const (
	AttrController  = "usernaut_controller"
	AttrBackend     = "usernaut_backend"
	AttrBackendType = "usernaut_backend_type"
	AttrStatus      = "usernaut_status"
	AttrOperation   = "usernaut_operation"
	AttrError       = "usernaut_error"
)

const (
	StatusSuccess = "success"
	StatusError   = "error"
)

func BuildMetricName(baseName, suffix string) string {
	prefixedName := "usernaut_" + baseName
	if suffix == "" {
		return prefixedName
	}
	return prefixedName + suffix
}

// creates attribute for controller name
func WithController(controller string) attribute.KeyValue {
	return attribute.String(AttrController, controller)
}

// creates attribute for backend name
func WithBackend(backend string) attribute.KeyValue {
	return attribute.String(AttrBackend, backend)
}

// creates attribute for backend type
func WithBackendType(backendType string) attribute.KeyValue {
	return attribute.String(AttrBackendType, backendType)
}

// creates attribute for status
func WithStatus(status string) attribute.KeyValue {
	return attribute.String(AttrStatus, status)
}

// creates  attribute for operation name
func WithOperation(operation string) attribute.KeyValue {
	return attribute.String(AttrOperation, operation)
}

// creates attribute for error type
func WithError(errType string) attribute.KeyValue {
	return attribute.String(AttrError, errType)
}
