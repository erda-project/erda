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
	"context"
	"encoding/json"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/structpb"
	"gopkg.in/yaml.v3"

	"github.com/erda-project/erda-infra/base/version"
	pb "github.com/erda-project/erda-proto-go/core/dicehub/extension/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/dicehub/extension/db"
	"github.com/erda-project/erda/modules/dicehub/service/apierrors"
	"github.com/erda-project/erda/pkg/common/apis"
)

type extensionService struct {
	p             *provider
	db            *db.ExtensionConfigDB
	bdl           *bundle.Bundle
	extensionMenu map[string][]string
}

func (s *extensionService) SearchExtensions(ctx context.Context, req *pb.ExtensionSearchRequest) (*pb.ExtensionSearchResponse, error) {
	result := map[string]*pb.ExtensionVersion{}
	for _, fullName := range req.Extensions {
		splits := strings.Split(fullName, "@")
		name := splits[0]
		version := ""
		if len(splits) > 1 {
			version = splits[1]
		}
		if version == "" {
			extVersion, _ := s.GetExtensionDefaultVersion(name, req.YamlFormat)
			result[fullName] = extVersion
		} else {
			extensionVersion, err := s.GetExtension(name, version, req.YamlFormat)
			if err != nil {
				result[fullName] = nil
			} else {
				result[fullName] = extensionVersion
			}
		}
	}

	return &pb.ExtensionSearchResponse{Data: result}, nil
}

func (s *extensionService) CreateExtension(ctx context.Context, req *pb.ExtensionCreateRequest) (*pb.ExtensionCreateResponse, error) {
	err := s.checkPushPermission(ctx)
	if err != nil {
		return nil, apierrors.ErrCreateExtensionVersion.AccessDenied()
	}

	if req.Type != "action" && req.Type != "addon" {
		return nil, apierrors.ErrCreateExtension.InvalidParameter("type")
	}

	result, err := s.Create(req)

	if err != nil {
		return nil, apierrors.ErrCreateExtension.InternalError(err)
	}

	return &pb.ExtensionCreateResponse{Data: result}, nil
}

func (s *extensionService) QueryExtensions(ctx context.Context, req *pb.QueryExtensionsRequest) (*pb.QueryExtensionsResponse, error) {
	result, err := s.QueryExtensionList(req.All, req.Type, req.Labels)
	if err != nil {
		return nil, apierrors.ErrQueryExtension.InternalError(err)
	}

	if req.Menu == "true" {
		menuExtResult := menuExt(result, s.extensionMenu)
		resp, err := ToProtoValue(menuExtResult)
		if err != nil {
			logrus.Errorf("fail transform interface to any type")
			return nil, apierrors.ErrQueryExtension.InternalError(err)
		}
		return &pb.QueryExtensionsResponse{Data: resp}, nil
	}

	resp, err := ToProtoValue(result)
	if err != nil {
		logrus.Errorf("fail transform interface to any type")
		return nil, apierrors.ErrQueryExtension.InternalError(err)
	}
	return &pb.QueryExtensionsResponse{Data: resp}, nil
}

func (s *extensionService) QueryExtensionsMenu(ctx context.Context, req *pb.QueryExtensionsMenuRequest) (*pb.QueryExtensionsMenuResponse, error) {
	result, err := s.QueryExtensionList(req.All, req.Type, req.Labels)
	if err != nil {
		return nil, apierrors.ErrQueryExtension.InternalError(err)
	}

	locale := s.bdl.GetLocale(apis.GetLang(ctx))

	menu := menuExtWithLocale(result, locale)

	resp := make(map[string]*structpb.Value)
	for k, v := range menu {
		val, err := ToProtoValue(v)
		if err != nil {
			return nil, apierrors.ErrQueryExtension.InternalError(err)
		}
		resp[k] = val
	}

	return &pb.QueryExtensionsMenuResponse{Data: resp}, nil
}
func (s *extensionService) CreateExtensionVersion(ctx context.Context, req *pb.ExtensionVersionCreateRequest) (*pb.ExtensionVersionCreateResponse, error) {
	err := s.checkPushPermission(ctx)
	if err != nil {
		return nil, apierrors.ErrCreateExtensionVersion.AccessDenied()
	}
	return s.CreateExtensionVersionByRequest(req)
}

