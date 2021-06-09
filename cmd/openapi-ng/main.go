// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/modcom"
	"github.com/erda-project/erda/conf"

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
	modcom.Run(&servicehub.RunOptions{
		ConfigFile: conf.OpenAPINGConfigFilePath,
		Content:    conf.OpenAPINGDefaultConfig,
	})
}
