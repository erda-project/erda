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

package executor

import (
	_ "github.com/erda-project/erda/modules/scheduler/executor/plugins/chronos"
	_ "github.com/erda-project/erda/modules/scheduler/executor/plugins/demo"
	_ "github.com/erda-project/erda/modules/scheduler/executor/plugins/edas"
	_ "github.com/erda-project/erda/modules/scheduler/executor/plugins/flink"
	_ "github.com/erda-project/erda/modules/scheduler/executor/plugins/k8s"
	_ "github.com/erda-project/erda/modules/scheduler/executor/plugins/k8sflink"
	_ "github.com/erda-project/erda/modules/scheduler/executor/plugins/k8sjob"
	_ "github.com/erda-project/erda/modules/scheduler/executor/plugins/k8sspark"
	_ "github.com/erda-project/erda/modules/scheduler/executor/plugins/marathon"
	_ "github.com/erda-project/erda/modules/scheduler/executor/plugins/metronome"
	_ "github.com/erda-project/erda/modules/scheduler/executor/plugins/spark"
)
