package mcp_server_director

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/erda-project/erda-infra/providers/redis"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/filters/cache"
	"github.com/erda-project/erda/internal/apps/ai-proxy/filters/directors/mcp-server-director/session"
	"github.com/erda-project/erda/pkg/reverseproxy"
	"github.com/sirupsen/logrus"
	"net/http"
	"regexp"
)

const McpServerInfoKey = "McpServerInfoKey"

var (
	_ reverseproxy.RequestFilter = (*McpServerDirector)(nil)
)

const Name = "mcp-server-director"

func init() {
	reverseproxy.RegisterFilterCreator(Name, func(message json.RawMessage) (reverseproxy.Filter, error) {
		return New(message, nil)
	})
}

type McpServerDirector struct {
	*reverseproxy.DefaultResponseFilter
	manager session.Manager
}

func New(config json.RawMessage, client redis.Interface) (reverseproxy.Filter, error) {
	var manager session.Manager
	manager = session.NewRemoteManager(client)

	return &McpServerDirector{
		DefaultResponseFilter: reverseproxy.NewDefaultResponseFilter(),
		manager:               manager,
	}, nil
}

func (m *McpServerDirector) OnResponseChunkImmutable(ctx context.Context, infor reverseproxy.HttpInfor, copiedChunk []byte) (signal reverseproxy.Signal, err error) {
	logrus.Debugf("method:%v response: %v", infor.Method(), string(copiedChunk))

	sessionId, err := parseSessionId(string(copiedChunk))
	if err != nil {
		return reverseproxy.Continue, nil
	}
	info, ok := ctxhelper.GetMcpInfo(ctx)
	if !ok {
		return reverseproxy.Intercept, fmt.Errorf("[Proxy Error] mcp info fail")
	}

	logrus.Debugf("Mcp server info: %v", info)
	err = m.manager.Save(sessionId, &session.ServerInfo{
		Host:   info.Host,
		Scheme: info.Scheme,
	})
	if err != nil {
		return reverseproxy.Intercept, fmt.Errorf("[Proxy Error] redis set fail: %v", err)
	}

	logrus.Infof("[Proxy Info] mcp server connect success %v", sessionId)
	return reverseproxy.Continue, nil
}

func (m *McpServerDirector) OnRequest(ctx context.Context, w http.ResponseWriter, infor reverseproxy.HttpInfor) (signal reverseproxy.Signal, err error) {
	req := infor.Request()
	routePath := req.URL.String()
	logrus.Debugf("%s", routePath)

	// handle /connect/mcp-server/{mcpName}/{version}/{transportType}
	name, version, transport, err := parseMcpPath(routePath)
	if err == nil {
		return m.OnConnect(ctx, name, version, transport, req)
	}

	// handle /message
	sessionId, err := parseSseMessagePath(routePath)
	if err == nil {
		return m.OnMessage(ctx, sessionId, req)
	}

	return reverseproxy.Intercept, fmt.Errorf("[Proxy Error] not supported router path %v", routePath)
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
		return nil, fmt.Errorf("[Proxy Error] failed to parse endpoint %s", endpoint)
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
		return nil, errors.New("[Proxy Error] missing scheme or host")
	}

	if port == "" {
		switch scheme {
		case "http":
			port = "80"
		case "https":
			port = "443"
		default:
			return nil, errors.New("[Proxy Error] unsupported scheme")
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

func parseMcpPath(path string) (name, version, transport string, err error) {
	re := regexp.MustCompile(`^/connect/mcp-server/([^/]+)/([^/]+)/([^/]+)$`)
	matches := re.FindStringSubmatch(path)

	if len(matches) != 4 {
		return "", "", "", fmt.Errorf("[Proxy Error] path does not match expected format")
	}

	return matches[1], matches[2], matches[3], nil
}

func parseSessionId(message string) (sessionId string, err error) {
	re := regexp.MustCompile("sessionId=([a-f0-9-]+)")
	matches := re.FindStringSubmatch(message)
	if matches == nil || len(matches) < 2 {
		return "", fmt.Errorf("[Proxy Error] failed to parse sessionId")
	}
	return matches[1], nil
}

func parseSseMessagePath(path string) (sessionId string, err error) {
	re := regexp.MustCompile("^/message\\?sessionId=([a-f0-9\\-]+)$")
	matches := re.FindStringSubmatch(path)
	if matches == nil || len(matches) < 2 {
		return "", fmt.Errorf("[Proxy Error] failed to parse sessionId")
	}
	return matches[1], nil
}

func (m *McpServerDirector) OnConnect(ctx context.Context, name, version, transport string, req *http.Request) (reverseproxy.Signal, error) {

	server, err := cache.GetMcpServer(name, version)
	if err != nil {
		logrus.Errorf("[Proxy Error] get cache server failed: %v", err)
		return reverseproxy.Intercept, err
	}

	if server.TransportType != transport {
		err := errors.New("[Proxy Error] transport type not match")
		logrus.Error(err)
		return reverseproxy.Intercept, err
	}

	parsedEndpoint, err := parseEndpoint(server.Endpoint)
	if err != nil {
		logrus.Error("[Proxy Error] parse endpoint failed", err)
		return reverseproxy.Intercept, err
	}

	endpoint := fmt.Sprintf("%s:%s", parsedEndpoint.Host, parsedEndpoint.Port)

	ctxhelper.PutMcpInfo(ctx, &ctxhelper.McpInfo{
		Name:          name,
		Version:       version,
		TransportType: transport,
		Host:          endpoint,
		Scheme:        parsedEndpoint.Scheme,
	})

	req.Host = endpoint
	req.URL.Scheme = parsedEndpoint.Scheme
	req.URL.Host = endpoint
	req.URL.Path = parsedEndpoint.Path
	req.Header.Set("Host", parsedEndpoint.Host)
	return reverseproxy.Continue, nil
}

func (m *McpServerDirector) OnMessage(ctx context.Context, session string, req *http.Request) (reverseproxy.Signal, error) {
	serverInfo, err := m.manager.Load(session)
	if err != nil {
		return reverseproxy.Intercept, fmt.Errorf("[Proxy Error] load server info failed: %v", err)
	}
	req.Host = serverInfo.Host
	req.URL.Scheme = serverInfo.Scheme
	req.URL.Host = serverInfo.Host
	req.URL.Path = "/message"

	return reverseproxy.Continue, nil
}
