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

package extension

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Masterminds/semver"
	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"
	"sigs.k8s.io/yaml"

	"github.com/erda-project/erda-infra/base/version"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/dicehub/conf"
	"github.com/erda-project/erda/modules/dicehub/dbclient"
	"github.com/erda-project/erda/modules/dicehub/service/apierrors"
	"github.com/erda-project/erda/modules/pipeline/pexpr"
	"github.com/erda-project/erda/pkg/i18n"
)

// Extension Extension
type Extension struct {
	db                  *dbclient.DBClient
	bdl                 *bundle.Bundle
	cacheExtensionSpecs sync.Map
}

// Option 定义 Extension 对象的配置选项
type Option func(*Extension)

// New 新建 Extension 实例，操作 Extension 资源
func New(options ...Option) *Extension {
	app := &Extension{}
	for _, op := range options {
		op(app)
	}
	return app
}

// WithDBClient 配置 db client
func WithDBClient(db *dbclient.DBClient) Option {
	return func(a *Extension) {
		a.db = db
	}
}

// WithBundle 配置 bundle
func WithBundle(bdl *bundle.Bundle) Option {
	return func(a *Extension) {
		a.bdl = bdl
	}
}

// Create 创建Extension
func (i *Extension) Create(req *apistructs.ExtensionCreateRequest) (*apistructs.Extension, error) {
	if req.Type != "addon" && req.Type != "action" {
		return nil, apierrors.ErrCreateExtension.InvalidParameter("type")
	}
	if req.Name == "" {
		return nil, apierrors.ErrCreateExtension.InvalidParameter("name")
	}
	ext := dbclient.Extension{
		Name:        req.Name,
		Type:        req.Type,
		Desc:        req.Desc,
		Category:    req.Category,
		DisplayName: req.DisplayName,
		LogoUrl:     req.LogoUrl,
		Public:      req.Public,
		Labels:      req.Labels,
	}
	err := i.db.CreateExtension(&ext)
	if err != nil {
		return nil, err
	}
	return ext.ToApiData(), nil
}

// SearchExtensions 批量查询扩展
func (i *Extension) SearchExtensions(req apistructs.ExtensionSearchRequest) (map[string]*apistructs.ExtensionVersion, error) {
	result := struct {
		mp map[string]*apistructs.ExtensionVersion
		sync.RWMutex
	}{mp: make(map[string]*apistructs.ExtensionVersion)}
	var wg sync.WaitGroup
	wg.Add(len(req.Extensions))
	for _, fullName := range req.Extensions {
		go func(fnm string) {
			defer wg.Done()
			splits := strings.SplitN(fnm, "@", 2)
			name := splits[0]
			version := ""
			if len(splits) > 1 {
				version = splits[1]
			}
			if version == "" {
				extVersion, _ := i.GetExtensionDefaultVersion(name, req.YamlFormat)
				result.Lock()
				result.mp[fnm] = extVersion
				result.Unlock()
			} else if strings.HasPrefix(version, "https://") || strings.HasPrefix(version, "http://") {
				extensionVersion, err := i.GetExtensionByGit(name, version, "spec.yml", "dice.yml", "README.md")
				result.Lock()
				if err != nil {
					result.mp[fnm] = nil
				} else {
					result.mp[fnm] = extensionVersion
				}
				result.Unlock()
			} else {
				extensionVersion, err := i.GetExtensionVersion(name, version, req.YamlFormat)
				result.Lock()
				if err != nil {
					result.mp[fnm] = nil
				} else {
					result.mp[fnm] = extensionVersion
				}
				result.Unlock()
			}
		}(fullName)
	}
	wg.Wait()
	return result.mp, nil
}

func (i *Extension) MenuExtWithLocale(extensions []*apistructs.Extension, locale *i18n.LocaleResource) (map[string][]apistructs.ExtensionMenu, error) {
	var result = map[string][]apistructs.ExtensionMenu{}

	var extensionName []string
	for _, v := range extensions {
		extensionName = append(extensionName, v.Name)
	}
	extensionVersionMap, err := i.db.ListExtensionVersions(extensionName)
	if err != nil {
		return nil, apierrors.ErrQueryExtension.InternalError(err)
	}

	extMap := i.extMap(extensions)
	// Traverse categories with large categories
	for categoryType, typeItems := range apistructs.CategoryTypes {
		// Each category belongs to a map
		menuList := result[categoryType]
		// Traverse subcategories in a large category
		for _, v := range typeItems {
			// Gets the object data of the extension of this subcategory
			extensionListWithKeyName, ok := extMap[v]
			if !ok {
				continue
			}

			// extension displayName desc Internationalization settings
			for _, extension := range extensionListWithKeyName {
				defaultExtensionVersion := getDefaultExtensionVersion("", extensionVersionMap[extension.Name])
				if defaultExtensionVersion.ID <= 0 {
					logrus.Errorf("extension %v not find default extension version", extension.Name)
					continue
				}

				// get from caches and set to caches
				localeDisplayName, localeDesc, err := i.getLocaleDisplayNameAndDesc(defaultExtensionVersion)
				if err != nil {
					return nil, err
				}

				if localeDisplayName != "" {
					extension.DisplayName = localeDisplayName
				}

				if localeDesc != "" {
					extension.Desc = localeDesc
				}
			}

			// Whether this subcategory is internationalized or not is the name of the word category
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
			// Assign these word categories to the array
			menuList = append(menuList, apistructs.ExtensionMenu{
				Name:        v,
				DisplayName: displayName,
				Items:       extensionListWithKeyName,
			})
		}
		// Set the array back into the map
		result[categoryType] = menuList
	}

	return result, nil
}

