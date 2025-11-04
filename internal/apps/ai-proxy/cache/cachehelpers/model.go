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

package cachehelpers

import (
	"context"
	"fmt"

	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda-proto-go/apps/aiproxy/model/pb"
	providerpb "github.com/erda-project/erda-proto-go/apps/aiproxy/service_provider/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/cache/cacheimpl/item_template"
	"github.com/erda-project/erda/internal/apps/ai-proxy/cache/cachetypes"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/template"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/template/templatetypes"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/metadata"
)

func GetRenderedModelByID(ctx context.Context, modelID string) (*pb.Model, error) {
	cache := ctxhelper.MustGetCacheManager(ctx).(cachetypes.Manager)
	// get model
	modelV, err := cache.GetByID(ctx, cachetypes.ItemTypeModel, modelID)
	if err != nil {
		return nil, err
	}
	model := modelV.(*pb.Model)
	// render template
	if model.TemplateId == "" {
		return model, nil
	}
	// get template
	tplV, err := cache.GetByID(ctx, cachetypes.ItemTypeTemplate, item_template.ConstructID(templatetypes.TemplateTypeModel, model.TemplateId))
	if err != nil {
		return nil, fmt.Errorf("failed to get template: %w", err)
	}
	tpl := tplV.(*item_template.TypeNamedTemplate).Tpl

	providerTemplateName := ""
	if model.ProviderId != "" {
		if providerV, providerErr := cache.GetByID(ctx, cachetypes.ItemTypeProvider, model.ProviderId); providerErr == nil {
			if provider, ok := providerV.(*providerpb.ServiceProvider); ok {
				providerTemplateName = provider.TemplateId
			}
		}
	}

	renderParams := make(map[string]string, len(model.TemplateParams)+1)
	for k, v := range model.TemplateParams {
		renderParams[k] = v
	}
	if providerTemplateName != "" {
		renderParams[template.ServiceProviderTemplateNameParamKey] = providerTemplateName
	}
	// render template
	if err := template.RenderTemplate(model.TemplateId, tpl, renderParams); err != nil {
		return nil, fmt.Errorf("failed to render template: %w", err)
	}
	// merge modelTemplate to model
	var modelTemplate pb.Model
	cputil.MustObjJSONTransfer(tpl.Config, &modelTemplate)
	mergedMetadata, err := metadata.OverrideMetadata(ctx, modelTemplate.Metadata, model.Metadata)
	if err != nil {
		return nil, err
	}
	model.Metadata = mergedMetadata
	model.ApiKey = modelTemplate.ApiKey
	model.Type = modelTemplate.Type
	return model, nil
}
