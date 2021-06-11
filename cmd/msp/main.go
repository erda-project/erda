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

	// modules and providers
	_ "github.com/erda-project/erda-infra/providers"
	_ "github.com/erda-project/erda-proto-go/msp/menu/client"
	_ "github.com/erda-project/erda/modules/msp/configcenter"
	_ "github.com/erda-project/erda/modules/msp/instance/permission"
	_ "github.com/erda-project/erda/modules/msp/menu"
	_ "github.com/erda-project/erda/pkg/common/permission"
)

func main() {
	modcom.Run(&servicehub.RunOptions{
		ConfigFile: conf.MSPConfigFilePath,
		Content:    conf.MSPDefaultConfig,
	})
}
