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

package handler_template

import (
	"context"
	"sort"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	modelpb "github.com/erda-project/erda-proto-go/apps/aiproxy/model/pb"
	providerpb "github.com/erda-project/erda-proto-go/apps/aiproxy/service_provider/pb"
	"github.com/erda-project/erda-proto-go/apps/aiproxy/template/pb"
	_ "github.com/erda-project/erda-proto-go/apps/aiproxy/template/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/cache/cachehelpers"
	"github.com/erda-project/erda/internal/apps/ai-proxy/cache/cachetypes"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/template"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/template/templatetypes"
	"github.com/erda-project/erda/internal/apps/ai-proxy/handlers/handler_i18n/i18n_services"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/i18n"
	"github.com/erda-project/erda/pkg/strutil"
)

type TemplateHandler struct {
	Cache cachetypes.Manager
}

func (t *TemplateHandler) ListServiceProviderTemplates(ctx context.Context, req *pb.TemplateListRequest) (*pb.TemplateListResponse, error) {
	_, allTemplatesV, err := t.Cache.ListAll(ctx, cachetypes.ItemTypeTemplate)
	if err != nil {
		return nil, err
	}
	allTemplates := allTemplatesV.([]*templatetypes.TypeNamedTemplate)
	spTemplateMap := make(map[string]*pb.Template)
	for _, tpl := range allTemplates {
		if tpl.Type != templatetypes.TemplateTypeServiceProvider {
			continue
		}
		if tpl.Tpl.GetDeprecated() {
			continue
		}
		spTemplateMap[tpl.Name] = tpl.Tpl
	}
	var result pb.TemplateListResponse

	// sort by name
	names := make([]string, 0, len(spTemplateMap))
	for name := range spTemplateMap {
		names = append(names, name)
	}
	sort.Strings(names)

	providersByTemplateID := make(map[string][]*providerpb.ServiceProvider)
	if req.CheckInstance {
		_, providersV, err := t.Cache.ListAll(ctx, cachetypes.ItemTypeProvider)
		if err != nil {
			return nil, err
		}
		providers := providersV.([]*providerpb.ServiceProvider)
		for _, provider := range providers {
			if provider.TemplateId == "" || provider.ClientId == "" || provider.ClientId != req.ClientId {
				continue
			}
			providersByTemplateID[provider.TemplateId] = append(providersByTemplateID[provider.TemplateId], provider)
		}
	}

	for _, name := range names {
		spTemplate := spTemplateMap[name]
		var providerConfig providerpb.ServiceProvider
		cputil.MustObjJSONTransfer(spTemplate.Config, &providerConfig)
		detail := pb.TemplateForCreate{
			Name:         name,
			Desc:         spTemplate.Desc, // use top-level desc
			Placeholders: convertPlaceholdersForCreate(spTemplate.Placeholders),
			Metadata:     spTemplate.Metadata,
			Config:       spTemplate.Config,
			Deprecated:   spTemplate.GetDeprecated(),
		}
		if req.RenderTemplate {
			params := make(map[string]string)
			for _, placeholder := range spTemplate.Placeholders {
				params[placeholder.Name] = strutil.FirstNoneEmpty(placeholder.Default, placeholder.Example)
			}
			if err := template.RenderTemplate(name, spTemplate, params); err != nil {
				return nil, err
			}
			detail.Config = spTemplate.Config
			detail.Desc = spTemplate.Desc
		}
		// calculate instance count
		if req.CheckInstance {
			instanceCount, enabledInstanceCount := 0, 0
			for _, instance := range providersByTemplateID[name] {
				instanceCount += 1
				if instance.IsEnabled != nil && *instance.IsEnabled {
					enabledInstanceCount += 1
				}
			}
			detail.InstanceCount = &[]int64{int64(instanceCount)}[0]
			detail.EnabledInstanceCount = &[]int64{int64(enabledInstanceCount)}[0]
		}
		result.List = append(result.List, &detail)
	}

	result.Total = int64(len(result.List))

	return &result, nil
}

