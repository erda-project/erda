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

package extension

import (
	"errors"
	"strings"
	"sync"

	"github.com/Masterminds/semver"
	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"
	"sigs.k8s.io/yaml"

	"github.com/erda-project/erda-infra/base/version"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/dicehub/dbclient"
	"github.com/erda-project/erda/modules/dicehub/service/apierrors"
)

// Extension Extension
type Extension struct {
	db  *dbclient.DBClient
	bdl *bundle.Bundle
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