func (s *extensionService) CreateExtensionVersionByRequest(req *pb.ExtensionVersionCreateRequest) (*pb.ExtensionVersionCreateResponse, error) {
	specData := apistructs.Spec{}
	err := yaml.Unmarshal([]byte(req.SpecYml), &specData)
	if err != nil {
		return nil, apierrors.ErrQueryExtension.InternalError(err)
	}
	if specData.DisplayName == "" {
		specData.DisplayName = specData.Name
	}
	// Non-semantic version cannot set public
	_, err = semver.NewVersion(specData.Version)
	if err != nil && req.Public {
		return nil, apierrors.ErrQueryExtension.InternalError(err)
	}

	if !specData.CheckDiceVersion(version.Version) {
		err := s.db.DeleteExtensionVersion(specData.Name, specData.Version)
		if err != nil {
			return nil, apierrors.ErrQueryExtension.InternalError(err)
		}
		count, err := s.db.GetExtensionVersionCount(specData.Name)
		if err != nil {
			return nil, apierrors.ErrQueryExtension.InternalError(err)
		}
		if count == 0 {
			err = s.db.DeleteExtension(specData.Name)
			if err != nil {
				return nil, apierrors.ErrQueryExtension.InternalError(err)
			}
		}
		s.triggerPushEvent(specData, "delete")
		return nil, nil
	}
	labels := ""
	if specData.Labels != nil {
		for k, v := range specData.Labels {
			labels += k + ":" + v + ","
		}
	}
	extModel, err := s.db.GetExtension(req.Name)
	var ext *pb.Extension
	if err == nil {
		ext = extModel.ToApiData()
	} else if err == gorm.ErrRecordNotFound {
		// no same name,create
		ext, err = s.Create(&pb.ExtensionCreateRequest{
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
			return nil, apierrors.ErrQueryExtension.InternalError(err)
		}
	} else {
		return nil, apierrors.ErrQueryExtension.InternalError(err)
	}

	version, err := s.db.GetExtensionVersion(req.Name, req.Version)

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			version = &db.ExtensionVersion{
				ExtensionId: ext.Id,
				Name:        specData.Name,
				Version:     specData.Version,
				Dice:        req.DiceYml,
				Spec:        req.SpecYml,
				Swagger:     req.SwaggerYml,
				Readme:      req.Readme,
				Public:      req.Public,
				IsDefault:   req.IsDefault,
			}
			err = s.db.CreateExtensionVersion(version)
			s.triggerPushEvent(specData, "create")
			if err != nil {
				return nil, apierrors.ErrQueryExtension.InternalError(err)
			}

			data, err := version.ToApiData(ext.Type, false)
			if err != nil {
				return nil, apierrors.ErrQueryExtension.InternalError(err)
			}
			return &pb.ExtensionVersionCreateResponse{Data: data}, nil
		} else {
			return nil, apierrors.ErrQueryExtension.InternalError(err)
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
			err := s.db.SetUnDefaultVersion(version.Name)
			if err != nil {
				return nil, apierrors.ErrQueryExtension.InternalError(err)
			}
		}
		err := s.db.Save(&version).Error
		if err != nil {
			return nil, apierrors.ErrQueryExtension.InternalError(err)
		}
		s.triggerPushEvent(specData, "update")
		if req.All {
			extModel.Category = specData.Category
			extModel.LogoUrl = specData.LogoUrl
			extModel.DisplayName = specData.DisplayName
			extModel.Desc = specData.Desc
			extModel.Public = req.Public
			extModel.Labels = labels
			err = s.db.Save(&extModel).Error
			if err != nil {
				return nil, apierrors.ErrQueryExtension.InternalError(err)
			}
		}
		data, err := version.ToApiData(ext.Type, false)
		if err != nil {
			return nil, apierrors.ErrQueryExtension.InternalError(err)
		}
		return &pb.ExtensionVersionCreateResponse{Data: data}, nil
	} else {
		return nil, apierrors.ErrQueryExtension.InternalError(errors.New("version already exist"))
	}
}