func (t *TemplateHandler) ListModelTemplates(ctx context.Context, req *pb.TemplateListRequest) (*pb.TemplateListResponse, error) {
	_, allTemplatesV, err := t.Cache.ListAll(ctx, cachetypes.ItemTypeTemplate)
	if err != nil {
		return nil, err
	}
	allTemplates := allTemplatesV.([]*templatetypes.TypeNamedTemplate)
	modelTemplateMap := make(map[string]*pb.Template)
	for _, tpl := range allTemplates {
		if tpl.Type != templatetypes.TemplateTypeModel {
			continue
		}
		if !req.ShowDeprecated && tpl.Tpl.GetDeprecated() {
			continue
		}
		modelTemplateMap[tpl.Name] = tpl.Tpl
	}
	var result pb.TemplateListResponse

	// sort by name
	names := make([]string, 0, len(modelTemplateMap))
	for name := range modelTemplateMap {
		names = append(names, name)
	}
	sort.Strings(names)

	modelsByTemplateID := make(map[string][]*modelpb.Model)
	if req.CheckInstance {
		var modelInstances []*modelpb.Model
		if req.ClientId != "" {
			allClientModels, err := cachehelpers.ListAllClientModels(ctx, req.ClientId)
			if err != nil {
				return nil, err
			}
			for _, clientModel := range allClientModels {
				modelInstances = append(modelInstances, clientModel.Model)
			}
		} else if ctxhelper.MustGetIsAdmin(ctx) {
			_, modelsV, err := t.Cache.ListAll(ctx, cachetypes.ItemTypeModel)
			if err != nil {
				return nil, err
			}
			modelInstances = modelsV.([]*modelpb.Model)
		}
		for _, instance := range modelInstances {
			if instance.TemplateId == "" {
				continue
			}
			modelsByTemplateID[instance.TemplateId] = append(modelsByTemplateID[instance.TemplateId], instance)
		}
	}

	var (
		enhanceService *i18n_services.MetadataEnhancerService
		locale         string
	)
	if req.RenderTemplate {
		enhanceService = i18n_services.NewMetadataEnhancerService(ctx, ctxhelper.MustGetDBClient(ctx))
		inputLang := string(i18n.LocaleDefault)
		if lang, ok := ctxhelper.GetAccessLang(ctx); ok {
			inputLang = lang
		}
		locale = i18n_services.GetLocaleFromContext(inputLang)
	}

	for _, name := range names {
		modelTemplate := modelTemplateMap[name]
		var modelTemplateConfig modelpb.Model
		cputil.MustObjJSONTransfer(modelTemplate.Config, &modelTemplateConfig)
		detail := pb.TemplateForCreate{
			Name:         name,
			Desc:         modelTemplate.Desc, // use top-level desc
			Placeholders: convertPlaceholdersForCreate(modelTemplate.Placeholders),
			Metadata:     modelTemplate.Metadata,
			Config:       modelTemplate.Config,
			Deprecated:   modelTemplate.GetDeprecated(),
		}
		// render template.config
		if req.RenderTemplate {
			params := make(map[string]string)
			for _, placeholder := range modelTemplate.Placeholders {
				params[placeholder.Name] = strutil.FirstNoneEmpty(placeholder.Default, placeholder.Example)
			}
			if err := template.RenderTemplate(name, modelTemplate, params); err != nil {
				return nil, err
			}
			detail.Config = modelTemplate.Config
			var renderedCfg modelpb.Model
			cputil.MustObjJSONTransfer(detail.Config, &renderedCfg)
			if enhanceService != nil && renderedCfg.Metadata != nil {
				tmpModel := &modelpb.Model{Metadata: renderedCfg.Metadata, Publisher: modelTemplateConfig.Publisher}
				enhanced := enhanceService.EnhanceModelMetadata(ctx, tmpModel, locale)
				renderedCfg.Metadata = enhanced.GetMetadata()
			}
			jsonBytes, err := protojson.MarshalOptions{UseProtoNames: true}.Marshal(&renderedCfg)
			if err != nil {
				return nil, err
			}
			cfgStruct := &structpb.Struct{}
			if err := protojson.Unmarshal(jsonBytes, cfgStruct); err != nil {
				return nil, err
			}
			detail.Config = cfgStruct
			detail.Desc = renderedCfg.Desc
		}
		// calculate instance count
		if req.CheckInstance {
			instanceCount, enabledInstanceCount := 0, 0
			for _, instance := range modelsByTemplateID[name] {
				instanceCount += 1
				if instance.IsEnabled != nil && *instance.IsEnabled {
					enabledInstanceCount += 1
				}
			}
			detail.InstanceCount = &[]int64{int64(instanceCount)}[0]
			detail.EnabledInstanceCount = &[]int64{int64(enabledInstanceCount)}[0]
		}
		result.List = append(result.List, &detail)
	}
	result.Total = int64(len(result.List))
	return &result, nil
}

func convertPlaceholdersForCreate(src []*pb.Placeholder) []*pb.PlaceholderForCreate {
	if len(src) == 0 {
		return nil
	}
	result := make([]*pb.PlaceholderForCreate, 0, len(src))
	for _, placeholder := range src {
		if placeholder == nil {
			continue
		}
		result = append(result, &pb.PlaceholderForCreate{
			Name:     placeholder.GetName(),
			Required: placeholder.GetRequired(),
			Desc:     placeholder.GetDesc(),
			Default:  placeholder.GetDefault(),
			Example:  placeholder.GetExample(),
			Mapping:  placeholder.GetMapping(),
		})
	}
	return result
}
