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

package publishitem

import (
	"encoding/json"
	"net/http"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/pkg/transport"
	muxserver "github.com/erda-project/erda-infra/providers/httpserver"
	"github.com/erda-project/erda-infra/providers/mysql"
	"github.com/erda-project/erda-proto-go/dop/publishitem/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/internal/apps/dop/providers/publishitem/db"
	"github.com/erda-project/erda/internal/apps/dop/services/apierrors"
	"github.com/erda-project/erda/internal/pkg/user"
	"github.com/erda-project/erda/pkg/common/apis"
	"github.com/erda-project/erda/pkg/database/dbengine"
	"github.com/erda-project/erda/pkg/http/httpserver"
)

type config struct {
	SiteUrl string `env:"SITE_URL"`
}

// +provider
type provider struct {
	Cfg                *config
	Log                logs.Logger
	Register           transport.Register
	MySQL              mysql.Interface
	publishItemService *PublishItemService
}

func (p *provider) Init(ctx servicehub.Context) error {
	bdl := bundle.New()
	p.publishItemService = &PublishItemService{p: p,
		db: &db.DBClient{DBEngine: &dbengine.DBEngine{
			DB: p.MySQL.DB()}},
		bdl: bdl,
	}
	if p.Register != nil {
		p.Register.Add(http.MethodGet, "/api/publish-items/{publishItemId}/distribution", func(rw http.ResponseWriter, r *http.Request) {
			var req pb.DistributePublishItemRequest
			vars := muxserver.Vars(r)
			publishItemId, err := getPublishItemId(vars)
			if err != nil {
				httpserver.WriteErr(rw, "400", apierrors.ErrCreateOffLinePublishItemVersion.InvalidParameter(err).Error())
				return
			}
			req.MobileType = r.URL.Query().Get("mobileType")
			req.PackageName = r.URL.Query().Get("packageName")
			req.PublishItemId = publishItemId
			result, err := p.publishItemService.GetPublishItemDistribution(req.PublishItemId, apistructs.ResourceType(req.MobileType), req.PackageName,
				rw, r)
			if err != nil {
				httpserver.WriteErr(rw, "400", apierrors.ErrGetPublishItem.InternalError(err).Error())
				return
			}
			httpserver.WriteData(rw, &pb.PublishItemDistributionResponse{
				Data: result,
			})
		})
		p.Register.Add(http.MethodPost, "/api/publish-items/{publishItemId}/versions/create-offline-version", func(rw http.ResponseWriter, r *http.Request) {
			vars := muxserver.Vars(r)
			// 获取上传文件
			_, fileHeader, err := r.FormFile("file")
			if err != nil {
				httpserver.WriteErr(rw, "400", apierrors.ErrCreateOffLinePublishItemVersion.InvalidParameter(err).Error())
				return
			}
			// 校验用户登录
			identityInfo, err := user.GetIdentityInfo(r)
			if err != nil {
				httpserver.WriteErr(rw, "400", apierrors.ErrCreateOffLinePublishItemVersion.NotLogin().Error())
				return
			}
			// 获取publishitemID
			itemID, err := getPublishItemId(vars)
			if err != nil {
				httpserver.WriteErr(rw, "400", apierrors.ErrCreateOffLinePublishItemVersion.InvalidParameter(err).Error())
				return
			}
			// 获取企业ID
			orgID, err := getPermissionHeader(r)
			if err != nil {
				httpserver.WriteErr(rw, "400", apierrors.ErrCreateOffLinePublishItemVersion.NotLogin().Error())
				return
			}

			req := apistructs.CreateOffLinePublishItemVersionRequest{
				Desc:          r.PostFormValue("desc"),
				FileHeader:    fileHeader,
				PublishItemID: itemID,
				IdentityInfo:  identityInfo,
				OrgID:         orgID,
			}

			mobileType, err := p.publishItemService.CreateOffLineVersion(req)
			if err != nil {
				httpserver.WriteErr(rw, "400", apierrors.ErrCreateOffLinePublishItemVersion.InternalError(err).Error())
			}

			httpserver.WriteData(rw, &pb.CreatePublishItemOfflineResponse{
				Data: mobileType,
			})
		})
		p.Register.Add(http.MethodPost, "/api/publish-items/actions/latest-versions", func(rw http.ResponseWriter, r *http.Request) {
			var req pb.GetPublishItemLatestVersionRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				httpserver.WriteErr(rw, "400", apierrors.ErrGetPublishItem.InvalidParameter(err).Error())
				return
			}
			if req.Ak == "" || req.Ai == "" {
				httpserver.WriteErr(rw, "400", apierrors.ErrGetPublishItem.MissingParameter("ak or ai").Error())
				return
			}
			results, err := p.publishItemService.GetPublicPublishItemLaststVersion(rw, r, req)
			if err != nil {
				httpserver.WriteErr(rw, "400", apierrors.ErrGetPublishItem.InternalError(err).Error())
				return
			}
			httpserver.WriteData(rw, &pb.GetPublishItemLatestVersionResponse{
				Data: results,
			})
		})
		pb.RegisterPublishItemServiceImp(p.Register, p.publishItemService, apis.Options())
	}
	return nil
}

func (p *provider) Provide(ctx servicehub.DependencyContext, args ...interface{}) interface{} {
	switch {
	case ctx.Service() == "erda.dop.publishitem.PublishItemService" || ctx.Type() == pb.PublishItemServiceServerType() || ctx.Type() == pb.PublishItemServiceHandlerType():
		return p.publishItemService
	}
	return p
}

func init() {
	servicehub.Register("erda.dop.publishitem", &servicehub.Spec{
		Services:             pb.ServiceNames(),
		Types:                pb.Types(),
		OptionalDependencies: []string{"service-register"},
		Description:          "",
		ConfigFunc: func() interface{} {
			return &config{}
		},
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
