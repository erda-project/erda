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
	"strings"

	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda-proto-go/apps/aiproxy/service_provider/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/cache/cacheimpl/item_template"
	"github.com/erda-project/erda/internal/apps/ai-proxy/cache/cachetypes"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/template"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/template/templatetypes"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/metadata"
)

func GetRenderedServiceProviderByID(ctx context.Context, providerID string) (*pb.ServiceProvider, error) {
	cache := ctxhelper.MustGetCacheManager(ctx).(cachetypes.Manager)
	// get provider
	providerV, err := cache.GetByID(ctx, cachetypes.ItemTypeProvider, providerID)
	if err != nil {
		return nil, err
	}
	provider := providerV.(*pb.ServiceProvider)
	// render template
	if provider.TemplateId == "" {
		return provider, nil
	}
	// get template
	tplV, err := cache.GetByID(ctx, cachetypes.ItemTypeTemplate, item_template.ConstructID(templatetypes.TemplateTypeServiceProvider, provider.TemplateId))
	if err != nil {
		return nil, fmt.Errorf("failed to get template: %w", err)
	}
	tpl := tplV.(*templatetypes.TypeNamedTemplate).Tpl
	renderParams := make(map[string]string)
	for k, v := range provider.TemplateParams {
		renderParams[k] = v
	}
	renderParams[template.PathMatcherParamKey] = ctxhelper.MustGetPathMatcher(ctx).Pattern
	// render
	if err := template.RenderTemplate(provider.TemplateId, tpl, renderParams); err != nil {
		return nil, fmt.Errorf("failed to render template: %w", err)
	}
	var providerTemplate pb.ServiceProvider
	cputil.MustObjJSONTransfer(tpl.Config, &providerTemplate)
	mergedMetadata, err := metadata.OverrideMetadata(ctx, providerTemplate.Metadata, provider.Metadata)
	if err != nil {
		return nil, err
	}
	provider.Metadata = mergedMetadata
	provider.ApiKey = providerTemplate.ApiKey
	provider.Type = providerTemplate.Type
	provider.Desc = strings.ReplaceAll(provider.Desc, template.TemplateDescPlaceholder, tpl.Desc)
	return provider, nil
}
