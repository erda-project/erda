// Author: recallsong
// Email: ruiguo.srg@alibaba-inc.com

package elasticsearchr

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync/atomic"

	"github.com/erda-project/erda-infra/base/servicehub"
	log "github.com/sirupsen/logrus"
)

type define struct{}

func (d *define) Service() []string      { return []string{"logs-elasticsearch-proxy"} }
func (d *define) Dependencies() []string { return []string{} }
func (d *define) Summary() string        { return "elasticsearch proxy" }
func (d *define) Description() string    { return d.Summary() }
func (d *define) Config() interface{}    { return &config{} }
func (d *define) Creator() servicehub.Creator {
	return func() servicehub.Provider {
		return &provider{}
	}
}

type config struct {
	Addr    string `file:"addr"`
	Targets string `file:"targets"`
}

type provider struct {
	C       *config
	L       logs.Logger
	server  *http.Server
	targets []string
	index   int32
}

func (p *provider) Init(ctx servicehub.Context) error {
	p.targets = strings.Split(p.C.Targets, ",")
	p.server = &http.Server{Addr: p.C.Addr, Handler: http.HandlerFunc(p.Handler)}
	return nil
}

func (p *provider) Start() error {
	p.L.Infof("start http proxy server at %s", p.C.Addr)
	return p.server.ListenAndServe()
}

func (p *provider) Close() error {
	return p.server.Close()
}

func (p *provider) getTargetURL() string {
	idx := atomic.AddInt32(&p.index, 1)
	return p.targets[int(idx)%len(p.targets)]
}

func (p *provider) Handler(w http.ResponseWriter, req *http.Request) {
	// target url
	path, err := url.PathUnescape(req.URL.Path)
	if err == nil {
		if strings.HasPrefix(path, "/<") && strings.HasSuffix(path, ">") {
			path = "/" + url.QueryEscape(path[1:])
		}
	} else {
		path = req.URL.Path
	}
	urlstr := p.getTargetURL() + path + "?" + req.URL.RawQuery
	u, err := url.Parse(urlstr)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Errorf("invalid proxy url: %s", urlstr)
		return
	}
	request := &http.Request{
		Method: req.Method,
		URL:    u,
		Header: req.Header,
		Body:   req.Body,
	}
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("%s", err)))
		log.Errorf("fail to proxy url: %s", err)
		return
	}
	for key, vals := range resp.Header {
		for _, val := range vals {
			w.Header().Add(key, val)
		}
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func init() {
	servicehub.RegisterProvider("logs-elasticsearch-proxy", &define{})
}
