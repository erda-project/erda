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

package marketplace

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/pkg/transport"
	_ "github.com/erda-project/erda-infra/providers/mysql/v2"
	"github.com/erda-project/erda-proto-go/admin/gallery/pb"
	commonPb "github.com/erda-project/erda-proto-go/common/pb"
	extensionPb "github.com/erda-project/erda-proto-go/core/dicehub/extension/pb"
	releasepb "github.com/erda-project/erda-proto-go/core/dicehub/release/pb"
	"github.com/erda-project/erda/modules/admin/services/cache"
	"github.com/erda-project/erda/modules/admin/services/marketplace/apierr"
	"github.com/erda-project/erda/modules/admin/services/marketplace/model"
	"github.com/erda-project/erda/pkg/common/apis"
)

const (
	ExtensionAction  GalleryType = "extension/action"
	ExtensionAddon   GalleryType = "extension/addon"
	ArtifactsProject GalleryType = "artifacts/project"
)

var (
	name = "erda.admin.gallery"
	spec = servicehub.Spec{
		Define:               nil,
		Services:             pb.ServiceNames(),
		Dependencies:         nil,
		OptionalDependencies: []string{"service-register"},
		DependenciesFunc:     nil,
		Summary:              "gallery service",
		Description:          "gallery service",
		ConfigFunc: func() interface{} {
			return new(struct{})
		},
		Types: pb.Types(),
		Creator: func() servicehub.Provider {
			return new(provider)
		},
	}
)

func init() {
	servicehub.Register(name, &spec)
}

// +provider
type provider struct {
	R  transport.Register `autowired:"service-register" required:"true"`
	DB *gorm.DB           `autowired:"mysql-gorm.v2-client"`

	// providers clients
	c *cache.Cache `autowired:"easy-memory-cache-client"`

	// gRPC clients
	extensionCli extensionPb.ExtensionServiceServer `autowired:"erda.core.dicehub.extension.ExtensionService"`
	releaseCli   releasepb.ReleaseServiceServer     `autowired:"erda.core.dicehub.release.ReleaseGetDiceService"`

	l *logrus.Entry
}

func (p *provider) Init(ctx servicehub.Context) error {
	p.l = logrus.WithField("provider", name)
	p.l.Infoln("Init")
	if p.R == nil {
		p.l.Infoln("register self")
		pb.RegisterMarketplaceImp(p.R, p, apis.Options())
	}
	return nil
}

func (p *provider) ListArtifacts(ctx context.Context, req *pb.ListArtifactsReq) (*pb.ListArtifactsResp, error) {
	//TODO implement me
	panic("implement me")
}

