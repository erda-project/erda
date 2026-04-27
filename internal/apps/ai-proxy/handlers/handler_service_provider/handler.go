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
	"strings"

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

const (
	volcengineVikingTemplateID      = "volcengine-viking"
	volcengineVikingAKTemplateParam = "access-key-id"
	volcengineVikingSKTemplateParam = "secret-access-key"
	volcengineVikingAPIKeyParam     = "api-key"
)

type ServiceProviderHandler struct {
	DAO dao.DAO
}

func (h *ServiceProviderHandler) Create(ctx context.Context, req *pb.ServiceProviderCreateRequest) (*pb.ServiceProvider, error) {
	if err := validateVolcengineVikingTemplateParams(req.TemplateId, req.TemplateParams); err != nil {
		return nil, err
	}

	// check template
	tpl, err := cachehelpers.GetAndCheckTemplateByTypeName(ctx, templatetypes.TemplateTypeServiceProvider, req.TemplateId, req.TemplateParams)
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
		Desc:           strutil.FirstNotEmpty(req.Desc, rawProviderTemplate.Desc),
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

	finalTemplateParams := current.TemplateParams
	if req.TemplateParams != nil {
		finalTemplateParams = req.TemplateParams
	}
	if err := validateVolcengineVikingTemplateParams(current.TemplateId, finalTemplateParams); err != nil {
		return nil, err
	}

	u := pb.ServiceProvider{
		Id:             req.Id,
		Name:           req.Name,
		Desc:           req.Desc,
		Metadata:       req.Metadata,
		ClientId:       req.ClientId,
		TemplateParams: finalTemplateParams,
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

func validateVolcengineVikingTemplateParams(templateID string, params map[string]string) error {
	if templateID != volcengineVikingTemplateID {
		return nil
	}
	if strings.TrimSpace(params[volcengineVikingAPIKeyParam]) != "" {
		return fmt.Errorf("template_params.api-key is not supported for volcengine-viking, use access-key-id/secret-access-key")
	}
	if strings.TrimSpace(params[volcengineVikingAKTemplateParam]) == "" {
		return fmt.Errorf("template_params.access-key-id is required for volcengine-viking")
	}
	if strings.TrimSpace(params[volcengineVikingSKTemplateParam]) == "" {
		return fmt.Errorf("template_params.secret-access-key is required for volcengine-viking")
	}
	return nil
}
