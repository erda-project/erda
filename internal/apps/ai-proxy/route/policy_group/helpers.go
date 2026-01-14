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

package policy_group

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	modelpb "github.com/erda-project/erda-proto-go/apps/aiproxy/model/pb"
	providerpb "github.com/erda-project/erda-proto-go/apps/aiproxy/service_provider/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/cache/cachehelpers"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/common_types"
	"github.com/erda-project/erda/pkg/strutil"
)

// BuildRoutingInstancesForClient fetches client models and converts to routing instances.
func BuildRoutingInstancesForClient(ctx context.Context, clientID string) ([]*RoutingModelInstance, error) {
	cfg := &cachehelpers.ClientModelConfig{OnlyEnabled: true}
	allClientModels, err := cachehelpers.ListAllClientModels(ctx, clientID, cfg)
	if err != nil {
		return nil, err
	}
	var routingInstances []*RoutingModelInstance
	for _, modelWithProvider := range allClientModels {
		labels := BuildLabels(modelWithProvider.Model, modelWithProvider.Provider)
		routingInstances = append(routingInstances, &RoutingModelInstance{
			ModelWithProvider: modelWithProvider,
			Labels:            labels,
		})
	}
	return routingInstances, nil
}

// BuildLabels collects base labels + metadata labels (public/secret) for selector use.
func BuildLabels(model *modelpb.Model, provider *providerpb.ServiceProvider) map[string]string {
	labels := make(map[string]string)

	// custom model labels
	for k, v := range model.Labels {
		labels[k] = v
	}

	// official model label
	labels[common_types.PolicyLabelKeyModelInstanceID] = model.Id
	labels[common_types.PolicyLabelKeyModelInstanceName] = model.Name
	labels[common_types.PolicyLabelKeyModelPublisher] = model.Publisher
	labels[common_types.PolicyLabelKeyModelTemplateID] = model.TemplateId
	labels[common_types.PolicyLabelKeyTemplate] = model.TemplateId
	labels[common_types.PolicyLabelKeyModelPublisherModelTemplateID] = model.Publisher + "/" + model.TemplateId
	labels[common_types.PolicyLabelKeyModelIsEnabled] = strconv.FormatBool(model.GetIsEnabled())

	// official service provider label
	labels[common_types.PolicyLabelKeyServiceProviderInstanceID] = provider.Id
	labels[common_types.PolicyLabelKeyServiceProviderInstanceName] = provider.Name
	labels[common_types.PolicyLabelKeyServiceProviderType] = provider.Type

	// location, region, country
	labels[common_types.PolicyLabelKeyLocation] = strutil.FirstNotEmpty(
		model.Metadata.Public[common_types.PolicyLabelKeyLocation].GetStringValue(),
		provider.Metadata.Public[common_types.PolicyLabelKeyLocation].GetStringValue(),
	)
	labels[common_types.PolicyLabelKeyRegion] = strutil.FirstNotEmpty(
		model.Metadata.Public[common_types.PolicyLabelKeyRegion].GetStringValue(),
		provider.Metadata.Public[common_types.PolicyLabelKeyRegion].GetStringValue(),
	)
	labels[common_types.PolicyLabelKeyCountry] = strutil.FirstNotEmpty(
		model.Metadata.Public[common_types.PolicyLabelKeyCountry].GetStringValue(),
		provider.Metadata.Public[common_types.PolicyLabelKeyCountry].GetStringValue(),
	)

	return labels
}

// BuildRequestMetaFromHeader copies headers (lowercased) into RequestMeta.Keys.
func BuildRequestMetaFromHeader(headers http.Header) RequestMeta {
	keys := make(map[string]string)
	for k := range headers {
		keys[common_types.StickyKeyPrefixFromReqHeader+strings.ToLower(k)] = headers.Get(k)
	}
	return RequestMeta{Keys: keys}
}
