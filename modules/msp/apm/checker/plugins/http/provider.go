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

package http

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-proto-go/msp/apm/checker/pb"
	"github.com/erda-project/erda/modules/msp/apm/checker/plugins"
)

type config struct {
	Headers http.Header `file:"headers"`
	Trace   struct {
		Enable bool `file:"enable" default:"true"`
		Rate   int  `file:"rate" default:"50"`
	} `file:"trace"`
}

// +provider
type provider struct {
	Cfg *config
	Log logs.Logger
}

func (p *provider) Init(ctx servicehub.Context) error { return nil }

func (p *provider) Validate(c *pb.Checker) error {
	urlstr := c.Config["url"]
	if len(urlstr) <= 0 {
		return fmt.Errorf("url must not be empdy")
	}
	_, err := url.Parse(urlstr)
	if err != nil {
		return fmt.Errorf("invalid url: %s", err)
	}
	return nil
}

func (p *provider) New(c *pb.Checker) (plugins.Handler, error) {
	// method
	method := c.Config["method"]
	if len(method) <= 0 {
		method = http.MethodGet
	}

	// url
	urlstr := c.Config["url"]
	if len(urlstr) <= 0 {
		return nil, fmt.Errorf("url must not be empdy")
	}
	_, err := url.Parse(urlstr)
	if err != nil {
		return nil, fmt.Errorf("invalid url: %s", err)
	}

	// headers
	headers := http.Header{}
	for k, vals := range p.Cfg.Headers {
		for _, v := range vals {
			headers.Add(k, v)
		}
	}
	for k, v := range c.Config {
		if strings.HasPrefix(k, "header_") {
			k = k[len("header_"):]
			headers.Add(k, v)
		}
	}

	// timeout
	var timeout time.Duration
	if val := c.Config["timeout"]; len(val) > 0 {
		timeout, err = time.ParseDuration(val)
		if err != nil {
			return nil, fmt.Errorf("invalid timeout: %s", err)
		}
	}
	if timeout <= 0 {
		timeout = 5 * time.Second
	}

	// trace
	var trace bool
	if val := c.Config["trace"]; len(val) > 0 {
		trace, err = strconv.ParseBool(val)
		if err != nil {
			return nil, fmt.Errorf("invalid trace: %s", err)
		}
	}

	return &httpHandler{
		p:       p,
		tags:    c.Tags,
		method:  method,
		url:     urlstr,
		headers: headers,
		body:    c.Config["body"],
		timeout: timeout,
		trace:   trace,
	}, nil
}

type httpHandler struct {
	p       *provider
	tags    map[string]string
	method  string
	url     string
	headers http.Header
	body    string
	timeout time.Duration
	trace   bool
}

func (h httpHandler) Do(ctx plugins.Context) error {
	select {
	case <-ctx.Done():
		return nil
	default:
	}

	tags, fields := make(map[string]string), make(map[string]interface{})
	for k, v := range h.tags {
		tags[k] = v
	}
	tags["url"] = h.url
	fields["retry"] = 0 // for compatibility

	req, err := http.NewRequestWithContext(ctx, h.method, h.url, strings.NewReader(h.body))
	if err != nil {
		fields["latency"] = 0
		fields["code"] = 601
		fields["message"] = err.Error()
		tags["status"] = "2"                 // for compatibility
		tags["status_name"] = "Major Outage" // for compatibility
		ctx.Report(&plugins.Metric{
			Name:   "status_page",
			Tags:   tags,
			Fields: fields,
		})
		return err
	}

	// setup trage if need
	if h.trace || h.p.shouldSample() {
		h.p.setupTrace(req, tags, fields)
	}

	// create client with timeout
	client := &http.Client{
		Timeout: h.timeout,
		Transport: &http.Transport{
			Dial: func(netw, addr string) (net.Conn, error) {
				c, err := net.DialTimeout(netw, addr, h.timeout*3)
				if err != nil {
					return nil, err
				}
				c.SetDeadline(time.Now().Add(h.timeout))
				return c, nil

			},
		},
	}

	// do request
	start := time.Now()
	resp, err := client.Do(req)
	latency := time.Now().Sub(start).Milliseconds()
	fields["latency"] = latency
	if err != nil {
		fields["code"] = 601
		fields["message"] = err.Error()
		tags["status"] = "2"                 // for compatibility
		tags["status_name"] = "Major Outage" // for compatibility
	} else {
		fields["code"] = resp.StatusCode
		if resp.StatusCode >= 400 {
			fields["message"] = http.StatusText(resp.StatusCode)
			tags["status"] = "2"                 // for compatibility
			tags["status_name"] = "Major Outage" // for compatibility
		} else {
			tags["status"] = "1"                // for compatibility
			tags["status_name"] = "Operational" // for compatibility
		}
	}

	ctx.Report(&plugins.Metric{
		Name:   "status_page",
		Tags:   tags,
		Fields: fields,
	})
	return nil
}

func init() {
	servicehub.Register("erda.msp.apm.checker.task.plugins.http", &servicehub.Spec{
		Services:   []string{"erda.msp.apm.checker.task.plugins.http"},
		ConfigFunc: func() interface{} { return &config{} },
		Creator:    func() servicehub.Provider { return &provider{} },
	})
}
