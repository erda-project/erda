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

package mcp_server_request

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"regexp"
	"strings"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-proto-go/apps/aiproxy/mcp_server/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filter_define"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filters/request/mcp-server-request/request"
	setrespbodychunksplitter "github.com/erda-project/erda/internal/apps/ai-proxy/route/filters/request/set-resp-body-chunk-splitter"
)

const Name = "mcp-server-request"

const TerminusRequestIdHeader = "Terminus-Request-Id"

var (
	_ filter_define.ProxyRequestRewriter = (*Filter)(nil)
)

func init() {
	filter_define.RegisterFilterCreator(Name, Creator)
}

type Filter struct{}

var Creator filter_define.RequestRewriterCreator = func(_ string, _ json.RawMessage) filter_define.ProxyRequestRewriter {
	return &Filter{}
}

func (f *Filter) OnProxyRequest(pr *httputil.ProxyRequest) error {
	req := pr.In
	routePath := req.URL.String()
	logger := ctxhelper.MustGetLogger(req.Context())
	logger.Debugf("%s", routePath)

	ctxhelper.PutRespBodyChunkSplitter(req.Context(), &setrespbodychunksplitter.NewLineSplitter{})

	if strings.HasPrefix(routePath, "/proxy/connect") {
		name, version, err := parseMcpPath(routePath)
		if err == nil {
			return f.OnConnect(logger, req.Context(), name, version, pr.Out)
		}
	}

	if strings.HasPrefix(routePath, "/proxy/message") {
		name, tag, messagePath, sessionId, err := parseSseMessagePath(routePath)
		if err == nil {
			return f.OnMessage(logger, req.Context(), name, tag, messagePath, sessionId, pr.Out)
		}
	}

	return fmt.Errorf("not supported router path %v", routePath)
}

func (f *Filter) OnConnect(logger logs.Logger, ctx context.Context, name, version string, req *http.Request) error {

	client := ctxhelper.MustGetDBClient(ctx)
	resp, err := client.MCPServerClient().Get(ctx, &pb.MCPServerGetRequest{
		Name:    name,
		Version: version,
	})

	if err != nil {
		logger.Errorf("get cache server failed: %v", err)
		return err
	}
	server := resp.GetData()

	parsedEndpoint, err := parseEndpoint(server.Endpoint)
	if err != nil {
		logger.Error("parse endpoint failed", err)
		return err
	}

	endpoint := fmt.Sprintf("%s:%s", parsedEndpoint.Host, parsedEndpoint.Port)

	ctxhelper.PutmcpMcpInfo(ctx, ctxhelper.McpInfo{
		Name:    name,
		Version: version,
		Host:    endpoint,
		Scheme:  parsedEndpoint.Scheme,
	})

	req.Host = endpoint
	req.URL.Scheme = parsedEndpoint.Scheme
	req.URL.Host = endpoint
	req.URL.Path = parsedEndpoint.Path
	req.Header.Set("Host", parsedEndpoint.Host)
	return nil
}

func (f *Filter) OnMessage(logger logs.Logger, ctx context.Context, name string, tag string, messagePath string, sessionId string, req *http.Request) error {
	client := ctxhelper.MustGetDBClient(ctx)
	resp, err := client.MCPServerClient().Get(ctx, &pb.MCPServerGetRequest{
		Name:    name,
		Version: tag,
	})

	if err != nil {
		logger.Errorf("get cache server failed: %v", err)
		return err
	}
	server := resp.GetData()

	parsedEndpoint, err := parseEndpoint(server.Endpoint)
	if err != nil {
		logger.Error("parse endpoint failed", err)
		return err
	}

	logger.Infof("server info: %+v", parsedEndpoint)

	raw, err := io.ReadAll(req.Body)
	req.Body.Close()
	if err != nil {
		return err
	}

	if err = filterTerminusTraceId(raw, req); err != nil {
		logger.Errorf("filter terminus trace id failed: %v", err)
		return err
	}

	logger.Infof("request body %s", string(raw))

	reader := bytes.NewReader(raw)
	closer := io.NopCloser(reader)
	req.Body = closer
	req.ContentLength = int64(len(raw))

	req.Host = parsedEndpoint.Host
	req.URL.Scheme = parsedEndpoint.Scheme
	req.URL.Host = parsedEndpoint.Host
	req.URL.Path = messagePath

	return nil
}

func parseMcpPath(path string) (name, version string, err error) {
	re := regexp.MustCompile(`^/proxy/connect/([^/]+)/([^/]+)$`)
	matches := re.FindStringSubmatch(path)

	if len(matches) != 3 {
		return "", "", fmt.Errorf("path does not match expected format")
	}

	return matches[1], matches[2], nil
}

func parseSseMessagePath(path string) (name, tag, messagePath, sessionId string, err error) {
	re := regexp.MustCompile(`^/proxy/message/([^/]+)/([^/]+)(?P<sub_path>/[^?]+)\?sessionId=(?P<sessionId>[0-9a-fA-F-]+)$`)
	matches := re.FindStringSubmatch(path)
	if matches == nil || len(matches) < 5 {
		return "", "", "", "", fmt.Errorf("failed to parse sessionId")
	}
	return matches[1], matches[2], matches[3], matches[4], nil
}

type ParsedEndpoint struct {
	Endpoint string
	Host     string
	Port     string
	Path     string
	Scheme   string
}

func parseEndpoint(endpoint string) (*ParsedEndpoint, error) {
	re := regexp.MustCompile(`^(?P<scheme>https?)://(?P<host>[^/:]+)(?::(?P<port>\d+))?(?P<path>/.*)?$`)
	matches := re.FindStringSubmatch(endpoint)
	if matches == nil {
		return nil, fmt.Errorf("failed to parse endpoint %s", endpoint)
	}

	names := re.SubexpNames()
	result := map[string]string{}
	for i, name := range names {
		if i != 0 && name != "" {
			result[name] = matches[i]
		}
	}

	scheme := result["scheme"]
	host := result["host"]
	port := result["port"]
	path := result["path"]

	if scheme == "" || host == "" {
		return nil, errors.New("missing scheme or host")
	}

	if port == "" {
		switch scheme {
		case "http":
			port = "80"
		case "https":
			port = "443"
		default:
			return nil, errors.New("unsupported scheme")
		}
	}

	if path == "" {
		path = "/"
	}

	return &ParsedEndpoint{
		Endpoint: endpoint,
		Host:     host,
		Port:     port,
		Path:     path,
		Scheme:   scheme,
	}, nil
}

func filterTerminusTraceId(body []byte, rawReq *http.Request) error {
	var req request.Request
	if err := json.Unmarshal(body, &req); err != nil {
		return err
	}

	var traceId string

	if id, exist := req.Params.Arguments[TerminusRequestIdHeader]; exist {
		traceId = fmt.Sprintf("%v", id)
		delete(req.Params.Arguments, TerminusRequestIdHeader)
	}
	jsonData, err := json.Marshal(req)
	if err != nil {
		return err
	}
	rawReq.Body = io.NopCloser(bytes.NewReader(jsonData))
	rawReq.ContentLength = int64(len(jsonData))

	if traceId != "" {
		rawReq.Header.Set(TerminusRequestIdHeader, traceId)
	}

	return nil
}
