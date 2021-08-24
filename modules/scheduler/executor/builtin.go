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

package executor

import (
	_ "github.com/erda-project/erda/modules/scheduler/executor/plugins/demo"
	_ "github.com/erda-project/erda/modules/scheduler/executor/plugins/edas"
	_ "github.com/erda-project/erda/modules/scheduler/executor/plugins/flink"
	_ "github.com/erda-project/erda/modules/scheduler/executor/plugins/k8s"

	//_ "github.com/erda-project/erda/modules/scheduler/executor/plugins/k8sflink"
	//_ "github.com/erda-project/erda/modules/scheduler/executor/plugins/k8sjob"
	//_ "github.com/erda-project/erda/modules/scheduler/executor/plugins/k8sspark"
	_ "github.com/erda-project/erda/modules/scheduler/executor/plugins/marathon"
	_ "github.com/erda-project/erda/modules/scheduler/executor/plugins/metronome"
	_ "github.com/erda-project/erda/modules/scheduler/executor/plugins/spark"
)
