package telemetry

import (
	"context"
	"fmt"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

type influxQueryHelper struct {
	client influxdb2.Client
	org    string
	bucket string
}

func newInfluxHelper(client influxdb2.Client, org, bucket string) *influxQueryHelper {
	return &influxQueryHelper{client: client, org: org, bucket: bucket}
}

func (h *influxQueryHelper) WriteRecord(rec TelemetryRecord) {
	if h.client == nil || h.org == "" || h.bucket == "" {
		return
	}
	writeAPI := h.client.WriteAPI(h.org, h.bucket)
	p := influxdb2.NewPointWithMeasurement("telemetry").
		AddTag("sensor_channel_id", fmt.Sprintf("%d", rec.SensorChannelID)).
		AddTag("metric_code", rec.MetricCode).
		AddTag("quality_flag", rec.QualityFlag).
		AddField("value", rec.Value).
		SetTime(rec.CollectedAt)
	if rec.RawValue != nil {
		p.AddField("raw_value", *rec.RawValue)
	}
	writeAPI.WritePoint(p)
	writeAPI.Flush()
}

func (h *influxQueryHelper) QueryHistory(channelID uint64, metricCode, startTime, endTime string, limit int) ([]map[string]interface{}, error) {
	if h.client == nil {
		return nil, fmt.Errorf("influx not configured")
	}

	queryAPI := h.client.QueryAPI(h.org)

	rangeStart := "-30d"
	if startTime != "" {
		if _, err := time.Parse(time.RFC3339, startTime); err == nil {
			rangeStart = startTime
		} else if _, err := time.Parse(time.RFC3339Nano, startTime); err == nil {
			rangeStart = startTime
		}
	}

	rangeStop := "now()"
	if endTime != "" {
		if _, err := time.Parse(time.RFC3339, endTime); err == nil {
			rangeStop = endTime
		} else if _, err := time.Parse(time.RFC3339Nano, endTime); err == nil {
			rangeStop = endTime
		}
	}

	filterStr := fmt.Sprintf(`r["sensor_channel_id"] == "%d"`, channelID)
	if metricCode != "" {
		filterStr += fmt.Sprintf(` and r["metric_code"] == "%s"`, metricCode)
	}

	flux := fmt.Sprintf(`
		from(bucket: "%s")
			|> range(start: %s, stop: %s)
			|> filter(fn: (r) => r["_measurement"] == "telemetry")
			|> filter(fn: (r) => %s)
			|> filter(fn: (r) => r["_field"] == "value")
			|> sort(columns: ["_time"], desc: true)
			|> limit(n: %d)
	`, h.bucket, rangeStart, rangeStop, filterStr, limit)

	result, err := queryAPI.Query(context.Background(), flux)
	if err != nil {
		return nil, err
	}

	// Parse result into records
	records := make(map[string][]int64) // metric_code -> times
	values := make(map[string]map[int64]float64)

	for result.Next() {
		record := result.Record()
		ts := record.Time().UnixMilli()
		metric := record.ValueByKey("metric_code").(string)

		if _, ok := records[metric]; !ok {
			records[metric] = make([]int64, 0)
			values[metric] = make(map[int64]float64)
		}
		records[metric] = append(records[metric], ts)
		if v, ok := record.Value().(float64); ok {
			values[metric][ts] = v
		}
	}

	if result.Err() != nil {
		return nil, result.Err()
	}

	// Flatten into response format
	var items []map[string]interface{}
	seen := make(map[string]bool)
	for metric, timestamps := range records {
		for _, ts := range timestamps {
			key := fmt.Sprintf("%s:%d", metric, ts)
			if seen[key] {
				continue
			}
			seen[key] = true

			t := time.UnixMilli(ts).UTC()
			items = append(items, map[string]interface{}{
				"sensor_channel_id": channelID,
				"metric_code":       metric,
				"value":             values[metric][ts],
				"collected_at":      t.Format(time.RFC3339),
			})
		}
	}

	return items, nil
}

func (h *influxQueryHelper) Close() {
	// Client lifecycle managed externally
}
