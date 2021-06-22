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

package logic

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pkg/errors"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/pkg/task_uuid"
	"github.com/erda-project/erda/modules/pipeline/spec"
	"github.com/erda-project/erda/pkg/parser/diceyml"
	"github.com/erda-project/erda/pkg/parser/pipelineyml/pipelineymlv1"
)

func TransferToSchedulerJob(task *spec.PipelineTask) (job apistructs.JobFromUser, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.Errorf("%v", r)
		}
	}()

	return apistructs.JobFromUser{
		Name: task_uuid.MakeJobID(task),
		Kind: func() string {
			switch task.Type {
			case string(pipelineymlv1.RES_TYPE_FLINK):
				return string(apistructs.Flink)
			case string(pipelineymlv1.RES_TYPE_SPARK):
				return string(apistructs.Spark)
			default:
				return ""
			}
		}(),
		Namespace: task.Extra.Namespace,
		ClusterName: func() string {
			if len(task.Extra.ClusterName) == 0 {
				panic(errors.New("missing cluster name in pipeline task"))
			}
			return task.Extra.ClusterName
		}(),
		Image:      task.Extra.Image,
		Cmd:        strings.Join(append([]string{task.Extra.Cmd}, task.Extra.CmdArgs...), " "),
		CPU:        task.Extra.RuntimeResource.CPU,
		Memory:     task.Extra.RuntimeResource.Memory,
		Binds:      task.Extra.Binds,
		Volumes:    MakeVolume(task),
		PreFetcher: task.Extra.PreFetcher,
		Env:        task.Extra.PublicEnvs,
		Labels:     task.Extra.Labels,
		// flink/spark
		Resource:  task.Extra.FlinkSparkConf.JarResource,
		MainClass: task.Extra.FlinkSparkConf.MainClass,
		MainArgs:  task.Extra.FlinkSparkConf.MainArgs,
		// 重试不依赖 scheduler，由 pipeline engine 自己实现，保证所有 action executor 均适用
		Params: task.Extra.Action.Params,
	}, nil
}

func MakeVolume(task *spec.PipelineTask) []diceyml.Volume {
	diceVolumes := make([]diceyml.Volume, 0)
	for _, vo := range task.Extra.Volumes {
		if vo.Type == string(spec.StoreTypeDiceVolumeFake) || vo.Type == string(spec.StoreTypeDiceCacheNFS) {
			// fake volume,没有实际挂载行为,不传给scheduler
			continue
		}
		diceVolume := diceyml.Volume{
			Path: vo.Value,
			Storage: func() string {
				switch vo.Type {
				case string(spec.StoreTypeDiceVolumeNFS):
					return "nfs"
				case string(spec.StoreTypeDiceVolumeLocal):
					return "local"
				default:
					panic(errors.Errorf("%q has not supported volume type: %s", vo.Name, vo.Type))
				}
			}(),
		}
		if vo.Labels != nil {
			if id, ok := vo.Labels["ID"]; ok {
				diceVolume.ID = &id
				goto AppendDiceVolume
			}
		}
		// labels == nil or labels["ID"] not exist
		// 如果 id 不存在，说明上一次没有生成 volume，并且是 optional 的，则不创建 diceVolume
		if vo.Optional {
			continue
		}
	AppendDiceVolume:
		diceVolumes = append(diceVolumes, diceVolume)
	}
	return diceVolumes
}

func GetBigDataConf(task *spec.PipelineTask) (apistructs.BigdataConf, error) {
	conf := apistructs.BigdataConf{
		BigdataMetadata: apistructs.BigdataMetadata{
			Name:      task.Name,
			Namespace: task.Extra.Namespace,
		},
		Spec: apistructs.BigdataSpec{},
	}
	value, ok := task.Extra.Action.Params["bigDataConf"]
	if !ok {
		return conf, fmt.Errorf("missing big data conf from task: %s", task.Name)
	}

	if err := json.Unmarshal([]byte(value.(string)), &conf.Spec); err != nil {
		return conf, fmt.Errorf("unmarshal bigdata config error: %s", err.Error())
	}
	return conf, nil
}
