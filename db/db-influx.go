package db

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

// InfluxDBClient wraps the InfluxDB client for easier usage
type InfluxDBClient struct {
	client influxdb2.Client
	org    string
	bucket string
}

// NewInfluxDBClient initializes a new InfluxDB client
func NewInfluxDBClient(url, token, org, bucket string) *InfluxDBClient {
	client := influxdb2.NewClient(url, token)
	return &InfluxDBClient{
		client: client,
		org:    org,
		bucket: bucket,
	}
}

// WritePLCData writes PLC data to the InfluxDB with retry logic and logging
func (i *InfluxDBClient) WritePLCData(ctx context.Context, measurement string, tags map[string]string, fields map[string]interface{}, timestamp time.Time) error {
	writeAPI := i.client.WriteAPIBlocking(i.org, i.bucket)
	var err error
	maxAttempts := 3

	// Helper to write a single point with retry
	writePoint := func(tags map[string]string, fields map[string]interface{}, ts time.Time) error {
		var err error
		point := influxdb2.NewPoint(measurement, tags, fields, ts)
		for attempt := 1; attempt <= maxAttempts; attempt++ {
			err = writeAPI.WritePoint(ctx, point)
			if err == nil {
				return nil
			}
			log.Printf("[InfluxDB] Write attempt %d failed: %v", attempt, err)
			backoff := time.Duration(500*attempt+rand.Intn(250)) * time.Millisecond
			time.Sleep(backoff)
		}
		return err
	}

	for k, v := range fields {
		switch val := v.(type) {
		case []float32:
			for idx, elem := range val {
				ftags := make(map[string]string, len(tags)+1)
				for tk, tv := range tags {
					ftags[tk] = tv
				}
				ftags["ArrayIndex"] = fmt.Sprintf("%d", idx)
				f := map[string]interface{}{k: float64(elem)}
				if err = writePoint(ftags, f, timestamp); err != nil {
					return fmt.Errorf("failed to write array field %q index %d: %w", k, idx, err)
				}
			}
		case []float64:
			for idx, elem := range val {
				ftags := make(map[string]string, len(tags)+1)
				for tk, tv := range tags {
					ftags[tk] = tv
				}
				ftags["ArrayIndex"] = fmt.Sprintf("%d", idx)
				f := map[string]interface{}{k: elem}
				if err = writePoint(ftags, f, timestamp); err != nil {
					return fmt.Errorf("failed to write array field %q index %d: %w", k, idx, err)
				}
			}
		case []int:
			for idx, elem := range val {
				ftags := make(map[string]string, len(tags)+1)
				for tk, tv := range tags {
					ftags[tk] = tv
				}
				ftags["ArrayIndex"] = fmt.Sprintf("%d", idx)
				f := map[string]interface{}{k: float64(elem)}
				if err = writePoint(ftags, f, timestamp); err != nil {
					return fmt.Errorf("failed to write array field %q index %d: %w", k, idx, err)
				}
			}
		case []int32:
			for idx, elem := range val {
				ftags := make(map[string]string, len(tags)+1)
				for tk, tv := range tags {
					ftags[tk] = tv
				}
				ftags["ArrayIndex"] = fmt.Sprintf("%d", idx)
				f := map[string]interface{}{k: float64(elem)}
				if err = writePoint(ftags, f, timestamp); err != nil {
					return fmt.Errorf("failed to write array field %q index %d: %w", k, idx, err)
				}
			}
		case []int64:
			for idx, elem := range val {
				ftags := make(map[string]string, len(tags)+1)
				for tk, tv := range tags {
					ftags[tk] = tv
				}
				ftags["ArrayIndex"] = fmt.Sprintf("%d", idx)
				f := map[string]interface{}{k: float64(elem)}
				if err = writePoint(ftags, f, timestamp); err != nil {
					return fmt.Errorf("failed to write array field %q index %d: %w", k, idx, err)
				}
			}
		default:
			// Scalar value: normalize numeric types to float64
			normalizedFields := make(map[string]interface{}, 1)
			switch v2 := v.(type) {
			case int:
				normalizedFields[k] = float64(v2)
			case int8:
				normalizedFields[k] = float64(v2)
			case int16:
				normalizedFields[k] = float64(v2)
			case int32:
				normalizedFields[k] = float64(v2)
			case int64:
				normalizedFields[k] = float64(v2)
			case uint:
				normalizedFields[k] = float64(v2)
			case uint8:
				normalizedFields[k] = float64(v2)
			case uint16:
				normalizedFields[k] = float64(v2)
			case uint32:
				normalizedFields[k] = float64(v2)
			case uint64:
				normalizedFields[k] = float64(v2)
			case float32:
				normalizedFields[k] = float64(v2)
			case float64:
				normalizedFields[k] = v2
			default:
				log.Printf("[InfluxDB] Warning: field %q is non-numeric type %T (value: %v)", k, v, v)
				normalizedFields[k] = v
			}
			if err = writePoint(tags, normalizedFields, timestamp); err != nil {
				return fmt.Errorf("failed to write field %q: %w", k, err)
			}
		}
	}
	return nil
}

// ReadPLCData queries PLC data from the InfluxDB with context timeout and logging
func (i *InfluxDBClient) ReadPLCData(ctx context.Context, query string) ([]map[string]any, error) {
	queryAPI := i.client.QueryAPI(i.org)

	// Set a default timeout if none is set
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
	}

	result, err := queryAPI.Query(ctx, query)
	if err != nil {
		log.Printf("[InfluxDB] Query error: %v", err)
		return nil, fmt.Errorf("failed to query PLC data: %w", err)
	}

	var records []map[string]any
	for result.Next() {
		record := make(map[string]any)
		for key, value := range result.Record().Values() {
			record[key] = value
		}
		records = append(records, record)
	}

	if result.Err() != nil {
		log.Printf("[InfluxDB] Query result error: %v", result.Err())
		return nil, fmt.Errorf("query result error: %w", result.Err())
	}

	return records, nil
}

// HealthCheck checks if the InfluxDB connection is alive
func (i *InfluxDBClient) HealthCheck(ctx context.Context) error {
	health, err := i.client.Health(ctx)
	if err != nil {
		return fmt.Errorf("influxdb health check failed: %w", err)
	}
	if health.Status != "pass" {
		return fmt.Errorf("influxdb health check status: %s", health.Status)
	}
	return nil
}

// Close closes the InfluxDB client
func (i *InfluxDBClient) Close() {
	i.client.Close()
}
