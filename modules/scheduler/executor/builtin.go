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
