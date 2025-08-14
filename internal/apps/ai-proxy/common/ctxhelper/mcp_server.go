package ctxhelper

import (
	"context"
	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
	"github.com/erda-project/erda/pkg/reverseproxy"
	"sync"
)

type McpInfo struct {
	Name          string
	Version       string
	TransportType string
	Host          string
	Scheme        string
}

func GetMcpInfo(ctx context.Context) (*McpInfo, bool) {
	info, ok := ctx.Value(reverseproxy.CtxKeyMap{}).(*sync.Map).Load(vars.MapKeyMcpInfo{})
	if info == nil || !ok {
		return nil, false
	}
	mcpInfo, ok := info.(*McpInfo)
	return mcpInfo, ok
}

func PutMcpInfo(ctx context.Context, info *McpInfo) {
	ctx.Value(reverseproxy.CtxKeyMap{}).(*sync.Map).Store(vars.MapKeyMcpInfo{}, info)
}