func (s *extensionService) GetExtensionVersion(ctx context.Context, req *pb.GetExtensionVersionRequest) (*pb.GetExtensionVersionResponse, error) {
	result, err := s.GetExtension(req.Name, req.Version, req.YamlFormat)

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apierrors.ErrQueryExtension.NotFound()
		}
		return nil, apierrors.ErrQueryExtension.InternalError(err)
	}

	return &pb.GetExtensionVersionResponse{Data: result}, nil
}
func (s *extensionService) QueryExtensionVersions(ctx context.Context, req *pb.ExtensionVersionQueryRequest) (*pb.ExtensionVersionQueryResponse, error) {
	ext, err := s.db.GetExtension(req.Name)
	if err != nil {
		return nil, err
	}
	versions, err := s.db.QueryExtensionVersions(req.Name, req.All)
	if err != nil {
		return nil, err
	}
	result := []*pb.ExtensionVersion{}
	for _, v := range versions {
		data, err := v.ToApiData(ext.Type, false)
		if err != nil {
			return nil, err
		}
		result = append(result, data)
	}
	return &pb.ExtensionVersionQueryResponse{Data: result}, nil
}

// Create Create Extension
func (s *extensionService) Create(req *pb.ExtensionCreateRequest) (*pb.Extension, error) {
	if req.Type != "addon" && req.Type != "action" {
		return nil, apierrors.ErrCreateExtension.InvalidParameter("type")
	}
	if req.Name == "" {
		return nil, apierrors.ErrCreateExtension.InvalidParameter("name")
	}
	ext := db.Extension{
		Name:        req.Name,
		Type:        req.Type,
		Desc:        req.Desc,
		Category:    req.Category,
		DisplayName: req.DisplayName,
		LogoUrl:     req.LogoUrl,
		Public:      req.Public,
		Labels:      req.Labels,
	}
	err := s.db.CreateExtension(&ext)
	if err != nil {
		return nil, err
	}
	return ext.ToApiData(), nil
}

// QueryExtensionList Get Extension List
func (s *extensionService) QueryExtensionList(all string, typ string, labels string) ([]*pb.Extension, error) {
	extensions, err := s.db.QueryExtensions(all, typ, labels)
	if err != nil {
		return nil, err
	}
	result := []*pb.Extension{}
	for _, v := range extensions {
		apiData := v.ToApiData()
		result = append(result, apiData)
	}
	return result, nil
}

// GetExtension Get extension with specified version
func (s *extensionService) GetExtension(name string, version string, yamlFormat bool) (*pb.ExtensionVersion, error) {
	ext, err := s.db.GetExtension(name)
	if err != nil {
		return nil, err
	}
	var extensionVersion *db.ExtensionVersion
	if version == "default" {
		extensionVersion, err = s.db.GetExtensionDefaultVersion(name)
	} else {
		extensionVersion, err = s.db.GetExtensionVersion(name, version)
	}
	if err != nil {
		return nil, err
	}
	return extensionVersion.ToApiData(ext.Type, yamlFormat)
}

// GetExtensionDefaultVersion Get extension default version
func (s *extensionService) GetExtensionDefaultVersion(name string, yamlFormat bool) (*pb.ExtensionVersion, error) {
	ext, err := s.db.GetExtension(name)
	if err != nil {
		return nil, err
	}
	extensionVersion, err := s.db.GetExtensionDefaultVersion(name)
	if err != nil {
		return nil, err
	}
	return extensionVersion.ToApiData(ext.Type, yamlFormat)
}

func (s *extensionService) triggerPushEvent(specData apistructs.Spec, action string) {
	if specData.Type != "addon" {
		return
	}
	go func() {
		err := s.bdl.CreateEvent(&apistructs.EventCreateRequest{
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

func (s *extensionService) checkPushPermission(ctx context.Context) error {
	userID := apis.GetUserID(ctx)
	if userID == "" {
		return errors.Errorf("failed to get permission(User-ID is empty)")
	}
	data, err := s.bdl.CheckPermission(&apistructs.PermissionCheckRequest{
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

func ToProtoValue(i interface{}) (*structpb.Value, error) {
	data, err := json.Marshal(i)
	if err != nil {
		return nil, err
	}
	v := &structpb.Value{}
	err = v.UnmarshalJSON(data)
	if err != nil {
		return nil, err
	}
	return v, nil
}
