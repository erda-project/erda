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

// Package httpclient impl http client
package httpclient

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/pkg/clusterdialer"
)

const (
	// DialTimeout 建立 tcp 连接的超时时间
	DialTimeout = 15 * time.Second
	// ClientDefaultTimeout 从建立 tcp 到读完 response body 超时时间
	ClientDefaultTimeout = 60 * time.Second
)

type BasicAuth struct {
	name     string
	password string
}

type HTTPClient struct {
	option *Option
	cli    *http.Client
	proto  string
	// set basic auth if necessary
	basicAuth *BasicAuth
	// add token auth if set, e.g. dcos cluster
	tokenAuth string
	// add bearer token auth if set. e.g. k8s cluster
	bearerTokenAuth string
}
type Option struct {
	isHTTPS         bool
	ca              *x509.CertPool
	keyPair         tls.Certificate
	debugWriter     io.Writer
	tracer          Tracer
	checkRedirect   func(req *http.Request, via []*http.Request) error
	loadingPrint    bool
	loadingDesc     string
	proxy           string
	clusterDialKey  string
	cookieJar       http.CookieJar
	dnscache        *DNSCache
	dialerKeepalive time.Duration
	acceptEncoding  string

	dialTimeout   time.Duration
	clientTimeout time.Duration

	enableAutoRetry bool
}

type OpOption func(*Option)

func WithClusterDialer(clusterKey string) OpOption {
	return func(op *Option) {
		// TODO: more elegant way to get current cluster key
		currentClusterKey := os.Getenv("DICE_CLUSTER_NAME")
		if clusterKey != currentClusterKey {
			op.clusterDialKey = clusterKey
		}
	}
}

func WithHTTPS() OpOption {
	return func(op *Option) {
		op.isHTTPS = true
	}
}

func WithAcceptEncoding(ae string) OpOption {
	return func(op *Option) {
		op.acceptEncoding = ae
	}
}

func WithProxy(proxy string) OpOption {
	return func(op *Option) {
		op.proxy = proxy
	}
}

func WithHttpsCertFromJSON(certFile, keyFile, caCrt []byte) OpOption {
	pair, err := tls.X509KeyPair(certFile, keyFile)
	if err != nil {
		logrus.Fatalf("LoadX509KeyPair: %v", err)
	}

	if caCrt == nil {
		return func(op *Option) {
			op.isHTTPS = true
			op.keyPair = pair
		}
	}

	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(caCrt)

	return func(op *Option) {
		op.isHTTPS = true
		op.ca = pool
		op.keyPair = pair
	}
}

func WithDebug(w io.Writer) OpOption {
	return func(op *Option) {
		op.debugWriter = w
	}
}

func WithTracer(w io.Writer, tracer Tracer) OpOption {
	return func(op *Option) {
		op.debugWriter = w
		op.tracer = tracer
	}
}

func WithDnsCache() OpOption {
	return func(op *Option) {
		op.dnscache = defaultDNSCache
	}
}

func WithCompleteRedirect() OpOption {
	return func(op *Option) {
		op.checkRedirect = func(req *http.Request, via []*http.Request) error {
			logrus.Infof("origin: %+v", req)
			oldest := via[0] // oldest first
			req.Header = oldest.Header
			req.Method = oldest.Method
			var err error
			if oldest.GetBody != nil {
				req.Body, err = oldest.GetBody()
				if err != nil {
					return err
				}
			}
			logrus.Infof("modified: %+v", req)
			return nil
		}
	}
}
func WithLoadingPrint(desc string) OpOption {
	return func(r *Option) {
		r.loadingPrint = true
		r.loadingDesc = desc
	}
}

func WithDialerKeepAlive(keepalive time.Duration) OpOption {
	return func(r *Option) {
		r.dialerKeepalive = keepalive
	}
}

func WithTimeout(dialTimeout, clientTimeout time.Duration) OpOption {
	return func(r *Option) {
		r.dialTimeout = dialTimeout
		r.clientTimeout = clientTimeout
	}
}

func WithCookieJar(jar http.CookieJar) OpOption {
	return func(r *Option) {
		r.cookieJar = jar
	}
}

func WithEnableAutoRetry(enableAutoRetry bool) OpOption {
	return func(r *Option) {
		r.enableAutoRetry = enableAutoRetry
	}
}

func mkDialContext(option *Option) func(ctx context.Context, network, addr string) (net.Conn, error) {
	raw := (&net.Dialer{
		Timeout:   option.dialTimeout,
		KeepAlive: option.dialerKeepalive,
	}).DialContext
	dialcontext := raw
	if option.clusterDialKey != "" {
		return clusterdialer.DialContext(option.clusterDialKey)
	}
	if option.dnscache != nil {
		dialcontext = func(ctx context.Context, network, addr string) (net.Conn, error) {
			var host string
			var remain string

			idx := strings.LastIndex(addr, ":")
			if idx == -1 {
				host = addr
				remain = ""
			} else {
				host = addr[:idx]
				remain = addr[idx:]
			}

			// 1. address 是 IP
			// 2. network 是 tcp.*
			// 3. network 是 udp.*
			if net.ParseIP(host) != nil || (!strings.HasPrefix(network, "tcp") && !strings.HasPrefix(network, "udp")) {
				return raw(ctx, network, addr)
			}
			ips, err := option.dnscache.lookup(host)
			if err != nil {
				return raw(ctx, network, addr)
			}
			for _, ip := range ips {
				conn, err := raw(ctx, network, ip.String()+remain)
				if err != nil {
					continue
				}
				return conn, err
			}
			return raw(ctx, network, addr)
		}
	}
	return dialcontext
}