func getDefaultExtensionVersion(version string, extensionVersions []dbclient.ExtensionVersion) dbclient.ExtensionVersion {
	var defaultVersion dbclient.ExtensionVersion
	if version == "" {
		for _, extensionVersion := range extensionVersions {
			if extensionVersion.IsDefault {
				defaultVersion = extensionVersion
				break
			}
		}
		if defaultVersion.ID <= 0 && len(extensionVersions) > 0 {
			defaultVersion = extensionVersions[0]
		}
	} else {
		for _, extensionVersion := range extensionVersions {
			if extensionVersion.Version == version {
				defaultVersion = extensionVersion
				break
			}
		}
	}
	return defaultVersion
}

func (s *Extension) getLocaleDisplayNameAndDesc(extensionVersion dbclient.ExtensionVersion) (string, string, error) {
	value, ok := s.cacheExtensionSpecs.Load(extensionVersion.Spec)
	var specData apistructs.Spec
	if !ok {
		err := yaml.Unmarshal([]byte(extensionVersion.Spec), &specData)
		if err != nil {
			return "", "", apierrors.ErrQueryExtension.InternalError(err)
		}
		s.cacheExtensionSpecs.Store(extensionVersion.Spec, specData)
		go func(spec string) {
			// caches expiration time
			time.AfterFunc(30*25*time.Hour, func() {
				s.cacheExtensionSpecs.Delete(spec)
			})
		}(extensionVersion.Spec)
	} else {
		specData = value.(apistructs.Spec)
	}

	displayName := specData.GetLocaleDisplayName(i18n.GetGoroutineBindLang())
	desc := specData.GetLocaleDesc(i18n.GetGoroutineBindLang())

	return displayName, desc, nil
}

func (i *Extension) MenuExt(extensions []*apistructs.Extension) interface{} {
	extMap := i.extMap(extensions)
	menuMap := &MenuMap{}
	for subMenuName, subMenuValues := range conf.ExtensionMenu() {
		subMenu := &MenuMap{}
		for _, v := range subMenuValues {
			params := strings.Split(v, ":")
			keyName := params[0]
			displayName := params[1]
			subMenus := i.getMapValue(extMap, keyName)
			if len(subMenus) > 0 {
				subMenu.Put(displayName, subMenus)
			}
		}
		menuMap.Put(subMenuName, subMenu)
	}
	return menuMap
}

func (i *Extension) getMapValue(extMap map[string][]*apistructs.Extension, key string) []*apistructs.Extension {
	extList, _ := extMap[key]
	return extList
}

