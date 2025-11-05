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

package common_types_util

import (
	serviceproviderpb "github.com/erda-project/erda-proto-go/apps/aiproxy/service_provider/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/common_types"
)

const metaKeyServiceProviderType = "service_provider_type"

// GetServiceProviderType
// - get from p.type if valid
// - backward to p.metadata.public.service_provider_type
func GetServiceProviderType(p *serviceproviderpb.ServiceProvider) string {
	if p == nil {
		return ""
	}
	typ := p.GetType()
	if common_types.ServiceProviderType(typ).IsValid() {
		return typ
	}
	metadata := p.GetMetadata()
	if metadata == nil {
		return ""
	}
	public := metadata.GetPublic()
	if public == nil {
		return ""
	}
	value, ok := public[metaKeyServiceProviderType]
	if !ok || value == nil {
		return ""
	}
	return value.GetStringValue()
}
