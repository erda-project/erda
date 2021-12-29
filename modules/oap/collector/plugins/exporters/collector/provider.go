package collector

import (
	"bytes"
	"fmt"
	"net/http"
	"time"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda/modules/oap/collector/common/compressor"
	"github.com/erda-project/erda/modules/oap/collector/core/model"
	"github.com/erda-project/erda/modules/oap/collector/plugins"
	"github.com/erda-project/erda/modules/oap/collector/plugins/exporters/collector/auth"
)

var providerName = plugins.WithPrefixExporter("collector")

type config struct {
	URL             string        `file:"url"`
	Timeout         time.Duration `file:"timeout" default:"3s"`
	Serializer      string        `file:"serializer" default:"json"`
	ContentEncoding string        `file:"content_encoding" default:"gzip"`
	Authentication  *struct {
		Type    string                 `file:"type" default:"token"`
		Options map[string]interface{} `file:"options"`
	} `file:"authentication"`
	Headers map[string]string `file:"headers"`
	// capability of old data format
	Compatibility bool `file:"compatibility" default:"true"`
}

// +provider
type provider struct {
	Cfg *config
	Log logs.Logger

	client *http.Client
	au     auth.Authenticator
	cp     compressor.Compressor
}

func (p *provider) ComponentID() model.ComponentID {
	return model.ComponentID(providerName)
}

func (p *provider) Connect() error {
	client := &http.Client{
		Timeout: p.Cfg.Timeout,
	}
	p.client = client

	headReq, err := http.NewRequest(http.MethodHead, p.Cfg.URL, nil)
	if err != nil {
		return err
	}
	code, err := doRequest(p.client, headReq)
	if err != nil {
		return err
	}
	if code >= 500 {
		return fmt.Errorf("invalid response code: %d", code)
	}
	return nil
}

func (p *provider) Export(data model.ObservableData) error {
	buf, err := doSerialize(p.Cfg.Serializer, data, p.Cfg.Compatibility)
	if err != nil {
		return fmt.Errorf("serialize err: %w", err)
	}
	buf, err = p.cp.Compress(buf)
	if err != nil {
		return fmt.Errorf("compress err: %w", err)
	}
	req, err := http.NewRequest(http.MethodPost, p.Cfg.URL, bytes.NewReader(buf))
	if err != nil {
		return fmt.Errorf("create request err: %w", err)
	}
	setHeaders(req, p.Cfg.Headers)
	p.au.Secure(req)
	code, err := doRequest(p.client, req)
	if err != nil {
		return fmt.Errorf("do request err: %w")
	}
	if code < 200 || code >= 300 {
		return fmt.Errorf("response status code %d is not success", code)
	}
	return nil
}

func (p *provider) Init(ctx servicehub.Context) error {
	if err := p.Connect(); err != nil {
		return fmt.Errorf("try connect to remote err: %w", err)
	}
	if err := p.createAuthenticator(); err != nil {
		return fmt.Errorf("createAuthenticator err: %w", err)
	}
	if err := p.createCompressor(); err != nil {
		return fmt.Errorf("createCompressor err: %w", err)
	}
	return nil
}

func (p *provider) createAuthenticator() error {
	cfg := p.Cfg.Authentication.Options
	switch auth.AuthenticationType(p.Cfg.Authentication.Type) {
	case auth.Basic:
		au, err := auth.NewBasicAuth(cfg)
		if err != nil {
			return err
		}
		p.au = au
	case auth.Token:
		au, err := auth.NewTokenAuth(cfg)
		if err != nil {
			return err
		}
		p.au = au
	default:
		return fmt.Errorf("invalid authentication type: %q", p.Cfg.Authentication.Type)
	}
	return nil
}

func (p *provider) createCompressor() error {
	switch p.Cfg.ContentEncoding {
	case GZIPEncoding:
		p.cp = compressor.NewGzipEncoder(3)
	default:
		return fmt.Errorf("invalid ContentEncoding: %q", p.Cfg.ContentEncoding)
	}
	return nil
}

func init() {
	servicehub.Register(providerName, &servicehub.Spec{
		Services: []string{
			providerName,
		},
		Description: "here is description of erda.oap.collector.exporter.collector",
		ConfigFunc: func() interface{} {
			return &config{}
		},
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
