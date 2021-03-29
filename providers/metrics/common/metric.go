package common

import "time"

type Metric struct {
	Name      string                 `json:"name"`
	Timestamp int64                  `json:"timestamp"`
	Tags      map[string]string      `json:"tags"`
	Fields    map[string]interface{} `json:"fields"`
}

type BulkMetricRequest []*Metric

func CreateBulkMetricRequest() *BulkMetricRequest {
	metricRequest := make(BulkMetricRequest, 0)
	return &metricRequest
}

func (b *BulkMetricRequest) Add(name string, tags map[string]string, fields map[string]interface{}) *BulkMetricRequest {
	*b = append(*b, &Metric{
		Name:      name,
		Timestamp: time.Now().UnixNano(),
		Tags:      tags,
		Fields:    fields,
	})
	return b
}

func (b *BulkMetricRequest) AddWithTime(name string, tags map[string]string, fields map[string]interface{}, timestamp time.Time) *BulkMetricRequest {
	*b = append(*b, &Metric{
		Name:      name,
		Timestamp: timestamp.UnixNano(),
		Tags:      tags,
		Fields:    fields,
	})
	return b
}
