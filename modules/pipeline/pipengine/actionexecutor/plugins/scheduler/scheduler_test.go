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

package scheduler

import (
	"context"
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/pipengine/actionexecutor/plugins/scheduler/executor"
	"github.com/erda-project/erda/modules/pipeline/pipengine/actionexecutor/plugins/scheduler/executor/plugins/k8sflink"
	"github.com/erda-project/erda/modules/pipeline/pipengine/actionexecutor/plugins/scheduler/executor/plugins/k8sjob"
	"github.com/erda-project/erda/modules/pipeline/pipengine/actionexecutor/plugins/scheduler/executor/plugins/k8sspark"
	"github.com/erda-project/erda/modules/pipeline/pipengine/actionexecutor/plugins/scheduler/executor/types"
	"github.com/erda-project/erda/modules/pipeline/spec"
	"github.com/erda-project/erda/pkg/parser/pipelineyml"
)

var (
	s           *Sched
	taskManager *executor.Manager
	clusters    map[string]apistructs.ClusterInfo
	executors   map[types.Name]types.TaskExecutor
)

func init() {
	clusters = map[string]apistructs.ClusterInfo{
		"terminus-dev": apistructs.ClusterInfo{
			Type: "k8s",
		},
		"dcos-cluster": apistructs.ClusterInfo{
			Type: "dcos",
		},
		"edas-cluster": apistructs.ClusterInfo{
			Type: "edas",
		},
	}

	executors = map[types.Name]types.TaskExecutor{
		types.Name("terminus-devfork8sjob"):   &k8sjob.K8sJob{},
		types.Name("terminus-devfork8sflink"): &k8sflink.K8sFlink{},
		types.Name("terminus-devfork8sspark"): &k8sspark.K8sSpark{},
		types.Name("edas-clusterfork8sjob"):   &k8sjob.K8sJob{},
	}

	s = &Sched{
		taskManager: taskManager,
	}
}

