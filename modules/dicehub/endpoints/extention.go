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

package endpoints

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"

	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dicehub/service/apierrors"
	"github.com/erda-project/erda/pkg/http/httpserver"
)

// CreateExtension 创建扩展
func (e *Endpoints) CreateExtension(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	err := e.checkPushPermission(r)
	if err != nil {
		return apierrors.ErrCreateExtensionVersion.AccessDenied().ToResp(), nil
	}
	var request apistructs.ExtensionCreateRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return apierrors.ErrCreateExtension.InvalidParameter(err).ToResp(), nil
	}

	if request.Type != "action" && request.Type != "addon" {
		return apierrors.ErrCreateExtension.InvalidParameter("type").ToResp(), nil
	}

	result, err := e.extension.Create(&request)

	if err != nil {
		return apierrors.ErrCreateExtension.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp(result)
}

// SearchExtensions 批量查询扩展列表
func (e *Endpoints) SearchExtensions(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {

	var request apistructs.ExtensionSearchRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return apierrors.ErrQueryExtension.InvalidParameter(err).ToResp(), nil
	}

	result, err := e.extension.SearchExtensions(request)
	if err != nil {
		return apierrors.ErrQueryExtension.InternalError(err).ToResp(), nil
	}
	return httpserver.OkResp(result)
}

// QueryExtensions 获取扩展列表
func (e *Endpoints) QueryExtensions(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	all := r.URL.Query().Get("all")
	typ := r.URL.Query().Get("type")
	labels := r.URL.Query().Get("labels")
	result, err := e.extension.QueryExtensions(all, typ, labels)
	if err != nil {
		return apierrors.ErrQueryExtension.InternalError(err).ToResp(), nil
	}

	data, err := e.extension.MenuExtWithLocale(result, e.bdl.GetLocaleByRequest(r))
	if err != nil {
		return apierrors.ErrQueryExtension.InternalError(err).ToResp(), nil
	}

	var newResult []*apistructs.Extension
	for _, menu := range data {
		for _, value := range menu {
			for _, extension := range value.Items {
				newResult = append(newResult, extension)
			}
		}
	}

	menuMode := r.URL.Query().Get("menu")
	if menuMode == "true" {
		return httpserver.OkResp(e.extension.MenuExt(newResult))
	}

	return httpserver.OkResp(newResult)
}

// QueryExtensions 获取扩展列表
func (e *Endpoints) QueryExtensionsMenu(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	all := r.URL.Query().Get("all")
	typ := r.URL.Query().Get("type")
	labels := r.URL.Query().Get("labels")
	if labels != "" {
		labelsUnescaped, err := url.QueryUnescape(labels)
		if err == nil {
			labels = labelsUnescaped
		}
	}
	result, err := e.extension.QueryExtensions(all, typ, labels)
	if err != nil {
		return apierrors.ErrQueryExtension.InternalError(err).ToResp(), nil
	}

	data, err := e.extension.MenuExtWithLocale(result, e.bdl.GetLocaleByRequest(r))
	if err != nil {
		return apierrors.ErrQueryExtension.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp(data)
}

// CreateExtensionVersion 创建扩展版本
func (e *Endpoints) CreateExtensionVersion(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	err := e.checkPushPermission(r)
	if err != nil {
		return apierrors.ErrCreateExtensionVersion.AccessDenied().ToResp(), nil
	}
	name, err := url.QueryUnescape(vars["name"])
	if err != nil {
		return apierrors.ErrCreateExtensionVersion.InvalidParameter("name").ToResp(), nil
	}

	var request apistructs.ExtensionVersionCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return apierrors.ErrCreateExtensionVersion.InvalidParameter(err).ToResp(), nil
	}

	request.Name = name
	result, err := e.extension.CreateExtensionVersion(&request)

	if err != nil {
		return apierrors.ErrCreateExtensionVersion.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp(result)
}

// GetExtensionVersion 获取指定版本扩展
func (e *Endpoints) GetExtensionVersion(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	name, err := url.QueryUnescape(vars["name"])
	if err != nil {
		return apierrors.ErrQueryExtension.InvalidParameter("name").ToResp(), nil
	}

	version, err := url.QueryUnescape(vars["version"])
	if err != nil {
		return apierrors.ErrQueryExtension.InvalidParameter("version").ToResp(), nil
	}

	yamlFormatStr := r.URL.Query().Get("yamlFormat")
	yamlFormat, _ := strconv.ParseBool(yamlFormatStr)

	result, err := e.extension.GetExtensionVersion(name, version, yamlFormat)

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return apierrors.ErrQueryExtension.NotFound().ToResp(), nil
		}
		return apierrors.ErrQueryExtension.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp(result)
}

// QueryExtensionVersions 查询扩展版本列表
func (e *Endpoints) QueryExtensionVersions(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	name, err := url.QueryUnescape(vars["name"])
	if err != nil {
		return apierrors.ErrQueryExtension.InvalidParameter("name").ToResp(), nil
	}
	request := apistructs.ExtensionVersionQueryRequest{}
	if err := e.queryStringDecoder.Decode(&request, r.URL.Query()); err != nil {
		return apierrors.ErrQueryExtension.InternalError(err).ToResp(), nil
	}
	request.Name = name

	result, err := e.extension.QueryExtensionVersions(&request)

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return apierrors.ErrQueryExtension.NotFound().ToResp(), nil
		}
		return apierrors.ErrQueryExtension.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp(result)
}

func (e *Endpoints) checkPushPermission(r *http.Request) error {
	userID := r.Header.Get("User-ID")
	if userID == "" {
		return errors.Errorf("failed to get permission(User-ID is empty)")
	}
	data, err := e.bdl.CheckPermission(&apistructs.PermissionCheckRequest{
		UserID:   userID,
		Scope:    apistructs.SysScope,
		ScopeID:  1,
		Action:   apistructs.CreateAction,
		Resource: apistructs.OrgResource,
	})
	if err != nil {
		return err
	}
	if !data.Access {
		return errors.New("no permission to push")
	}
	return nil
}

type MenuItem struct {
	Name string `json:"name"`
}
