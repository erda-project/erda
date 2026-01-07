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

package handler_model

import (
	"context"
	"fmt"

	"github.com/mohae/deepcopy"

	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	clientmodelrelationpb "github.com/erda-project/erda-proto-go/apps/aiproxy/client_model_relation/pb"
	"github.com/erda-project/erda-proto-go/apps/aiproxy/model/pb"
	providerpb "github.com/erda-project/erda-proto-go/apps/aiproxy/service_provider/pb"
	templatepb "github.com/erda-project/erda-proto-go/apps/aiproxy/template/pb"
	commonpb "github.com/erda-project/erda-proto-go/common/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/cache/cachehelpers"
	"github.com/erda-project/erda/internal/apps/ai-proxy/cache/cachetypes"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/auth"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/template"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/template/templatetypes"
	"github.com/erda-project/erda/internal/apps/ai-proxy/handlers/handler_i18n/i18n_services"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/i18n"
	"github.com/erda-project/erda/internal/apps/ai-proxy/providers/dao"
	"github.com/erda-project/erda/pkg/strutil"
)

type ModelHandler struct {
	DAO dao.DAO
}

func (h *ModelHandler) Create(ctx context.Context, req *pb.ModelCreateRequest) (*pb.Model, error) {
	// check template
	tpl, err := cachehelpers.GetAndCheckTemplateByTypeName(ctx, templatetypes.TemplateTypeModel, req.TemplateId, req.TemplateParams)
	if err != nil {
		return nil, fmt.Errorf("failed to get template: %w", err)
	}
	// convert template.config for easy-use
	var rawModelTemplate pb.Model
	cputil.MustObjJSONTransfer(tpl.Config, &rawModelTemplate)

	renderedTpl := deepcopy.Copy(tpl).(*templatepb.Template)
	if err := template.RenderTemplate(req.TemplateId, renderedTpl, req.TemplateParams); err != nil {
		return nil, fmt.Errorf("failed to render template: %w", err)
	}
	// convert template.config for easy-use
	var renderedModelTemplate pb.Model
	cputil.MustObjJSONTransfer(renderedTpl.Config, &renderedModelTemplate)

	// check provider
	cache := ctxhelper.MustGetCacheManager(ctx).(cachetypes.Manager)
	providerV, err := cache.GetByID(ctx, cachetypes.ItemTypeProvider, req.ProviderId)
	if err != nil {
		return nil, err
	}
	provider := providerV.(*providerpb.ServiceProvider)
	// check client-id
	if req.ClientId != "" {
		if provider.ClientId != req.ClientId {
			return nil, fmt.Errorf("provider doesn't belong to client")
		}
	} else {
		// use provider's client-id as default
		req.ClientId = provider.ClientId
	}

	model := &pb.Model{
		Name:           req.Name,
		Desc:           strutil.FirstNotEmpty(req.Desc, rawModelTemplate.Desc),
		Type:           renderedModelTemplate.Type, // use rendered value for paging
		ProviderId:     req.ProviderId,
		ApiKey:         rawModelTemplate.ApiKey,
		Metadata:       req.Metadata,                    // only store requested metadata; all metadata will be merged when display or use
		Publisher:      renderedModelTemplate.Publisher, // use rendered value for paging
		ClientId:       req.ClientId,
		TemplateId:     req.TemplateId,
		TemplateParams: req.TemplateParams,
		IsEnabled:      req.IsEnabled,
	}
	result, err := h.DAO.ModelClient().Create(ctx, model)
	if err != nil {
		return nil, err
	}
	// trigger cache refresh
	go ctxhelper.MustGetCacheManager(ctx).(cachetypes.Manager).TriggerRefresh(ctx, cachetypes.ItemTypeModel)
	return result, nil
}

func (h *ModelHandler) Get(ctx context.Context, req *pb.ModelGetRequest) (*pb.Model, error) {
	clientModel, err := cachehelpers.GetOneClientModel(ctx, req.ClientId, req.Id, nil)
	if err != nil {
		return nil, err
	}
	model := clientModel.Model
	if req.RenderTemplate {
		model, err = cachehelpers.GetRenderedModelByID(ctx, model.Id)
		if err != nil {
			return nil, err
		}
	}
	enhanceService := i18n_services.NewMetadataEnhancerService(ctx, ctxhelper.MustGetDBClient(ctx))
	inputLang := string(i18n.LocaleDefault)
	if lang, ok := ctxhelper.GetAccessLang(ctx); ok {
		inputLang = lang
	}
	locale := i18n_services.GetLocaleFromContext(inputLang)
	model = enhanceService.EnhanceModelMetadata(ctx, model, locale)
	// data sensitive
	desensitizeModel(ctx, model)
	return model, nil
}

