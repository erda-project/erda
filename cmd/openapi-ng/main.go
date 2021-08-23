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

package main

import (
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda/conf"
	"github.com/erda-project/erda/pkg/common"

	// providers and modules
	_ "github.com/erda-project/erda-infra/providers"
	_ "github.com/erda-project/erda/modules/openapi-ng"
	_ "github.com/erda-project/erda/modules/openapi-ng/interceptors/audit"
	_ "github.com/erda-project/erda/modules/openapi-ng/interceptors/auth"
	_ "github.com/erda-project/erda/modules/openapi-ng/interceptors/auth/legacy"
	_ "github.com/erda-project/erda/modules/openapi-ng/interceptors/common"
	_ "github.com/erda-project/erda/modules/openapi-ng/interceptors/csrf"
	_ "github.com/erda-project/erda/modules/openapi-ng/services"
	_ "github.com/erda-project/erda/providers/service-discover/erda-discover"
	_ "github.com/erda-project/erda/providers/service-discover/fixed-discover"
)

func main() {
	common.Run(&servicehub.RunOptions{
		ConfigFile: conf.OpenAPINGConfigFilePath,
		Content:    conf.OpenAPINGDefaultConfig,
	})
}
