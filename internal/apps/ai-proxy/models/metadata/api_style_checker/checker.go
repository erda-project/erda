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

package api_style_checker

import (
	"strings"

	providerpb "github.com/erda-project/erda-proto-go/apps/aiproxy/model_provider/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/metadata"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/metadata/api_style"
)

func CheckIsOpenAICompatibleByProvider(provider *providerpb.ModelProvider) bool {
	if provider == nil {
		return false
	}
	providerNormalMeta := metadata.FromProtobuf(provider.Metadata)
	providerMeta := providerNormalMeta.MustToModelProviderMeta()
	return providerMeta.Public.API != nil &&
		strings.EqualFold(string(providerMeta.Public.API.APIStyle), string(api_style.APIStyleOpenAICompatible))
}
