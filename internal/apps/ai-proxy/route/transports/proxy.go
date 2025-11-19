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

package transports

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/net/http/httpproxy"
	"golang.org/x/net/proxy"
)

const (
	envKeyForwardProxyHosts = "FORWARD_PROXY_HOSTS"
	envKeyForwardHttpProxy  = "FORWARD_HTTP_PROXY"
	envKeyForwardHttpsProxy = "FORWARD_HTTPS_PROXY"
	envKeyNoProxy           = "NO_PROXY"
)

// shared base dialer for all outbound connections
var baseDialer = &net.Dialer{
	Timeout:   10 * time.Second, // Timeout for establishing a TCP connection, even when routing through proxies.
	KeepAlive: 60 * time.Second, // Keep-alive interval to keep long-lived TCP connections reusable.
}

// forwardProxyHosts caches the FORWARD_PROXY_HOSTS list for both HTTP and SOCKS proxy decisions.
var forwardProxyHosts = func() []string {
	hosts := strings.Split(os.Getenv(envKeyForwardProxyHosts), ",")
	var cleaned []string
	for _, h := range hosts {
		h = strings.TrimSpace(h)
		if h == "" {
			continue
		}
		cleaned = append(cleaned, h)
	}
	return cleaned
}()

// socksProxyURL and socksDialer are non-nil only when FORWARD_HTTP_PROXY / FORWARD_HTTPS_PROXY
// is configured with a socks5:// scheme.
var (
	socksProxyURL *url.URL
	socksDialer   proxy.Dialer
)

// init forward proxy configuration (HTTP/HTTPS) and optional SOCKS5 dialer.
func init() {
	httpProxy := os.Getenv(envKeyForwardHttpProxy)
	httpsProxy := os.Getenv(envKeyForwardHttpsProxy)

	// if both are set and not equal, fail fast.
	if httpProxy != "" && httpsProxy != "" && httpProxy != httpsProxy {
		panic(fmt.Sprintf("%s (%q) and %s (%q) must be equal when both are set", envKeyForwardHttpProxy, envKeyForwardHttpsProxy, httpProxy, httpsProxy))
	}

	// normalize: if only one is set, copy it to the other so that the effective config is consistent.
	if httpProxy == "" {
		httpProxy = httpsProxy
	}
	if httpsProxy == "" {
		httpsProxy = httpProxy
	}

	ProxyConfig = &httpproxy.Config{
		HTTPProxy:  httpProxy,
		HTTPSProxy: httpsProxy,
		NoProxy:    os.Getenv(envKeyNoProxy),
		CGI:        os.Getenv("REQUEST_METHOD") != "",
	}

	// initialize SOCKS5 dialer if the effective proxy is socks5://
	raw := httpsProxy
	if raw == "" {
		raw = httpProxy
	}
	if raw != "" {
		if u, err := url.Parse(raw); err == nil && strings.EqualFold(u.Scheme, "socks5") {
			auth := &proxy.Auth{}
			if u.User != nil {
				auth.User = u.User.Username()
				if pw, ok := u.User.Password(); ok {
					auth.Password = pw
				}
			}
			if d, err := proxy.SOCKS5("tcp", u.Host, auth, baseDialer); err == nil {
				socksProxyURL = u
				socksDialer = d
			} else {
				logrus.WithError(err).Errorf("failed to init socks5 proxy for %s", raw)
			}
		}
	}

	// print final proxy configuration at startup.
	socksEnabled := socksProxyEnabled()
	logrus.Infof(
		"ai-proxy forward proxy config: HTTP=%q HTTPS=%q NO_PROXY=%q SOCKS5_ENABLED=%t SOCKS5_URL=%v HOSTS=%v",
		ProxyConfig.HTTPProxy,
		ProxyConfig.HTTPSProxy,
		ProxyConfig.NoProxy,
		socksEnabled,
		socksProxyURL,
		forwardProxyHosts,
	)
}

func socksProxyEnabled() bool {
	return socksDialer != nil && socksProxyURL != nil
}

// ProxyConfig is forward proxy configuration, i.e., proxy configuration for transport outbound traffic
var ProxyConfig = &httpproxy.Config{}
