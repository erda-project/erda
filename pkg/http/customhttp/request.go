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

package customhttp

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/pkg/errors"

	"github.com/erda-project/erda/pkg/discover"
)

var (
	ErrInvalidAddr = Error{"customhttp: invalid inetaddr"}
)

type Error struct {
	msg string
}

func (e Error) Error() string {
	return e.msg
}

var inetAddr string
var mtx sync.Mutex

const (
	portalHostHeader = "X-Portal-Host"
	portalDestHeader = "X-Portal-Dest"
)

func SetInetAddr(addr string) {
	mtx.Lock()
	defer mtx.Unlock()
	inetAddr = addr
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
	mtx.Lock()
	if inetAddr == "" {
		inetAddr = discover.ClusterDialer()
	}
	mtx.Unlock()
	inetAddr = strings.TrimPrefix(inetAddr, "http://")
	portalHost, portalDest, portalUrl, _, err := parseInetUrl(url)
	if err != nil {
		return nil, err
	}
	url = fmt.Sprintf("http://%s/%s", inetAddr, portalUrl)
	request, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	request.Header.Set(portalHostHeader, portalHost)
	request.Header.Set(portalDestHeader, portalDest)
	return request, nil
}
