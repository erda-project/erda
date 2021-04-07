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

package query

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var (
	metricApiV1 = "api/metrics"
)

type MetricQuery interface {
	QueryMetric(req *MetricQueryRequest) (*MetricQueryResponse, error)

	SetTimeout(duration time.Duration)
}

type queryClient struct {
	endpoint string
	timeout  time.Duration

	httpClient *http.Client
}

func (client *queryClient) SetTimeout(duration time.Duration) {
	client.httpClient.Timeout = duration
}

func (client *queryClient) QueryMetric(req *MetricQueryRequest) (*MetricQueryResponse, error) {
	response := &MetricQueryResponse{}
	err := client.doAction(req, response)
	return response, err
}

func (client *queryClient) doAction(req *MetricQueryRequest, resp *MetricQueryResponse) error {
	if req.diagram == "histogram" && req.point == 0 {
		return errors.New("point must be limit when diagram is histogram")
	}
	// 1. marshal QueryMetricRequest to http.Request
	// 2. wait for response
	// 3. unmarshal http.Response to QueryMetricResponse
	path := []string{metricApiV1, req.scope}
	if req.diagram != "" {
		path = append(path, req.diagram)
	}
	dst, err := url.Parse(client.endpoint)
	if err != nil {
		return err
	}
	dst.Path = strings.Join(path, "/")
	dst.RawQuery = req.ConstructParam().Encode()

	request, err := http.NewRequest("GET", dst.String(), nil)
	if err != nil {
		return err
	}
	response, err := client.httpClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	return marshalResponse(response, resp)
}

func marshalResponse(response *http.Response, resp *MetricQueryResponse) error {
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}
	resp.StatusCode = response.StatusCode
	if resp.StatusCode != http.StatusOK {
		serverError := NewServerError(body)
		return serverError
	}
	resp.Body = body
	return nil
}
