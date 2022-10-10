package main

import (
	"context"
	"fmt"
	"log"
	"time"

	_ "go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/metric/instrument"
	"go.opentelemetry.io/otel/sdk/instrumentation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/aggregation"
	"go.opentelemetry.io/otel/sdk/metric/view"
)

const meterName = "metername_1"

func main() {

	ctx := context.Background()
	fmt.Println(ctx)

	endpoint := "otelmetrics-pp-observability.us-central1.gcp.dev.paypalinc.com:30706"

	opts := []otlpmetrichttp.Option{
		// otlpmetrichttp.WithInsecure(),
		otlpmetrichttp.WithEndpoint(endpoint),
	}

	exporter, err := otlpmetrichttp.New(ctx, opts...)

	if err != nil {
		log.Fatal("Error -->", err)

	}

	// View to customize histogram buckets and rename a single histogram instrument.
	customBucketsView, err := view.New(
		// Match* to match instruments
		view.MatchInstrumentName("custom_histogram"),
		view.MatchInstrumentationScope(instrumentation.Scope{Name: meterName}),

		// With* to modify instruments
		view.WithSetAggregation(aggregation.ExplicitBucketHistogram{
			Boundaries: []float64{64, 128, 256, 512, 1024, 2048, 4096},
		}),
		view.WithRename("bar"),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Default view to keep all instruments
	defaultView, err := view.New(view.MatchInstrumentName("*"))
	if err != nil {
		log.Fatal(err)
	}

	meterProvider := metric.NewMeterProvider(metric.WithReader(
		metric.NewPeriodicReader(exporter, metric.WithInterval(5*time.Second)),
		customBucketsView,
		defaultView))

	defer func() {
		if err := meterProvider.Shutdown(ctx); err != nil {
			panic(err)
		}
	}()
	global.SetMeterProvider(meterProvider)

	meter := global.Meter(meterName)

	attrs := []attribute.KeyValue{
		attribute.Key("dim1").String("dim1Value"),
		attribute.Key("dim2").String("dim2Value"),
	}

	counter, err := meter.SyncFloat64().Counter("hera.test.goalpharelease.http.counter", instrument.WithDescription("a simple counter to test GO SDK Alpha release"))
	for i := 0; i < 10; i++ {
		counter.Add(ctx, 5, attrs...)
		time.Sleep(6 * time.Second)

	}

}
