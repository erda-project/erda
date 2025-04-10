// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package promremotewrite

import (
	"context"
	"fmt"
	"io"
	"math"
	"sync"
	"time"

	"github.com/golang/snappy"
	pmodel "github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/prompb"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/internal/tools/monitor/core/metric"
)

const CollectorGroupTag = "collector_group"

func ParseStream(r io.Reader, metricsChan chan *metric.Metric) error {
	buf, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	buf, err = snappy.Decode(nil, buf)
	if err != nil {
		return err
	}

	wr := &prompb.WriteRequest{}
	err = wr.Unmarshal(buf)
	if err != nil {
		return fmt.Errorf("unmarshal WriteRequest: %w", err)
	}

	return parseWriteRequest(wr, metricsChan)
}

type MetricTagHash string

func parseWriteRequest(wr *prompb.WriteRequest, metricsChan chan *metric.Metric) error {
	now := time.Now() // receive time
	for _, ts := range wr.Timeseries {
		tags := map[string]string{}
		for _, l := range ts.Labels {
			tags[l.Name] = l.Value
		}
		metricName := tags[pmodel.MetricNameLabel]
		if metricName == "" {
			return fmt.Errorf("%q not found in tags or empty", pmodel.MetricNameLabel)
		}
		delete(tags, pmodel.MetricNameLabel)

		// set pmodel.JobLabel as  name
		job := tags[pmodel.JobLabel]
		if job == "" {
			return fmt.Errorf("%q not found in tags or empty", pmodel.JobLabel)
		}
		delete(tags, pmodel.JobLabel)

		for _, s := range ts.Samples {
			fields := make(map[string]interface{})
			if math.IsNaN(s.Value) {
				continue
			}
			fields[metricName] = s.Value

			// converting to metric
			t := now
			if s.Timestamp > 0 {
				t = time.Unix(0, s.Timestamp*time.Millisecond.Nanoseconds())
			}

			m := &metric.Metric{
				Name:      job,
				Timestamp: t.UnixNano(),
				Tags:      tags,
				Fields:    fields,
			}
			metricsChan <- m

			//err := callback(&m)
			//if err != nil {
			//	return fmt.Errorf("callback: %w", err)
			//}
		}
	}
	metricsChan <- nil
	return nil
}

type GroupMetricsOptions struct {
	MinSize        int
	RetentionRatio float64
	GroupTagName   string
	MetricsChan    chan *metric.Metric
	Callback       func(record *metric.Metric) error
}

func DealGroupMetrics(ctx context.Context, options GroupMetricsOptions) {
	var (
		metricTagGroup = NewGroupMetricList()
		lastDealTime   = time.Now()
		lock           sync.Mutex
	)
	cancelCtx, cancelFunc := context.WithCancel(ctx)
	defer cancelFunc()
	go func(ctx context.Context) {
		ticker := time.NewTicker(time.Minute)
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				lock.Lock()
				flag := time.Since(lastDealTime) > time.Minute
				lock.Unlock()
				logrus.Infof("deal by ticker, flag: %v", flag)
				if flag {
					lock.Lock()
					lastDealTime = time.Now()
					lock.Unlock()
					metrics := metricTagGroup.PopAllMetrics()
					for i := range metrics {
						if err := options.Callback(metrics[i]); err != nil {
							logrus.Errorf("deal group metric error %v", err)
						}
					}
				}
			}
		}
	}(cancelCtx)

	for {
		select {
		case <-ctx.Done():
			logrus.Infof("deal group metric done")
			metrics := metricTagGroup.PopAllMetrics()
			for i := range metrics {
				if err := options.Callback(metrics[i]); err != nil {
					logrus.Errorf("deal group metric error %v", err)
				}
			}
			return
		case m := <-options.MetricsChan:
			if m != nil {
				if _, ok := m.Tags[options.GroupTagName]; !ok || options.GroupTagName == "" {
					if err := options.Callback(m); err != nil {
						logrus.Errorf("metric callback error: %v", err)
					}
					continue
				}

				metricTagGroup.Append(m)
				continue
			}

			length := metricTagGroup.Length()
			retentionNum := int(float64(length) * options.RetentionRatio)
			// don't deal any metric, leave it to a timed task.
			if retentionNum >= length {
				retentionNum = length
				continue
			}

			if length < options.MinSize {
				retentionNum = 1
			}

			lock.Lock()
			lastDealTime = time.Now()
			lock.Unlock()

			for i := 0; i < length-retentionNum; i++ {
				metrics := metricTagGroup.PopMetrics()
				for j := range metrics {
					if err := options.Callback(metrics[j]); err != nil {
						logrus.Errorf("deal group metric error %v", err)
					}
				}
			}
		}
	}
}
