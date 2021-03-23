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

type QueryAction struct {
	endpoint string
	timeout  time.Duration

	httpClient *http.Client
}

func (client *QueryAction) SetTimeout(duration time.Duration) {
	client.httpClient.Timeout = duration
}

func (client *QueryAction) QueryMetric(req *MetricQueryRequest) (*MetricQueryResponse, error) {
	response := &MetricQueryResponse{}
	err := client.doAction(req, response)
	return response, err
}

func (client *QueryAction) doAction(req *MetricQueryRequest, resp *MetricQueryResponse) error {
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