func (p *provider) CreateArtifacts(ctx context.Context, req *pb.CreateArtifactsReq) (*commonPb.VoidResponse, error) {
	// get org info
	orgID := apis.GetOrgID(ctx)
	orgDTO, ok := p.c.GetOrgByOrgID(orgID)
	if !ok {
		return nil, apierr.CreateGallery.InvalidParameter("invalid organization: " + orgID)
	}

	// get user info
	userID := apis.GetUserID(ctx)

	// todo: 鉴权

	// do not support other gallery type yet
	if ArtifactsProject.String() != req.GetType() {
		return nil, apierr.CreateGallery.InvalidParameter("invalid artifacts type: " + req.GetType())
	}
	if req.GetName() == "" {
		return nil, apierr.CreateGallery.InvalidParameter("invalid artifacts name")
	}
	if req.GetVersion() == "" {
		return nil, apierr.CreateGallery.InvalidParameter("invalid artifacts name")
	}
	var spec = new(pb.CreateArtifactsReq_ReleaseSpec)
	if err := req.GetSpec().UnmarshalTo(spec); err != nil {
		return nil, apierr.CreateGallery.InvalidParameter("invalid spec")
	}

	// check the record if is exists
	switch err := p.DB.Where(map[string]interface{}{
		"org_id":  orgID,
		"name":    req.GetName(),
		"type":    req.GetType(),
		"version": req.GetVersion(),
	}).Find(new(model.MarketplaceGalleryArtifacts)).Error; {
	case err == nil:
		return nil, apierr.CreateGallery.InvalidParameter(fmt.Sprintf("the gallery is exists, name: %s, type: %s, version: %s",
			req.GetName(), req.GetType(), req.GetVersion()))
	case errors.Is(err, gorm.ErrRecordNotFound):
	default:
		return nil, apierr.CreateGallery.InternalError(err)
	}

	// check the release if is not exists
	release, err := p.releaseCli.GetRelease(ctx, &releasepb.ReleaseGetRequest{ReleaseID: spec.GetReleaseID()})
	if err != nil {
		p.l.WithField("releaseID", spec.GetReleaseID()).Errorln("failed to p.releaseCli.GetRelease")
		return nil, apierr.CreateGallery.InternalError(errors.Wrap(err, "failed to get release"))
	}
	if release == nil || release.Data == nil {
		p.l.WithField("releaseID", spec.GetReleaseID()).Warnln("release data not found")
		return nil, apierr.CreateGallery.NotFound()
	}

	tx := p.DB.Begin()
	defer func() {
		if err == nil {
			tx.Commit()
		} else {
			tx.Rollback()
		}
	}()

	if err = tx.Model(new(model.MarketplaceGalleryArtifacts)).
		Where(map[string]interface{}{
			"org_id": orgID,
			"name":   req.GetName(),
			"type":   req.GetType(),
		}).
		Updates(map[string]interface{}{
			"is_default": false,
			"updater_id": userID,
		}).
		Error; err != nil {
		return nil, apierr.CreateGallery.InternalError(err)
	}

	var m = &model.MarketplaceGalleryArtifacts{
		Common: model.Common{
			OrgID:     uint32(orgDTO.ID),
			OrgName:   orgDTO.Name,
			CreatorID: userID,
			UpdaterID: userID,
		},
		ReleaseID:   spec.GetReleaseID(),
		Name:        req.GetName(),
		DisplayName: "", // todo: 查询项目, 以项目 displaceName 为值
		Version:     req.GetVersion(),
		Type:        req.GetType(),
		Changelog:   release.Data.Changelog,
		IsDefault:   true,
	}
	if err = tx.Create(m).Error; err != nil {
		return nil, apierr.CreateGallery.InternalError(err)
	}

	if _, err = p.releaseCli.PublishArtifacts(ctx, &releasepb.ReleaseGetRequest{ReleaseID: spec.GetReleaseID()}); err != nil {
		return nil, apierr.CreateGallery.InternalError(errors.Wrap(err, "failed to PublishArtifacts"))
	}

	return new(commonPb.VoidResponse), nil
}

func (p *provider) GetArtifacts(ctx context.Context, req *pb.GetArtifactsReq) (*pb.GetArtifactsResp, error) {
	if req.GetName() == "" {
		return nil, apierr.GetGallery.NotFound()
	}

	switch GalleryType(req.GetType()) {
	case ExtensionAction:
		extensions, err := p.getExtensions(ctx, req.GetName(), "action", req.GetVersion())
		if err != nil {
			return nil, apierr.GetGallery.InternalError(err)
		}
		return &pb.GetArtifactsResp{Data: &pb.GetArtifactsRespData{
			Total: int32(len(extensions)),
			List:  extensions,
		}}, nil
	case ExtensionAddon:
		galleries, err := p.getExtensions(ctx, req.GetName(), "addon", req.GetVersion())
		if err != nil {
			return nil, apierr.GetGallery.InternalError(err)
		}
		return &pb.GetArtifactsResp{Data: &pb.GetArtifactsRespData{
			Total: int32(len(galleries)),
			List:  galleries,
		}}, nil
	case ArtifactsProject:
		galleries, err := p.getProjectArtifacts(ctx, apis.GetOrgID(ctx), req.GetName(), req.GetVersion())
		if err != nil {
			return nil, apierr.GetGallery.InternalError(err)
		}
		var resp = &pb.GetArtifactsResp{Data: &pb.GetArtifactsRespData{
			Total: int32(len(galleries)),
			List:  galleries,
		}}
		for _, item := range resp.Data.List {
			resp.UserIDs = append(resp.UserIDs, item.GetPublisher().GetId())
		}
		return resp, nil
	default:
		return nil, apierr.GetGallery.InvalidParameter("invalid gallery type")
	}
}

