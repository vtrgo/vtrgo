package db

import (
	"context"
	"fmt"
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

// WritePLCData writes PLC data to the InfluxDB
func (i *InfluxDBClient) WritePLCData(ctx context.Context, measurement string, tags map[string]string, fields map[string]interface{}, timestamp time.Time) error {
	writeAPI := i.client.WriteAPIBlocking(i.org, i.bucket)

	point := influxdb2.NewPoint(measurement, tags, fields, timestamp)
	err := writeAPI.WritePoint(ctx, point)
	if err != nil {
		return fmt.Errorf("failed to write PLC data: %w", err)
	}
	return nil
}

// ReadPLCData queries PLC data from the InfluxDB
func (i *InfluxDBClient) ReadPLCData(ctx context.Context, query string) ([]map[string]interface{}, error) {
	queryAPI := i.client.QueryAPI(i.org)

	result, err := queryAPI.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query PLC data: %w", err)
	}

	var records []map[string]interface{}
	for result.Next() {
		record := make(map[string]interface{})
		for key, value := range result.Record().Values() {
			record[key] = value
		}
		records = append(records, record)
	}

	if result.Err() != nil {
		return nil, fmt.Errorf("query result error: %w", result.Err())
	}

	return records, nil
}

// Close closes the InfluxDB client
func (i *InfluxDBClient) Close() {
	i.client.Close()
}
