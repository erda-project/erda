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

package impl

import (
	_ "github.com/erda-project/erda/internal/tools/orchestrator/hepa/apipolicy/micro-service-engine/built-in"
	_ "github.com/erda-project/erda/internal/tools/orchestrator/hepa/apipolicy/micro-service-engine/cors"
	_ "github.com/erda-project/erda/internal/tools/orchestrator/hepa/apipolicy/micro-service-engine/csrf"
	_ "github.com/erda-project/erda/internal/tools/orchestrator/hepa/apipolicy/micro-service-engine/ip"
	_ "github.com/erda-project/erda/internal/tools/orchestrator/hepa/apipolicy/micro-service-engine/proxy"
	_ "github.com/erda-project/erda/internal/tools/orchestrator/hepa/apipolicy/micro-service-engine/sbac"
	_ "github.com/erda-project/erda/internal/tools/orchestrator/hepa/apipolicy/micro-service-engine/server-guard"
	_ "github.com/erda-project/erda/internal/tools/orchestrator/hepa/apipolicy/micro-service-engine/waf"
	_ "github.com/erda-project/erda/internal/tools/orchestrator/hepa/apipolicy/nginx-kong-engine/built-in"
	_ "github.com/erda-project/erda/internal/tools/orchestrator/hepa/apipolicy/nginx-kong-engine/cors"
	_ "github.com/erda-project/erda/internal/tools/orchestrator/hepa/apipolicy/nginx-kong-engine/csrf"
	_ "github.com/erda-project/erda/internal/tools/orchestrator/hepa/apipolicy/nginx-kong-engine/custom"
	_ "github.com/erda-project/erda/internal/tools/orchestrator/hepa/apipolicy/nginx-kong-engine/ip"
	_ "github.com/erda-project/erda/internal/tools/orchestrator/hepa/apipolicy/nginx-kong-engine/proxy"
	_ "github.com/erda-project/erda/internal/tools/orchestrator/hepa/apipolicy/nginx-kong-engine/sbac"
	_ "github.com/erda-project/erda/internal/tools/orchestrator/hepa/apipolicy/nginx-kong-engine/server-guard"
	_ "github.com/erda-project/erda/internal/tools/orchestrator/hepa/apipolicy/nginx-kong-engine/waf"
)
