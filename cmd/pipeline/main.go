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

	// providers and modules
	_ "github.com/erda-project/erda-infra/providers/mysqlxorm"
	_ "github.com/erda-project/erda-infra/providers/serviceregister"
	_ "github.com/erda-project/erda/modules/pipeline"
	_ "github.com/erda-project/erda/modules/pipeline/providers/aop"
	_ "github.com/erda-project/erda/modules/pipeline/providers/aop/plugins/pipeline/apitest_report"
	_ "github.com/erda-project/erda/modules/pipeline/providers/aop/plugins/pipeline/basic"
	_ "github.com/erda-project/erda/modules/pipeline/providers/aop/plugins/pipeline/echo"
	_ "github.com/erda-project/erda/modules/pipeline/providers/aop/plugins/pipeline/precheck_before_pop"
	_ "github.com/erda-project/erda/modules/pipeline/providers/aop/plugins/pipeline/project"
	_ "github.com/erda-project/erda/modules/pipeline/providers/aop/plugins/pipeline/scene_after"
	_ "github.com/erda-project/erda/modules/pipeline/providers/aop/plugins/pipeline/scene_before"
	_ "github.com/erda-project/erda/modules/pipeline/providers/aop/plugins/task/autotest_cookie_keep_after"
	_ "github.com/erda-project/erda/modules/pipeline/providers/aop/plugins/task/autotest_cookie_keep_before"
	_ "github.com/erda-project/erda/modules/pipeline/providers/aop/plugins/task/echo"
	_ "github.com/erda-project/erda/modules/pipeline/providers/aop/plugins/task/unit_test_report"
	_ "github.com/erda-project/erda/modules/pipeline/providers/cms"
)

func main() {
	common.Run(&servicehub.RunOptions{
		ConfigFile: "conf/pipeline/pipeline.yaml",
	})
}