func New(ops ...OpOption) *HTTPClient {
	option := &Option{}
	option.dialTimeout = DialTimeout
	option.clientTimeout = ClientDefaultTimeout
	for _, op := range ops {
		op(option)
	}

	tr := &http.Transport{
		DialContext:  mkDialContext(option),
		MaxIdleConns: 2,
	}
	if option.proxy != "" {
		tr.Proxy = func(request *http.Request) (u *url.URL, err error) {
			return url.Parse(option.proxy)
		}
	}
	if option.dialerKeepalive != 0 {
		tr.IdleConnTimeout = option.dialerKeepalive
	}
	tr.ExpectContinueTimeout = 1 * time.Second
	proto := "http"
	if option.isHTTPS {
		proto = "https"
		if option.ca != nil {
			tr.TLSClientConfig = &tls.Config{
				RootCAs:      option.ca,
				Certificates: []tls.Certificate{option.keyPair},
			}
		} else {
			tr.TLSClientConfig = &tls.Config{
				InsecureSkipVerify: true,
				Certificates:       []tls.Certificate{option.keyPair},
			}
		}
	}

	return &HTTPClient{
		proto: proto,
		cli: &http.Client{
			Transport:     tr,
			Timeout:       option.clientTimeout,
			CheckRedirect: option.checkRedirect,
			Jar:           option.cookieJar,
		},
		option: option,
	}
}

func (c *HTTPClient) BasicAuth(user, password string) *HTTPClient {
	c.basicAuth = &BasicAuth{user, password}
	return c
}

func (c *HTTPClient) TokenAuth(token string) *HTTPClient {
	c.tokenAuth = token
	return c
}

func (c *HTTPClient) BearerTokenAuth(token string) *HTTPClient {
	c.bearerTokenAuth = token
	return c
}

func (c *HTTPClient) Get(host string, retry ...RetryOption) *Request {
	if retry == nil && c.option.enableAutoRetry {
		retry = []RetryOption{RetryErrResp, Retry5XX}
	}
	req := c.newRequest(retry)
	req.method = http.MethodGet
	req.host = host
	return req
}

func (c *HTTPClient) Post(host string, retry ...RetryOption) *Request {
	req := c.newRequest(retry)
	req.method = http.MethodPost
	req.host = host
	return req
}

func (c *HTTPClient) Delete(host string, retry ...RetryOption) *Request {
	req := c.newRequest(retry)
	req.method = http.MethodDelete
	req.host = host
	return req
}

func (c *HTTPClient) Put(host string, retry ...RetryOption) *Request {
	req := c.newRequest(retry)
	req.method = http.MethodPut
	req.host = host
	return req
}

func (c *HTTPClient) Patch(host string, retry ...RetryOption) *Request {
	req := c.newRequest(retry)
	req.method = http.MethodPatch
	req.host = host
	return req
}

func (c *HTTPClient) Head(host string, retry ...RetryOption) *Request {
	if retry == nil {
		retry = []RetryOption{RetryErrResp}
	}
	req := c.newRequest(retry)
	req.method = http.MethodHead
	req.host = host
	return req
}

func (c *HTTPClient) Method(method string, host string, retry ...RetryOption) *Request {
	switch method {
	case http.MethodGet:
		return c.Get(host, retry...)
	case http.MethodPost:
		return c.Post(host, retry...)
	case http.MethodDelete:
		return c.Delete(host, retry...)
	case http.MethodPut:
		return c.Put(host, retry...)
	case http.MethodPatch:
		return c.Patch(host, retry...)
	case http.MethodHead:
		return c.Head(host, retry...)
	default:
		panic(fmt.Errorf("unsupported http method: %s", method))
	}
}

func (c *HTTPClient) newRequest(retry []RetryOption) *Request {
	header := make(map[string]string)
	if c.basicAuth != nil {
		header["Authorization"] = "Basic " + constructBasicAuth(c.basicAuth.name, c.basicAuth.password)
	}

	if len(c.tokenAuth) > 0 {
		header["Authorization"] = "token=" + c.tokenAuth
	}

	if len(c.bearerTokenAuth) > 0 {
		header["Authorization"] = "Bearer " + c.bearerTokenAuth
	}

	if c.option.isHTTPS {
		header["X-Portal-Scheme"] = "https"
	}

	if len(c.option.acceptEncoding) > 0 {
		header["Accept-Encoding"] = c.option.acceptEncoding
	}

	request := &Request{
		cli:         c.cli,
		proto:       c.proto,
		params:      make(url.Values),
		header:      header,
		option:      c.option,
		retryOption: mergeRetryOptions(retry),
	}
	if request.option.debugWriter != nil {
		tracer := Tracer(NewDefaultTracer(request.option.debugWriter))
		if request.option.tracer != nil {
			tracer = request.option.tracer
		}
		request.tracer = tracer
	}
	return request
}

func constructBasicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func (c *HTTPClient) BackendClient() *http.Client {
	return c.cli
}
