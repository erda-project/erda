package ctxhelper

import (
	"context"
	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
	"sync"
)

type McpInfo struct {
	Name           string
	Version        string
	Host           string
	Scheme         string
	NeedTerminusId bool
}

func GetMcpInfo(ctx context.Context) (*McpInfo, bool) {
	info, ok := ctx.Value(CtxKeyMap{}).(*sync.Map).Load(vars.MapKeyMcpInfo{})
	if info == nil || !ok {
		return nil, false
	}
	mcpInfo, ok := info.(*McpInfo)
	return mcpInfo, ok
}

func PutMcpInfo(ctx context.Context, info *McpInfo) {
	ctx.Value(CtxKeyMap{}).(*sync.Map).Store(vars.MapKeyMcpInfo{}, info)
}
