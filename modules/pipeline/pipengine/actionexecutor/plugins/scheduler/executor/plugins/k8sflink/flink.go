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

package k8sflink

import (
	"github.com/erda-project/erda/modules/pipeline/pipengine/actionexecutor/plugins/scheduler/executor/types"
)

var Kind = types.Kind("k8sflink")

//func init() {
//	types.MustRegister(Kind, func(name types.Name, clusterName string, options map[string]string) (types.TaskExecutor, error) {
//		k, err := New(name, clusterName, options)
//		if err != nil {
//			return nil, err
//		}
//		return k, nil
//	})
//}

//func (k *K8sFlink) Status(ctx context.Context, task *spec.PipelineTask) (apistructs.StatusDesc, error) {
//	k.client.ClientSet.
//}