func (p *provider) getExtensions(ctx context.Context, name, type_, version string) ([]*pb.GetArtifactsRespDataItem, error) {
	versions, err := p.extensionCli.QueryExtensionVersions(ctx, &extensionPb.ExtensionVersionQueryRequest{
		Name:               name,
		YamlFormat:         true,
		All:                true,
		OrderByVersionDesc: true,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to QueryExtensionVersions")
	}
	var (
		result    []*pb.GetArtifactsRespDataItem
		publisher = &pb.User{
			Id:       "0",
			Name:     "erda",
			NickName: "Erda",
		}
	)
	for _, v := range versions.Data {
		if type_ != v.GetType() {
			continue
		}
		if version != "" && version != v.GetVersion() {
			continue
		}
		item := &pb.GetArtifactsRespDataItem{
			Name:        v.GetName(),
			DisplayName: "", // todo: 需要解析 spec.yml
			Version:     v.GetVersion(),
			Type:        v.GetType(),
			Category:    "", // todo: 需要解析 spec.yml
			LogoUrl:     "", // todo:
			Dice:        nil,
			Spec:        nil,
			Readme:      v.GetReadme(),
			IsDefault:   v.GetIsDefault(),
			CreatedAt:   v.GetCreatedAt(),
			UpdatedAt:   v.GetUpdatedAt(),
			Publisher:   publisher,
			Params:      nil, // todo:
			Outputs:     nil, // todo:
		}
		item.Dice = new(anypb.Any)
		item.Spec = new(anypb.Any)
		if err := item.Dice.MarshalFrom(v.GetDice()); err != nil {
			return nil, errors.Wrap(err, "failed to item.Dice.MarshalFrom")
		}
		if err := item.Spec.MarshalFrom(v.GetSpec()); err != nil {
			return nil, errors.Wrap(err, "failed to item.Dice.MarshalFrom")
		}
		result = append(result)
	}
	return result, nil
}

func (p *provider) getProjectArtifacts(_ context.Context, orgID, name, version string) ([]*pb.GetArtifactsRespDataItem, error) {
	var artifacts []model.MarketplaceGalleryArtifacts
	where := p.DB.Where("name = ?", name)
	if version != "" {
		where = where.Where("version = ?", version)
	}
	where = where.Where("org_id = ?", orgID)
	if err := where.Find(&artifacts).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.Wrap(err, "failed to Find")
	}
	var result []*pb.GetArtifactsRespDataItem
	for _, a := range artifacts {
		var item = &pb.GetArtifactsRespDataItem{
			Name:        a.Name,
			DisplayName: a.DisplayName,
			Version:     a.Version,
			Type:        ArtifactsProject.String(),
			ChangeLog:   a.Changelog,
			IsDefault:   a.IsDefault,
			CreatedAt:   timestamppb.New(a.CreatedAt),
			UpdatedAt:   timestamppb.New(a.UpdatedAt),
			Publisher:   &pb.User{Id: a.CreatorID},
		}
		result = append(result, item)
	}
	return result, nil
}

func (p *provider) DeleteArtifacts(ctx context.Context, req *pb.DeleteArtifactsReq) (*commonPb.VoidResponse, error) {
	// get org info
	orgID := apis.GetOrgID(ctx)

	// get user info
	userID := apis.GetUserID(ctx)

	// todo: 鉴权

	// do not support other artifacts type yet
	if ArtifactsProject.String() != req.GetType() {
		return nil, apierr.DeleteGallery.InvalidParameter("invalid artifacts type: " + req.GetType())
	}
	if req.GetName() == "" {
		return nil, apierr.DeleteGallery.InvalidParameter("invalid artifacts name")
	}
	if req.GetVersion() == "" {
		return nil, apierr.DeleteGallery.InvalidParameter("invalid artifacts version")
	}

	var (
		tx  = p.DB.Begin()
		err error
	)
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	if err = tx.Where(map[string]interface{}{
		"org_id":  orgID,
		"name":    req.GetName(),
		"type":    req.GetType(),
		"version": req.GetVersion(),
	}).
		Delete(new(model.MarketplaceGalleryArtifacts)).
		Error; err != nil {
		return nil, apierr.DeleteGallery.InternalError(err)
	}

	// set default

	// tag unpublish in dicehub

	//TODO implement me
	panic("implement me")
}

type GalleryType string

func (t GalleryType) String() string {
	return string(t)
}
