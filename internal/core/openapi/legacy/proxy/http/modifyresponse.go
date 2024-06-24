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

package http

import (
	"bytes"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/labstack/echo"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/internal/core/openapi/legacy/api"
	apispec "github.com/erda-project/erda/internal/core/openapi/legacy/api/spec"
	"github.com/erda-project/erda/internal/core/openapi/legacy/hooks"
	"github.com/erda-project/erda/internal/core/openapi/legacy/monitor"
)

func getModifyResponse(rw http.ResponseWriter) func(*http.Response) error {
	return func(res *http.Response) error {

		spec := api.API.FindOriginPath(res.Request)
		if spec == nil {
			// unreachable
			logrus.Errorf("failed to modifyResponse: not found spec (unreachable):%v", res.Request.URL)
			return nil
		}
		defer func() {
			if res.StatusCode/100 == 5 {
				monitor.Notify(monitor.Info{
					Tp:     monitor.API50xCount,
					Detail: spec.Path.String(),
				})
			} else if res.StatusCode/100 == 4 {
				monitor.Notify(monitor.Info{
					Tp:     monitor.API40xCount,
					Detail: spec.Path.String(),
				})
			}
		}()
		var err error
		if spec.CustomResponse != nil {
			err = spec.CustomResponse(res)
		}
		if err != nil {
			logrus.Errorf("failed to modifyResponse: %v", err)
			return err
		}
		if !spec.ChunkAPI {
			if hooks.Enable {
				// if err = posthandle.InjectUserInfo(res, spec.NeedDesensitize); err != nil {
				// 	logrus.Errorf("failed to inject userinfo: %v", err)
				// 	return err
				// }
			}
		}

		if spec.Audit != nil && res.StatusCode/100 == 2 {
			request := res.Request
			reqBody := request.Context().Value("reqBody").(io.ReadCloser)
			bdl := request.Context().Value("bundle").(*bundle.Bundle)
			beginTime := request.Context().Value("beginTime").(time.Time)
			cache := request.Context().Value("cache").(*sync.Map)
			request.Body = reqBody
			resBody, err := io.ReadAll(res.Body)
			if err != nil {
				return err
			}
			res.Body = io.NopCloser(bytes.NewReader(resBody))
			orgIDStr := request.Header.Get("Org-ID")
			orgID, _ := strconv.ParseInt(orgIDStr, 10, 64)
			auditContext := &apispec.AuditContext{
				UrlParams: spec.Path.Vars(request.RequestURI),
				UserID:    request.Header.Get("User-ID"),
				OrgID:     orgID,
				Request:   request,
				Response:  res,
				Bundle:    bdl,
				BeginTime: beginTime,
				EndTime:   time.Now(),
				Result:    apistructs.SuccessfulResult,
				UserAgent: request.Header.Get("User-Agent"),
				ClientIP:  getRealIP(request),
				Cache:     cache,
			}

			go func() {
				defer func() {
					if err := recover(); err != nil {
						logrus.Errorf("audit panic url:%s , err:%s", request.RequestURI, err)
					}
				}()
				auditContext.Response.Body = io.NopCloser(bytes.NewReader(resBody))
				err := spec.Audit(auditContext)
				if err != nil {
					logrus.Errorf("audit failed url:%s , err:%s", request.RequestURI, err)
				}
			}()
		}

		handleReverseRespHeader(rw, res)

		return err
	}
}

func handleReverseRespHeader(rw http.ResponseWriter, res *http.Response) {
	// compatible original CORS
	res.Header.Set(echo.HeaderAccessControlAllowOrigin, "*")
}

// getRealIP .
func getRealIP(request *http.Request) string {
	ra := request.RemoteAddr
	if ip := request.Header.Get(echo.HeaderXForwardedFor); ip != "" {
		// filter space like '42.120.75.131,::ffff:10.112.1.1, 10.112.3.224'
		forwardedFilterSpace := strings.ReplaceAll(ip, " ", "")
		ra = strings.Split(forwardedFilterSpace, ",")[0]
	} else if ip := request.Header.Get(echo.HeaderXRealIP); ip != "" {
		ra = ip
	} else {
		ra, _, _ = net.SplitHostPort(ra)
	}
	return ra
}
