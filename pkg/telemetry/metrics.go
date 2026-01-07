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

	"go.opentelemetry.io/otel/attribute"
	otelmetric "go.opentelemetry.io/otel/metric"
)

type MetricOptions struct {
	Name        string
	Description string
	Unit        string
	Attributes  []attribute.KeyValue
}

type Counter struct {
	counter otelmetric.Int64Counter
}

func NewCounter(meter otelmetric.Meter, opts MetricOptions) (*Counter, error) {
	counter, err := meter.Int64Counter(
		opts.Name,
		otelmetric.WithDescription(opts.Description),
		otelmetric.WithUnit(opts.Unit),
	)
	if err != nil {
		return nil, err
	}

	return &Counter{counter: counter}, nil
}

func (c *Counter) Add(ctx context.Context, value int64, attrs ...attribute.KeyValue) {
	c.counter.Add(ctx, value, otelmetric.WithAttributes(attrs...))
}

func (c *Counter) Inc(ctx context.Context, attrs ...attribute.KeyValue) {
	c.Add(ctx, 1, attrs...)
}

type Histogram struct {
	histogram otelmetric.Float64Histogram
}

func NewHistogram(meter otelmetric.Meter, opts MetricOptions) (*Histogram, error) {
	histogram, err := meter.Float64Histogram(
		opts.Name,
		otelmetric.WithDescription(opts.Description),
		otelmetric.WithUnit(opts.Unit),
	)
	if err != nil {
		return nil, err
	}

	return &Histogram{histogram: histogram}, nil
}

func (h *Histogram) Record(ctx context.Context, value float64, attrs ...attribute.KeyValue) {
	h.histogram.Record(ctx, value, otelmetric.WithAttributes(attrs...))
}

type GaugeCallback func(context.Context) (float64, []attribute.KeyValue)

type Gauge struct {
	gauge otelmetric.Float64ObservableGauge
}

func NewGauge(meter otelmetric.Meter, opts MetricOptions, callback GaugeCallback) (*Gauge, error) {
	gauge, err := meter.Float64ObservableGauge(
		opts.Name,
		otelmetric.WithDescription(opts.Description),
		otelmetric.WithUnit(opts.Unit),
		otelmetric.WithFloat64Callback(func(ctx context.Context, observer otelmetric.Float64Observer) error {
			value, attrs := callback(ctx)
			observer.Observe(value, otelmetric.WithAttributes(attrs...))
			return nil
		}),
	)
	if err != nil {
		return nil, err
	}

	return &Gauge{gauge: gauge}, nil
}

type UpDownCounter struct {
	counter otelmetric.Int64UpDownCounter
}

func NewUpDownCounter(meter otelmetric.Meter, opts MetricOptions) (*UpDownCounter, error) {
	counter, err := meter.Int64UpDownCounter(
		opts.Name,
		otelmetric.WithDescription(opts.Description),
		otelmetric.WithUnit(opts.Unit),
	)
	if err != nil {
		return nil, err
	}

	return &UpDownCounter{counter: counter}, nil
}

func (u *UpDownCounter) Add(ctx context.Context, value int64, attrs ...attribute.KeyValue) {
	u.counter.Add(ctx, value, otelmetric.WithAttributes(attrs...))
}

func (u *UpDownCounter) Inc(ctx context.Context, attrs ...attribute.KeyValue) {
	u.Add(ctx, 1, attrs...)
}

func (u *UpDownCounter) Dec(ctx context.Context, attrs ...attribute.KeyValue) {
	u.Add(ctx, -1, attrs...)
}
