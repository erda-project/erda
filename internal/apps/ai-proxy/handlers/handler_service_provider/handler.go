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

package handler_service_provider

import (
	"context"
	"fmt"

	"github.com/mohae/deepcopy"

	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	modelpb "github.com/erda-project/erda-proto-go/apps/aiproxy/model/pb"
	"github.com/erda-project/erda-proto-go/apps/aiproxy/service_provider/pb"
	templatepb "github.com/erda-project/erda-proto-go/apps/aiproxy/template/pb"
	commonpb "github.com/erda-project/erda-proto-go/common/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/cache/cachehelpers"
	"github.com/erda-project/erda/internal/apps/ai-proxy/cache/cachetypes"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/auth"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/common_types"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/template"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/template/templatetypes"
	"github.com/erda-project/erda/internal/apps/ai-proxy/providers/dao"
	"github.com/erda-project/erda/pkg/strutil"
)

type ServiceProviderHandler struct {
	DAO dao.DAO
}

func (h *ServiceProviderHandler) Create(ctx context.Context, req *pb.ServiceProviderCreateRequest) (*pb.ServiceProvider, error) {
	// check template
	tpl, err := cachehelpers.GetTemplateByTypeName(ctx, templatetypes.TemplateTypeServiceProvider, req.TemplateId, req.TemplateParams)
	if err != nil {
		return nil, fmt.Errorf("failed to get template: %w", err)
	}
	// convert template.config for easy-use
	var rawProviderTemplate pb.ServiceProvider
	cputil.MustObjJSONTransfer(tpl.Config, &rawProviderTemplate)

	renderedTpl := deepcopy.Copy(tpl).(*templatepb.Template)
	if err := template.RenderTemplate(req.TemplateId, renderedTpl, req.TemplateParams); err != nil {
		return nil, fmt.Errorf("failed to render template: %w", err)
	}
	// convert template.config for easy-use
	var renderedProviderTemplate pb.ServiceProvider
	cputil.MustObjJSONTransfer(renderedTpl.Config, &renderedProviderTemplate)

	// type
	if !common_types.ServiceProviderType(renderedProviderTemplate.Type).IsValid() {
		return nil, fmt.Errorf("invalid provider type: %s", rawProviderTemplate.Type)
	}

	c := pb.ServiceProvider{
		Name:           req.Name,
		Desc:           strutil.FirstNoneEmpty(req.Desc, rawProviderTemplate.Desc),
		Type:           renderedProviderTemplate.Type, // use rendered value for paging
		ApiKey:         rawProviderTemplate.ApiKey,
		Metadata:       req.Metadata, // only store requested metadata; all metadata will be merged when display or use
		ClientId:       req.ClientId,
		TemplateId:     req.TemplateId,
		TemplateParams: req.TemplateParams,
	}
	result, err := h.DAO.ServiceProviderClient().Create(ctx, &c)
	if err != nil {
		return nil, err
	}
	// trigger cache refresh
	go ctxhelper.MustGetCacheManager(ctx).(cachetypes.Manager).TriggerRefresh(ctx, cachetypes.ItemTypeProvider)
	return result, nil
}

func (h *ServiceProviderHandler) Get(ctx context.Context, req *pb.ServiceProviderGetRequest) (*pb.ServiceProvider, error) {
	provider, err := h.DAO.ServiceProviderClient().Get(ctx, req)
	if err != nil {
		return nil, err
	}
	if req.RenderTemplate {
		provider, err = cachehelpers.GetRenderedServiceProviderByID(ctx, provider.Id)
		if err != nil {
			return nil, err
		}
	}
	// data sensitive
	desensitizeProvider(ctx, provider)
	return provider, nil
}

