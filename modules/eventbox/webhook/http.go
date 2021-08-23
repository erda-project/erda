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

package webhook

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	stypes "github.com/erda-project/erda/modules/eventbox/server/types"
)

const (
	BadRequestCode        = "WH400"
	InternalServerErrCode = "WH500"
	OtherErrCode          = "WH600"
)

func toCode(e error) string {
	switch e {
	case BadRequestErr:
		return BadRequestCode
	case InternalServerErr:
		return InternalServerErrCode
	}
	return OtherErrCode
}

type WebHookHTTP struct {
	impl *WebHookImpl
}

func NewWebHookHTTP() (*WebHookHTTP, error) {
	impl, err := NewWebHookImpl()
	if err != nil {
		return nil, err
	}
	if err := MakeSureBuiltinHooks(impl); err != nil {
		return nil, err
	}
	return &WebHookHTTP{
		impl: impl,
	}, nil
}
func (w *WebHookHTTP) ListHooks(ctx context.Context, req *http.Request, vars map[string]string) (stypes.Responser, error) {
	location, err := extractHookLocation(req)
	if err != nil {
		return stypes.HTTPResponse{
			Error: &stypes.ErrorResponse{
				Code: BadRequestCode,
				Msg:  err.Error(),
			},
			Compose: true,
		}, nil
	}
	if location.Org == "" {
		return stypes.HTTPResponse{
			Error: &stypes.ErrorResponse{
				Code: BadRequestCode,
				Msg:  "not provide org",
			},
			Compose: true,
		}, nil
	}
	r, err := w.impl.ListHooks(location)
	if err != nil {
		return stypes.HTTPResponse{
			Error: &stypes.ErrorResponse{
				Code: toCode(errors.Cause(err)),
				Msg:  err.Error(),
			},
			Compose: true,
		}, nil
	}
	return stypes.HTTPResponse{
		Compose: true,
		Content: r,
	}, nil
}

// not require 'org' in query param,
// and it should return not found when found-hook's org not match orgID(in header)
func (w *WebHookHTTP) InspectHook(ctx context.Context, req *http.Request, vars map[string]string) (stypes.Responser, error) {
	id := vars["id"]
	orgID := extractOrgIDHeader(req)

	r, err := w.impl.InspectHook(orgID, id)
	if err != nil {
		return stypes.HTTPResponse{
			Compose: true,
			Error: &stypes.ErrorResponse{
				Code: toCode(errors.Cause(err)),
				Msg:  err.Error(),
			},
		}, nil
	}
	return stypes.HTTPResponse{
		Compose: true,
		Content: r,
	}, nil
}