func (i *Extension) extMap(extensions []*apistructs.Extension) map[string][]*apistructs.Extension {
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

// QueryExtensions 查询Extension列表
func (i *Extension) QueryExtensions(all string, typ string, labels string) ([]*apistructs.Extension, error) {
	extensions, err := i.db.QueryExtensions(all, typ, labels)
	if err != nil {
		return nil, err
	}

	result := []*apistructs.Extension{}
	for _, v := range extensions {
		apiData := v.ToApiData()
		result = append(result, apiData)
	}
	return result, nil
}

// GetExtensionVersion 获取指定版本extension
func (i *Extension) GetExtensionVersion(name string, version string, yamlFormat bool) (*apistructs.ExtensionVersion, error) {
	ext, err := i.db.GetExtension(name)
	if err != nil {
		return nil, err
	}
	var extensionVersion *dbclient.ExtensionVersion
	if version == "default" {
		extensionVersion, err = i.db.GetExtensionDefaultVersion(name)
	} else {
		extensionVersion, err = i.db.GetExtensionVersion(name, version)
	}
	if err != nil {
		return nil, err
	}
	return extensionVersion.ToApiData(ext.Type, yamlFormat), nil
}

// GetExtensionDefaultVersion 获取extension默认版本
func (i *Extension) GetExtensionDefaultVersion(name string, yamlFormat bool) (*apistructs.ExtensionVersion, error) {
	ext, err := i.db.GetExtension(name)
	if err != nil {
		return nil, err
	}
	extensionVersion, err := i.db.GetExtensionDefaultVersion(name)
	if err != nil {
		return nil, err
	}
	return extensionVersion.ToApiData(ext.Type, yamlFormat), nil
}

// CreateExtensionVersion 创建插件版本
func (i *Extension) CreateExtensionVersion(req *apistructs.ExtensionVersionCreateRequest) (*apistructs.ExtensionVersion, error) {
	specData := apistructs.Spec{}
	err := yaml.Unmarshal([]byte(req.SpecYml), &specData)
	if err != nil {
		return nil, err
	}
	invalidPhs := pexpr.FindInvalidPlaceholders(req.SpecYml)
	if len(invalidPhs) > 0 {
		return nil, fmt.Errorf("invalid i18n expression, found invalid placeholders: %s (must match: %s)", strings.Join(invalidPhs, ", "), pexpr.PhRe.String())
	}

	if specData.DisplayName == "" {
		specData.DisplayName = specData.Name
	}
	// 非语义化版本不能设置public
	_, err = semver.NewVersion(specData.Version)
	if err != nil && req.Public {
		return nil, err
	}

	if !specData.CheckDiceVersion(version.Version) {
		err := i.db.DeleteExtensionVersion(specData.Name, specData.Version)
		if err != nil {
			return nil, err
		}
		count, err := i.db.GetExtensionVersionCount(specData.Name)
		if err != nil {
			return nil, err
		}
		if count == 0 {
			err = i.db.DeleteExtension(specData.Name)
			if err != nil {
				return nil, err
			}
		}
		i.triggerPushEvent(specData, "delete")
		return &apistructs.ExtensionVersion{}, nil
	}
	labels := ""
	if specData.Labels != nil {
		for k, v := range specData.Labels {
			labels += k + ":" + v + ","
		}
	}
	extModel, err := i.db.GetExtension(req.Name)
	var ext *apistructs.Extension
	if err == nil {
		ext = extModel.ToApiData()
	} else if err == gorm.ErrRecordNotFound {
		//不存在name 自动创建
		ext, err = i.Create(&apistructs.ExtensionCreateRequest{
			Type:        specData.Type,
			Name:        req.Name,
			DisplayName: specData.DisplayName,
			Desc:        specData.Desc,
			Category:    specData.Category,
			LogoUrl:     specData.LogoUrl,
			Public:      req.Public,
			Labels:      labels,
		})
		if err != nil {
			return nil, err
		}
	} else {
		return nil, err
	}

	version, err := i.db.GetExtensionVersion(req.Name, req.Version)

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			version = &dbclient.ExtensionVersion{
				ExtensionId: ext.ID,
				Name:        specData.Name,
				Version:     specData.Version,
				Dice:        req.DiceYml,
				Spec:        req.SpecYml,
				Swagger:     req.SwaggerYml,
				Readme:      req.Readme,
				Public:      req.Public,
				IsDefault:   req.IsDefault,
			}
			err = i.db.CreateExtensionVersion(version)
			i.triggerPushEvent(specData, "create")
			if err != nil {
				return nil, err
			}
			return version.ToApiData(ext.Type, false), nil
		} else {
			return nil, err
		}
	}
	if req.ForceUpdate {
		version.Spec = req.SpecYml
		version.Dice = req.DiceYml
		version.Swagger = req.SwaggerYml
		version.Readme = req.Readme
		version.Public = req.Public
		version.IsDefault = req.IsDefault
		if version.IsDefault {
			err := i.db.SetUnDefaultVersion(version.Name)
			if err != nil {
				return nil, err
			}
		}
		err := i.db.Save(&version).Error
		if err != nil {
			return nil, err
		}
		i.triggerPushEvent(specData, "update")
		if req.All {
			extModel.Category = specData.Category
			extModel.LogoUrl = specData.LogoUrl
			extModel.DisplayName = specData.DisplayName
			extModel.Desc = specData.Desc
			extModel.Public = req.Public
			extModel.Labels = labels
			err = i.db.Save(&extModel).Error
			if err != nil {
				return nil, err
			}
		}

		return version.ToApiData(ext.Type, false), err
	} else {
		return nil, errors.New("version already exist")
	}

}

// QueryExtensionVersion 查询扩展版本
func (i *Extension) QueryExtensionVersions(req *apistructs.ExtensionVersionQueryRequest) ([]*apistructs.ExtensionVersion, error) {
	ext, err := i.db.GetExtension(req.Name)
	if err != nil {
		return nil, err
	}
	versions, err := i.db.QueryExtensionVersions(req.Name, req.All)
	if err != nil {
		return nil, err
	}
	result := []*apistructs.ExtensionVersion{}
	for _, v := range versions {
		result = append(result, v.ToApiData(ext.Type, req.YamlFormat))
	}
	return result, nil
}

func (i *Extension) triggerPushEvent(specData apistructs.Spec, action string) {
	if specData.Type != "addon" {
		return
	}
	go func() {
		err := i.bdl.CreateEvent(&apistructs.EventCreateRequest{
			EventHeader: apistructs.EventHeader{
				Event:  "addon_extension_push",
				Action: action,
			},
			Sender: "dicehub",
			Content: apistructs.ExtensionPushEventData{
				Name:    specData.Name,
				Version: specData.Version,
				Type:    specData.Type,
			},
		})
		if err != nil {
			logrus.Errorf("failed to create event :%v", err)
		}
	}()
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
