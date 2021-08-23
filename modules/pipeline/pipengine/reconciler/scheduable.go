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

package reconciler

import (
	"sort"

	"github.com/erda-project/erda/modules/pipeline/pipengine/reconciler/rlog"
	"github.com/erda-project/erda/modules/pipeline/spec"
	"github.com/erda-project/erda/pkg/dag"
	"github.com/erda-project/erda/pkg/strutil"
)

// getSchedulableTasks return the list of schedulable tasks.
// tasks in list can be schedule concurrently.
func (r *Reconciler) getSchedulableTasks(p *spec.Pipeline, tasks []*spec.PipelineTask) ([]*spec.PipelineTask, error) {

	// construct DAG
	dagNodes := make([]dag.NamedNode, 0, len(tasks))
	for _, task := range tasks {
		dagNodes = append(dagNodes, task)
	}
	_dag, err := dag.New(dagNodes,
		// pipeline DAG 中目前可以禁用任意节点，即 dag.WithAllowMarkArbitraryNodesAsDone=true
		dag.WithAllowMarkArbitraryNodesAsDone(true),
		// 不做 cycle check，因为 pipeline.yml 写法保证一定无环，即 dag.WithAllowNotCheckCycle=true
		dag.WithAllowNotCheckCycle(true),
	)
	if err != nil {
		return nil, err
	}

	// calculate schedulable nodes according to dag and current done tasks
	schedulableNodeFromDAG, err := _dag.GetSchedulable((&spec.PipelineWithTasks{Tasks: tasks}).DoneTasks()...)
	if err != nil {
		return nil, err
	}

	// transfer schedulable nodes to tasks
	taskMap := make(map[string]*spec.PipelineTask)
	for _, task := range tasks {
		taskMap[task.Name] = task
	}
	var schedulableTasks []*spec.PipelineTask
	for nodeName := range schedulableNodeFromDAG {
		// get task by nodeName
		task := taskMap[nodeName]
		// if task is already processing by another goroutine, skip
		if _, alreadyProcessing := r.processingTasks.LoadOrStore(task.ID, true); alreadyProcessing {
			continue
		}
		schedulableTasks = append(schedulableTasks, task)
	}

	// nothing can be schedule
	if len(schedulableTasks) == 0 {
		rlog.PInfof(p.ID, "no schedulable tasks, wait for another task done")
		return nil, nil
	}

	// sort by names
	var schedulableTaskNames []string
	for _, task := range schedulableTasks {
		schedulableTaskNames = append(schedulableTaskNames, task.Name)
	}
	sort.Strings(schedulableTaskNames)
	rlog.PInfof(p.ID, "schedulable tasks: %s", strutil.Join(schedulableTaskNames, ", ", true))

	return schedulableTasks, nil
}