func Test_GetTaskExecutor(t *testing.T) {
	p := monkey.PatchInstanceMethod(reflect.TypeOf(taskManager), "GetCluster", func(_ *executor.Manager, clusterName string) (apistructs.ClusterInfo, error) {
		if len(clusterName) == 0 {
			return apistructs.ClusterInfo{}, errors.Errorf("clusterName is empty")
		}
		cluster, ok := clusters[clusterName]
		if !ok {
			return apistructs.ClusterInfo{}, errors.Errorf("failed to get cluster info by clusterName: %s", clusterName)
		}
		return cluster, nil
	})
	defer p.Unpatch()

	m := monkey.PatchInstanceMethod(reflect.TypeOf(taskManager), "Get", func(_ *executor.Manager, name types.Name) (types.TaskExecutor, error) {
		if len(name) == 0 {
			return nil, errors.Errorf("executor name is empty")
		}
		e, ok := executors[name]
		if !ok {
			return nil, errors.Errorf("not found action executor [%s]", name)
		}
		return e, nil
	})
	defer m.Unpatch()

	_, err := s.taskManager.GetCluster("terminus-dev")
	assert.NoError(t, err)

	k8sjobTask := &spec.PipelineTask{
		Type: "git-checkout",
		Extra: spec.PipelineTaskExtra{
			ClusterName: "terminus-dev",
		},
	}
	k8sflinkTask := &spec.PipelineTask{
		Type: "k8sflink",
		Extra: spec.PipelineTaskExtra{
			Action: pipelineyml.Action{
				Params: map[string]interface{}{
					"bigDataConf": "{\"args\":[\"--ACTION_OSS_ENDPOINT\",\"oss-cn-hangzhou.aliyuncs.com\",\"--ACTION_OSS_ACCESS_KEY_ID\",\"\",\"--ACTION_OSS_ACCESS_KEY_SECRET\",\"\",\"--ACTION_OSS_BUCKET\",\"dice-files\",\"--ACTION_SCRIPT_PATH\",\"terminus-dev/1/108/customer_CustomerBO_delta_sync\"],\"class\":\"io.terminus.dice.fdp.FlinkSqlMain\",\"envs\":[{\"name\":\"test\",\"value\":\"test\"}],\"flinkConf\":{\"jobManagerResource\":{\"cpu\":\"500.0m\",\"memory\":\"1024Mi\",\"replica\":1},\"kind\":\"FlinkJob\",\"parallelism\":1,\"taskManagerResource\":{\"cpu\":\"500.0m\",\"memory\":\"1024Mi\",\"replica\":1}},\"image\":\"registry.cn-hangzhou.aliyuncs.com/terminus/dice-flink:1.12.0\",\"properties\":{\"classloader.resolve-order\":\"parent-first\",\"s3.access-key\":\"minio\",\"s3.endpoint\":\"http://minio.minio-fdp.svc.cluster.local:9000\",\"s3.path.style.access\":\"true\",\"s3.secret-key\":\"minio123\",\"taskmanager.memory.flink.size\":\"1024M\",\"taskmanager.numberOfTaskSlots\":\"2\"},\"resource\":\"http://dice-files.oss-cn-hangzhou.aliyuncs.com/terminus-dev/flink/jar/fdp-flink-sql-20210630.jar\"}",
				},
			},
			ClusterName: "terminus-dev",
		},
	}
	k8sSparkTask := &spec.PipelineTask{
		Type: "k8sspark",
		Extra: spec.PipelineTaskExtra{
			Action: pipelineyml.Action{
				Params: map[string]interface{}{
					"bigDataConf": "{\"class\":\"io.terminus.dice.fdp.SparkSqlMain\",\"envs\":[{\"name\":\"ACTION_OSS_ENDPOINT\",\"value\":\"oss-cn-hangzhou.aliyuncs.com\"},{\"name\":\"ACTION_OSS_ACCESS_KEY_ID\",\"value\":\"\"},{\"name\":\"ACTION_OSS_ACCESS_KEY_SECRET\",\"value\":\"gUdhK5t2WdoC5RhPUgdJsMXpgncx6L\"},{\"name\":\"ACTION_OSS_BUCKET\",\"value\":\"dice-files\"},{\"name\":\"ACTION_SCRIPT_PATH\",\"value\":\"terminus-test/1/149/customer_CustomerBO_delta_sync\"}],\"image\":\"registry.cn-hangzhou.aliyuncs.com/terminus/spark:v3.0.0\",\"resource\":\"http://dice-files.oss-cn-hangzhou.aliyuncs.com/terminus-dev/spark/jar/fdp-spark-sql-20210302.jar\",\"sparkConf\":{\"driverResource\":{\"cpu\":\"1\",\"memory\":\"512\",\"replica\":1},\"executorResource\":{\"cpu\":\"1\",\"memory\":\"1024\",\"replica\":1},\"kind\":\"cluster\",\"type\":\"Java\"}}",
				},
			},
			ClusterName: "terminus-dev",
		},
	}
	flinkTask := &spec.PipelineTask{
		Type: "flink",
		Extra: spec.PipelineTaskExtra{
			ClusterName: "dcos-cluster",
		},
	}
	sparkTask := &spec.PipelineTask{
		Type: "spark",
		Extra: spec.PipelineTaskExtra{
			ClusterName: "demo-cluster",
		},
	}
	edasTask := &spec.PipelineTask{
		Type: "cdp",
		Extra: spec.PipelineTaskExtra{
			ClusterName: "edas-cluster",
		},
	}
	shouldDispatch, _, err := s.GetTaskExecutor(k8sjobTask.Type, k8sjobTask.Extra.ClusterName, k8sjobTask)
	assert.Equal(t, nil, err)
	assert.Equal(t, false, shouldDispatch)

	shouldDispatch, _, err = s.GetTaskExecutor(k8sflinkTask.Type, k8sflinkTask.Extra.ClusterName, k8sflinkTask)
	assert.Equal(t, nil, err)
	assert.Equal(t, false, shouldDispatch)

	shouldDispatch, _, err = s.GetTaskExecutor(k8sSparkTask.Type, k8sSparkTask.Extra.ClusterName, k8sSparkTask)
	assert.Equal(t, nil, err)
	assert.Equal(t, false, shouldDispatch)

	shouldDispatch, _, err = s.GetTaskExecutor(flinkTask.Type, flinkTask.Extra.ClusterName, flinkTask)
	assert.Equal(t, nil, err)
	assert.Equal(t, true, shouldDispatch)

	shouldDispatch, _, err = s.GetTaskExecutor(sparkTask.Type, sparkTask.Extra.ClusterName, sparkTask)
	assert.Error(t, err)
	assert.Equal(t, false, shouldDispatch)

	shouldDispatch, _, err = s.GetTaskExecutor(edasTask.Type, edasTask.Extra.ClusterName, edasTask)
	assert.Equal(t, nil, err)
	assert.Equal(t, false, shouldDispatch)
}

func TestCancel(t *testing.T) {
	s = &Sched{
		taskManager: taskManager,
	}
	m := monkey.PatchInstanceMethod(reflect.TypeOf(s), "GetTaskExecutor", func(_ *Sched, executorType string, clusterName string, task *spec.PipelineTask) (bool, types.TaskExecutor, error) {
		return false, &k8sjob.K8sJob{}, nil
	})
	defer m.Unpatch()

	p := monkey.PatchInstanceMethod(reflect.TypeOf(&k8sjob.K8sJob{}), "Remove", func(_ *k8sjob.K8sJob, ctx context.Context, task *spec.PipelineTask) (data interface{}, err error) {
		return nil, nil
	})
	defer p.Unpatch()

	uuid := "pipeline-123456"
	ctx := context.Background()
	task := &spec.PipelineTask{
		Extra: spec.PipelineTaskExtra{
			Namespace: "pipeline-1",
			UUID:      uuid,
			LoopOptions: &apistructs.PipelineTaskLoopOptions{
				LoopedTimes: 3,
			},
		},
	}
	_, err := s.Cancel(ctx, task)
	assert.Equal(t, nil, err)
	assert.Equal(t, uuid, task.Extra.UUID)
}