func (h *ModelHandler) Update(ctx context.Context, req *pb.ModelUpdateRequest) (*pb.Model, error) {
	// get current
	current, err := h.DAO.ModelClient().Get(ctx, &pb.ModelGetRequest{Id: req.Id})
	if err != nil {
		return nil, err
	}
	u := &pb.Model{
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
	resp, err := h.DAO.ModelClient().Update(ctx, u)
	if err != nil {
		return nil, err
	}
	desensitizeModel(ctx, resp)
	// trigger cache refresh
	go ctxhelper.MustGetCacheManager(ctx).(cachetypes.Manager).TriggerRefresh(ctx, cachetypes.ItemTypeModel)
	return resp, nil
}

func (h *ModelHandler) Delete(ctx context.Context, req *pb.ModelDeleteRequest) (*commonpb.VoidResponse, error) {
	// check client-model-relation first
	relationPagingResp, err := h.DAO.ClientModelRelationClient().Paging(ctx, &clientmodelrelationpb.PagingRequest{
		PageNum:  1,
		PageSize: 1,
		ModelIds: []string{req.Id},
	})
	if err != nil {
		return nil, err
	}
	if relationPagingResp.Total > 0 {
		return nil, fmt.Errorf("model is assigned to clients, can not delete")
	}
	result, err := h.DAO.ModelClient().Delete(ctx, req)
	if err != nil {
		return nil, err
	}
	// trigger cache refresh
	go ctxhelper.MustGetCacheManager(ctx).(cachetypes.Manager).TriggerRefresh(ctx, cachetypes.ItemTypeModel)
	return result, nil
}

func (h *ModelHandler) Paging(ctx context.Context, req *pb.ModelPagingRequest) (resp *pb.ModelPagingResponse, err error) {
	if req.ViaDB {
		resp, err = h.pagingViaDB(ctx, req)
	} else {
		resp, err = h.pagingViaCache(ctx, req)
	}
	if err != nil {
		return nil, err
	}
	enhanceService := i18n_services.NewMetadataEnhancerService(ctx, ctxhelper.MustGetDBClient(ctx))
	inputLang := string(i18n.LocaleDefault)
	if lang, ok := ctxhelper.GetAccessLang(ctx); ok {
		inputLang = lang
	}
	locale := i18n_services.GetLocaleFromContext(inputLang)
	// data sensitive
	for i, item := range resp.List {
		if req.RenderTemplate {
			item, err = cachehelpers.GetRenderedModelByID(ctx, item.Id)
			if err != nil {
				return nil, err
			}
			resp.List[i] = item
		}
		item = enhanceService.EnhanceModelMetadata(ctx, item, locale)
		resp.List[i] = item
		desensitizeModel(ctx, item)
	}
	return resp, nil
}

func (h *ModelHandler) UpdateModelAbilitiesInfo(ctx context.Context, req *pb.ModelAbilitiesInfoUpdateRequest) (*commonpb.VoidResponse, error) {
	return h.DAO.ModelClient().UpdateModelAbilitiesInfo(ctx, req)
}

func (h *ModelHandler) LabelModel(ctx context.Context, req *pb.ModelLabelRequest) (*pb.Model, error) {
	return h.DAO.ModelClient().LabelModel(ctx, req)
}

func (h *ModelHandler) SetEnabled(ctx context.Context, req *pb.ModelSetEnabledRequest) (*pb.Model, error) {
	resp, err := h.DAO.ModelClient().SetEnabled(ctx, req)
	if err != nil {
		return nil, err
	}
	desensitizeModel(ctx, resp)
	// trigger cache refresh
	go ctxhelper.MustGetCacheManager(ctx).(cachetypes.Manager).TriggerRefresh(ctx, cachetypes.ItemTypeModel)
	return resp, nil
}

func desensitizeModel(ctx context.Context, item *pb.Model) {
	// pass for: admin
	if auth.IsAdmin(ctx) {
		return
	}
	// if a model is client-belonged, pass for the client
	if item.ClientId != "" && auth.IsClient(ctx) {
		return
	}
	// hide sensitive data for non-admin
	item.ApiKey = ""
	item.Metadata.Secret = nil
	item.TemplateParams = nil
}
