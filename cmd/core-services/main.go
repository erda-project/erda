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
	"github.com/erda-project/erda/pkg/common"

	// providers
	_ "github.com/erda-project/erda-infra/providers/grpcclient"
	_ "github.com/erda-project/erda-proto-go/core/pipeline/cms/client"
	_ "github.com/erda-project/erda/modules/core-services"
	_ "github.com/erda-project/erda/modules/core/services/accesskey"
	// infra
	_ "github.com/erda-project/erda-infra/providers/grpcserver"
	_ "github.com/erda-project/erda-infra/providers/httpserver"
	_ "github.com/erda-project/erda-infra/providers/mysql"
	_ "github.com/erda-project/erda-infra/providers/serviceregister"
)

func main() {
	common.Run(&servicehub.RunOptions{
		ConfigFile: "conf/core-services/core-services.yaml",
	})
}
