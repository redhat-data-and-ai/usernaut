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
	"context"
	"sync"

	otelmetric "go.opentelemetry.io/otel/metric"
)

var (
	reconciliationMetrics     *ReconciliationMetrics
	reconciliationMetricsOnce sync.Once
)

type ReconciliationMetrics struct {
	CountTotal *Counter
	ErrorTotal *Counter
}

func InitReconciliationMetrics(meter otelmetric.Meter) error {
	var initErr error
	reconciliationMetricsOnce.Do(func() {

		countTotal, err := NewCounter(meter, MetricOptions{
			Name: BuildMetricName("reconciliation_count", MetricNameSuffixTotal),
			Description: "total number of reconciliation attempts. " +
				"It will help us in calculations ahead when we get " +
				"either the error total or the success total",
			Unit: "1",
		})
		if err != nil {
			initErr = err
			return
		}

		errorTotal, err := NewCounter(meter, MetricOptions{
			Name: BuildMetricName("reconciliation_error", MetricNameSuffixTotal),
			Description: "total number of reconciliation errors. " +
				"It can be used to calculate reconciliation success rate in the metrics backend. " +
				"rate calculation: error% = usernaut_reconciliation_error_total / usernaut_reconciliation_count_total, " +
				"success% = 1 - error%",
			Unit: "1",
		})
		if err != nil {
			initErr = err
			return
		}

		reconciliationMetrics = &ReconciliationMetrics{
			CountTotal: countTotal,
			ErrorTotal: errorTotal,
		}
	})

	return initErr
}

func GetReconciliationMetrics() *ReconciliationMetrics {
	return reconciliationMetrics
}

func (rm *ReconciliationMetrics) RecordReconciliationStart(ctx context.Context, controller string) {
	if rm == nil {
		return
	}
	rm.CountTotal.Inc(ctx, WithController(controller))
}

func (rm *ReconciliationMetrics) RecordReconciliationError(ctx context.Context, controller string) {
	if rm == nil {
		return
	}
	rm.ErrorTotal.Inc(ctx, WithController(controller))
}

func (rm *ReconciliationMetrics) RecordReconciliation(ctx context.Context, controller string, err error) {
	if rm == nil {
		return
	}
	rm.RecordReconciliationStart(ctx, controller)
	if err != nil {
		rm.RecordReconciliationError(ctx, controller)
	}
}
