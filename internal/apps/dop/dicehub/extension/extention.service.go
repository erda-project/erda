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
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/erda-project/erda-proto-go/core/dicehub/extension/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/apps/dop/dicehub/service/apierrors"
	"github.com/erda-project/erda/pkg/common/apis"
	"github.com/erda-project/erda/pkg/limit_sync_group"
)

func (s *provider) SearchExtensions(ctx context.Context, req *pb.ExtensionSearchRequest) (*pb.ExtensionSearchResponse, error) {
	result := map[string]*pb.ExtensionVersion{}

	worker := limit_sync_group.NewWorker(5)
	for _, fullName := range req.Extensions {
		worker.AddFunc(func(locker *limit_sync_group.Locker, i ...interface{}) error {
			fnm := i[0].(string)
			splits := strings.SplitN(fnm, "@", 2)
			name := splits[0]
			version := ""
			if len(splits) > 1 {
				version = splits[1]
			}
			if version == "" {
				extVersion, _ := s.ExtensionSvc.GetExtensionDefaultVersion(name, req.YamlFormat)

				locker.Lock()
				result[fnm] = extVersion
				locker.Unlock()
			} else if strings.HasPrefix(version, "https://") || strings.HasPrefix(version, "http://") {
				extensionVersion, err := s.ExtensionSvc.GetExtensionByGit(name, version, "spec.yml", "dice.yml", "README.md")

				locker.Lock()
				if err != nil {
					result[fnm] = nil
				} else {
					result[fnm] = extensionVersion
				}
				locker.Unlock()
			} else {
				extensionVersion, err := s.ExtensionSvc.GetExtension(name, version, req.YamlFormat)

				locker.Lock()
				if err != nil {
					result[fnm] = nil
				} else {
					result[fnm] = extensionVersion
				}
				locker.Unlock()
			}
			return nil
		}, fullName)
	}

	err := worker.Do().Error()
	if err != nil {
		return nil, err
	}

	return &pb.ExtensionSearchResponse{Data: result}, nil
}

func (s *provider) CreateExtension(ctx context.Context, req *pb.ExtensionCreateRequest) (*pb.ExtensionCreateResponse, error) {
	err := s.checkPushPermission(ctx)
	if err != nil {
		return nil, apierrors.ErrCreateExtensionVersion.AccessDenied()
	}

	if req.Type != "action" && req.Type != "addon" {
		return nil, apierrors.ErrCreateExtension.InvalidParameter("type")
	}

	result, err := s.ExtensionSvc.Create(req)

	if err != nil {
		return nil, apierrors.ErrCreateExtension.InternalError(err)
	}

	return &pb.ExtensionCreateResponse{Data: result}, nil
}

func (s *provider) QueryExtensions(ctx context.Context, req *pb.QueryExtensionsRequest) (*pb.QueryExtensionsResponse, error) {
	result, err := s.ExtensionSvc.QueryExtensionList(req.All, req.Type, req.Labels)
	if err != nil {
		return nil, apierrors.ErrQueryExtension.InternalError(err)
	}

	locale := s.bdl.GetLocale(apis.GetLang(ctx))
	data, err := s.ExtensionSvc.MenuExtWithLocale(result, locale, req.All)
	if err != nil {
		return nil, apierrors.ErrQueryExtension.InternalError(err)
	}

	var newResult []*pb.Extension
	for _, menu := range data {
		for _, value := range menu {
			for _, extension := range value.Items {
				newResult = append(newResult, extension)
			}
		}
	}

	if req.Menu == "true" {
		menuExtResult := s.ExtensionSvc.MenuExt(newResult, s.Cfg.ExtensionMenu)
		resp, err := s.ExtensionSvc.ToProtoValue(menuExtResult)
		if err != nil {
			logrus.Errorf("fail transform interface to any type")
			return nil, apierrors.ErrQueryExtension.InternalError(err)
		}
		return &pb.QueryExtensionsResponse{Data: resp}, nil
	}

	resp, err := s.ExtensionSvc.ToProtoValue(newResult)
	if err != nil {
		logrus.Errorf("fail transform interface to any type")
		return nil, apierrors.ErrQueryExtension.InternalError(err)
	}
	return &pb.QueryExtensionsResponse{Data: resp}, nil
}

func (s *provider) QueryExtensionsMenu(ctx context.Context, req *pb.QueryExtensionsMenuRequest) (*pb.QueryExtensionsMenuResponse, error) {
	locale := s.bdl.GetLocale(apis.GetLang(ctx))
	result, err := s.ExtensionSvc.QueryExtensionList(req.All, req.Type, req.Labels)
	if err != nil {
		return nil, apierrors.ErrQueryExtension.InternalError(err)
	}

	data, err := s.ExtensionSvc.MenuExtWithLocale(result, locale, req.All)
	if err != nil {
		return nil, apierrors.ErrQueryExtension.InternalError(err)
	}

	resp := make(map[string]*structpb.Value)
	for k, v := range data {
		val, err := s.ExtensionSvc.ToProtoValue(v)
		if err != nil {
			return nil, apierrors.ErrQueryExtension.InternalError(err)
		}
		resp[k] = val
	}

	return &pb.QueryExtensionsMenuResponse{Data: resp}, nil
}

func (s *provider) CreateExtensionVersion(ctx context.Context, req *pb.ExtensionVersionCreateRequest) (*pb.ExtensionVersionCreateResponse, error) {
	err := s.checkPushPermission(ctx)
	if err != nil {
		return nil, apierrors.ErrCreateExtensionVersion.AccessDenied()
	}
	return s.ExtensionSvc.CreateExtensionVersionByRequest(req)
}

func (s *provider) GetExtensionVersion(ctx context.Context, req *pb.GetExtensionVersionRequest) (*pb.GetExtensionVersionResponse, error) {
	result, err := s.ExtensionSvc.GetExtension(req.Name, req.Version, req.YamlFormat)

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apierrors.ErrQueryExtension.NotFound()
		}
		return nil, apierrors.ErrQueryExtension.InternalError(err)
	}

	return &pb.GetExtensionVersionResponse{Data: result}, nil
}

func (s *provider) QueryExtensionVersions(ctx context.Context, req *pb.ExtensionVersionQueryRequest) (*pb.ExtensionVersionQueryResponse, error) {
	res, err := s.ExtensionSvc.QueryExtensionVersions(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s *provider) checkPushPermission(ctx context.Context) error {
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