// org, project will be provided in request body
func (w *WebHookHTTP) CreateHook(ctx context.Context, req *http.Request, vars map[string]string) (stypes.Responser, error) {
	h := CreateHookRequest{}
	if err := json.NewDecoder(req.Body).Decode(&h); err != nil {
		err := fmt.Errorf("createhook: decode fail: %v", err)
		logrus.Error(err)
		return stypes.HTTPResponse{
			Error: &stypes.ErrorResponse{
				Code: BadRequestCode,
				Msg:  err.Error(),
			},
			Compose: true,
		}, nil
	}
	r, err := w.impl.CreateHook(extractOrgIDHeader(req), h)
	if err != nil {
		logrus.Error(err)
		return stypes.HTTPResponse{
			Error: &stypes.ErrorResponse{
				Code: toCode(errors.Cause(err)),
				Msg:  err.Error(),
			},
			Compose: true,
		}, nil
	}
	return stypes.HTTPResponse{
		Content: r,
		Compose: true,
	}, nil
}
func (w *WebHookHTTP) EditHook(ctx context.Context, req *http.Request, vars map[string]string) (stypes.Responser, error) {
	id := vars["id"]
	e := EditHookRequest{}
	if err := json.NewDecoder(req.Body).Decode(&e); err != nil {
		err := fmt.Errorf("edithook: decode fail: %v", err)
		logrus.Error(err)
		return stypes.HTTPResponse{
			Error: &stypes.ErrorResponse{
				Code: BadRequestCode,
				Msg:  err.Error(),
			},
			Compose: true,
		}, nil
	}
	orgID := extractOrgIDHeader(req)
	r, err := w.impl.EditHook(orgID, id, e)
	if err != nil {
		logrus.Error(err)
		return stypes.HTTPResponse{
			Error: &stypes.ErrorResponse{
				Code: toCode(errors.Cause(err)),
				Msg:  err.Error(),
			},
			Compose: true,
		}, nil
	}
	return stypes.HTTPResponse{
		Content: r,
		Compose: true,
	}, nil
}
func (w *WebHookHTTP) PingHook(ctx context.Context, req *http.Request, vars map[string]string) (stypes.Responser, error) {
	id := vars["id"]
	orgID := extractOrgIDHeader(req)
	if err := w.impl.PingHook(orgID, id); err != nil {
		logrus.Error(err)
		return stypes.HTTPResponse{
			Error: &stypes.ErrorResponse{
				Code: toCode(errors.Cause(err)),
				Msg:  err.Error(),
			},
			Compose: true,
		}, nil
	}
	return stypes.HTTPResponse{
		Content: "",
		Compose: true,
	}, nil
}
func (w *WebHookHTTP) DeleteHook(ctx context.Context, req *http.Request, vars map[string]string) (stypes.Responser, error) {
	id := vars["id"]
	orgID := extractOrgIDHeader(req)
	if err := w.impl.DeleteHook(orgID, id); err != nil {
		logrus.Error(err)
		return stypes.HTTPResponse{
			Error: &stypes.ErrorResponse{
				Code: toCode(errors.Cause(err)),
				Msg:  err.Error(),
			},
			Compose: true,
		}, nil
	}
	return stypes.HTTPResponse{
		Content: "",
		Compose: true,
	}, nil
}

type ListHookEventsResponse = apistructs.WebhookListEventsResponseData

func (w *WebHookHTTP) ListHookEvents(ctx context.Context, req *http.Request, vars map[string]string) (stypes.Responser, error) {
	return stypes.HTTPResponse{
		Content: ListHookEventsResponse([]struct {
			Key   string `json:"key"`
			Title string `json:"title"`
			Desc  string `json:"desc"`
		}{
			// {"application", "application", "应用的创建，删除"},
			{"runtime", "runtime", "runtime 的创建，删除"},
			{"pipeline", "pipeline", "pipeline 的创建(create)，停止(stop)，完成(finish)"},
			// {"deployment", "deployment", "runtime部署过程：初始化，申请addon，部署service"},
			// {"domain-changed", "domain changed", "runtime域名的改变"},
			// {"config-modified", "config modified", "配置变化"},
		}),
		Compose: true,
	}, nil

}

func extractHookLocation(req *http.Request) (apistructs.HookLocation, error) {
	org := queryGetByOrder(req, "orgID", "orgId", "orgid")
	project := queryGetByOrder(req, "projectID", "projectId", "projectid")
	application := queryGetByOrder(req, "applicationID", "applicationId", "applicationid")
	envstr := queryGetByOrder(req, "env")
	var env []string
	if envstr != "" {
		envraw := strings.Split(envstr, ",")
		for i := range envraw {
			if envraw[i] != "" {
				normalizedEnv := strings.ToLower(strings.TrimSpace(envraw[i]))
				if normalizedEnv != "dev" && normalizedEnv != "test" && normalizedEnv != "staging" && normalizedEnv != "prod" {
					return apistructs.HookLocation{}, fmt.Errorf("bad env param: %v", envraw[i])
				}
				env = append(env, strings.ToLower(strings.TrimSpace(envraw[i])))
			}
		}
	}
	return apistructs.HookLocation{
		Org:         org,
		Project:     project,
		Application: application,
		Env:         env,
	}, nil
}

