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

package ai_proxy

import (
	_ "github.com/erda-project/erda/internal/apps/ai-proxy/filters/assets"
	_ "github.com/erda-project/erda/internal/apps/ai-proxy/filters/audit-after-llm-director"
	_ "github.com/erda-project/erda/internal/apps/ai-proxy/filters/audit-before-llm-director"
	_ "github.com/erda-project/erda/internal/apps/ai-proxy/filters/body-size-limit"
	_ "github.com/erda-project/erda/internal/apps/ai-proxy/filters/context"
	_ "github.com/erda-project/erda/internal/apps/ai-proxy/filters/context-assistant"
	_ "github.com/erda-project/erda/internal/apps/ai-proxy/filters/context-audio"
	_ "github.com/erda-project/erda/internal/apps/ai-proxy/filters/context-chat"
	_ "github.com/erda-project/erda/internal/apps/ai-proxy/filters/context-embedding"
	_ "github.com/erda-project/erda/internal/apps/ai-proxy/filters/context-file"
	_ "github.com/erda-project/erda/internal/apps/ai-proxy/filters/context-image"
	_ "github.com/erda-project/erda/internal/apps/ai-proxy/filters/context-responses"
	_ "github.com/erda-project/erda/internal/apps/ai-proxy/filters/dashscope-director"
	_ "github.com/erda-project/erda/internal/apps/ai-proxy/filters/directors/azure-director"
	_ "github.com/erda-project/erda/internal/apps/ai-proxy/filters/directors/openai-compatible-director"
	_ "github.com/erda-project/erda/internal/apps/ai-proxy/filters/directors/openai-director"
	_ "github.com/erda-project/erda/internal/apps/ai-proxy/filters/erda-auth"
	_ "github.com/erda-project/erda/internal/apps/ai-proxy/filters/extra-body"
	_ "github.com/erda-project/erda/internal/apps/ai-proxy/filters/finalize"
	_ "github.com/erda-project/erda/internal/apps/ai-proxy/filters/initialize"
	_ "github.com/erda-project/erda/internal/apps/ai-proxy/filters/log-http"
	_ "github.com/erda-project/erda/internal/apps/ai-proxy/filters/openai-v1-models"
	_ "github.com/erda-project/erda/internal/apps/ai-proxy/filters/prometheus-collector"
	_ "github.com/erda-project/erda/internal/apps/ai-proxy/filters/rate-limit"
)