//import (
//	"context"
//	"encoding/json"
//	"fmt"
//	"testing"
//
//	"github.com/stretchr/testify/require"
//
//	"github.com/erda-project/erda/apistructs"
//	"github.com/erda-project/erda/modules/pipeline/spec"
//)
//
//var (
//	s      *Sched
//	action *spec.PipelineTask
//)
//
//func init() {
//	s = &Sched{
//		name:    "test",
//		addr:    "scheduler.default.svc.cluster.local:9091",
//		options: nil,
//	}
//	action = &spec.PipelineTask{
//		Name:         "pipeline-local-test",
//		Type:         "custom-script",
//		ExecutorKind: "SCHEDULER",
//		Status:       apistructs.PipelineStatusAnalyzed,
//		Extra: spec.PipelineTaskExtra{
//			Namespace:    "pipeline-test-namespace-1",
//			ExecutorName: "scheduler",
//			ClusterName:  "terminus-dev",
//			PrivateEnvs: map[string]string{
//				"AGENT_PRE_FETCHER_DEST_DIR": "/opt/emptydir",
//				"WORKDIR":                    "/.pipeline/container/context/custom-script",
//				"CONTEXTDIR":                 "/.pipeline/container/context",
//				"METAFILE":                   "/.pipeline/container/metadata/custom-script/metadata",
//				"DICE_OPENAPI_ADDR":          "openapi.default.svc.cluster.local:9529",
//				"DICE_OPENAPI_TOKEN":         "xxx",
//			},
//			Image:   "registry.cn-hangzhou.aliyuncs.com/dice/default-action-image:3.4.0-20190704-1765035",
//			Cmd:     "/opt/emptydir/agent",
//			CmdArgs: []string{"eyJjb21tYW5kcyI6WyJlY2hvIGhlbGxvIHBpcGVsaW5lISIsInNsZWVwIDFoIl0sImNvbnRleHQiOnsib3V0U3RvcmFnZXMiOlt7Im5hbWUiOiJjdXN0b20tc2NyaXB0IiwidmFsdWUiOiIvLnBpcGVsaW5lL2NvbnRleHQvY3VzdG9tLXNjcmlwdCIsInR5cGUiOiJkaWNlLW5mcy12b2x1bWUifV19LCJwaXBlbGluZUlEIjoxOTUsInBpcGVsaW5lVGFza0lEIjo3MzB9"},
//			RuntimeResource: spec.RuntimeResource{
//				CPU:    0.1,
//				Memory: 32,
//			},
//			UUID: "uuid1",
//			PreFetcher: &apistructs.PreFetcher{
//				FileFromImage: "registry.cn-hangzhou.aliyuncs.com/dice/action-agent:3.7-20191022-fcc24a74bd",
//				FileFromHost:  "/netdata/devops/ci/action-agent",
//				ContainerPath: "/opt/emptydir",
//			},
//			Volumes: []apistructs.MetadataField{
//				{
//					Name:  "custom-script",
//					Value: "/.pipeline/context/custom-script",
//					Type:  "dice-nfs-volume",
//				},
//			},
//		},
//	}
//}
//
//func TestSched_Create(t *testing.T) {
//	//job, err := s.Create(context.Background(), action)
//	//require.NoError(t, err)
//	//b, _ := json.MarshalIndent(job, "", "  ")
//	//fmt.Println("created:\n" + string(b))
//
//	job, err := s.Start(context.Background(), action)
//	require.NoError(t, err)
//	b, _ := json.MarshalIndent(job, "", "  ")
//	fmt.Println("started:\n" + string(b))
//
//	//data, err := s.Cancel(context.Background(), action)
//	//require.NoError(t, err)
//	//spew.Dump(data)
//}
//
//func TestSched_DeleteNamespace(t *testing.T) {
//	fmt.Println(s.DeleteNamespace(context.Background(), action))
//}
//
//func TestSched_Status(t *testing.T) {
//	fmt.Println(s.Status(context.Background(), action))
//}
//
//func TestSched_BatchDelete(t *testing.T) {
//	actions := []*spec.PipelineTask{action, action}
//	fmt.Println(s.BatchDelete(context.Background(), actions))
//}
