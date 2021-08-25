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
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dicehub/conf"
	"github.com/erda-project/erda/modules/dicehub/service/apierrors"
	"github.com/erda-project/erda/pkg/http/httpserver"
	"github.com/erda-project/erda/pkg/i18n"
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

	menuMode := r.URL.Query().Get("menu")
	if menuMode == "true" {
		return httpserver.OkResp(menuExt(result))
	}
	return httpserver.OkResp(result)
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

	locale := e.bdl.GetLocaleByRequest(r)

	return httpserver.OkResp(menuExtWithLocale(result, locale))
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

	all := r.URL.Query().Get("all")

	request := apistructs.ExtensionVersionQueryRequest{
		Name: name,
		All:  all,
	}

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

func menuExtWithLocale(extensions []*apistructs.Extension, locale *i18n.LocaleResource) map[string][]apistructs.ExtensionMenu {
	var result = map[string][]apistructs.ExtensionMenu{}

	extMap := extMap(extensions)
	//遍历category大的类别
	for categoryType, typeItems := range apistructs.CategoryTypes {
		//每个类别归属于一个map
		menuList := result[categoryType]
		//遍历大类别中的子类别
		for _, v := range typeItems {
			//获取这种子类别的Extension的对象数据
			extensionListWithKeyName, ok := extMap[v]
			if !ok {
				continue
			}
			//这个子类别是否有国际化，没有就是字类别的名称
			var displayName string
			if locale != nil {
				displayNameTemplate := locale.GetTemplate(apistructs.DicehubExtensionsMenu + "." + categoryType + "." + v)
				if displayNameTemplate != nil {
					displayName = displayNameTemplate.Content()
				}
			}

			if displayName == "" {
				displayName = v
			}
			//将这些字类别归属于数组中
			menuList = append(menuList, apistructs.ExtensionMenu{
				Name:        v,
				DisplayName: displayName,
				Items:       extensionListWithKeyName,
			})
		}
		//将数组设置回map中
		result[categoryType] = menuList
	}

	return result
}

func menuExt(extensions []*apistructs.Extension) interface{} {
	extMap := extMap(extensions)
	menuMap := &MenuMap{}
	for subMenuName, subMenuValues := range conf.ExtensionMenu() {
		subMenu := &MenuMap{}
		for _, v := range subMenuValues {
			params := strings.Split(v, ":")
			keyName := params[0]
			displayName := params[1]
			subMenus := getMapValue(extMap, keyName)
			if len(subMenus) > 0 {
				subMenu.Put(displayName, subMenus)
			}
		}
		menuMap.Put(subMenuName, subMenu)
	}
	return menuMap
}

func getMapValue(extMap map[string][]*apistructs.Extension, key string) []*apistructs.Extension {
	extList, _ := extMap[key]
	return extList
}

func extMap(extensions []*apistructs.Extension) map[string][]*apistructs.Extension {
	extMap := map[string][]*apistructs.Extension{}
	for _, v := range extensions {
		extList, exist := extMap[v.Category]
		if exist {
			extList = append(extList, v)
		} else {
			extList = []*apistructs.Extension{v}
		}
		extMap[v.Category] = extList
	}
	return extMap
}

type MenuMap []*SortMapNode

type SortMapNode struct {
	Key string
	Val interface{}
}

func (m *MenuMap) Put(key string, val interface{}) {
	index, _, ok := m.get(key)
	if ok {
		(*m)[index].Val = val
	} else {
		node := &SortMapNode{Key: key, Val: val}
		*m = append(*m, node)
	}
}

func (m *MenuMap) Get(key string) (interface{}, bool) {
	_, val, ok := m.get(key)
	return val, ok
}

func (m *MenuMap) get(key string) (int, interface{}, bool) {
	for index, node := range *m {
		if node.Key == key {
			return index, node.Val, true
		}
	}
	return -1, nil, false
}
func (m *MenuMap) MarshalJSON() ([]byte, error) {
	mapJson := m.ToSortedMapJson(m)
	return []byte(mapJson), nil
}

func (m *MenuMap) ToSortedMapJson(smap *MenuMap) string {
	s := "{"
	for _, node := range *smap {
		v := node.Val
		isSamp := false
		str := ""
		switch v.(type) {
		case *MenuMap:
			isSamp = true
			str = smap.ToSortedMapJson(v.(*MenuMap))
		}

		if !isSamp {
			b, _ := json.Marshal(node.Val)
			str = string(b)
		}

		s = fmt.Sprintf("%s\"%s\":%s,", s, node.Key, str)
	}
	s = strings.TrimRight(s, ",")
	s = fmt.Sprintf("%s}", s)
	return s
}
