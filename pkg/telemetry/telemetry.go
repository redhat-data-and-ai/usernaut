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
	"fmt"
	"strings"
	"sync"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	otelmetric "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.27.0"
)

// Note: This package requires the following dependencies:
//   go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp
//   go.opentelemetry.io/otel/sdk/metric
//   go.opentelemetry.io/otel/semconv/v1.27.0
// These should be added to go.mod if not already present.

var (
	meterProvider     *metric.MeterProvider
	meterProviderOnce sync.Once
	shutdownOnce      sync.Once
)

type Config struct {
	ServiceName    string
	ServiceVersion string
	OTLPEndpoint   string
	//default:false
	Insecure bool
	// default;true
	Enabled bool
}

func Init(ctx context.Context, config Config) error {
	var initErr error
	meterProviderOnce.Do(func() {
		if !config.Enabled {
			// no=op meter provider usage whenever telemetry is disabled
			meterProvider = metric.NewMeterProvider()
			otel.SetMeterProvider(meterProvider)
			return
		}
		if config.ServiceName == "" {
			initErr = fmt.Errorf("service name is required")
			return
		}
		if config.OTLPEndpoint == "" {
			initErr = fmt.Errorf("OTLP endpoint is required")
			return
		}
		res, err := resource.New(ctx,
			resource.WithAttributes(
				semconv.ServiceName(config.ServiceName),
				semconv.ServiceVersion(config.ServiceVersion),
			),
		)
		if err != nil {
			initErr = fmt.Errorf("failed to create resource: %w", err)
			return
		}

		//  OTLP http exporter
		endpoint := strings.TrimSpace(config.OTLPEndpoint)
		endpoint = strings.TrimPrefix(strings.ToLower(endpoint), "http://")
		endpoint = strings.TrimPrefix(endpoint, "https://")
		if idx := strings.Index(endpoint, "/"); idx != -1 {
			endpoint = endpoint[:idx]
		}
		if idx := strings.Index(endpoint, "?"); idx != -1 {
			endpoint = endpoint[:idx]
		}
		exporterOpts := []otlpmetrichttp.Option{
			otlpmetrichttp.WithEndpoint(endpoint),
		}
		if config.Insecure {
			exporterOpts = append(exporterOpts, otlpmetrichttp.WithInsecure())
		}
		exporter, err := otlpmetrichttp.New(ctx, exporterOpts...)
		if err != nil {
			initErr = fmt.Errorf("failed to create OTLP exporter: %w", err)
			return
		}
		reader := metric.NewPeriodicReader(exporter)
		meterProvider = metric.NewMeterProvider(
			metric.WithResource(res),
			metric.WithReader(reader),
		)
		otel.SetMeterProvider(meterProvider)
	})

	return initErr
}

func Shutdown(ctx context.Context) error {
	var shutdownErr error
	shutdownOnce.Do(func() {
		if meterProvider != nil {
			shutdownErr = meterProvider.Shutdown(ctx)
		}
	})
	return shutdownErr
}

func GetMeter(name string, opts ...otelmetric.MeterOption) otelmetric.Meter {
	return otel.Meter(name, opts...)
}
