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

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/pkg/transport"
	"github.com/erda-project/erda-proto-go/admin/marketplace/pb"
	commonPb "github.com/erda-project/erda-proto-go/common/pb"
	extensionPb "github.com/erda-project/erda-proto-go/core/dicehub/extension/pb"
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
	name = "erda.admin.marketplace"
	spec = servicehub.Spec{
		Define:               nil,
		Services:             pb.ServiceNames(),
		Dependencies:         nil,
		OptionalDependencies: []string{"service-register"},
		DependenciesFunc:     nil,
		Summary:              "marketplace service",
		Description:          "marketplace service",
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

	// gRPC clients
	extensionCli extensionPb.ExtensionServiceServer `autowired:"erda.core.dicehub.extension.ExtensionService"`

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

func (p *provider) ListGalleries(ctx context.Context, req *pb.ListGalleriesReq) (*pb.ListGalleriesResp, error) {
	//TODO implement me
	panic("implement me")
}

func (p *provider) CreateGallery(ctx context.Context, req *pb.CreateGalleryReq) (*commonPb.VoidResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (p *provider) GetGallery(ctx context.Context, req *pb.GetGalleryReq) (*pb.GetGalleryResp, error) {
	if req.GetName() == "" {
		return nil, apierr.GetGallery.NotFound()
	}

	switch GalleryType(req.GetType()) {
	case ExtensionAction:
		extensions, err := p.getExtensions(ctx, req.GetName(), "action", req.GetVersion())
		if err != nil {
			return nil, apierr.GetGallery.InternalError(err)
		}
		return &pb.GetGalleryResp{Data: &pb.GetGalleryRespData{
			Total: int32(len(extensions)),
			List:  extensions,
		}}, nil
	case ExtensionAddon:
		galleries, err := p.getExtensions(ctx, req.GetName(), "addon", req.GetVersion())
		if err != nil {
			return nil, apierr.GetGallery.InternalError(err)
		}
		return &pb.GetGalleryResp{Data: &pb.GetGalleryRespData{
			Total: int32(len(galleries)),
			List:  galleries,
		}}, nil
	case ArtifactsProject:
		galleries, err := p.getProjectArtifacts(ctx, apis.GetOrgID(ctx), req.GetName(), req.GetVersion())
		if err != nil {
			return nil, apierr.GetGallery.InternalError(err)
		}
		var resp = &pb.GetGalleryResp{Data: &pb.GetGalleryRespData{
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

func (p *provider) getExtensions(ctx context.Context, name, type_, version string) ([]*pb.GetGalleryRespDataItem, error) {
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
		result    []*pb.GetGalleryRespDataItem
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
		item := &pb.GetGalleryRespDataItem{
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

func (p *provider) getProjectArtifacts(_ context.Context, orgID, name, version string) ([]*pb.GetGalleryRespDataItem, error) {
	var artifacts []model.MarketplaceGalleryArtifacts
	where := p.DB.Where("name = ?", name)
	if version != "" {
		where = where.Where("version = ?", version)
	}
	where = where.Where("org_id = ?", orgID)
	if err := where.Find(&artifacts).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.Wrap(err, "failed to Find")
	}
	var result []*pb.GetGalleryRespDataItem
	for _, a := range artifacts {
		var item = &pb.GetGalleryRespDataItem{
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

func (p *provider) DeleteGallery(ctx context.Context, req *pb.DeleteGalleryReq) (*commonPb.VoidResponse, error) {
	//TODO implement me
	panic("implement me")
}

type GalleryType string

func (t GalleryType) String() string {
	return string(t)
}
