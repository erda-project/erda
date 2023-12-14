package aiproxyclient

import (
	"context"
	"fmt"
	"reflect"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/pkg/transport"
	sessionpb "github.com/erda-project/erda-proto-go/apps/aiproxy/session/pb"
	"github.com/erda-project/erda/pkg/http/httputil"
)

type Interface interface {
	AIEnabled() bool
	Context(yourCtx ...context.Context) context.Context
	Session() sessionpb.SessionServiceServer
}

var Instance Interface

var ErrorAINotEnabled = fmt.Errorf("AI not enabled")

type config struct {
	URL      string `file:"url" env:"AI_PROXY_URL"`
	ClientAK string `file:"client_ak" env:"AI_PROXY_CLIENT_AK"`
}

type provider struct {
	Cfg *config
	Log logs.Logger

	AIProxySession sessionpb.SessionServiceServer `optional:"false"`
}

func (p *provider) AIEnabled() bool {
	return p.Cfg.URL != "" && p.Cfg.ClientAK != ""
}

func (p *provider) Context(yourCtx ...context.Context) context.Context {
	ctx := context.Background()
	if len(yourCtx) > 0 {
		ctx = yourCtx[0]
	}

	currentHeader := transport.ContextHeader(ctx)
	if currentHeader == nil {
		currentHeader = transport.Header{}
	}

	// authorization
	if v := currentHeader.Get(httputil.HeaderKeyAuthorization); len(v) == 0 {
		currentHeader.Set(httputil.HeaderKeyAuthorization, p.Cfg.ClientAK)
	}

	return transport.WithHeader(ctx, currentHeader)
}

func (p *provider) Session() sessionpb.SessionServiceServer {
	return p.AIProxySession
}

func (p *provider) Init(ctx servicehub.Context) error {
	p.Log.Info("URL: ", p.Cfg.URL)
	p.Log.Info("ClientAK: ", p.Cfg.ClientAK)

	Instance = ctx.Service(serviceName).(Interface)
	return nil
}

func (p *provider) Run(ctx context.Context) error {
	return nil
}

const (
	serviceName = "ai-proxy-client"
)

func init() {
	servicehub.Register(serviceName, &servicehub.Spec{
		Services:   []string{serviceName},
		Types:      []reflect.Type{reflect.TypeOf((*Interface)(nil)).Elem()},
		ConfigFunc: func() interface{} { return &config{} },
		Creator:    func() servicehub.Provider { return &provider{} },
	})
}