func (h *ServiceProviderHandler) Delete(ctx context.Context, req *pb.ServiceProviderDeleteRequest) (*commonpb.VoidResponse, error) {
	// check models
	modelPagingResp, err := h.DAO.ModelClient().Paging(ctx, &modelpb.ModelPagingRequest{
		PageNum:    1,
		PageSize:   1,
		ProviderId: req.Id,
	})
	if err != nil {
		return nil, err
	}
	if modelPagingResp.Total > 0 {
		return nil, fmt.Errorf("provider has related models, can not delete")
	}
	result, err := h.DAO.ServiceProviderClient().Delete(ctx, req)
	if err != nil {
		return nil, err
	}
	// trigger cache refresh
	go ctxhelper.MustGetCacheManager(ctx).(cachetypes.Manager).TriggerRefresh(ctx, cachetypes.ItemTypeProvider)
	return result, nil
}

func (h *ServiceProviderHandler) Update(ctx context.Context, req *pb.ServiceProviderUpdateRequest) (*pb.ServiceProvider, error) {
	// get current
	current, err := h.DAO.ServiceProviderClient().Get(ctx, &pb.ServiceProviderGetRequest{Id: req.Id})
	if err != nil {
		return nil, err
	}
	u := pb.ServiceProvider{
		Id:             req.Id,
		Name:           req.Name,
		Desc:           req.Desc,
		Metadata:       req.Metadata,
		ClientId:       req.ClientId,
		TemplateParams: req.TemplateParams,
	}
	if req.IsEnabled == nil {
		u.IsEnabled = current.IsEnabled
	} else {
		u.IsEnabled = req.IsEnabled
	}
	result, err := h.DAO.ServiceProviderClient().Update(ctx, &u)
	if err != nil {
		return nil, err
	}
	// trigger cache refresh
	go ctxhelper.MustGetCacheManager(ctx).(cachetypes.Manager).TriggerRefresh(ctx, cachetypes.ItemTypeProvider)
	return result, nil
}

func (h *ServiceProviderHandler) Paging(ctx context.Context, req *pb.ServiceProviderPagingRequest) (*pb.ServiceProviderPagingResponse, error) {
	resp, err := h.DAO.ServiceProviderClient().Paging(ctx, req)
	if err != nil {
		return nil, err
	}
	// data sensitive
	for i, item := range resp.List {
		if req.RenderTemplate {
			item, err = cachehelpers.GetRenderedServiceProviderByID(ctx, item.Id)
			if err != nil {
				return nil, err
			}
			resp.List[i] = item
		}
		desensitizeProvider(ctx, item)
	}
	return resp, nil
}

func desensitizeProvider(ctx context.Context, item *pb.ServiceProvider) {
	// pass for: admin
	if auth.IsAdmin(ctx) {
		return
	}
	// if a provider is client-belonged, pass for the client
	if item.ClientId != "" && auth.IsClient(ctx) {
		return
	}
	// hide sensitive data
	item.ApiKey = ""
	item.Metadata.Secret = nil
	item.TemplateParams = nil
}

//func renderTemplate(ctx context.Context, provider *pb.ServiceProvider) error {
//	// get template
//	tpl, err := cachehelpers.GetTemplateByTypeName(ctx, templatetypes.TemplateTypeServiceProvider, provider.TemplateId, provider.TemplateParams)
//	if err != nil {
//		return fmt.Errorf("failed to get template: %w", err)
//	}
//	// render template
//	if err := template.RenderTemplate(tpl, provider.TemplateParams); err != nil {
//		return fmt.Errorf("failed to render template: %w", err)
//	}
//	// merge providerTemplate to provider
//	var providerTemplate pb.ServiceProvider
//	cputil.MustObjJSONTransfer(tpl.Config, &providerTemplate)
//	provider.ApiKey = providerTemplate.ApiKey
//	mergedMetadata, err := metadata.OverrideMetadata(ctx, providerTemplate.Metadata, provider.Metadata)
//	if err != nil {
//		return err
//	}
//	provider.Metadata = mergedMetadata
//	return nil
//}
