// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package customhttp

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/pkg/errors"

	"github.com/erda-project/erda/pkg/discover"
)

var (
	ErrAddrMiss    = Error{"customhttp: inetaddr miss"}
	ErrInvalidAddr = Error{"customhttp: invalid inetaddr"}
)

type Error struct {
	msg string
}

func (e Error) Error() string {
	return e.msg
}

var dialerAddr string

const (
	clusterDialerAddr       = "CLUSTER_DIALER_ADDR"
	portalPassthroughHeader = "X-Portal-Passthrough"
	portalDirectHeader      = "X-Portal-Direct"
	portalSSLHeader         = "X-Portal-SSL"
	portalSSEHeader         = "X-Portal-SSE"
	portalWSHeader          = "X-Portal-Websocket"
	portalHostHeader        = "X-Portal-Host"
	portalDestHeader        = "X-Portal-Dest"
)

// 用于覆盖根据环境变量取的值
func SetInetAddr(addr string) {
	dialerAddr = addr
}

func init() {
	dialerAddr = discover.ClusterDialer()
}

func parseInetUrl(url string) (portalHost string, portalDest string, portalUrl string, portalArgs map[string]string, err error) {
	url = strings.TrimPrefix(url, "inet://")
	url = strings.Replace(url, "//", "/", -1)
	portalArgs = map[string]string{}
	parts := strings.SplitN(url, "/", 3)
	if len(parts) < 2 {
		err = errors.Wrapf(ErrInvalidAddr, "addr:%s", url)
		return
	}
	portalHost = parts[0]
	portalDest = parts[1]
	portalUrl = ""
	if len(parts) > 2 {
		portalUrl = parts[2]
	}
	portalArgsPos := strings.Index(portalHost, "?")
	if portalArgsPos == -1 {
		return
	}
	argStr := portalHost[portalArgsPos+1:]
	portalHost = portalHost[:portalArgsPos]
	argPairs := strings.Split(argStr, "&")
	for _, pair := range argPairs {
		argParts := strings.Split(pair, "=")
		if len(argParts) == 2 {
			portalArgs[argParts[0]] = argParts[1]
		}
	}
	return
}

func NewRequest(method, url string, body io.Reader) (*http.Request, error) {
	if !strings.HasPrefix(url, "inet://") {
		return http.NewRequest(method, url, body)
	}
	if dialerAddr == "" {
		return nil, errors.WithStack(ErrAddrMiss)
	}
	dialerAddr = strings.TrimPrefix(dialerAddr, "http://")
	portalHost, portalDest, portalUrl, portalArgs, err := parseInetUrl(url)
	if err != nil {
		return nil, err
	}
	url = fmt.Sprintf("http://%s/%s", dialerAddr, portalUrl)
	request, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	request.Header.Set(portalHostHeader, portalHost)
	request.Header.Set(portalDestHeader, portalDest)
	if portalArgs["ssl"] == "on" {
		request.Header.Set(portalSSLHeader, "on")
	}
	if portalArgs["sse"] == "on" {
		request.Header.Set(portalSSEHeader, "on")
	} else if portalArgs["ws"] == "on" {
		request.Header.Set(portalWSHeader, "on")
	}
	if portalArgs["passthrough"] == "on" {
		request.Header.Set(portalPassthroughHeader, "on")
	}
	// Support for cluster-dialer
	request.Host = portalDest
	return request, nil
}
