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

package admin_tools

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"

	"github.com/erda-project/erda-infra/base/version"
	"github.com/erda-project/erda-infra/modcom/api"
)

func (p *provider) showEnvs(w http.ResponseWriter, param struct {
	Format string `query:"format"`
	Key    string `query:"key"`
}) interface{} {
	if param.Format == "json" {
		if len(param.Key) > 0 {
			return api.Success(os.Getenv(param.Key))
		}
		envs := os.Environ()
		sort.Strings(envs)
		return api.Success(envs)
	}

	var text string
	if len(param.Key) > 0 {
		text = os.Getenv(param.Key)
	} else {
		text = strings.Join(os.Environ(), "\n")
	}
	w.Write([]byte(text))
	return nil
}

func (p *provider) showVersionInfo() interface{} {
	return api.Success(map[string]interface{}{
		"version":      version.Version,
		"commit_id":    version.CommitID,
		"go_version":   version.GoVersion,
		"build_time":   version.BuildTime,
		"docker_image": version.DockerImage,
	})
}

// proxy .
func (p *provider) proxy(req *http.Request, w http.ResponseWriter) interface{} {
	// target url
	target := req.Header.Get("X-Proxy-Target")
	if len(target) <= 0 {
		return api.Errors.InvalidParameter("invalid proxy target")
	}
	queryString := req.URL.RawQuery
	path := req.URL.Path[len("/api/admin/proxy"):]
	urlstr := fmt.Sprintf("%s%s?%s", target, path, queryString)
	u, err := url.Parse(urlstr)
	if err != nil {
		return api.Errors.InvalidParameter("invalid proxy url", urlstr)
	}
	req.Header.Del("X-Proxy-Target")
	request := &http.Request{
		Method: req.Method,
		URL:    u,
		Header: req.Header,
		Body:   req.Body,
	}
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return api.Failure(api.CodedError(http.StatusBadGateway, "BadGateway"), fmt.Errorf("fail to proxy: %s", err), urlstr)
	}
	for key, vals := range resp.Header {
		w.Header().Del(key)
		for _, val := range vals {
			w.Header().Add(key, val)
		}
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
	return nil
}

//func (p *provider) setGlobalLoglevel(params struct {
//	Provider string `query:"provider" validate:"required"`
//	Level    string `query:"level" default:"info"`
//}) interface{} {
//	level, err := logrus.ParseLevel(params.Level)
//	if err != nil {
//		return api.Errors.Internal(err)
//	}
//	service := p.ctx.Service(params.Provider)
//	if service == nil {
//		return api.Success(fmt.Sprintf("no such service: %s", params.Provider))
//	}
//	prov, ok := service.(servicehub.Provider)
//	if !ok {
//		return api.Success(fmt.Sprintf("service %s isn't a provider", params.Provider))
//	}
//	logger := servicehub.GetLoggerFromProvider(prov)
//	if logger != nil {
//		logger.SetLevel(params.Level)
//	} else {
//		api.Success("can not find logger")
//	}
//	return api.Success(fmt.Sprintf("new-level=%s", level))
//}
