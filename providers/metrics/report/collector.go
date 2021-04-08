// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package report

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/pkg/errors"

	"github.com/erda-project/erda-infra/pkg/transport/http/compress"
)

type ReportClient struct {
	CFG        *config
	HttpClient *http.Client
}

type MetricReport interface {
	SetCFG(cfg *config)
	Send(in []*Metric) error
	CreateReportClient(addr, username, password string) *ReportClient
}

type NamedMetrics struct {
	Name    string
	Metrics Metrics
}

type Metrics []*Metric

func (c *ReportClient) SetCFG(cfg *config) {
	c.CFG = cfg
}

func (c *ReportClient) CreateReportClient(addr, username, password string) *ReportClient {
	return &ReportClient{
		CFG: &config{
			ReportConfig: &ReportConfig{
				Collector: &CollectorConfig{
					Addr:     addr,
					UserName: username,
					Password: password,
				},
			},
		}, HttpClient: new(http.Client),
	}
}

func (c *ReportClient) Send(in []*Metric) error {
	groups := c.group(in)
	for _, group := range groups {
		if len(group.Metrics) == 0 {
			continue
		}
		requestBuffer, err := c.serialize(group)
		if err != nil {
			continue
		}
		for i := 0; i < c.CFG.ReportConfig.Collector.Retry; i++ {
			if err = c.write(group.Name, requestBuffer); err == nil {
				break
			}
			fmt.Printf("%s E! Retry %d # report in to collector error %s /n", time.Now().Format("2006-01-02 15:04:05"), i, err.Error())
		}
	}
	return nil
}

func (c *ReportClient) serialize(group *NamedMetrics) (io.Reader, error) {
	requestContent, err := json.Marshal(map[string]interface{}{group.Name: group.Metrics})
	if err != nil {
		return nil, err
	}
	base64Content := make([]byte, base64.StdEncoding.EncodedLen(len(requestContent)))
	base64.StdEncoding.Encode(base64Content, requestContent)
	return compress.CompressWithGzip(bytes.NewBuffer(base64Content))
}

func (c *ReportClient) group(in []*Metric) []*NamedMetrics {
	metrics := &NamedMetrics{
		Name:    "metrics",
		Metrics: make([]*Metric, 0),
	}
	trace := &NamedMetrics{
		Name:    "trace",
		Metrics: make([]*Metric, 0),
	}
	errorG := &NamedMetrics{
		Name:    "error",
		Metrics: make([]*Metric, 0),
	}
	for _, m := range in {
		switch m.Name {
		case "trace":
		case "span":
			trace.Metrics = append(trace.Metrics, m)
			break
		case "error":
			errorG.Metrics = append(errorG.Metrics, m)
			break
		default:
			metrics.Metrics = append(metrics.Metrics, m)
		}
	}
	return []*NamedMetrics{metrics, trace, errorG}
}

func (c *ReportClient) write(name string, requestBuffer io.Reader) error {
	req, err := http.NewRequest(http.MethodPost, c.formatRoute(name), requestBuffer)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Custom-Content-Encoding", "base64")
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(c.CFG.ReportConfig.Collector.UserName, c.CFG.ReportConfig.Collector.UserName)
	resp, err := c.HttpClient.Do(req)
	if err == nil && (resp.StatusCode < 200 || resp.StatusCode >= 300) {
		err = errors.Errorf("when writing to [%s] received status code: %d/n", c.formatRoute(name), resp.StatusCode)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Printf("%s error! close response body error %s", time.Now().Format("2006-01-02 15:04:05"), err)
		}
	}()
	return err
}

func (c *ReportClient) formatRoute(name string) string {
	return fmt.Sprintf("http://%s/collect/%s", c.CFG.ReportConfig.Collector.Addr, name)
}