func queryGetByOrder(req *http.Request, key ...string) string {
	for _, k := range key {
		v := req.URL.Query().Get(k)
		if v != "" {
			return v
		}
	}
	return ""
}

func extractOrgIDHeader(req *http.Request) string {
	return req.Header.Get("Org-ID")
}

func invalidResponse(msg ...string) (stypes.Responser, error) {
	message := "invalid query"
	if len(msg) > 0 {
		message = fmt.Sprintf("invalid query: %s", msg[0])
	}
	return stypes.HTTPResponse{
		Error: &stypes.ErrorResponse{
			Code: BadRequestCode,
			Msg:  message,
		},
		Compose: true,
	}, nil
}

func check(h func(context.Context, *http.Request, map[string]string) (stypes.Responser, error)) func(context.Context, *http.Request, map[string]string) (stypes.Responser, error) {
	bdl := bundle.New(bundle.WithCoreServices())
	f := func(ctx context.Context, req *http.Request, vars map[string]string) (stypes.Responser, error) {
		userid := req.Header.Get("User-ID")
		if userid != "" {
			org := queryGetByOrder(req, "orgID", "orgId", "orgid")
			orgid, err := strconv.ParseUint(org, 10, 64)
			if err != nil {
				return invalidResponse("org")
			}

			checkresult, err := bdl.CheckPermission(&apistructs.PermissionCheckRequest{
				UserID:   userid,
				Scope:    apistructs.OrgScope,
				ScopeID:  orgid,
				Resource: "webhook",
				Action:   "OPERATE",
			})
			if err != nil {
				return invalidResponse(err.Error())
			}
			if !checkresult.Access {
				return invalidResponse("permission denied")
			}
			project := queryGetByOrder(req, "projectID", "projectId", "projectid")
			if project != "" {
				projectid, err := strconv.ParseUint(project, 10, 64)
				if err != nil {
					return invalidResponse("project")
				}
				checkresult, err := bdl.CheckPermission(&apistructs.PermissionCheckRequest{
					UserID:   userid,
					Scope:    apistructs.ProjectScope,
					ScopeID:  projectid,
					Resource: "webhook",
					Action:   "OPERATE",
				})
				if err != nil {
					return invalidResponse(err.Error())
				}
				if !checkresult.Access {
					return invalidResponse("permission denied")
				}
			}
			application := queryGetByOrder(req, "applicationID", "applicationId", "applicationid")
			if application != "" {
				appid, err := strconv.ParseUint(application, 10, 64)
				if err != nil {
					return invalidResponse("application")
				}
				checkresult, err := bdl.CheckPermission(&apistructs.PermissionCheckRequest{
					UserID:   userid,
					Scope:    apistructs.AppScope,
					ScopeID:  appid,
					Resource: "webhook",
					Action:   "OPERATE",
				})
				if err != nil {
					return invalidResponse(err.Error())
				}
				if !checkresult.Access {
					return invalidResponse("permission denied")
				}
			}
		}
		return h(ctx, req, vars)
	}
	return f
}

func (w *WebHookHTTP) GetHTTPEndPoints() []stypes.Endpoint {
	return []stypes.Endpoint{
		{"/webhooks", http.MethodGet, check(w.ListHooks)},
		{"/webhooks/{id}", http.MethodGet, check(w.InspectHook)},
		{"/webhooks", http.MethodPost, check(w.CreateHook)},
		{"/webhooks/{id}", http.MethodPut, check(w.EditHook)},
		{"/webhooks/{id}/actions/ping", http.MethodPost, check(w.PingHook)},
		{"/webhooks/{id}", http.MethodDelete, check(w.DeleteHook)},
		{"/webhook_events", http.MethodGet, w.ListHookEvents},
	}
}
