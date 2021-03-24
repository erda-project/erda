package promxp

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	dto "github.com/prometheus/client_model/go"

	"github.com/erda-project/erda/pkg/telemetry/common"
	"github.com/erda-project/erda/pkg/telemetry/report"
)

type publisher struct {
	ticker    *time.Ticker
	disruptor report.Disruptor
}

func init() {
	//report := common.DEFAULT_REPORTER
	publisher := &publisher{
		ticker:    time.NewTicker(30 * time.Second),
		disruptor: report.DefaultDisruptor,
	}
	go publisher.tick()
}

func (d *publisher) tick() {
	for {
		select {
		case <-d.ticker.C:
			gatherer := promxpGatherer
			if metricFamilies, err := gatherer.Gather(); err == nil {
				for _, input := range metricFamilies {
					if metrics := d.convert(input); metrics != nil && len(metrics) > 0 {
						if err = d.disruptor.In(metrics...); err != nil {
							fmt.Printf("%s E! report metricFamilies error %s /n", time.Now().Format("2006-01-02 15:04:05"), err.Error())
						}
					}
				}
			}
		}
	}
}

func (d *publisher) convert(input *dto.MetricFamily) []*common.Metric {
	timestamp := time.Now().UnixNano()
	var metrics []*common.Metric
	switch input.GetType() {
	case dto.MetricType_COUNTER:
		metrics = convertCounter(input, timestamp)
		break
	case dto.MetricType_GAUGE:
		metrics = convertGauge(input, timestamp)
		break
	case dto.MetricType_HISTOGRAM:
		metrics = convertHistogram(input, timestamp)
		break
	case dto.MetricType_SUMMARY:
		metrics = convertSummary(input, timestamp)
		break
	case dto.MetricType_UNTYPED:
		if len(input.GetMetric()) > 0 {
			for _, label := range input.GetMetric()[0].GetLabel() {
				if label.GetName() == "_custom_prome_type" && label.GetValue() == "meter" {
					metrics = convertMeter(input, timestamp)
					break
				}
			}
		}
		break
	default:
		break
	}
	if metrics != nil {
		handleMetricName(metrics)
	}
	return metrics
}

func convertCounter(input *dto.MetricFamily, timestamp int64) []*common.Metric {
	metrics := make([]*common.Metric, 0)
	for _, m := range input.GetMetric() {
		metric := &common.Metric{
			Name:      input.GetName(),
			Timestamp: timestamp,
			Tags:      convertLabelToTags(m.GetLabel()),
			Fields: map[string]interface{}{
				"count": m.GetCounter().GetValue(),
			},
		}
		metrics = append(metrics, metric)
	}
	return metrics
}

func convertGauge(input *dto.MetricFamily, timestamp int64) []*common.Metric {
	metrics := make([]*common.Metric, 0)
	for _, m := range input.GetMetric() {
		metric := &common.Metric{
			Name:      input.GetName(),
			Timestamp: timestamp,
			Tags:      convertLabelToTags(m.GetLabel()),
			Fields: map[string]interface{}{
				"value": m.GetGauge().GetValue(),
			},
		}
		metrics = append(metrics, metric)
	}
	return metrics
}

func convertHistogram(input *dto.MetricFamily, timestamp int64) []*common.Metric {
	metrics := make([]*common.Metric, 0)
	for _, m := range input.GetMetric() {
		tags := convertLabelToTags(m.GetLabel())
		fields := make(map[string]interface{})
		histogram := m.GetHistogram()
		fields["count"] = histogram.GetSampleCount()
		fields["sum"] = histogram.GetSampleSum()
		for _, bucket := range histogram.GetBucket() {
			key := "le_" + strings.Replace(strconv.FormatFloat(bucket.GetUpperBound(), 'f', -1, 64), ".", "_", -1)
			fields[key] = bucket.GetCumulativeCount()
		}
		metric := &common.Metric{
			Name:      input.GetName(),
			Timestamp: timestamp,
			Tags:      tags,
			Fields:    fields,
		}
		metrics = append(metrics, metric)
	}
	return metrics
}

func convertSummary(input *dto.MetricFamily, timestamp int64) []*common.Metric {
	metrics := make([]*common.Metric, 0)
	for _, m := range input.GetMetric() {
		tags := convertLabelToTags(m.GetLabel())
		fields := make(map[string]interface{})
		summary := m.GetSummary()
		fields["count"] = summary.GetSampleCount()
		fields["sum"] = summary.GetSampleSum()
		for _, quantile := range summary.GetQuantile() {
			key := "quantile_" + strings.Replace(strconv.FormatFloat(quantile.GetQuantile(), 'f', -1, 64), ".", "_", -1)
			fields[key] = quantile.GetValue()
		}
		metric := &common.Metric{
			Name:      input.GetName(),
			Timestamp: timestamp,
			Tags:      tags,
			Fields:    fields,
		}
		metrics = append(metrics, metric)
	}
	return metrics
}

func convertMeter(input *dto.MetricFamily, timestamp int64) []*common.Metric {
	metrics := make([]*common.Metric, 0)
	for _, m := range input.GetMetric() {
		metric := &common.Metric{
			Name:      input.GetName(),
			Timestamp: timestamp,
			Tags:      convertLabelToTags(m.GetLabel()),
			Fields: map[string]interface{}{
				"rate_1": m.GetUntyped().GetValue(),
			},
		}
		metrics = append(metrics, metric)
	}
	return metrics
}

func convertLabelToTags(labels []*dto.LabelPair) map[string]string {
	tags := make(map[string]string)
	for _, label := range labels {
		tags[label.GetName()] = label.GetValue()
	}
	return tags
}

func handleMetricName(metrics []*common.Metric) {
	labels := common.GetGlobalLabels()
	if component, ok := labels["component"]; ok {
		component = "dice_" + component
		for _, metric := range metrics {
			field := metric.Name
			idx := strings.Index(field, component)
			if idx > -1 {
				field = strings.TrimPrefix(metric.Name, component+"_")
			}
			metric.Tags["field"] = field
			metric.Name = component
		}
	}
}
