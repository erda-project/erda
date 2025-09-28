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

package mcp

import (
	"context"
	"errors"
	"fmt"
	"github.com/erda-project/erda/pkg/clusterdialer"
	"io"
	"net/http"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
	"github.com/mark3labs/mcp-go/mcp"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/wrapperspb"
	corev1 "k8s.io/api/core/v1"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-proto-go/apps/aiproxy/mcp_server/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/handlers/handler_mcp_server"
	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
	"github.com/erda-project/erda/pkg/http/customhttp"
)

type Register struct {
	handler *handler_mcp_server.MCPHandler
	logger  logs.Logger
}

func NewRegister(handler *handler_mcp_server.MCPHandler, logger logs.Logger) *Register {
	return &Register{
		handler: handler,
		logger:  logger,
	}
}

func (r *Register) register(ctx context.Context, svc *corev1.Service, clusterName string) error {
	name, ok := svc.Labels[vars.LabelMcpErdaCloudName]
	if !ok {
		return errors.New("service label mcp.erda.cloud/name not found")
	}
	version, ok := svc.Labels[vars.LabelMcpErdaCloudVersion]
	if !ok {
		return errors.New("service label mcp.erda.cloud/version not found")
	}

	isPublished, ok := svc.Labels[vars.LabelMcpErdaCloudIsPublished]
	if !ok {
		return errors.New("service label mcp.erda.cloud/is_published not found")
	}

	isDefault, ok := svc.Labels[vars.LabelMcpErdaCloudIsDefault]
	if !ok {
		return errors.New("service label mcp.erda.cloud/is_default not found")
	}

	port, ok := svc.Labels[vars.LabelMcpErdaCloudServicePort]
	if !ok {
		return errors.New("service label mcp.erda.cloud/is_default not found")
	}

	transportType, ok := svc.Labels[vars.LabelMcpErdaCloudTransportType]
	if !ok {
		transportType = "sse"
	}

	if transportType != "sse" && transportType != "streamable" {
		return fmt.Errorf("service label mcp.erda.cloud/transport_type %s not supported", transportType)
	}

	description, ok := svc.Annotations[vars.AnnotationMcpErdaCloudDescription]
	if !ok {
		description = ""
	}

	uri, ok := svc.Annotations[vars.AnnotationMcpErdaCloudConnectUri]

	scopeId := svc.Labels[vars.LabelMcpErdaCloudServiceScopeId]
	if scopeId == "" {
		scopeId = "0"
	}

	scopeType := svc.Labels[vars.LabelMcpErdaCloudServiceScopeType]
	if scopeType == "" {
		scopeType = "org"
	}

	svcHost := fmt.Sprintf("%s.%s.svc.cluster.local:%s", svc.Name, svc.Namespace, port)

	url := fmt.Sprintf("inet://%s/%s%s", clusterName, svcHost, uri)

	endpoint := fmt.Sprintf("http://%s%s", svcHost, uri)

	r.logger.Infof("endpoint: %s, clusterName: %s, uri: %s", endpoint, clusterName, uri)

	tools, err := r.listTools(ctx, transportType, endpoint, clusterName)
	if err != nil {
		return err
	}

	serverConfig, err := r.requestServerInfo(clusterName, svcHost)
	if err != nil {
		return err
	}

	req := &pb.MCPServerRegisterRequest{
		Name:             name,
		Version:          version,
		IsPublished:      wrapperspb.Bool(isPublished == "true"),
		IsDefaultVersion: wrapperspb.Bool(isDefault == "true"),
		TransportType:    transportType,
		Description:      description,
		Tools:            tools,
		Endpoint:         url,
		ScopeType:        &scopeType,
		ScopeId:          &scopeId,
		ServerConfig:     serverConfig,
	}
	_, err = r.handler.Register(ctx, req)
	return err
}

func (r *Register) listTools(ctx context.Context, transportType string, url string, clusterName string) ([]*pb.MCPServerTool, error) {
	var mcpClient *client.Client
	var err error

	httpClient := &http.Client{
		Transport: &http.Transport{
			DialContext: clusterdialer.DialContext(clusterName),
		},
	}

	switch transportType {
	case "sse":
		mcpClient, err = client.NewSSEMCPClient(url, transport.WithHTTPClient(httpClient))
		if err != nil {
			return nil, err
		}
	case "streamable":
		mcpClient, err = client.NewStreamableHttpClient(url, transport.WithHTTPBasicClient(httpClient))
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unknown transport type %s", transportType)
	}
	defer mcpClient.Close()

	if err = mcpClient.Start(ctx); err != nil {
		return nil, err
	}

	request := mcp.InitializeRequest{}
	request.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	request.Params.ClientInfo = mcp.Implementation{
		Name:    "mcp-proxy",
		Version: "1.0.0",
	}
	_, err = mcpClient.Initialize(context.Background(), request)
	if err != nil {
		return nil, fmt.Errorf("mcp initialize error: %s", err)
	}

	tools, err := mcpClient.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		return nil, fmt.Errorf("mcp list tools error: %s", err)
	}

	removeAnyOf(tools)

	mcpTools := make([]*pb.MCPServerTool, 0)
	for _, tool := range tools.Tools {
		properties := make(map[string]*structpb.Struct)

		for key, value := range tool.InputSchema.Properties {
			m, ok := value.(map[string]any)
			if !ok {
				r.logger.Errorf("invalid input schema type: %s", tool.InputSchema.Type)
				continue
			}
			newStruct, err := structpb.NewStruct(m)
			if err != nil {
				r.logger.Errorf("structpb.NewStruct failed: %v", err)
				continue
			}
			properties[key] = newStruct
		}

		mcpTools = append(mcpTools, &pb.MCPServerTool{
			Name:        tool.Name,
			Description: tool.Description,
			InputSchema: &pb.MCPServerToolInputSchema{
				Type:       tool.InputSchema.Type,
				Properties: properties,
				Required:   tool.InputSchema.Required,
			},
		})
	}

	return mcpTools, nil
}

func (r *Register) requestServerInfo(clusterName string, host string) (string, error) {
	url := fmt.Sprintf("inet://%s/%s%s", clusterName, host, "/server/config")
	request, err := customhttp.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		r.logger.Infof("haven't set any server config")
		return "", nil
	}
	all, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	return string(all), nil
}

// removeAnyOf removes anyOf fields from tool schemas
func removeAnyOf(tools *mcp.ListToolsResult) {
	for _, tool := range tools.Tools {
		ProcessAnyOf(tool.InputSchema.Properties)
	}
}

// ProcessAnyOf processes anyOf fields in tool schemas
func ProcessAnyOf(obj interface{}) {
	switch v := obj.(type) {
	case map[string]interface{}:
		// if exist `anyOf`
		if anyOf, ok := v["anyOf"]; ok {
			if list, ok := anyOf.([]interface{}); ok && len(list) > 0 {
				if first, ok := list[0].(map[string]interface{}); ok {
					// remove anyOf
					delete(v, "anyOf")
					// replace the first item
					for k, val := range first {
						v[k] = val
					}
				}
			}
		}
		for _, val := range v {
			ProcessAnyOf(val)
		}
	case []interface{}:
		for _, item := range v {
			ProcessAnyOf(item)
		}
	}
}
