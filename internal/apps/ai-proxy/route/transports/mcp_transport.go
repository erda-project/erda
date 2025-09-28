package transports

import (
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/pkg/clusterdialer"
	"github.com/sirupsen/logrus"
	"net/http"
)

type McpTransport struct {
	http.RoundTripper
}

func NewMcpTransport() *McpTransport {
	return &McpTransport{
		BaseTransport,
	}
}

func (t *McpTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	transport, ok := t.RoundTripper.(*http.Transport)
	if !ok {
		logrus.Infof("failed to cast transport to http.Transport")
		return t.RoundTripper.RoundTrip(req)
	}

	info, ok := ctxhelper.GetMcpInfo(req.Context())
	if !ok {
		logrus.Infof("failed to cast context to mcp info")
		return t.RoundTripper.RoundTrip(req)
	}

	logrus.Infof("with clusterdialer, cluster name: %s", info.ClusterName)
	transport.DialContext = clusterdialer.DialContext(info.ClusterName)

	return transport.RoundTrip(req)
}
